// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package basic is a suite of basic functionality tests for the catalog service
package basic

import (
	"context"
	"fmt"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/open-edge-platform/orch-library/go/pkg/grpc/retry"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"path/filepath"
	"testing"
)

const (
	ActiveProjectID = "ActiveProjectID"
	restAddress     = "catalog-service-rest-proxy:8081/"
)

// TestSuite is the basic test suite
type TestSuite struct {
	suite.Suite
	ctx               context.Context
	client            catalogv3.CatalogServiceClient
	restClient        *restapi.ClientWithResponses
	KeycloakServer    string
	CatalogServer     string
	CatalogRESTServer string
	ArtifactFilename  string
	token             string
	NoClear           bool
}

func (s *TestSuite) Context() context.Context {
	return s.ctx
}

func (s *TestSuite) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *TestSuite) findCharts(chartName string) string {
	matches, err := filepath.Glob("build/_output/" + chartName + "-*.tgz")
	s.NoError(err)
	return matches[0]
}

// SetupSuite sets-up the integration tests for the application catalog basic test suite
func (s *TestSuite) SetupSuite() {
	s.KeycloakServer = "keycloak.orch-system.svc"
	s.CatalogServer = "catalog-service-grpc-server:8080"
	s.CatalogRESTServer = restAddress
	s.ArtifactFilename = "test/basic/1x1.png"
}

// SetupTest sets up for each integration test
func (s *TestSuite) SetupTest() {
	conn, err := createConnection(s.CatalogServer)
	s.NoError(err)
	restClient, err := restapi.NewClientWithResponses("http://" + s.CatalogRESTServer)
	s.NoError(err)
	s.restClient = restClient
	s.client = catalogv3.NewCatalogServiceClient(conn)
	s.token = auth.SetUpAccessToken(s.T(), s.KeycloakServer)
	authToken = s.token
	if !s.NoClear {
		fmt.Printf("Clearing out old data\n")
		s.WipeOutData()
		fmt.Printf("Old data cleared out\n")
	}
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TearDownTest tears down remnants of each integration test
func (s *TestSuite) TearDownTest(ctx context.Context) {
}

func (s *TestSuite) CheckStatus(name string) {
	if s.T().Failed() {
		fmt.Printf("%s Failed!\n", name)
	} else {
		fmt.Printf("%s Passed!\n", name)
	}
}

// AddHeaders adds authentication and project ID headers
func (s *TestSuite) AddHeaders(projectUUID string) context.Context {
	return auth.AddGrpcAuthHeader(s.Context(), s.token, projectUUID)
}

// ProjectID adds project UUID to the context metadata
func (s *TestSuite) ProjectID(projectUUID string) context.Context {
	return metadata.AppendToOutgoingContext(s.Context(), ActiveProjectID, projectUUID)
}

func createConnection(server string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(retry.RetryingUnaryClientInterceptor()),
	}

	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
