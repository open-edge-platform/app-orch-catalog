// SPDX-FileCopyrightText: (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/migrate"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"github.com/open-edge-platform/orch-library/go/pkg/northbound"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	// pq is Postgres driver for the database/sql package
	_ "github.com/lib/pq"

	ent "github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	service "github.com/open-edge-platform/app-orch-catalog/internal/northbound"
)

var log = dazl.GetPackageLogger()

const (
	databaseUser = "DATABASE_USER"
	databasePwd  = "DATABASE_PWD"

	// OIDCServerURL - address of an OpenID Connect server
	OIDCServerURL = "OIDC_SERVER_URL"

	opaHostname = "localhost"
	opaPort     = 8181
	opaScheme   = "http"
)

// Config is a manager configuration
type Config struct {
	CAPath                   string
	KeyPath                  string
	CertPath                 string
	GRPCPort                 int
	DatabaseHostname         string
	DatabasePort             int
	DatabaseSslmode          bool
	DatabaseDisableMigration bool
	DatabaseDriver           string
	DatabaseName             string
	MigrationsDir            string
	DefaultProjectUUID       string
}

// NewManager creates a new manager
func NewManager(config Config) *Manager {
	log.Infof("Creating Manager with config: %+v", config)
	return &Manager{
		Config: config,
	}
}

// Manager single point of entry for the application-catalog system.
type Manager struct {
	Config   Config
	dbClient *ent.Client
}

// Run starts a synchronizer based on the devices and the northbound services.
func (m *Manager) Run() {
	workingDirectory := os.TempDir()
	log.Infof("Starting Manager in %s", workingDirectory)
	_ = os.Chdir(workingDirectory)
	if err := m.Start(); err != nil {
		log.Fatalw("Unable to run Manager", dazl.Error(err))
	}
}

// Start starts the manager
func (m *Manager) Start() error {
	sslModeStr := "disable"
	if m.Config.DatabaseSslmode {
		sslModeStr = "require"
	}
	// TODO: Replace with call to get Secrets directly
	dbUser, ok := os.LookupEnv(databaseUser)
	if !ok {
		log.Fatalf("%s env var is not set", databaseUser)
	}
	dbPwd, ok := os.LookupEnv(databasePwd)
	if !ok {
		log.Fatalf("%s env var is not set", databasePwd)
	}
	var err error
	m.dbClient, err = ent.Open(m.Config.DatabaseDriver,
		fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			m.Config.DatabaseHostname, m.Config.DatabasePort,
			dbUser, m.Config.DatabaseName, dbPwd,
			sslModeStr))
	if err != nil {
		log.Fatalf("failed opening connection to %s %s:%d:%s for %s: %v",
			m.Config.DatabaseDriver, m.Config.DatabaseHostname, m.Config.DatabasePort, m.Config.DatabaseName, dbUser, err)
	}
	log.Infof("Connected to %s %s:%d:%s as %s",
		m.Config.DatabaseDriver, m.Config.DatabaseHostname, m.Config.DatabasePort, m.Config.DatabaseName, dbUser)

	// Unless the database schema migration has been disabled, attempt to run it explicitly via Atlas.
	if !m.Config.DatabaseDisableMigration {
		// Initialize Atlas migration tracking for databases migration
		// This handles upgrades where Atlas migration files were previously unavailable
		if err := migrate.InitializeAtlasTracking(dbPath(m.Config, dbUser, dbPwd)); err != nil {
			log.Fatalf("ATTENTION: failed to initialize Atlas migration tracking: %v", err)
		}

		if output, err := migrate.RunAtlasMigrations(dbPath(m.Config, dbUser, dbPwd), m.Config.MigrationsDir); err != nil {
			// Log the full Atlas output to help diagnose migration failures
			log.Fatalf("ATTENTION: failed to apply migrations: %v\nAtlas output: %s", err, string(output))
		}

		// Trigger the second phase of release-specific database migration.
		if m.Config.DefaultProjectUUID != "" {
			if err = newMigration(m.dbClient, m.Config.DefaultProjectUUID).run(context.Background()); err != nil {
				log.Errorf("ATTENTION: failed to migrate project: %v", err)
			}
		}
		log.Infof("Database migration complete")
	}

	err = m.startNorthboundServer()
	if err != nil {
		return err
	}
	return nil
}

// Composes data base path, e.g. postgres://postgres:pass@localhost:5432/database?sslmode=disable
func dbPath(cfg Config, dbUser string, dbPwd string) string {
	sslMode := "disable"
	if cfg.DatabaseSslmode {
		sslMode = "require"
	}
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DatabaseDriver, dbUser, dbPwd, cfg.DatabaseHostname, cfg.DatabasePort, cfg.DatabaseName, sslMode)
}

// startNorthboundServer starts the northbound gRPC server
func (m *Manager) startNorthboundServer() error {
	serverConfig := northbound.NewInsecureServerConfig(int16(m.Config.GRPCPort))

	if oidcURL := os.Getenv(OIDCServerURL); oidcURL != "" {
		serverConfig.SecurityCfg = &northbound.SecurityConfig{
			AuthenticationEnabled: true,
			AuthorizationEnabled:  true,
		}
		log.Infof("Authentication enabled. %s=%s", OIDCServerURL, oidcURL)
	} else {
		log.Infof("Authentication not enabled %s not set", OIDCServerURL)
	}

	s := northbound.NewServer(serverConfig)

	serverAddr := fmt.Sprintf("%s://%s:%d", opaScheme, opaHostname, opaPort)

	var opaClient openpolicyagent.ClientWithResponsesInterface
	var err error
	if serverConfig.SecurityCfg.AuthorizationEnabled {
		opaClient, err = openpolicyagent.NewClientWithResponses(serverAddr)
		if err != nil {
			log.Fatalf("OPA server cannot be created %v", err)
		}
	}

	// TODO: Determine whether this is required in the future: s.AddService(dazl.Service{})
	s.AddService(service.NewService(m.dbClient, opaClient))
	s.AddService(HealthCheck{})

	// It's observed that idle connections are leading to lingering goroutines for as long as
	// two hours. The keepalive settings below will send keepalive requests after 30 seconds
	// of idle time. Even if a client is responding to keepalives (there is some evidence the
	// rest-proxy may be doing this on idle connections), then we will still terminate the
	// connection after 2 minutes of idle time.

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle: 120 * time.Second, // If a client is idle for 2 minutes, send a GOAWAY
		Time:              30 * time.Second,  // Server pings client after 30s of inactivity
		Timeout:           15 * time.Second,  // Server waits 15s for client ACK
	}

	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		}, grpc.KeepaliveParams(kasp))
		if err != nil {
			m.dbClient.Close()
			doneCh <- err
		}
	}()
	return <-doneCh
}

// Close kills the channels and manager related objects
func (m *Manager) Close() {
	m.dbClient.Close()
	log.Info("Closing Manager")
}

// HealthCheck is a struct receiver implementing onos northbound Register interface.
type HealthCheck struct{}

// Register is a method to register a health check gRPC service.
func (h HealthCheck) Register(s *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
}
