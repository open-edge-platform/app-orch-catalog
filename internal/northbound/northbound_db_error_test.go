// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	sqldb "database/sql"
	"entgo.io/ent/dialect/sql"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	ent "github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sync"
	"testing"
	"time"
)

// Suite of northbound tests for checking the handling of database errors
type NorthBoundDBErrTestSuite struct {
	suite.Suite

	ctx       context.Context
	cancel    context.CancelFunc
	db        *sqldb.DB
	mock      sqlmock.Sqlmock
	entClient *ent.Client
	server    Server
}

func (s *NorthBoundDBErrTestSuite) SetupSuite() {
}

func (s *NorthBoundDBErrTestSuite) TearDownSuite() {
}

func (s *NorthBoundDBErrTestSuite) SetupTest() {
	var err error
	s.db, s.mock, err = sqlmock.New()
	if err != nil {
		s.T().Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	driver := sql.OpenDB("mysql", s.db)

	driverOption := ent.Driver(driver)
	s.entClient = ent.NewClient(driverOption)

	s.server = Server{
		UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
		databaseClient:                    s.entClient,
		listeners: &EventListeners{
			lock:                       sync.RWMutex{},
			registryListeners:          map[chan *catalogv3.WatchRegistriesResponse]*catalogv3.WatchRegistriesRequest{},
			artifactListeners:          map[chan *catalogv3.WatchArtifactsResponse]*catalogv3.WatchArtifactsRequest{},
			applicationListeners:       map[chan *catalogv3.WatchApplicationsResponse]*catalogv3.WatchApplicationsRequest{},
			deploymentPackageListeners: map[chan *catalogv3.WatchDeploymentPackagesResponse]*catalogv3.WatchDeploymentPackagesRequest{},
		},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{ActiveProjectID: footen}))
	s.ctx, s.cancel = context.WithTimeout(ctx, 5*time.Minute)
}

func (s *NorthBoundDBErrTestSuite) TearDownTest() {
	// make sure that the database mock is fully consumed
	s.NoError(s.mock.ExpectationsWereMet())
	_ = s.entClient.Close()
	_ = s.db.Close()
	s.cancel()
}

func (s *NorthBoundDBErrTestSuite) validateError(err error, code codes.Code, contains string, r any) {
	s.Error(err)
	s.Equal(code, status.Code(err))
	s.Contains(err.Error(), contains)
	s.Nil(r)
}

func (s *NorthBoundDBErrTestSuite) validateDBError(err error, r any) {
	s.Error(err)
	s.Equal(codes.Internal, status.Code(err))
	s.Nil(r)
	s.ErrorIs(err, status.Errorf(codes.Internal, `an internal database error occurred`))
}

func (s *NorthBoundDBErrTestSuite) validateInvalidArgumentError(err error, r any) {
	s.Error(err)
	s.Equal(codes.InvalidArgument, status.Code(err))
	s.Nil(r)
}

func (s *NorthBoundDBErrTestSuite) addMockedEmptyQueryRows(count int) {
	s.addMockedQueryRowsWithResult(count, 0)
}

func (s *NorthBoundDBErrTestSuite) addMockedQueryRowsWithResult(count int, result int) {
	var rows *sqlmock.Rows
	for i := 1; i <= count; i++ {
		rows = sqlmock.NewRows([]string{"count"}).AddRow(result)
		s.mock.ExpectQuery("SELECT .*").WillReturnRows(rows)
	}
}

func TestNorthBoundDBErr(t *testing.T) {
	suite.Run(t, &NorthBoundDBErrTestSuite{})
}

type testServerStream struct{}

func (*testServerStream) SetHeader(_ metadata.MD) error  { return nil }
func (*testServerStream) SendHeader(_ metadata.MD) error { return nil }
func (*testServerStream) SetTrailer(_ metadata.MD)       {}
func (*testServerStream) Context() context.Context       { return context.TODO() }
func (*testServerStream) SendMsg(_ any) error            { return nil }
func (*testServerStream) RecvMsg(_ any) error            { return nil }
