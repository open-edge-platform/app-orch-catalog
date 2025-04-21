// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/open-edge-platform/app-orch-catalog/internal/manager"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/version"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	"github.com/open-edge-platform/orch-library/go/dazl"
	_ "github.com/open-edge-platform/orch-library/go/dazl/zap"
	"os"
	"strconv"
)

const (
	malwareScannerAddressEnv    = "MALWARE_SCANNER_ADDRESS"
	malwareScannerPermissiveEnv = "MALWARE_SCANNER_PERMISSIVE"
)

// log
var log = dazl.GetLogger()

func main() {
	var err error
	caPath := flag.String("caPath", "", "path to CA certificate")
	keyPath := flag.String("keyPath", "", "path to client private key")
	certPath := flag.String("certPath", "", "path to client certificate")
	databaseDriver := flag.String("databaseDriver", "postgres", "database driver")
	databaseHostname := flag.String("databaseHostname", "localhost", "database hostname")
	databasePort := flag.String("databasePort", "5432", "database network port")
	databaseName := flag.String("databaseName", "postgres", "database network port")
	databaseSslMode := flag.Bool("databaseSslMode", true, "database SSL mode")
	databaseDisableMigration := flag.Bool("databaseDisableMigration", true, "disable database migration")
	useSecretsService := flag.Bool("useSecretsService", false, "use secrets service for sensitive data")
	migrationsDir := flag.String("migrationsDir", "/usr/share/migrations", "directory containing database schema migrations")
	defaultProjectUUID := flag.String("defaultProjectUUID", "28e65b24-522d-4462-9477-79d9c0bf6e8f", "default project UUID")
	vaultServerAddress := flag.String("vaultServerAddress", "", "vault server address")

	ready := make(chan bool)
	flag.Parse()
	errors.Init()

	if malwareScannerAddress := os.Getenv(malwareScannerAddressEnv); malwareScannerAddress != "" {
		permissive := false
		if permissiveStr := os.Getenv(malwareScannerPermissiveEnv); permissiveStr != "" {
			permissive, err = strconv.ParseBool(permissiveStr)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Infof("Enabling malware scanner at %s. Permissive=%v", malwareScannerAddress, permissive)
		malware.DefaultScanner = malware.NewScanner(malwareScannerAddress, malware.DefaultScannerTimeout, permissive)
	} else {
		log.Warn("Malware scanning is not enabled")
	}

	northbound.UseSecretService = *useSecretsService
	northbound.VaultServerAddress = *vaultServerAddress

	log.Info("Starting application-catalog")
	version.LogVersion("  ")

	databasePortInt, err := strconv.Atoi(*databasePort)
	if err != nil {
		log.Fatalf("Unable to convert database port to Int %s %v", databasePort, err)
	}
	cfg := manager.Config{
		CAPath:                   *caPath,
		KeyPath:                  *keyPath,
		CertPath:                 *certPath,
		GRPCPort:                 8080,
		DatabaseDriver:           *databaseDriver,
		DatabaseHostname:         *databaseHostname,
		DatabasePort:             databasePortInt,
		DatabaseName:             *databaseName,
		DatabaseSslmode:          *databaseSslMode,
		DatabaseDisableMigration: *databaseDisableMigration,
		MigrationsDir:            *migrationsDir,
		DefaultProjectUUID:       *defaultProjectUUID,
	}

	mgr := manager.NewManager(cfg)
	mgr.Run()
	<-ready
}
