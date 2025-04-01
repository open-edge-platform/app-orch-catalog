// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	ent "github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/enttest"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Suite of proxy tests
type ProxyTestSuite struct {
	suite.Suite
	cfg   *Config
	proxy *RESTProxy

	ctx    context.Context
	cancel context.CancelFunc

	ws *websocket.Conn

	dbClient   *generated.Client
	conn       *grpc.ClientConn
	client     catalogv3.CatalogServiceClient
	opa        openpolicyagent.ClientWithResponsesInterface
	httpClient http.Client
}

func (s *ProxyTestSuite) SetupSuite() {
	s.cfg = &Config{
		Port:               6942,
		GRPCEndpoint:       "localhost:6943",
		BasePath:           "/",
		SpecFilePath:       "../../api/spec/openapi.yaml",
		AllowedCorsOrigins: "http://localhost:6943,http://localhost:8081",
		OIDCExternal:       "",
	}

	mockController := gomock.NewController(s.T())
	s.dbClient = enttest.Open(s.T(), "sqlite3", "file:ent?mode=memory&_fk=1")

	ctx := metadata.AppendToOutgoingContext(context.Background(), "activeprojectid", "project")
	s.ctx, s.cancel = context.WithTimeout(ctx, 10*time.Minute)

	opaMock := openpolicyagent.NewMockClientWithResponsesInterface(mockController)
	result := openpolicyagent.OpaResponse_Result{}
	err := result.FromOpaResponseResult1(true)
	s.NoError(err)
	opaMock.EXPECT().PostV1DataPackageRuleWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&openpolicyagent.PostV1DataPackageRuleResponse{
			JSON200: &openpolicyagent.OpaResponse{
				DecisionId: nil,
				Metrics:    nil,
				Result:     result,
			},
		}, nil,
	).AnyTimes()
	s.opa = opaMock

	// Prepare a connection for interacting with the catalog service for testing purposes
	s.conn = createServerConnection(s.T(), s.dbClient, s.opa)
	s.client = catalogv3.NewCatalogServiceClient(s.conn)
	s.httpClient = http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}

	s.StartRESTProxy()
}

func (s *ProxyTestSuite) TearDownSuite() {
	if s.conn != nil {
		_ = s.conn.Close()
		_ = s.dbClient.Close()
		s.cancel()
	}
	s.conn = nil
}

func (s *ProxyTestSuite) SetupTest() {
}

func (s *ProxyTestSuite) TearDownTest() {
	s.cleanUpTestRegistry()
}

func (s *ProxyTestSuite) setupWebSocket() {
	var err error
	// Prepare WS connection for serving as a client subscribing for and receiving async events
	u := url.URL{Scheme: "ws", Host: "localhost:6942", Path: "/catalog.orchestrator.apis/events"}
	s.ws, _, err = websocket.DefaultDialer.Dial(u.String(), map[string][]string{ActiveProjectID: {"project"}})
	s.NoError(err)
	s.NotNil(s.ws)
	s.NoError(s.ws.SetReadDeadline(time.Now().Add(30 * time.Second)))
	_ = s.ws.WriteMessage(websocket.PingMessage, []byte{0, 1, 2, 3})
}

func (s *ProxyTestSuite) closeWebSocket() {
	if s.ws != nil {
		_ = s.ws.Close()
	}
}

func (s *ProxyTestSuite) StartRESTProxy() {
	var err error
	s.proxy, err = NewRESTProxy(s.cfg)
	s.NoError(err)
	go func() { _ = s.proxy.Run() }()
}

func TestNorthBound(t *testing.T) {
	suite.Run(t, &ProxyTestSuite{})
}

func newTestService(dbClient *ent.Client, opaClient openpolicyagent.ClientWithResponsesInterface) (northbound.Service, error) {
	return northbound.Service{DatabaseClient: dbClient, OpaClient: opaClient}, nil
}

func createServerConnection(t *testing.T, dbClient *ent.Client, opaClient openpolicyagent.ClientWithResponsesInterface) *grpc.ClientConn {
	lis, err := net.Listen("tcp", "localhost:6943")
	assert.NoError(t, err)

	s, err := newTestService(dbClient, opaClient)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	server := grpc.NewServer()
	s.Register(server)

	go func() {
		if err := server.Serve(lis); err != nil {
			assert.NoError(t, err, "Server exited with error: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost:6943", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}
