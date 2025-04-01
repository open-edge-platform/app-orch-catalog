// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	_ "github.com/mattn/go-sqlite3"
	ent "github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/enttest"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"strings"
	"testing"
)

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func newTestService(dbClient *ent.Client, opaClient openpolicyagent.ClientWithResponsesInterface) (Service, error) {
	return Service{DatabaseClient: dbClient, OpaClient: opaClient}, nil
}

func createServerConnection(t *testing.T, dbClient *ent.Client, opaClient openpolicyagent.ClientWithResponsesInterface) *grpc.ClientConn {
	lis = bufconn.Listen(1024 * 1024)
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
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}

func TestNewService(t *testing.T) {
	dbClient := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer dbClient.Close()

	s := NewService(dbClient, nil)
	assert.NotNil(t, s)
}

func TestServiceUnimplemented(t *testing.T) {
	t.Skip()
	dbClient := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer dbClient.Close()

	conn := createServerConnection(t, dbClient, nil)
	defer func() { _ = conn.Close() }()

	client := catalogv3.NewCatalogServiceClient(conn)

	_, err := client.GetDeploymentPackage(context.Background(), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "test application",
	})
	assert.ErrorIs(t, status.Errorf(codes.Unimplemented, "method GetDeploymentPackage not implemented"), err)
}

func checkFilters(t *testing.T, filters []*filter, wantedFieldList string, wantedValuesList string) {
	if len(filters) == 0 && wantedValuesList == "" {
		return
	}
	wantedFields := strings.Split(wantedFieldList, ",")
	wantedValues := strings.Split(wantedValuesList, ",")

	assert.Len(t, filters, len(wantedFields))
	for i := range wantedFields {
		assert.Equal(t, wantedFields[i], filters[i].name)
		assert.Equal(t, wantedValues[i], filters[i].value)
	}
}

func TestFiltersParsing(t *testing.T) {
	tests := map[string]struct {
		filter           string
		wantedFieldList  string
		wantedValuesList string
		expectedError    string
	}{
		"none":            {filter: "", wantedValuesList: "", wantedFieldList: ""},
		"single":          {filter: "field1=value1", wantedFieldList: "field1", wantedValuesList: "value1"},
		"double":          {filter: "name=acme OR description=widget company", wantedFieldList: "name,description", wantedValuesList: "acme,widget company"},
		"triple":          {filter: "f1=v1 OR f2=v2 OR f3=v3", wantedFieldList: "f1,f2,f3", wantedValuesList: "v1,v2,v3"},
		"equals error":    {filter: "=", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"two equals":      {filter: "= =", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"no field":        {filter: "=v1", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"no value":        {filter: "f1=", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"no equals":       {filter: "f1 v1", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"just OR":         {filter: "OR", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"hanging OR":      {filter: "f1=v1 OR f2=v2 OR", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
		"OR no left side": {filter: "OR f2=v2", wantedFieldList: "", wantedValuesList: "", expectedError: "invalid filter request"},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := parseFilter(testCase.filter, errors.RegistryType)
			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				checkFilters(t, resp, testCase.wantedFieldList, testCase.wantedValuesList)
			}
		})
	}
}

func TestComputePageRange(t *testing.T) {
	tests := map[string]struct {
		pageSize      int32
		offset        int32
		totalCount    int
		expectedStart int
		expectedEnd   int
		expectedError string
	}{
		"zeros":                  {},
		"whole array no page":    {pageSize: 0, totalCount: 10, expectedStart: 0, expectedEnd: 9},
		"whole array":            {pageSize: 10, offset: 0, totalCount: 10, expectedStart: 0, expectedEnd: 9},
		"page larger than array": {pageSize: 10, offset: 0, totalCount: 5, expectedStart: 0, expectedEnd: 4},
		"first page":             {pageSize: 10, offset: 0, totalCount: 35, expectedStart: 0, expectedEnd: 9},
		"second page":            {pageSize: 10, offset: 10, totalCount: 35, expectedStart: 10, expectedEnd: 19},
		"last page":              {pageSize: 10, offset: 30, totalCount: 35, expectedStart: 30, expectedEnd: 34},
		"0 pagesize":             {offset: 10, totalCount: 25, expectedStart: 10, expectedEnd: 24},
		"negative offset":        {pageSize: 10, offset: -10, totalCount: 35, expectedStart: 10, expectedEnd: 19, expectedError: "pagination: offset must not be negative"},
		"negative pageSize":      {pageSize: -10, offset: 30, totalCount: 35, expectedStart: 30, expectedEnd: 34, expectedError: "pagination: pageSize must not be negative"},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			start, end, totalElements, err := computePageRange(testCase.pageSize, testCase.offset, testCase.totalCount)
			if testCase.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedStart, start)
				assert.Equal(t, testCase.expectedEnd, end)
				assert.Equal(t, testCase.totalCount, int(totalElements))
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
			}
		})
	}
}
