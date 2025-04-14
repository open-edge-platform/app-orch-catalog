// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
)

func (s *NorthBoundTestSuite) TestCreateDeploymentPackage() {
	created, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			Kind:        catalogv3.Kind_KIND_NORMAL,
			DisplayName: "Test bundle",
			Description: "This is a Test",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "foo", Version: "v0.1.0"},
				{Name: "bar", Version: "v0.2.0"},
				{Name: "goo", Version: "v0.1.2"},
			},
			ApplicationDependencies: []*catalogv3.ApplicationDependency{
				{Name: "foo", Requires: "bar"}, {Name: "foo", Requires: "goo"},
			},
			DefaultNamespaces: map[string]string{
				"foo": "foons", "bar": "barns",
			},
			Profiles: []*catalogv3.DeploymentProfile{
				{
					Name:                "cp-1",
					DisplayName:         "CP1",
					Description:         "The profile",
					ApplicationProfiles: map[string]string{"foo": "p2", "bar": "p3", "goo": "p2"},
				},
			},
			DefaultProfileName: "cp-1",
			Extensions: []*catalogv3.APIExtension{
				{
					Name:        "ext1",
					Version:     "v0.1.1",
					DisplayName: "Extension 1",
					Description: "First extension",
					Endpoints: []*catalogv3.Endpoint{
						{ServiceName: "svc1", ExternalPath: "blah/blah", InternalPath: "yada/yada", Scheme: "http", AuthType: "insecure", AppName: "app1"},
					},
				},
				{
					Name:        "ext2",
					Version:     "v0.1.2",
					DisplayName: "Extension 2",
					Description: "Second extension",
					Endpoints: []*catalogv3.Endpoint{
						{ServiceName: "svc1a", ExternalPath: "sure/yeah", InternalPath: "whatever", Scheme: "http", AuthType: "insecure", AppName: "app1a"},
						{ServiceName: "svc2", ExternalPath: "uhm/no", InternalPath: "whatever", Scheme: "https", AuthType: "tls", AppName: "app2"},
					},
					UiExtension: &catalogv3.UIExtension{
						Label:       "Awesome",
						ServiceName: "svc2",
						Description: "Awesome description",
						FileName:    "index.html",
						AppName:     "svc2.exe",
						ModuleName:  "awesome-module",
					},
				},
			},
			Artifacts: []*catalogv3.ArtifactReference{
				{Name: "icon", Purpose: "ui-icon"},
				{Name: "thumb", Purpose: "ui-thumbnail"},
			},
			ForbidsMultipleDeployments: true,
		},
	})
	s.validateResponse(err, created)
	s.validateDeploymentPkg(created.DeploymentPackage, "test-ca", "v0.1.0", "Test bundle",
		"This is a Test", "icon", "thumb", 3, 2, 1, "cp-1", 2, false)
	s.Less(s.startTime, created.DeploymentPackage.CreateTime.AsTime())
	s.Len(created.DeploymentPackage.Extensions, 2)
	s.Len(created.DeploymentPackage.Artifacts, 2)
	s.True(created.DeploymentPackage.ForbidsMultipleDeployments)
	s.Equal(catalogv3.Kind_KIND_NORMAL, created.DeploymentPackage.Kind)

	// Fetch the newly created deployment package and test some basics.
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "test-ca", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "test-ca", "v0.1.0", "Test bundle",
		"This is a Test", "icon", "thumb", 3, 2, 1, "cp-1", 2, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 2)
	s.Len(resp.DeploymentPackage.Artifacts, 2)
	s.True(resp.DeploymentPackage.ForbidsMultipleDeployments)
	s.Equal(catalogv3.Kind_KIND_NORMAL, resp.DeploymentPackage.Kind)

	// Validate presence of the UI extension
	ext1 := resp.DeploymentPackage.Extensions[0]
	ext2 := resp.DeploymentPackage.Extensions[1]
	if ext2.Name != "ext2" {
		ext1, ext2 = ext2, ext1
	}
	s.Nil(ext1.UiExtension)
	s.NotNil(ext2.UiExtension)
	s.Equal("Awesome", ext2.UiExtension.Label)
	s.Equal("svc2", ext2.UiExtension.ServiceName)
	s.Equal("Awesome description", ext2.UiExtension.Description)
	s.Equal("index.html", ext2.UiExtension.FileName)
	s.Equal("svc2.exe", ext2.UiExtension.AppName)
	s.Equal("awesome-module", ext2.UiExtension.ModuleName)

	// Validate the scheme and insecure fields
	s.Len(ext1.Endpoints, 1)
	s.Equal("http", ext1.Endpoints[0].Scheme)
	s.Equal("insecure", ext1.Endpoints[0].AuthType)

	// Create one with duplicated name and version
	_, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "test-ca", Version: "v0.1.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`deployment-package test-ca:v0.1.0 invalid: deployment package test-ca already exists`))

	// Delete the application we created
	deleted, err := s.client.DeleteDeploymentPackage(s.ProjectID(footen), &catalogv3.DeleteDeploymentPackageRequest{
		DeploymentPackageName: "test-ca",
		Version:               "v0.1.0",
	})
	s.validateResponse(err, deleted)
}

func (s *NorthBoundTestSuite) TestCreateDeploymentPackageInvalidName() {
	// Create one with invalid name
	_, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "Another DeploymentPackage", Version: "v0.1.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`deployment-package invalid: invalid DeploymentPackage.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$"`))

	// Create one with invalid version
	_, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "bar", Version: "V 1",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`deployment-package invalid: invalid DeploymentPackage.Version: value does not match regex pattern "^[a-z0-9][a-z0-9-.]{0,18}[a-z0-9]{0,1}$"`))

	// Create one with invalid application
	_, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "bar", Version: "v0.3.0",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "non-existent-application", Version: "v0.2.0"},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application-reference non-existent-application:v0.2.0 invalid: application reference not found`))
}

func (s *NorthBoundTestSuite) TestCreateDeploymentPackageWithBadDependencies() {
	// Bad app named as a source of dependency
	_, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "test-ca", Version: "v0.1.0",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "foo", Version: "v0.1.0"},
				{Name: "bar", Version: "v0.2.0"},
				{Name: "goo", Version: "v0.1.2"},
			},
			ApplicationDependencies: []*catalogv3.ApplicationDependency{
				{Name: "none", Requires: "bar"}, {Name: "foo", Requires: "goo"},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, `deployment-package test-ca:v0.1.0 invalid: dependency for application none does not exist`))

	// Bad app named as a requirement
	_, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "test-ca", Version: "v0.1.0",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "foo", Version: "v0.1.0"},
				{Name: "bar", Version: "v0.2.0"},
				{Name: "goo", Version: "v0.1.2"},
			},
			ApplicationDependencies: []*catalogv3.ApplicationDependency{
				{Name: "foo", Requires: "none"}, {Name: "foo", Requires: "goo"},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, `deployment-package test-ca:v0.1.0 invalid: dependency requirement none does not exist`))

	// App named as itself
	_, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "test-ca", Version: "v0.1.0",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "foo", Version: "v0.1.0"},
				{Name: "bar", Version: "v0.2.0"},
				{Name: "goo", Version: "v0.1.2"},
			},
			ApplicationDependencies: []*catalogv3.ApplicationDependency{
				{Name: "foo", Requires: "foo"}, {Name: "foo", Requires: "goo"},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, `deployment-package test-ca:v0.1.0 invalid: application foo cannot depend on itself`))
}

func (s *NorthBoundTestSuite) TestCreateDeploymentPackageDisplayName() {
	// Creating two deployment_packages with blank display name should work
	resp, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{Name: "app1", Version: "1"},
	})
	s.validateResponse(err, resp)
	resp, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{Name: "app2", Version: "1"},
	})
	s.validateResponse(err, resp)

	// Creating two deployment_packages with the same, non-blank display name should not work
	resp, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{Name: "app3", Version: "1", DisplayName: "Bundle"},
	})
	s.validateResponse(err, resp)
	_, err = s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{Name: "app4", Version: "1", DisplayName: "bundle"},
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "deployment-package app4 already exists: deployment package app4 display name bundle is not unique"))
}

func (s *NorthBoundTestSuite) TestListDeploymentPackages() {
	pkgs, err := s.client.ListDeploymentPackages(s.ProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{})
	s.validateResponse(err, pkgs)
	s.Len(pkgs.DeploymentPackages, 3)

	// Test with invalid publisher
	pkgs, err = s.client.ListDeploymentPackages(s.ProjectID("non-existent"), &catalogv3.ListDeploymentPackagesRequest{})
	s.validateResponse(err, pkgs)
	s.Len(pkgs.DeploymentPackages, 0)

	// Test with kind filter
	pkgs, err = s.client.ListDeploymentPackages(s.ProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{
		Kinds: []catalogv3.Kind{catalogv3.Kind_KIND_EXTENSION, catalogv3.Kind_KIND_NORMAL},
	})
	s.validateResponse(err, pkgs)
	s.Equal(3, len(pkgs.DeploymentPackages))

	// Test with kind filter
	pkgs, err = s.client.ListDeploymentPackages(s.ProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{
		Kinds: []catalogv3.Kind{catalogv3.Kind_KIND_ADDON},
	})
	s.validateResponse(err, pkgs)
	s.Equal(0, len(pkgs.DeploymentPackages))
}

func (s *NorthBoundTestSuite) checkListDeploymentPackages(dps *catalogv3.ListDeploymentPackagesResponse, err error, values string, count int32, onlyLength bool) {
	s.validateResponse(err, dps)
	s.Equal(count, dps.GetTotalElements())
	if values == "" {
		s.Len(dps.DeploymentPackages, 0)
		return
	}
	expected := strings.Split(values, ",")
	s.Equal(len(expected), len(dps.DeploymentPackages))
	if !onlyLength {
		for i, name := range expected {
			dp := dps.DeploymentPackages[i]
			s.Equal(name, dp.Name)
		}
	}
}

func (s *NorthBoundTestSuite) generateDeploymentPackages(count int) {
	format := ""
	if count < 10 {
		format = "a%d"
	} else {
		format = "a%02d"
	}
	for i := 1; i <= count; i++ {
		dp := &catalogv3.DeploymentPackage{
			Name:        fmt.Sprintf(format, i),
			Description: "XXX",
			Version:     fmt.Sprintf("%d.0.0", i),
		}
		resp, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{DeploymentPackage: dp})
		s.NoError(err)
		s.NotNil(resp)
	}
}

func (s *NorthBoundTestSuite) TestListDeploymentPackagesWithOrderBy() {
	tests := map[string]struct {
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":             {orderBy: "", wantedList: "ca-gigi,ca-gigi,ca-fifi,a1,a2,a3"},
		"default":          {orderBy: "version", wantedList: "a1,a2,a3,ca-fifi,ca-gigi,ca-gigi"},
		"asc":              {orderBy: "name asc", wantedList: "a1,a2,a3,ca-fifi,ca-gigi,ca-gigi"},
		"desc":             {orderBy: "name desc", wantedList: "ca-gigi,ca-gigi,ca-fifi,a3,a2,a1"},
		"camel case field": {orderBy: "displayName desc", wantedList: "a3,a2,a1,ca-gigi,ca-gigi,ca-fifi"},
		"multi":            {orderBy: "description desc, name desc", wantedList: "a3,a2,a1,ca-gigi,ca-gigi,ca-fifi"},
		"too many":         {orderBy: "description asc desc", wantedList: "", expectedError: "invalid:"},
		"bad direction":    {orderBy: "description ascdesc", wantedList: "", expectedError: "invalid:"},
		"bad column":       {orderBy: "descriptionXXX", wantedList: "", expectedError: "invalid:"},
	}
	s.generateDeploymentPackages(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			dps, err := s.client.ListDeploymentPackages(s.ProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListDeploymentPackages(dps, err, testCase.wantedList, int32(len(dps.DeploymentPackages)), testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListDeploymentPackagesWithFilter() {
	tests := map[string]struct {
		filter        string
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":              {filter: "", wantedList: "a1,a2,a3,ca-fifi,ca-gigi,ca-gigi", orderBy: "name asc"},
		"single":            {filter: "name=a3", wantedList: "a3", orderBy: "name asc"},
		"camel case field":  {filter: "displayName=a3", wantedList: "a3", orderBy: "name asc"},
		"1 wildcard":        {filter: "name=*3", wantedList: "a3", orderBy: "name asc"},
		"2 wildcard":        {filter: "name=*ca-*", wantedList: "ca-fifi,ca-gigi,ca-gigi", orderBy: "name asc"},
		"match all":         {filter: "name=*", wantedList: "a1,a2,a3,ca-fifi,ca-gigi,ca-gigi", orderBy: "name asc"},
		"match all no sort": {filter: "name=*", wantedList: "ca-gigi,ca-gigi,ca-fifi,a1,a2,a3"},
		"or operation":      {filter: "name=*2* OR name=*gi*", wantedList: "a2,ca-gigi,ca-gigi", orderBy: "name asc"},
		"bad column":        {filter: "bad=filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
		"bad filter":        {filter: "bad filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
	}
	s.generateDeploymentPackages(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			dps, err := s.client.ListDeploymentPackages(s.ProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{Filter: testCase.filter, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListDeploymentPackages(dps, err, testCase.wantedList, int32(len(dps.DeploymentPackages)), testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListDeploymentPackagesWithPagination() {
	tests := map[string]struct {
		pageSize      int32
		offset        int32
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"first ten":         {pageSize: 10, offset: 0, wantedList: "a01,a02,a03,a04,a05,a06,a07,a08,a09,a10", orderBy: "name asc"},
		"second ten":        {pageSize: 10, offset: 10, wantedList: "a11,a12,a13,a14,a15,a16,a17,a18,a19,a20", orderBy: "name asc"},
		"last five":         {pageSize: 5, offset: 28, wantedList: "a29,a30,ca-fifi,ca-gigi,ca-gigi", orderBy: "name asc"},
		"0 pagesize":        {offset: 28, wantedList: "a29,a30,ca-fifi,ca-gigi,ca-gigi", orderBy: "name asc"},
		"default page size": {wantedList: "a01,a02,a03,a04,a05,a06,a07,a08,a09,a10,a11,a12,a13,a14,a15,a16,a17,a18,a19,a20", orderBy: "name asc"},
		"page size too big": {pageSize: 1000, expectedError: "must not exceed"},
		"negative offset":   {pageSize: 5, offset: -29, expectedError: "negative"},
		"negative pageSize": {pageSize: -5, offset: 29, expectedError: "negative"},
		"bad offset":        {pageSize: 10, offset: 41},
	}
	s.generateDeploymentPackages(30)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			dps, err := s.client.ListDeploymentPackages(s.ProjectID(footen),
				&catalogv3.ListDeploymentPackagesRequest{PageSize: testCase.pageSize, Offset: testCase.offset, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListDeploymentPackages(dps, err, testCase.wantedList, 33, testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestGetDeploymentPackageVersions() {
	apps, err := s.client.GetDeploymentPackageVersions(s.ProjectID(footen), &catalogv3.GetDeploymentPackageVersionsRequest{
		DeploymentPackageName: "ca-gigi",
	})
	s.validateResponse(err, apps)
	s.Len(apps.DeploymentPackages, 2)
}

func (s *NorthBoundTestSuite) TestGetDeploymentPackage() {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.3.4", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 3, 0, 2, "cp-1", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Equal(catalogv3.Kind_KIND_NORMAL, resp.DeploymentPackage.Kind)

	// Try one that does not exist
	_, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "not-present", Version: "v0.1.0",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "deployment-package not-present:v0.1.0 not found"))

	// Try version that does not exist
	_, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.4.2",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "deployment-package ca-gigi:v0.4.2 not found"))

	// Try publisher that does not exist
	_, err = s.client.GetDeploymentPackage(s.ProjectID("non-existent"), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1"})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "deployment-package ca-gigi:v0.2.1 not found"))
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackage() {
	// Get the deployment-package to update...
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.3.4", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 3, 0, 2, "cp-1", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())

	// Add application dependencies, extensions and artifact references.
	pkg := resp.DeploymentPackage
	pkg.ApplicationDependencies = []*catalogv3.ApplicationDependency{
		{Name: "bar", Requires: "foo"},
		{Name: "bar", Requires: "goo"},
	}
	pkg.DefaultNamespaces = map[string]string{"foo": "foons", "bar": "barns"}
	pkg.Extensions = []*catalogv3.APIExtension{
		{
			Name:        "ext1",
			Version:     "v0.1.1",
			DisplayName: "Extension 1",
			Description: "First extension",
			Endpoints: []*catalogv3.Endpoint{
				{ServiceName: "svc1", ExternalPath: "blah/blah", InternalPath: "yada/yada", Scheme: "http", AuthType: "insecure", AppName: "app1"},
			},
		},
		{
			Name:        "ext2",
			Version:     "v0.1.2",
			DisplayName: "Extension 2",
			Description: "Second extension",
			Endpoints: []*catalogv3.Endpoint{
				{ServiceName: "svc1a", ExternalPath: "sure/yeah", InternalPath: "whatever", AppName: "app1a"},
				{ServiceName: "svc2", ExternalPath: "uhm/no", InternalPath: "whatever", AppName: "app2"},
			},
			UiExtension: &catalogv3.UIExtension{
				Label:       "Awesome",
				ServiceName: "svc2",
				Description: "Awesome description",
				FileName:    "index.html",
				AppName:     "svc2.exe",
				ModuleName:  "awesome-module",
			},
		},
	}

	pkg.Artifacts = []*catalogv3.ArtifactReference{
		{Name: "icon", Purpose: "ui-icon"},
		{Name: "thumb", Purpose: "ui-thumbnail"},
	}
	pkg.Kind = catalogv3.Kind_KIND_EXTENSION

	update, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4", DeploymentPackage: pkg,
	})
	s.validateResponse(err, update)

	// Get the deployment-package to validate the update...
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.3.4", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 3, 2, 2, "cp-1", 2, false)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 2)
	s.Len(resp.DeploymentPackage.Artifacts, 2)
	s.Equal(catalogv3.Kind_KIND_EXTENSION, resp.DeploymentPackage.Kind)

	// Update a deployment-package changing the default profile
	pkg.DefaultProfileName = "cp-2"
	pkg.IsDeployed = true

	update, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4", DeploymentPackage: pkg,
	})
	s.validateResponse(err, update)

	// Get the application to validate the update...
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.3.4", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 3, 2, 2, "cp-2", 2, true)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())

	// Validate presence of the UI extension
	ext1 := resp.DeploymentPackage.Extensions[0]
	ext2 := resp.DeploymentPackage.Extensions[1]
	if ext2.Name != "ext2" {
		ext1, ext2 = ext2, ext1
	}
	s.Nil(ext1.UiExtension)
	s.NotNil(ext2.UiExtension)
	s.Equal("Awesome", ext2.UiExtension.Label)
	s.Equal("svc2", ext2.UiExtension.ServiceName)
	s.Equal("Awesome description", ext2.UiExtension.Description)
	s.Equal("index.html", ext2.UiExtension.FileName)
	s.Equal("svc2.exe", ext2.UiExtension.AppName)
	s.Equal("awesome-module", ext2.UiExtension.ModuleName)

	// Validate the scheme and insecure fields
	s.Len(ext1.Endpoints, 1)
	s.Equal("http", ext1.Endpoints[0].Scheme)
	s.Equal("insecure", ext1.Endpoints[0].AuthType)

	//================

	// Update with no changes - other than deployment status - should succeed
	update, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4", DeploymentPackage: pkg,
	})
	s.validateResponse(err, update)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.3.4", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 3, 2, 2, "cp-2", 2, true)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())

	// Update with a change should fail when deployed
	pkg.Description = "Boom"
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4", DeploymentPackage: pkg,
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition, "deployment-package ca-gigi:v0.3.4 failed precondition: cannot modify deployed package"))

	//================

	// Change deployed back to false
	pkg.Description = "New description"
	pkg.IsDeployed = false

	update, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4", DeploymentPackage: pkg,
	})
	s.validateResponse(err, update)

	// Get the application to validate the update...
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.3.4",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.3.4", "Deployment Package ca-gigi",
		"New description", "icon", "thumb", 3, 2, 2, "cp-2", 2, false)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())

	// Try one that does not exist
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "not-present", Version: "v0.2.0",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: "not-present", Version: "v0.2.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "deployment-package not-present:v0.2.0 not found"))
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageKindInDeployedState() {
	// Get a deployment package and mark it as deployed
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)

	resp.DeploymentPackage.IsDeployed = true
	none, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: resp.DeploymentPackage,
	})
	s.validateResponse(err, none)

	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.True(resp.DeploymentPackage.IsDeployed)
	s.Equal(catalogv3.Kind_KIND_NORMAL, resp.DeploymentPackage.Kind)

	resp.DeploymentPackage.Kind = catalogv3.Kind_KIND_EXTENSION
	none, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: resp.DeploymentPackage,
	})
	s.validateResponse(err, none)

	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.True(resp.DeploymentPackage.IsDeployed)
	s.Equal(catalogv3.Kind_KIND_EXTENSION, resp.DeploymentPackage.Kind)
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageWithIllegalInputs() {
	// Try to update dp name
	_, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi",
		Version:               "v0.3.4",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:    "some-other-name",
			Version: "v0.3.4",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package invalid: name cannot be changed ca-gigi != some-other-name"))

	// Try to update with no details
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package invalid: incomplete request"))

	// Try to update with invalid display name
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi",
		Version:               "v0.3.4",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			DisplayName: "    this display name has spaces    ",
			Name:        "ca-gigi",
			Version:     "v0.3.4",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package ca-gigi:v0.3.4 invalid: display name cannot contain leading or trailing spaces"))

	// Try to update dp version
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi",
		Version:               "v0.3.4",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:    "ca-gigi",
			Version: "v0.3.5",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package invalid: version cannot be changed v0.3.4 != v0.3.5"))

	// Try to update dp name
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi",
		Version:               "v0.3.4",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:    "ca-gigi-new",
			Version: "v0.3.4",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package invalid: name cannot be changed ca-gigi != ca-gigi-new"))

	// Try with missing default profile
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi",
		Version:               "v0.3.4",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:    "ca-gigi",
			Version: "v0.3.4",
			Profiles: []*catalogv3.DeploymentProfile{
				{
					Name: "p1",
				},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package invalid: default profile name must be specified"))

	// Try with default profile not found
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi",
		Version:               "v0.3.4",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			DefaultProfileName: "notfound",
			Name:               "ca-gigi",
			Version:            "v0.3.4",
			Profiles: []*catalogv3.DeploymentProfile{
				{
					Name: "p1",
				},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package ca-gigi:v0.3.4 invalid: deployment-profile notfound not found"))
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageWithNewReferences() {
	// Get the deployment-package to update...
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 2, 0, 2, "cp-1", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 0)
	s.Len(resp.DeploymentPackage.Artifacts, 0)

	// Update an deployment-package with a new set of app references, extensions and artifacts
	app := resp.DeploymentPackage
	app.ApplicationReferences = []*catalogv3.ApplicationReference{{Name: "foo", Version: "v0.1.0"}}
	app.ApplicationDependencies = []*catalogv3.ApplicationDependency{}
	delete(app.Profiles[0].ApplicationProfiles, "bar")
	delete(app.Profiles[1].ApplicationProfiles, "bar")
	app.DefaultNamespaces = map[string]string{"foo": "foons"}

	app.Extensions = []*catalogv3.APIExtension{
		{
			Name:        "ext1",
			Version:     "v0.1.1",
			DisplayName: "Extension 1",
			Description: "First extension",
			Endpoints: []*catalogv3.Endpoint{
				{ServiceName: "svc1", ExternalPath: "blah/blah", InternalPath: "yada/yada", AppName: "app1"},
			},
		},
		{
			Name:        "ext2",
			Version:     "v0.1.2",
			DisplayName: "Extension 2",
			Description: "Second extension",
			Endpoints: []*catalogv3.Endpoint{
				{ServiceName: "svc1a", ExternalPath: "sure/yeah", InternalPath: "whatever", AppName: "app1a"},
				{ServiceName: "svc2", ExternalPath: "uhm/no", InternalPath: "whatever", AppName: "app2"},
			},
			UiExtension: &catalogv3.UIExtension{
				Label:       "Awesome",
				ServiceName: "svc2",
				Description: "Awesome description",
				FileName:    "index.html",
				AppName:     "svc2.exe",
				ModuleName:  "awesome-module",
			},
		},
	}

	app.Artifacts = []*catalogv3.ArtifactReference{
		{Name: "icon", Purpose: "ui-icon"},
		{Name: "thumb", Purpose: "ui-thumbnail"},
	}

	update, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: app,
	})
	s.validateResponse(err, update)

	// Get the deployment-package to validate the update...
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 1, 0, 2, "cp-1", 1, false)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 2)
	s.Len(resp.DeploymentPackage.Artifacts, 2)

	// Update with empty list of extensions and artifacts
	app.Extensions = []*catalogv3.APIExtension{}
	app.Artifacts = []*catalogv3.ArtifactReference{}

	update, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: app,
	})
	s.validateResponse(err, update)

	// Get the deployment-package to validate the update...
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 1, 0, 2, "cp-1", 1, false)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 0)
	s.Len(resp.DeploymentPackage.Artifacts, 0)
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageRemoveApplication() {
	// Get the deployment-package to update...
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 2, 0, 2, "cp-1", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 0)
	s.Len(resp.DeploymentPackage.Artifacts, 0)

	// Update an application with a new set of namespaces
	dp := resp.DeploymentPackage
	dp.DefaultNamespaces = map[string]string{"foo": "foons", "bar": "bar-ns"}
	delete(dp.Profiles[0].ApplicationProfiles, "bar")
	delete(dp.Profiles[1].ApplicationProfiles, "bar")

	update, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: dp,
	})
	s.validateResponse(err, update)

	// Get the application to validate the update...
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 2, 0, 2, "cp-1", 2, false)
	s.Less(resp.DeploymentPackage.CreateTime.AsTime(), resp.DeploymentPackage.UpdateTime.AsTime())

	// Now delete app-ref foo, but do not remove it from namespaces - should break
	dp.ApplicationReferences = []*catalogv3.ApplicationReference{dp.ApplicationReferences[1]}

	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: dp,
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package ca-gigi:v0.2.1 invalid: application foo does not exist"))
}

// Test that we can delete all profiles and that Default Profile can be blank in that case
func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageDefaultProfile() {
	// Get the deployment-package to update...
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 2, 0, 2, "cp-1", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 0)
	s.Len(resp.DeploymentPackage.Artifacts, 0)

	// Remove the deployment profiles
	s.deleteDeploymentProfiles(footen, "ca-gigi", "v0.2.1", "cp-1", "cp-2")

	// Verify profile count has dropped to 1 (implicit profile) and default profile is the name of the implicit profile
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 2, 0, 1, "implicit-default", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 0)
	s.Len(resp.DeploymentPackage.Artifacts, 0)

	// Ensure deployment-package can still be updated
	dp := resp.DeploymentPackage
	dp.Description = "update after profiles removed"
	update, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: dp,
	})
	s.validateResponse(err, update)
	// Verify display name has been changed
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"update after profiles removed", "icon", "thumb", 2, 0, 1, "implicit-default", 0, false)
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageWithFewerReferencesSameDependencies() {
	// Get the application to update...
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, "ca-gigi", "v0.2.1", "Deployment Package ca-gigi",
		"This is deployment package ca-gigi", "icon", "thumb", 2, 0, 2, "cp-1", 0, false)
	s.Less(s.startTime, resp.DeploymentPackage.CreateTime.AsTime())
	s.Len(resp.DeploymentPackage.Extensions, 0)
	s.Len(resp.DeploymentPackage.Artifacts, 0)

	// Add dependencies for foo to bar
	app := resp.DeploymentPackage
	app.ApplicationDependencies = []*catalogv3.ApplicationDependency{{Name: "foo", Requires: "bar"}}
	update, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: app,
	})
	s.validateResponse(err, update)

	// Try to update an application without bar app reference, but same dependencies
	app.ApplicationReferences = []*catalogv3.ApplicationReference{{Name: "foo", Version: "v0.1.0"}}
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: app,
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package ca-gigi:v0.2.1 invalid: dependency target bar not found"))

	// Try to update an application without foo app reference, but same dependencies
	app.ApplicationReferences = []*catalogv3.ApplicationReference{{Name: "bar", Version: "v0.2.1"}}
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: app,
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "deployment-package ca-gigi:v0.2.1 invalid: dependency source foo not found"))
}

func (s *NorthBoundTestSuite) TestDeploymentPackageCreateNewProfileWhenDeployed() {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.NoError(err)

	resp.DeploymentPackage.IsDeployed = true
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: resp.DeploymentPackage,
	})
	s.NoError(err)

	resp.DeploymentPackage.Profiles = append(resp.DeploymentPackage.Profiles, &catalogv3.DeploymentProfile{
		Name:                "cp-new",
		ApplicationProfiles: map[string]string{"foo": "p1", "bar": "p2"},
	})
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: resp.DeploymentPackage,
	})
	s.NoError(err)

	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.NoError(err)
	s.Len(resp.DeploymentPackage.Profiles, 3)

	resp.DeploymentPackage.Profiles[0].DisplayName = "change"
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: resp.DeploymentPackage,
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition, "deployment-package ca-gigi:v0.2.1 failed precondition: cannot modify deployed package"))
}

func (s *NorthBoundTestSuite) TestDeleteDeploymentPackage() {
	deleted, err := s.client.DeleteDeploymentPackage(s.ProjectID(footen), &catalogv3.DeleteDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, deleted)

	_, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "deployment-package ca-gigi:v0.2.1 not found"))

	// Try deleting it again, i.e. non-existent - should return NotFound
	deleted, err = s.client.DeleteDeploymentPackage(s.ProjectID(footen), &catalogv3.DeleteDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateNotFound(err, deleted)
}

func (s *NorthBoundTestSuite) TestDeleteDeploymentPackageWithDependencies() {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)

	resp.DeploymentPackage.ApplicationDependencies = []*catalogv3.ApplicationDependency{{Name: "foo", Requires: "bar"}}
	// Mark it as deployed, to we can test the Updating and Deletion with - setting and unsetting
	resp.DeploymentPackage.IsDeployed = true

	updated, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
		DeploymentPackage: resp.DeploymentPackage,
	})
	s.validateResponse(err, updated)

	_, err = s.client.DeleteDeploymentPackage(s.ProjectID(footen), &catalogv3.DeleteDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition,
		"deployment-package ca-gigi:v0.2.1 failed precondition: cannot modify deployed package"))

	// Unmark as deployed
	resp.DeploymentPackage.IsDeployed = false

	updated, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
		DeploymentPackage: resp.DeploymentPackage,
	})
	s.validateResponse(err, updated)

	deleted, err := s.client.DeleteDeploymentPackage(s.ProjectID(footen), &catalogv3.DeleteDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, deleted)
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageNoProfile() {
	created, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "foo", Version: "v0.1.0"},
				{Name: "bar", Version: "v0.2.0"},
				{Name: "goo", Version: "v0.1.2"},
			},
		},
	})
	s.validateResponse(err, created)
	s.validateDeploymentPkg(created.DeploymentPackage, "test-ca", "v0.1.0", "Test bundle", "",
		"", "", 3, 0, 0, "", 0, false)

	created.DeploymentPackage.IsDeployed = true
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "test-ca", Version: "v0.1.0", DeploymentPackage: created.DeploymentPackage,
	})
	s.NoError(err)
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageAddProfile() {
	created, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
		},
	})
	s.validateResponse(err, created)

	update, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "test-ca", Version: "v0.1.0",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
			Profiles: []*catalogv3.DeploymentProfile{
				{
					Name:        "cp-1",
					DisplayName: "CP1",
					Description: "The profile",
				},
			},
			DefaultProfileName: "cp-1",
		},
	})
	s.validateResponse(err, update)
}

func (s *NorthBoundTestSuite) TestDeploymentPackageWithSameAppNames() {
	created, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "bar", Version: "v0.2.0"},
				{Name: "bar", Version: "v0.2.1"},
			},
			Profiles: []*catalogv3.DeploymentProfile{
				{Name: "cp1", ApplicationProfiles: map[string]string{
					"bar:v0.2.0": "p1",
					"bar:v0.2.1": "p2",
				}},
			},
			DefaultProfileName: "cp1",
		},
	})
	s.validateResponse(err, created)

	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "test-ca", Version: "v0.1.0",
	})
	s.validateResponse(err, created)
	s.Len(resp.DeploymentPackage.Profiles, 1)
	s.Len(resp.DeploymentPackage.Profiles[0].ApplicationProfiles, 2)
}

func (s *NorthBoundTestSuite) TestCreateDeploymentPackageWithoutFullyQualifiedProfiles() {
	_, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "bar", Version: "v0.2.0"},
				{Name: "bar", Version: "v0.2.1"},
			},
			Profiles: []*catalogv3.DeploymentProfile{
				{Name: "cp1", ApplicationProfiles: map[string]string{
					"bar": "p3",
				}},
			},
			DefaultProfileName: "cp1",
		},
	})
	s.ErrorContains(err, "fully qualified")
}

func (s *NorthBoundTestSuite) TestUpdateDeploymentPackageWithoutFullyQualifiedProfiles() {
	created, err := s.client.CreateDeploymentPackage(s.ProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "bar", Version: "v0.2.0"},
			},
			Profiles: []*catalogv3.DeploymentProfile{
				{Name: "cp1", ApplicationProfiles: map[string]string{
					"bar": "p3",
				}},
			},
			DefaultProfileName: "cp1",
		},
	})
	s.validateResponse(err, created)

	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "test-ca", Version: "v0.1.0",
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:        "test-ca",
			Version:     "v0.1.0",
			DisplayName: "Test bundle",
			ApplicationReferences: []*catalogv3.ApplicationReference{
				{Name: "bar", Version: "v0.2.0"},
				{Name: "bar", Version: "v0.2.1"},
			},
			Profiles: []*catalogv3.DeploymentProfile{
				{Name: "cp1", ApplicationProfiles: map[string]string{
					"bar": "p3",
				}},
			},
			DefaultProfileName: "cp1",
		},
	})
	s.ErrorContains(err, "fully qualified")
}

func (s *NorthBoundTestSuite) TestDeploymentPackageEvents() {
	ctx, cancel := context.WithCancel(s.ProjectID(footen))
	stream, err := s.client.WatchDeploymentPackages(ctx, &catalogv3.WatchDeploymentPackagesRequest{NoReplay: true})
	s.NoError(err)

	pkg := s.createDeploymentPkg(footen, fooreg, "newpkg", "0.1.1", "foo:v0.1.0", "bar:v0.2.1:barten")

	resp, err := stream.Recv()
	s.NoError(err)
	s.Equal(CreatedEvent, EventType(resp.Event.Type))
	s.validateDeploymentPkg(resp.DeploymentPackage, pkg.Name, pkg.Version, pkg.DisplayName, pkg.Description, "", "",
		2, 0, 0, "", 0, false)

	pkg.DisplayName = "New Package"
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: pkg.Name, Version: pkg.Version, DeploymentPackage: pkg,
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(UpdatedEvent, EventType(resp.Event.Type))
	s.validateDeploymentPkg(resp.DeploymentPackage, pkg.Name, pkg.Version, pkg.DisplayName, pkg.Description, "", "",
		2, 0, 0, "", 0, false)

	_, err = s.client.DeleteDeploymentPackage(s.ProjectID(footen), &catalogv3.DeleteDeploymentPackageRequest{
		DeploymentPackageName: pkg.Name, Version: pkg.Version,
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(DeletedEvent, EventType(resp.Event.Type))
	s.validateDeploymentPkg(resp.DeploymentPackage, pkg.Name, pkg.Version, "", "", "", "",
		0, 0, 0, "", 0, false)

	// Make sure we get an error back for a Recv() on a closed channel
	cancel()
	s.createDeploymentPkg(footen, fooreg, "newpkg1", "0.1.1", "foo:v0.1.0", "bar:v0.2.1:barten")
	resp, err = stream.Recv()
	s.Error(err)
	s.Nil(resp)
}

func (s *NorthBoundDBErrTestSuite) TestDPWatchInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.WatchDeploymentPackagesRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			err := s.server.WatchDeploymentPackages(testCase.req, nil)
			s.validateInvalidArgumentError(err, nil)
		})
	}
}

func (s *NorthBoundTestSuite) TestGetNoPackage() {
	_, err := s.client.GetDeploymentPackageVersions(s.ProjectID(footen), &catalogv3.GetDeploymentPackageVersionsRequest{
		DeploymentPackageName: "XXXX",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, `deployment-package XXXX not found`))
}

// FuzzCreateDeploymentPackage - fuzz test creating a deployment-package
//
// In this case we are calling the Test Suite to create a Publisher and Applications through gRPC
// but calling the function-under-test directly
//
// Invoke with:
//
//	go test ./internal/northbound -fuzz FuzzCreateDeploymentPackage -fuzztime=60s
func FuzzCreateDeploymentPackage(f *testing.F) {
	f.Add("test-deployment-package", "Test Registry", "v1.0.0")
	f.Add("test-deployment-package", " space at start", "v1.0.0")
	f.Add("test-deployment-package", "space at end ", "v1.0.0")
	f.Add("-", "starts with hyphen", "v1.0.0")
	f.Add("a", "Single letter OK", "v1.0.0")
	f.Add("a.", "contains .", "v1.0.0")
	f.Add("aaaaa-bbbb-cccc-dddd-eeee-ffff-gggg-hhhhh", "name too long > 40", "v1.0.0")
	f.Add("test-deployment-package", "display name is too long at 40 chars - here", "v1.0.0")
	f.Add("test-deployment-package", `display name contains
new line`, "v1.0.0")
	f.Add("test-deployment-package", "version too short", "")
	f.Add("test-deployment-package", "version invalid chars", "V 1")
	f.Add("test-deployment-package", "version too long", "v10000000.00000000.000000")

	s := &NorthBoundTestSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()

	f.Fuzz(func(t *testing.T, name string, displayName string, version string) {
		s.SetT(t)
		s.SetupTest() // SetupTest cannot be called until here because it depends on T's Assertions
		defer s.TearDownTest()

		server := Server{
			UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
			databaseClient:                    s.dbClient,
			listeners:                         NewEventListeners(),
		}

		// We call the function directly - not through gRPC
		created, err := server.CreateDeploymentPackage(s.ServerProjectID(footen), &catalogv3.CreateDeploymentPackageRequest{
			DeploymentPackage: &catalogv3.DeploymentPackage{
				Name:        name,
				DisplayName: displayName,
				Description: strings.Repeat(displayName, 20),
				Version:     version,
				ApplicationReferences: []*catalogv3.ApplicationReference{
					{
						Name:    "foo",
						Version: "v0.1.0",
					},
					{
						Name:    "goo",
						Version: "v0.1.2",
					},
				},
			},
		})
		if err != nil || created == nil {
			if err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: invalid DeploymentPackage.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$"` &&
				err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: invalid DeploymentPackage.DisplayName: value length must be between 0 and 40 runes, inclusive` &&
				err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: display name cannot contain leading or trailing spaces` &&
				err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: invalid DeploymentPackage.Name: value length must be between 1 and 40 runes, inclusive` &&
				err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: invalid DeploymentPackage.DisplayName: value does not match regex pattern "^\\PC*$"` &&
				err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: invalid DeploymentPackage.Version: value length must be between 1 and 20 runes, inclusive` &&
				err.Error() != `rpc error: code = InvalidArgument desc = deployment-package invalid: invalid DeploymentPackage.Version: value does not match regex pattern "^[a-z0-9][a-z0-9-.]{0,18}[a-z0-9]{0,1}$"` {
				t.Errorf("%v Name: %v DisplayName: %v", err.Error(), name, displayName)
			}
		}
	})

}

func (s *NorthBoundTestSuite) TestDeploymentPackageAuthErrors() {
	var err error
	server := s.newMockOPAServer()

	createResp, err := server.CreateDeploymentPackage(s.ServerProjectID(footen),
		&catalogv3.CreateDeploymentPackageRequest{
			DeploymentPackage: &catalogv3.DeploymentPackage{
				Name:    "ddd",
				Version: "111.222.333",
			},
		})
	s.Nil(createResp)
	s.ErrorIs(err, expectedAuthError)

	updateResp, err := server.UpdateDeploymentPackage(s.ServerProjectID(footen),
		&catalogv3.UpdateDeploymentPackageRequest{
			DeploymentPackageName: "ddd",
			Version:               "111.222.333",
			DeploymentPackage: &catalogv3.DeploymentPackage{
				Name:    "ddd",
				Version: "111.222.333",
			},
		})
	s.Nil(updateResp)
	s.ErrorIs(err, expectedAuthError)

	delResp, err := server.DeleteDeploymentPackage(s.ServerProjectID(footen),
		&catalogv3.DeleteDeploymentPackageRequest{DeploymentPackageName: "ddd", Version: "111.222.333"})
	s.Nil(delResp)
	s.ErrorIs(err, expectedAuthError)

	getResp, err := server.GetDeploymentPackage(s.ServerProjectID(footen),
		&catalogv3.GetDeploymentPackageRequest{DeploymentPackageName: "ddd", Version: "111.222.333"})
	s.Nil(getResp)
	s.ErrorIs(err, expectedAuthError)

	getVersionResp, err := server.GetDeploymentPackageVersions(s.ServerProjectID(footen),
		&catalogv3.GetDeploymentPackageVersionsRequest{DeploymentPackageName: "ddd"})
	s.Nil(getVersionResp)
	s.ErrorIs(err, expectedAuthError)

	listResp, err := server.ListDeploymentPackages(s.ServerProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{})
	s.Nil(listResp)
	s.ErrorIs(err, expectedAuthError)
}

func (s *NorthBoundDBErrTestSuite) TestDeploymentPkgCreateInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.CreateDeploymentPackageRequest
	}{
		"nil request":            {req: nil},
		"nil deployment package": {req: &catalogv3.CreateDeploymentPackageRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.CreateDeploymentPackage(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestDeploymentPackageListInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.ListDeploymentPackagesRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.ListDeploymentPackages(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestDeploymentPackageGetInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.GetDeploymentPackageRequest
	}{
		"nil request":        {req: nil},
		"empty package name": {req: &catalogv3.GetDeploymentPackageRequest{}},
		"empty version":      {req: &catalogv3.GetDeploymentPackageRequest{DeploymentPackageName: "p"}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.GetDeploymentPackage(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestDeploymentPackageUpdateInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.UpdateDeploymentPackageRequest
	}{
		"nil request":                   {req: nil},
		"empty deployment package":      {req: &catalogv3.UpdateDeploymentPackageRequest{}},
		"empty deployment package name": {req: &catalogv3.UpdateDeploymentPackageRequest{DeploymentPackage: &catalogv3.DeploymentPackage{}}},
		"empty version":                 {req: &catalogv3.UpdateDeploymentPackageRequest{DeploymentPackage: &catalogv3.DeploymentPackage{}, DeploymentPackageName: "p"}},
		"bad validate":                  {req: &catalogv3.UpdateDeploymentPackageRequest{DeploymentPackage: &catalogv3.DeploymentPackage{}, DeploymentPackageName: "p", Version: "1.2.3"}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.UpdateDeploymentPackage(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestDeploymentPackageDeleteInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.DeleteDeploymentPackageRequest
	}{
		"nil request":     {req: nil},
		"nil publisher":   {req: &catalogv3.DeleteDeploymentPackageRequest{}},
		"nil dep package": {req: &catalogv3.DeleteDeploymentPackageRequest{}},
		"nil version":     {req: &catalogv3.DeleteDeploymentPackageRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.DeleteDeploymentPackage(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestDeploymentPackageIncorrect() {
	deploymentPackageRequest := &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:     "p",
			Version:  "66.77.88",
			Profiles: []*catalogv3.DeploymentProfile{{Name: "p"}},
		},
	}
	s.mock.ExpectBegin()

	resp, err := s.server.CreateDeploymentPackage(s.ctx, deploymentPackageRequest)
	s.Nil(resp)
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`deployment-package invalid: default profile name must be specified`))
}

func (s *NorthBoundDBErrTestSuite) TestGetDeploymentPackageVersionInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.GetDeploymentPackageVersionsRequest
	}{
		"nil request": {req: nil},
		"no package":  {req: &catalogv3.GetDeploymentPackageVersionsRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			_, err := s.server.GetDeploymentPackageVersions(s.ctx, testCase.req)
			s.Error(err)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestDeleteDeploymentPackageDatabaseErrors() {
	deleteRequest := &catalogv3.DeleteDeploymentPackageRequest{DeploymentPackageName: "p", Version: "23"}

	// Test unable to start a transaction
	r, err := s.server.DeleteDeploymentPackage(s.ctx, deleteRequest)
	s.validateDBError(err, r)

	// Test can't query DP
	s.mock.ExpectBegin()
	r, err = s.server.DeleteDeploymentPackage(s.ctx, deleteRequest)
	s.validateDBError(err, r)

	// Test database error on delete
	s.mock.ExpectBegin()
	s.addMockedQueryRowsWithResult(1, 1)
	r, err = s.server.DeleteDeploymentPackage(s.ctx, deleteRequest)
	s.validateDBError(err, r)

	// Test nothing to delete
	s.mock.ExpectBegin()
	s.addMockedQueryRowsWithResult(1, 1)
	s.mock.ExpectExec("DELETE FROM .*").WillReturnResult(sqlmock.NewResult(1, 0))
	r, err = s.server.DeleteDeploymentPackage(s.ctx, deleteRequest)
	s.validateError(err, codes.NotFound, `deployment-package p:23 not found`, r)

	s.NoError(s.mock.ExpectationsWereMet())
}

func TestEmptyDepAppsDBQuery(t *testing.T) {
	s := &NorthBoundTestSuite{populateDB: false}
	s.SetT(t)
	s.SetupTest()

	apps, err := s.client.ListDeploymentPackages(s.ProjectID(footen), &catalogv3.ListDeploymentPackagesRequest{})
	s.validateResponse(err, apps)
	s.Equal(0, len(apps.DeploymentPackages))

	publisher, err := s.client.GetDeploymentPackage(s.ProjectID("nobody"), &catalogv3.GetDeploymentPackageRequest{
		Version:               "1.0",
		DeploymentPackageName: "none",
	})
	s.Nil(publisher)
	s.Error(err)
	s.Contains(err.Error(), "not found")
}
