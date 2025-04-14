// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"errors"
	"fmt"
	nberrors "github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
	"time"
)

var vaultError = status.Errorf(codes.Internal, `failed to access secret service`)

var errorOnRead bool
var errorOnWrite bool
var errorOnDelete bool

type TestSecretService struct {
}

func (TestSecretService) ReadSecret(_ context.Context, _ string) (string, error) {
	var err error
	if errorOnRead {
		err = errors.New("nothing to read here")
	}
	return "", err
}
func (TestSecretService) WriteSecret(_ context.Context, _ string, _ string) error {
	var err error
	if errorOnWrite {
		err = errors.New("nothing to write here")
	}
	return err
}
func (TestSecretService) DeleteSecret(_ context.Context, _ string) error {
	var err error
	if errorOnDelete {
		err = errors.New("nothing to delete here")
	}
	return err
}

func (TestSecretService) Logout(_ context.Context) {}

func testSecretServiceFactory(_ context.Context) (SecretService, error) {
	testService := &TestSecretService{}
	return testService, nil
}

func testSecretServiceFactoryError(_ context.Context) (SecretService, error) {
	return nil, errors.New("can't create secret service")
}

func (s *NorthBoundTestSuite) TestCreateRegistry() {
	created, err := s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{
			Name:        "test-registry",
			DisplayName: "Test registry",
			Description: "This is a Test",
			RootUrl:     "https://raw.githubusercontent.com/intel/DevcloudContent-helm/dev",
			Username:    "user",
			AuthToken:   "token",
			Cacerts:     "cacerts",
			Type:        helmType,
		},
	})
	s.validateResponse(err, created)
	s.validateRegistry(created.Registry, "test-registry", "Test registry", "This is a Test",
		"https://raw.githubusercontent.com/intel/DevcloudContent-helm/dev", "user", "token", "cacerts")

	resp, err := s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: "test-registry"})
	s.validateResponse(err, resp)
	s.validateRegistry(created.Registry, "test-registry", "Test registry", "This is a Test", "https://raw.githubusercontent.com/intel/DevcloudContent-helm/dev", "user", "token", "cacerts")

	// Create one with duplicated name
	_, err = s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{
			Name:    "test-registry",
			RootUrl: "http://test-reg.org",
			Type:    helmType,
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`registry test-registry invalid: registry test-registry already exists`))
}

func (s *NorthBoundTestSuite) TestCreateRegistryInvalidName() {
	// Create one with invalid name
	_, err := s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{Registry: &catalogv3.Registry{Name: "Third registry"}})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`registry invalid: invalid Registry.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$"`))
}

func (s *NorthBoundTestSuite) TestCreateRegistryDisplayName() {
	// Creating two registries with blank display name should work
	resp, err := s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{Name: "reg1", RootUrl: "https://test.com/", Type: helmType},
	})
	s.validateResponse(err, resp)
	resp, err = s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{Name: "reg2", RootUrl: "http://test.com?abc=def&ghi=jkl", Type: helmType},
	})
	s.validateResponse(err, resp)

	// Creating two registries with the same, non-blank display name should not work
	resp, err = s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{Name: "reg3", DisplayName: "Registry", RootUrl: "http://test", Type: helmType},
	})
	s.validateResponse(err, resp)
	_, err = s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{Name: "reg4", DisplayName: "registry", RootUrl: "http://test", Type: helmType},
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "registry reg4 already exists: registry reg4 display name registry is not unique"))
}

func (s *NorthBoundTestSuite) TestListRegistries() {
	registries, err := s.client.ListRegistries(s.ProjectID(barten), &catalogv3.ListRegistriesRequest{ShowSensitiveInfo: true})
	s.validateResponse(err, registries)
	s.Equal(2, len(registries.Registries))
	for _, r := range registries.Registries {
		s.True(strings.Contains(r.Name, "reg"), r.String())
		s.True(strings.HasPrefix(r.DisplayName, "Registry "), r.String())
		s.Equal("admin", r.Username)
		s.Equal("token", r.AuthToken)
		s.Equal("cacerts", r.Cacerts)
	}

	// Test without sensitive info included
	registries, err = s.client.ListRegistries(s.ProjectID(barten), &catalogv3.ListRegistriesRequest{})
	s.validateResponse(err, registries)
	s.Equal(2, len(registries.Registries))
	for _, r := range registries.Registries {
		s.True(strings.Contains(r.Name, "reg"), r.String())
		s.True(strings.HasPrefix(r.DisplayName, "Registry "), r.String())
		s.Equal("", r.Username)
		s.Equal("", r.AuthToken)
		s.Equal("", r.Cacerts)
	}

	// Test without specifying a project
	// FIXME: Remove this comment when ready to enforce empty activeprojectid metadata
	//_, err = s.client.ListRegistries(s.ctx, &catalogv3.ListRegistriesRequest{ShowSensitiveInfo: true})
	//s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "invalid: incomplete request: missing activeprojectid metadata"))
}

func (s *NorthBoundTestSuite) checkListRegistries(registries *catalogv3.ListRegistriesResponse, err error, values string, count int32, onlyLength bool) {
	s.validateResponse(err, registries)
	s.Equal(count, registries.GetTotalElements())
	if values == "" {
		s.Len(registries.Registries, 0)
		return
	}
	expected := strings.Split(values, ",")
	s.Len(registries.Registries, len(expected))
	if !onlyLength {
		for i, name := range expected {
			reg := registries.Registries[i]
			s.Equal(name, reg.Name)
		}
	}
}

func (s *NorthBoundTestSuite) generateRegistries(count int) {
	format := ""
	if count < 10 {
		format = "r%d"
	} else {
		format = "r%02d"
	}
	for i := 1; i <= count; i++ {
		reg := &catalogv3.Registry{
			Name:        fmt.Sprintf(format, i),
			Description: "XXX",
			RootUrl:     "http://intel.com",
			Type:        helmType,
		}
		resp, err := s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{Registry: reg})
		s.NoError(err)
		s.NotNil(resp)
	}
}

func (s *NorthBoundTestSuite) TestListRegistriesWithOrderBy() {
	tests := map[string]struct {
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":                 {orderBy: "", wantedList: "fooreg,fooregalt,r1,r2,r3"},
		"default":              {orderBy: "description", wantedList: "fooreg,fooregalt,r1,r2,r3"},
		"asc":                  {orderBy: "description asc", wantedList: "fooreg,fooregalt,r1,r2,r3"},
		"desc":                 {orderBy: "name desc", wantedList: "r3,r2,r1,fooregalt,fooreg"},
		"camel case field":     {orderBy: "displayName desc", wantedList: "r3,r2,r1,fooregalt,fooreg"},
		"multi":                {orderBy: "description asc, name desc", wantedList: "fooreg,fooregalt,r3,r2,r1"},
		"too many":             {orderBy: "description asc desc", wantedList: "", expectedError: "invalid:"},
		"bad direction":        {orderBy: "description ascdesc", wantedList: "", expectedError: "invalid:"},
		"bad column":           {orderBy: "descriptionXXX", wantedList: "", expectedError: "invalid:"},
		"not sortable rootUrl": {orderBy: "rootUrl", wantedList: "", expectedError: "cannot orderBy on attribute: rootUrl"},
		"not sortable cacerts": {orderBy: "cacerts", wantedList: "", expectedError: "cannot orderBy on attribute: cacerts"},
	}
	s.generateRegistries(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			registries, err := s.client.ListRegistries(s.ProjectID(footen), &catalogv3.ListRegistriesRequest{OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else if registries != nil {
				s.checkListRegistries(registries, err, testCase.wantedList, int32(len(registries.Registries)), testCase.orderBy == "")
			} else {
				s.Fail("unexpected test error: response is nil")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListRegistriesWithFilter() {
	tests := map[string]struct {
		filter        string
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":                 {filter: "", wantedList: "fooreg,fooregalt,r1,r2,r3", orderBy: "name asc"},
		"single":               {filter: "name=r1", wantedList: "r1", orderBy: "name asc"},
		"camel case field":     {filter: "displayName=r1", wantedList: "r1", orderBy: "name asc"},
		"1 wildcard":           {filter: "name=*1", wantedList: "r1", orderBy: "name asc"},
		"2 wildcard":           {filter: "name=*reg*", wantedList: "fooreg,fooregalt", orderBy: "name asc"},
		"match all":            {filter: "name=*", wantedList: "fooreg,fooregalt,r1,r2,r3", orderBy: "name asc"},
		"match all no sort":    {filter: "name=*", wantedList: "fooreg,fooregalt,r1,r2,r3"},
		"or operation":         {filter: "name=*2* OR name=*alt*", wantedList: "fooregalt,r2", orderBy: "name asc"},
		"contains":             {filter: "name=reg", wantedList: "fooreg,fooregalt", orderBy: "name asc"},
		"bad column":           {filter: "bad=filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
		"bad filter":           {filter: "bad filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
		"not sortable rootUrl": {filter: "rootUrl=rootUrl", wantedList: "", expectedError: "cannot filter on attribute: rootUrl"},
		"not sortable cacerts": {filter: "cacerts=cacert", wantedList: "", expectedError: "cannot filter on attribute: cacerts"},
	}
	s.generateRegistries(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			registries, err := s.client.ListRegistries(s.ProjectID(footen), &catalogv3.ListRegistriesRequest{Filter: testCase.filter, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListRegistries(registries, err, testCase.wantedList, int32(len(registries.Registries)), testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListRegistriesWithPagination() {
	tests := map[string]struct {
		pageSize      int32
		offset        int32
		orderBy       string
		wantedList    string
		expectedCount int32
		expectedError string
	}{
		"first ten":         {pageSize: 10, offset: 0, wantedList: "r30,r29,r28,r27,r26,r25,r24,r23,r22,r21", orderBy: "name desc", expectedCount: 32},
		"second ten":        {pageSize: 10, offset: 10, wantedList: "r20,r19,r18,r17,r16,r15,r14,r13,r12,r11", orderBy: "name desc", expectedCount: 32},
		"last two":          {pageSize: 5, offset: 30, wantedList: "fooregalt,fooreg", orderBy: "name desc", expectedCount: 32},
		"0 pagesize":        {offset: 30, wantedList: "fooregalt,fooreg", orderBy: "name desc", expectedCount: 32},
		"default page size": {wantedList: "r30,r29,r28,r27,r26,r25,r24,r23,r22,r21,r20,r19,r18,r17,r16,r15,r14,r13,r12,r11", orderBy: "name desc", expectedCount: 32},
		"page size too big": {pageSize: 1000, expectedError: "must not exceed"},
		"negative offset":   {pageSize: 5, offset: -29, wantedList: "a30,bar,bar,foo,goo", orderBy: "name asc", expectedError: "negative"},
		"negative pageSize": {pageSize: -5, offset: 29, wantedList: "a30,bar,bar,foo,goo", orderBy: "name asc", expectedError: "negative"},
		"bad offset":        {pageSize: 10, offset: 41, expectedCount: 32},
	}
	s.generateRegistries(30)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			registries, err := s.client.ListRegistries(s.ProjectID(footen),
				&catalogv3.ListRegistriesRequest{PageSize: testCase.pageSize, Offset: testCase.offset, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListRegistries(registries, err, testCase.wantedList, testCase.expectedCount, testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestGetRegistry() {
	resp, err := s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{
		RegistryName: fooreg, ShowSensitiveInfo: true,
	})
	s.validateResponse(err, resp)
	s.validateRegistry(resp.Registry, "fooreg", "Registry fooreg", "Registry that holds fooreg", "http://footen.com/fooreg", "admin", "token", "cacerts")
	s.Less(s.startTime, resp.Registry.CreateTime.AsTime())
	s.Less(s.startTime, resp.Registry.UpdateTime.AsTime())

	// Without sensitive info
	resp, err = s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.validateResponse(err, resp)
	s.validateRegistry(resp.Registry, "fooreg", "Registry fooreg", "Registry that holds fooreg", "http://footen.com/fooreg", "", "", "")

	// Try one that does not exist
	_, err = s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: "non-existent"})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "registry non-existent not found"))

	// Try project that does not exist
	_, err = s.client.GetRegistry(s.ProjectID("non-existent"), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "registry fooreg not found"))
}

func (s *NorthBoundTestSuite) TestUpdateRegistry() {
	_, err := s.client.UpdateRegistry(s.ProjectID(footen), &catalogv3.UpdateRegistryRequest{
		RegistryName: fooreg,
		Registry: &catalogv3.Registry{
			Name:        fooreg,
			DisplayName: "new display name",
			Description: "new description",
			RootUrl:     "http://new-url.org",
			Username:    "new user",
			AuthToken:   "new token",
			Cacerts:     "new certs",
			Type:        helmType,
		},
	})
	s.NoError(err)

	resp, err := s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg, ShowSensitiveInfo: true})
	s.validateResponse(err, resp)
	s.validateRegistry(resp.Registry, fooreg, "new display name", "new description", "http://new-url.org", "new user", "new token", "new certs")
	s.Less(resp.Registry.CreateTime.AsTime(), resp.Registry.UpdateTime.AsTime())

	// Try one that does not exist
	_, err = s.client.UpdateRegistry(s.ProjectID(footen), &catalogv3.UpdateRegistryRequest{
		RegistryName: "not-present",
		Registry: &catalogv3.Registry{
			Name:    "not-present",
			RootUrl: "http://test",
			Type:    helmType,
		},
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "registry not-present not found"))
}

func (s *NorthBoundTestSuite) TestUpdateRegistryType() {
	// Attempt to change the type of registry that is used already; it should fail
	_, err := s.client.UpdateRegistry(s.ProjectID(footen), &catalogv3.UpdateRegistryRequest{
		RegistryName: fooreg,
		Registry: &catalogv3.Registry{
			Name:        fooreg,
			DisplayName: "new display name",
			Description: "new description",
			RootUrl:     "http://new-url.org",
			Username:    "new user",
			AuthToken:   "new token",
			Cacerts:     "new certs",
			Type:        imageType,
		},
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition, "registry fooreg failed precondition: cannot change registry type to IMAGE"))

	// Attempt to change the registry name should be forbidden
	_, err = s.client.UpdateRegistry(s.ProjectID(footen), &catalogv3.UpdateRegistryRequest{
		RegistryName: fooreg,
		Registry: &catalogv3.Registry{
			Name:        "new-name",
			DisplayName: "new display name",
			Description: "new description",
			RootUrl:     "http://new-url.org",
			Username:    "new user",
			AuthToken:   "new token",
			Cacerts:     "new certs",
			Type:        helmType,
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "registry invalid: name cannot be changed fooreg != new-name"))

	// Attempt to change the type of registry that is not used yet; it should succeed
	_, err = s.client.UpdateRegistry(s.ProjectID(footen), &catalogv3.UpdateRegistryRequest{
		RegistryName: fooregalt,
		Registry: &catalogv3.Registry{
			Name:        fooregalt,
			DisplayName: "new display name",
			Description: "new description",
			RootUrl:     "http://new-url.org",
			Username:    "new user",
			AuthToken:   "new token",
			Cacerts:     "new certs",
			Type:        helmType,
		},
	})
	s.NoError(err)

	resp, err := s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooregalt, ShowSensitiveInfo: true})
	s.validateResponse(err, resp)
	s.validateRegistry(resp.Registry, fooregalt, "new display name", "new description", "http://new-url.org", "new user", "new token", "new certs")
	s.Less(resp.Registry.CreateTime.AsTime(), resp.Registry.UpdateTime.AsTime())
}

func (s *NorthBoundTestSuite) TestDeleteRegistry() {
	deleted, err := s.client.DeleteRegistry(s.ProjectID(axeten), &catalogv3.DeleteRegistryRequest{RegistryName: axereg})
	s.validateResponse(err, deleted)

	_, err = s.client.GetRegistry(s.ProjectID(axeten), &catalogv3.GetRegistryRequest{RegistryName: axereg})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "registry axereg not found"))

	// Try deleting non-existent - should return NotFound
	deleted, err = s.client.DeleteRegistry(s.ProjectID(axeten), &catalogv3.DeleteRegistryRequest{RegistryName: axereg})
	s.validateNotFound(err, deleted)

	// Set the image registry on foo app - get it first
	fooAppResp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
	})
	s.NoError(err)
	fooApp := fooAppResp.GetApplication()
	fooApp.ImageRegistryName = fooregalt
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application:     fooApp,
	})
	s.NoError(err)

	// Try deleting registry still in use by an application as HELM registry - should return InvalidArgument
	deleted, err = s.client.DeleteRegistry(s.ProjectID(footen), &catalogv3.DeleteRegistryRequest{RegistryName: fooreg})
	s.validateFailedPrecondition(err, deleted)

	// Try deleting registry still in use by an application as IMAGE registry - should return InvalidArgument
	deleted, err = s.client.DeleteRegistry(s.ProjectID(footen), &catalogv3.DeleteRegistryRequest{RegistryName: fooregalt})
	s.validateFailedPrecondition(err, deleted)
}

func (s *NorthBoundTestSuite) TestRegistryEvents() {
	ctx, cancel := context.WithCancel(s.ProjectID(barten))
	stream, err := s.client.WatchRegistries(ctx, &catalogv3.WatchRegistriesRequest{NoReplay: true})
	s.NoError(err)
	time.Sleep(100 * time.Millisecond) // Give the subscription a chance to take place

	reg := s.createRegistry(barten, "newreg", helmType)

	resp, err := stream.Recv()
	s.NoError(err)
	s.Equal(CreatedEvent, EventType(resp.Event.Type))
	s.validateRegistry(resp.Registry, reg.Name, reg.DisplayName, reg.Description, reg.RootUrl, reg.Username, reg.AuthToken, reg.Cacerts)

	reg.DisplayName = "New Registry"
	_, err = s.client.UpdateRegistry(s.ProjectID(barten), &catalogv3.UpdateRegistryRequest{
		RegistryName: "newreg", Registry: reg,
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(UpdatedEvent, EventType(resp.Event.Type))
	s.validateRegistry(resp.Registry, reg.Name, reg.DisplayName, reg.Description, reg.RootUrl, reg.Username, reg.AuthToken, reg.Cacerts)

	_, err = s.client.DeleteRegistry(s.ProjectID(barten), &catalogv3.DeleteRegistryRequest{RegistryName: "newreg"})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(DeletedEvent, EventType(resp.Event.Type))
	s.validateRegistry(resp.Registry, "newreg", "", "", "", "", "", "")

	// Make sure we get an error back for a Recv() on a closed channel
	cancel()
	s.createRegistry(barten, "newreg2", helmType)
	resp, err = stream.Recv()
	s.Error(err)
	s.Nil(resp)
}

func (s *NorthBoundDBErrTestSuite) TestRegistryWatchInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.WatchRegistriesRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			err := s.server.WatchRegistries(testCase.req, nil)
			s.validateInvalidArgumentError(err, nil)
		})
	}
}

// FuzzCreateRegistry - fuzz test creating a registry
//
// In this case we are calling the Test Suite to create a Publisher through gRPC
// but calling the function-under-test directly
//
// Invoke with:
//
//	go test ./internal/northbound -fuzz FuzzCreateRegistry -fuzztime=60s
func FuzzCreateRegistry(f *testing.F) {
	f.Add("test-registry", "Test Registry")
	f.Add("test-registry", " space at start")
	f.Add("test-registry", "space at end ")
	f.Add("-", "starts with hyphen")
	f.Add("a", "Single letter OK")
	f.Add("a.", "contains .")
	f.Add("aaaaa-bbbb-cccc-dddd-eeee-ffff-gggg-hhhhh", "name too long > 40")
	f.Add("test-registry", "display name is too long at 40 chars - here")
	f.Add("test-registry", `display name contains
new line`)

	s := &NorthBoundTestSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()

	f.Fuzz(func(t *testing.T, name string, displayName string) {
		s.SetT(t)
		s.SetupTest() // SetupTest cannot be called until here because it depends on T's Assertions
		defer s.TearDownTest()

		server := Server{
			UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
			databaseClient:                    s.dbClient,
			listeners:                         NewEventListeners(),
		}

		// We call the function directly - not through gRPC
		created, err := server.CreateRegistry(s.ServerProjectID(footen), &catalogv3.CreateRegistryRequest{
			Registry: &catalogv3.Registry{
				Name:        name,
				DisplayName: displayName,
				Description: strings.Repeat(displayName, 20),
				RootUrl:     fmt.Sprintf("https://%s.%s?%s=%s", name, name, name, name),
				Type:        helmType,
			},
		})
		if err != nil || created == nil {
			if err.Error() != `rpc error: code = InvalidArgument desc = registry invalid: invalid Registry.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$"` &&
				err.Error() != `rpc error: code = InvalidArgument desc = registry invalid: invalid Registry.Name: value length must be between 1 and 40 runes, inclusive` &&
				err.Error() != `rpc error: code = InvalidArgument desc = registry invalid: display name cannot contain leading or trailing spaces` &&
				err.Error() != `rpc error: code = InvalidArgument desc = registry invalid: invalid Registry.DisplayName: value length must be between 0 and 40 runes, inclusive` &&
				err.Error() != `rpc error: code = InvalidArgument desc = registry invalid: invalid Registry.DisplayName: value does not match regex pattern "^\\PC*$"` {
				t.Errorf("%v Name: %v DisplayName: %v", err.Error(), name, displayName)
			}
		}
	})

}

func (s *NorthBoundTestSuite) TestRegistryCreateErrorSecretWrite() {
	saveSecretFactory := SecretServiceFactory
	saveUseSecretService := UseSecretService
	defer func() {
		SecretServiceFactory = saveSecretFactory
		UseSecretService = saveUseSecretService
	}()
	server := Server{
		UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
		databaseClient:                    s.dbClient,
	}

	SecretServiceFactory = testSecretServiceFactory
	UseSecretService = true
	errorOnWrite = true
	var err error

	_, err = server.CreateRegistry(s.ServerProjectID(footen),
		&catalogv3.CreateRegistryRequest{
			Registry: &catalogv3.Registry{
				Name:    fooreg,
				RootUrl: "http://x.x",
				Type:    helmType,
			}})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "registry fooreg invalid: registry fooreg already exists"))
}

func getBaseUpdateRequest() *catalogv3.UpdateRegistryRequest {
	return &catalogv3.UpdateRegistryRequest{
		RegistryName: fooreg,
		Registry: &catalogv3.Registry{
			Name:        fooreg,
			DisplayName: "new display name",
			Description: "new description",
			RootUrl:     "http://new-url.org",
			Username:    "new user",
			AuthToken:   "new token",
			Cacerts:     "new certs",
			Type:        helmType,
		},
	}
}

func (s *NorthBoundTestSuite) TestRegistrySecretErrors() {
	saveSecretFactory := SecretServiceFactory
	saveUseSecretService := UseSecretService
	defer func() {
		SecretServiceFactory = saveSecretFactory
		UseSecretService = saveUseSecretService
	}()

	SecretServiceFactory = testSecretServiceFactory
	UseSecretService = true
	var err error

	errorOnWrite = true
	_, err = s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{
			Name:    "badpub",
			RootUrl: "http://badpub.com",
			Type:    helmType,
		}})
	s.ErrorIs(err, vaultError)

	errorOnRead = true
	errorOnWrite = false
	_, err = s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.ErrorIs(err, vaultError)

	errorOnRead = false
	errorOnWrite = true
	_, err = s.client.UpdateRegistry(s.ProjectID(footen), getBaseUpdateRequest())
	s.ErrorIs(err, vaultError)

	errorOnRead = true
	_, err = s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.ErrorIs(err, vaultError)

	errorOnRead = true
	_, err = s.client.ListRegistries(s.ProjectID(footen), &catalogv3.ListRegistriesRequest{})
	s.ErrorIs(err, vaultError)

	errorOnRead = false
	errorOnDelete = true
	deleteRequest := &catalogv3.DeleteRegistryRequest{RegistryName: axereg}
	// Test unable to start a transaction
	_, err = s.client.DeleteRegistry(s.ProjectID(axeten), deleteRequest)
	s.ErrorIs(err, vaultError)
}

func (s *NorthBoundDBErrTestSuite) TestRegistryCreateInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.CreateRegistryRequest
	}{
		"nil request":  {req: nil},
		"nil registry": {req: &catalogv3.CreateRegistryRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.CreateRegistry(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestRegistryUpdateInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.UpdateRegistryRequest
	}{
		"nil request":         {req: nil},
		"nil registry":        {req: &catalogv3.UpdateRegistryRequest{}},
		"empty registry name": {req: &catalogv3.UpdateRegistryRequest{RegistryName: ""}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.UpdateRegistry(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundTestSuite) TestRegistryUpdateErrors() {
	tests := map[string]struct {
		topLevelRegistryName string
		registryName         string
		displayName          string
		errorString          string
		errorCode            codes.Code
	}{
		"change registry name":  {topLevelRegistryName: "not-the-one", errorString: "registry invalid: name cannot be changed not-the-one != fooreg", errorCode: codes.InvalidArgument},
		"invalid registry name": {registryName: "invalid name!", errorString: `registry invalid: invalid Registry.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$"`, errorCode: codes.InvalidArgument},
		"invalid display name":  {displayName: "   invalid name!   ", errorString: "registry fooreg invalid: display name cannot contain leading or trailing spaces", errorCode: codes.InvalidArgument},
		"unique display name":   {registryName: fooreg, topLevelRegistryName: fooreg, displayName: "Registry fooregalt", errorString: "registry fooreg already exists: registry fooreg display name Registry fooregalt is not unique", errorCode: codes.AlreadyExists},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			req := getBaseUpdateRequest()
			if testCase.registryName != "" {
				req.Registry.Name = testCase.registryName
			}
			if testCase.topLevelRegistryName != "" {
				req.RegistryName = testCase.topLevelRegistryName
			}
			if testCase.displayName != "" {
				req.Registry.DisplayName = testCase.displayName
			}
			_, err := s.client.UpdateRegistry(s.ProjectID(footen), req)

			s.ErrorIs(err, status.Error(testCase.errorCode, testCase.errorString))
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestRegistryDeleteInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.DeleteRegistryRequest
	}{
		"nil request":         {req: nil},
		"nil registry":        {req: &catalogv3.DeleteRegistryRequest{}},
		"empty registry name": {req: &catalogv3.DeleteRegistryRequest{RegistryName: ""}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.DeleteRegistry(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestRegistryGetInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.GetRegistryRequest
	}{
		"nil request":         {req: nil},
		"nil registry":        {req: &catalogv3.GetRegistryRequest{}},
		"empty registry name": {req: &catalogv3.GetRegistryRequest{RegistryName: ""}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.GetRegistry(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestRegistryListInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.ListRegistriesRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.ListRegistries(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundTestSuite) TestRegistryAuthErrors() {
	var err error
	server := s.newMockOPAServer()

	_, err = server.CreateRegistry(s.ServerProjectID(footen),
		&catalogv3.CreateRegistryRequest{
			Registry: &catalogv3.Registry{
				Name:    fooreg,
				RootUrl: "http://x.x",
				Type:    helmType,
			}})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.UpdateRegistry(s.ServerProjectID(footen),
		&catalogv3.UpdateRegistryRequest{
			RegistryName: fooreg,
			Registry: &catalogv3.Registry{
				Name:    fooreg,
				RootUrl: "http://x.x",
				Type:    helmType,
			}})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.DeleteRegistry(s.ServerProjectID(footen), &catalogv3.DeleteRegistryRequest{RegistryName: fooreg})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.GetRegistry(s.ServerProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.ListRegistries(s.ServerProjectID(footen), &catalogv3.ListRegistriesRequest{})
	s.ErrorIs(err, expectedAuthError)
}

func (s *NorthBoundTestSuite) TestRegistrySecretServerCreationErrors() {
	saveSecretFactory := SecretServiceFactory
	saveUseSecretService := UseSecretService
	defer func() {
		SecretServiceFactory = saveSecretFactory
		UseSecretService = saveUseSecretService
	}()
	server := Server{
		UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
		databaseClient:                    s.dbClient,
	}

	SecretServiceFactory = testSecretServiceFactoryError
	UseSecretService = true
	var err error

	_, err = server.CreateRegistry(s.ProjectID(footen),
		&catalogv3.CreateRegistryRequest{
			Registry: &catalogv3.Registry{
				Name:    "xyzzy",
				RootUrl: "http://x.x",
				Type:    helmType,
			}})
	s.Error(err)

	_, err = s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.Error(err)

	_, err = s.client.UpdateRegistry(s.ProjectID(footen), &catalogv3.UpdateRegistryRequest{
		RegistryName: fooreg,
		Registry: &catalogv3.Registry{
			Name:        fooreg,
			DisplayName: "new display name",
			Description: "new description",
			RootUrl:     "http://new-url.org",
			Username:    "new user",
			AuthToken:   "new token",
			Cacerts:     "new certs",
			Type:        helmType,
		},
	})
	s.Error(err)

	_, err = s.client.ListRegistries(s.ctx, &catalogv3.ListRegistriesRequest{})
	s.Error(err)

	_, err = server.DeleteRegistry(s.ProjectID(axeten),
		&catalogv3.DeleteRegistryRequest{RegistryName: axereg})
	s.Error(err)
}

type base64Error struct{}

var errCannotDecodeError = errors.New("base64 cannot decode error")

func (b *base64Error) EncodeBase64(_ registrySecretData) string {
	return ""
}

func (b *base64Error) DecodeBase64(_ *registrySecretData, _ string) error {
	return errCannotDecodeError
}

func newBase64Error() Base64Strings {
	bs := &base64Error{}
	return bs
}

func (s *NorthBoundTestSuite) TestRegistryBase64Errors() {
	err := Base64Factory().DecodeBase64(&registrySecretData{}, "this is not Base64")
	s.Error(err)

	saveBase64Factory := Base64Factory
	defer func() { Base64Factory = saveBase64Factory }()
	Base64Factory = newBase64Error
	listResp, err := s.client.ListRegistries(s.ProjectID(footen), &catalogv3.ListRegistriesRequest{})
	s.Error(err)
	s.validateError(err, codes.Internal, listResp)

	getResp, err := s.client.GetRegistry(s.ProjectID(footen), &catalogv3.GetRegistryRequest{RegistryName: fooreg})
	s.validateError(err, codes.Internal, getResp)
}

var secretsMap = make(map[string]string)

type mappingSecretsService struct {
}

func newSecretsMap(_ context.Context) (SecretService, error) {
	return &mappingSecretsService{}, nil
}

func (m *mappingSecretsService) ReadSecret(_ context.Context, path string) (string, error) {
	value, ok := secretsMap[path]
	if ok {
		return value, nil
	}
	return "", nberrors.NewNotFound()
}

func (m *mappingSecretsService) WriteSecret(_ context.Context, path string, value string) error {
	secretsMap[path] = value
	return nil
}
func (m *mappingSecretsService) DeleteSecret(_ context.Context, path string) error {
	delete(secretsMap, path)
	return nil
}

func (m *mappingSecretsService) Logout(_ context.Context) {}

func (s *NorthBoundTestSuite) TestPublisherSecretOverlap() {
	var err error
	saveSecretFactory := SecretServiceFactory
	saveUseSecretService := UseSecretService
	defer func() {
		SecretServiceFactory = saveSecretFactory
		UseSecretService = saveUseSecretService
	}()
	server := Server{
		UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
		databaseClient:                    s.dbClient,
		listeners:                         NewEventListeners(),
	}

	SecretServiceFactory = newSecretsMap
	UseSecretService = true

	const project1 = "a"
	const registry1 = "b-c"
	const project2 = "a-b"
	const registry2 = "c"
	_, err = server.CreateRegistry(s.ServerProjectID(project1), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{
			Name:    registry1,
			RootUrl: "https://a.com",
			Type:    helmType,
		},
	})
	s.NoError(err)

	_, err = server.CreateRegistry(s.ServerProjectID(project2), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{
			Name:    registry2,
			RootUrl: "https://a-b.com",
			Type:    helmType,
		},
	})
	s.NoError(err)

	reg, err := server.GetRegistry(s.ServerProjectID(project1), &catalogv3.GetRegistryRequest{
		RegistryName: registry1,
	})
	s.NoError(err)
	s.Equal("https://a.com", reg.Registry.RootUrl)

}

func TestEmptyRegistriesDBQuery(t *testing.T) {
	s := &NorthBoundTestSuite{populateDB: false}
	s.SetT(t)
	s.SetupTest()

	apps, err := s.client.ListRegistries(s.ProjectID(footen), &catalogv3.ListRegistriesRequest{})
	s.validateResponse(err, apps)
	s.Equal(0, len(apps.Registries))

	project, err := s.client.GetRegistry(s.ProjectID("nobody"), &catalogv3.GetRegistryRequest{
		RegistryName: "none",
	})
	s.Nil(project)
	s.Error(err)
	s.Contains(err.Error(), "not found")
}

func (s *NorthBoundTestSuite) TestOCIRegistry() {
	name := "reg"
	reg := &catalogv3.Registry{
		Name:        name,
		DisplayName: fmt.Sprintf("Registry %s", name),
		Description: fmt.Sprintf("Registry that holds %s", name),
		RootUrl:     fmt.Sprintf("oci://%s.com/%s", footen, name),
		Username:    "admin",
		AuthToken:   "token",
		Cacerts:     "cacerts",
		Type:        helmType,
	}
	resp, err := s.client.CreateRegistry(s.ProjectID(footen), &catalogv3.CreateRegistryRequest{Registry: reg})
	s.NoError(err)
	s.NotNil(resp)
}
