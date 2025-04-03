// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *NorthBoundTestSuite) TestCreateApplication() {
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Kind:               catalogv3.Kind_KIND_NORMAL,
			Name:               "test-application",
			DisplayName:        "Test application",
			Description:        "This is a Test",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "profile-1",
			Profiles: []*catalogv3.Profile{
				{
					Name:        "profile-1",
					DisplayName: "Profile 1",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
				{
					Name:        "profile-2",
					DisplayName: "Profile 2",
					ChartValues: "key1b: value1b\nkey2b: value2b\n",
				},
			},
			IgnoredResources: []*catalogv3.ResourceReference{
				{Name: "foo", Kind: "ConfigMap"}, {Name: "bar", Kind: "SomeKind"},
			},
		},
	})
	s.validateResponse(err, created)
	s.validateApp(created.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Len(created.Application.IgnoredResources, 2)
	s.Equal(fooregalt, created.Application.ImageRegistryName)
	s.Equal(catalogv3.Kind_KIND_NORMAL, created.Application.Kind)
	s.Equal(1, len(created.Application.Profiles[0].DeploymentRequirement)+len(created.Application.Profiles[1].DeploymentRequirement))

	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Len(created.Application.IgnoredResources, 2)
	s.Equal(fooregalt, resp.Application.ImageRegistryName)
	s.Equal(catalogv3.Kind_KIND_NORMAL, created.Application.Kind)
	s.Equal(1, len(resp.Application.Profiles[0].DeploymentRequirement)+len(resp.Application.Profiles[1].DeploymentRequirement))

	// Create one with duplicated name and version
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "test-application",
			DisplayName:      "test application duplicate",
			Version:          "0.1.0",
			ChartName:        "test-chart",
			ChartVersion:     "v0.1.1",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application test-application invalid: deployment application test-application already exists`))

	// Test with no profiles, but default profile set
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			Name:               "test-application-minimal",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p1",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application test-application-minimal:0.1.0 invalid: could not update default profile p1`))

	// Test with profile, but not default-profile
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "test-application",
			DisplayName:        "Test application",
			Description:        "This is a Test",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "",
			Profiles: []*catalogv3.Profile{
				{
					Name:        "profile-1",
					DisplayName: "Profile 1",
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
				{
					Name:        "profile-2",
					DisplayName: "Profile 2",
					ChartValues: "key1b: value1b\nkey2b: value2b\n",
				},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: default profile name must be specified`))

	// Test with invalid display name
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "test-application-minimal",
			DisplayName:      "      this has spaces   ",
			Version:          "0.1.0",
			ChartName:        "test-chart",
			ChartVersion:     "0.1.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: display name cannot contain leading or trailing spaces`))

	// test empty
	// FIXME: Remove this comment when ready to enforce empty activeprojectid metadata
	//_, err = s.client.CreateApplication(s.ctx, &catalogv3.CreateApplicationRequest{})
	//s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
	//	`invalid: incomplete request: missing activeprojectid metadata`))
}

func (s *NorthBoundTestSuite) TestCreateApplicationInvalidName() {
	// Create one with invalid name
	_, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{Name: "Another Application", Version: ""},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: invalid Application.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,24}[a-z0-9]{0,1}$"`))

	// Create one with invalid version
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{Name: "another-application", Version: "V 1"},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: invalid Application.Version: value does not match regex pattern "^[a-z0-9][a-z0-9-.]{0,18}[a-z0-9]{0,1}$"`))

	// Create one with invalid registry
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName: "non-existent-registry",
			Name:             "fifth-application",
			DisplayName:      "A fifth application",
			Version:          "0.1.0",
			ChartName:        "test-chart",
			ChartVersion:     "v0.1.1",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, `application invalid: helm registry non-existent-registry not found`))
}

func (s *NorthBoundTestSuite) TestCreateWithLongChartVersion() {
	longChartVersion := "1.3.5678901234567890123456789012345678901234567890123"
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Kind:               catalogv3.Kind_KIND_NORMAL,
			Name:               "test-application",
			DisplayName:        "Test application",
			Description:        "This is a Test",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       longChartVersion,
			DefaultProfileName: "profile-1",
			Profiles: []*catalogv3.Profile{
				{
					Name:        "profile-1",
					DisplayName: "Profile 1",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
				{
					Name:        "profile-2",
					DisplayName: "Profile 2",
					ChartValues: "key1b: value1b\nkey2b: value2b\n",
				},
			},
			IgnoredResources: []*catalogv3.ResourceReference{
				{Name: "foo", Kind: "ConfigMap"}, {Name: "bar", Kind: "SomeKind"},
			},
		},
	})
	s.validateResponse(err, created)
	s.validateApp(created.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", longChartVersion, fooreg)
	s.Len(created.Application.IgnoredResources, 2)
	s.Equal(fooregalt, created.Application.ImageRegistryName)
	s.Equal(catalogv3.Kind_KIND_NORMAL, created.Application.Kind)
	s.Equal(1, len(created.Application.Profiles[0].DeploymentRequirement)+len(created.Application.Profiles[1].DeploymentRequirement))

	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", longChartVersion, fooreg)
	s.Len(created.Application.IgnoredResources, 2)
	s.Equal(fooregalt, resp.Application.ImageRegistryName)
	s.Equal(catalogv3.Kind_KIND_NORMAL, created.Application.Kind)
	s.Equal(1, len(resp.Application.Profiles[0].DeploymentRequirement)+len(resp.Application.Profiles[1].DeploymentRequirement))

	// Create one with duplicated name and version
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "test-application",
			DisplayName:      "test application duplicate",
			Version:          "0.1.0",
			ChartName:        "test-chart",
			ChartVersion:     "v0.1.1",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application test-application invalid: deployment application test-application already exists`))

	// Test with no profiles, but default profile set
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			Name:               "test-application-minimal",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p1",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application test-application-minimal:0.1.0 invalid: could not update default profile p1`))

	// Test with profile, but not default-profile
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "test-application",
			DisplayName:        "Test application",
			Description:        "This is a Test",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "",
			Profiles: []*catalogv3.Profile{
				{
					Name:        "profile-1",
					DisplayName: "Profile 1",
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
				{
					Name:        "profile-2",
					DisplayName: "Profile 2",
					ChartValues: "key1b: value1b\nkey2b: value2b\n",
				},
			},
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: default profile name must be specified`))

	// Test with invalid display name
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "test-application-minimal",
			DisplayName:      "      this has spaces   ",
			Version:          "0.1.0",
			ChartName:        "test-chart",
			ChartVersion:     "0.1.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: display name cannot contain leading or trailing spaces`))

	// test empty
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`application invalid: incomplete request`))
}

func (s *NorthBoundTestSuite) TestCreateApplicationDisplayName() {
	// Creating two applications with blank display name should work
	resp, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{Name: "app1", Version: "1",
			ChartName: "c", ChartVersion: "1", HelmRegistryName: fooreg},
	})
	s.validateResponse(err, resp)
	resp, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{Name: "app2", Version: "1",
			ChartName: "c", ChartVersion: "1", HelmRegistryName: fooreg},
	})
	s.validateResponse(err, resp)

	// Creating two applications with the same, non-blank display name should not work
	resp, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{Name: "app3", Version: "1", DisplayName: "Application",
			ChartName: "c", ChartVersion: "1", HelmRegistryName: fooreg},
	})
	s.validateResponse(err, resp)
	_, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{Name: "app4", Version: "1", DisplayName: "application",
			ChartName: "c", ChartVersion: "1", HelmRegistryName: fooreg},
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "application app4:1 already exists: display name already exists"))
}

func (s *NorthBoundTestSuite) TestCreateApplicationsWithSameDeploymentPackageDependency() {
	// Create an app specifying a package/profile dependency.
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "app1",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p1",
			Profiles: []*catalogv3.Profile{
				{
					Name: "p1",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.validateResponse(err, created)

	// Create another app specifying the same package/profile dependency.
	created, err = s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "app2",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p2",
			Profiles: []*catalogv3.Profile{
				{
					Name: "p2", DisplayName: "p2", Description: "p2",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.validateResponse(err, created)

	// Make sure that the dependency is indeed there.
	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "app2", Version: "0.1.0",
	})
	s.validateResponse(err, resp)
	if s.Len(resp.Application.Profiles, 1) {
		if s.Len(resp.Application.Profiles[0].DeploymentRequirement, 1) {
			s.Equal("cp-2", resp.Application.Profiles[0].DeploymentRequirement[0].DeploymentProfileName)
		}
	}

	// Update the second app to update the package/profile dependency to a different profile.
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "app2", Version: "0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "app2",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p2",
			Profiles: []*catalogv3.Profile{
				{
					Name: "p2", DisplayName: "p2", Description: "p2",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-1"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.NoError(err)

	// Make sure that the dependency is indeed updated.
	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "app2", Version: "0.1.0",
	})
	s.validateResponse(err, resp)
	if s.Len(resp.Application.Profiles, 1) {
		if s.Len(resp.Application.Profiles[0].DeploymentRequirement, 1) {
			s.Equal("cp-1", resp.Application.Profiles[0].DeploymentRequirement[0].DeploymentProfileName)
		}
	}

	// Update the second app to remove the package/profile dependency.
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "app2", Version: "0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "app2",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p2",
			Profiles: []*catalogv3.Profile{
				{
					Name: "p2", DisplayName: "p2", Description: "p2",
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.NoError(err)

	// Make sure that the dependency is indeed gone.
	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "app2", Version: "0.1.0",
	})
	s.validateResponse(err, resp)
	if s.Len(resp.Application.Profiles, 1) {
		s.Len(resp.Application.Profiles[0].DeploymentRequirement, 0)
	}
}

func (s *NorthBoundTestSuite) TestDeleteApplicationWithDeploymentRequirements() {
	// Create an app specifying a package/profile dependency.
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "app1",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "p1",
			Profiles: []*catalogv3.Profile{
				{
					Name: "p1",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.validateResponse(err, created)

	// Now remove that deployment package
	deleted, err := s.client.DeleteApplication(s.ProjectID(footen), &catalogv3.DeleteApplicationRequest{
		ApplicationName: "app1", Version: "0.1.0",
	})
	s.validateResponse(err, deleted)
}

func (s *NorthBoundTestSuite) TestListApplications() {
	applications, err := s.client.ListApplications(s.ProjectID(footen), &catalogv3.ListApplicationsRequest{})
	s.validateResponse(err, applications)
	s.Len(applications.Applications, 4)

	// Test with invalid publisher
	applications, err = s.client.ListApplications(s.ProjectID("non-existent"), &catalogv3.ListApplicationsRequest{})
	s.validateResponse(err, applications)
	s.Len(applications.Applications, 0)

	// Test with kind filter
	applications, err = s.client.ListApplications(s.ProjectID(footen), &catalogv3.ListApplicationsRequest{
		Kinds: []catalogv3.Kind{catalogv3.Kind_KIND_EXTENSION, catalogv3.Kind_KIND_NORMAL},
	})
	s.validateResponse(err, applications)
	s.Len(applications.Applications, 4)

	// Test with kind filter
	applications, err = s.client.ListApplications(s.ProjectID(footen), &catalogv3.ListApplicationsRequest{
		Kinds: []catalogv3.Kind{catalogv3.Kind_KIND_ADDON},
	})
	s.validateResponse(err, applications)
	s.Equal(0, len(applications.Applications))
}

func (s *NorthBoundTestSuite) checkListApplications(applications *catalogv3.ListApplicationsResponse, err error, values string, count int32, onlyLength bool) {
	s.validateResponse(err, applications)
	s.Equal(count, applications.GetTotalElements())
	if values == "" {
		s.Len(applications.Applications, 0)
		return
	}
	expected := strings.Split(values, ",")
	s.Len(applications.Applications, len(expected))
	if !onlyLength {
		for i, name := range expected {
			app := applications.Applications[i]
			s.Equal(name, app.Name)
		}
	}
}

func (s *NorthBoundTestSuite) generateApplications(count int) {
	format := ""
	if count < 10 {
		format = "a%d"
	} else {
		format = "a%02d"
	}
	for i := 1; i <= count; i++ {
		app := &catalogv3.Application{
			Name:             fmt.Sprintf(format, i),
			Description:      "XXX",
			Version:          "1.2.3",
			ChartVersion:     "1.2.3",
			ChartName:        fmt.Sprintf("chart%d", i),
			HelmRegistryName: fooreg,
		}
		resp, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{Application: app})
		s.NoError(err)
		s.NotNil(resp)
	}
}

func (s *NorthBoundTestSuite) TestListApplicationsWithOrderBy() {
	tests := map[string]struct {
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":             {orderBy: "", wantedList: "foo,bar,bar,goo,a1,a2,a3"},
		"default":          {orderBy: "description", wantedList: "bar,bar,foo,goo,a1,a2,a3"},
		"asc":              {orderBy: "description asc", wantedList: "bar,bar,foo,goo,a1,a2,a3"},
		"desc":             {orderBy: "name desc", wantedList: "goo,foo,bar,bar,a3,a2,a1"},
		"camel case field": {orderBy: "displayName", wantedList: "bar,bar,foo,goo,a1,a2,a3"},
		"multi":            {orderBy: "description asc, name desc", wantedList: "bar,bar,foo,goo,a3,a2,a1"},
		"too many":         {orderBy: "description asc desc", wantedList: "", expectedError: "invalid:"},
		"bad direction":    {orderBy: "description ascdesc", wantedList: "", expectedError: "invalid:"},
		"bad column":       {orderBy: "descriptionXXX", wantedList: "", expectedError: "invalid:"},
	}
	s.generateApplications(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			applications, err := s.client.ListApplications(s.ProjectID(footen), &catalogv3.ListApplicationsRequest{OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListApplications(applications, err, testCase.wantedList, 7, testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListApplicationsWithFilter() {
	tests := map[string]struct {
		filter        string
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":                   {filter: "", wantedList: "a1,a2,a3,bar,bar,foo,goo", orderBy: "name asc"},
		"single":                 {filter: "name=a1", wantedList: "a1", orderBy: "name asc"},
		"camel case field":       {filter: "displayName=a1", wantedList: "a1", orderBy: "name asc"},
		"1 wildcard":             {filter: "name=*1", wantedList: "a1", orderBy: "name asc"},
		"2 wildcard":             {filter: "name=*a*", wantedList: "a1,a2,a3,bar,bar", orderBy: "name asc"},
		"match all":              {filter: "name=*", wantedList: "a1,a2,a3,bar,bar,foo,goo", orderBy: "name asc"},
		"match all no sort":      {filter: "name=*", wantedList: "foo,bar,bar,goo,a1,a2,a3"},
		"or operation":           {filter: "name=*ar OR name=*oo", wantedList: "bar,bar,foo,goo", orderBy: "name asc"},
		"contains":               {filter: "name=o", wantedList: "foo,goo", orderBy: "name asc"},
		"bad column":             {filter: "bad=filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
		"bad filter":             {filter: "bad filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
		"case insensitive data":  {filter: "displayName=GOO", wantedList: "goo"},
		"case insensitive query": {filter: "displayName=application goo", wantedList: "goo"},
	}
	s.generateApplications(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			applications, err := s.client.ListApplications(s.ProjectID(footen), &catalogv3.ListApplicationsRequest{Filter: testCase.filter, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListApplications(applications, err, testCase.wantedList, int32(len(applications.Applications)), testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListApplicationsWithPagination() {
	tests := map[string]struct {
		pageSize      int32
		offset        int32
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"first ten":         {pageSize: 10, offset: 0, wantedList: "a01,a02,a03,a04,a05,a06,a07,a08,a09,a10", orderBy: "name asc"},
		"second ten":        {pageSize: 10, offset: 10, wantedList: "a11,a12,a13,a14,a15,a16,a17,a18,a19,a20", orderBy: "name asc"},
		"last five":         {pageSize: 5, offset: 29, wantedList: "a30,bar,bar,foo,goo", orderBy: "name asc"},
		"0 page size":       {offset: 29, wantedList: "a30,bar,bar,foo,goo", orderBy: "name asc"},
		"default page size": {wantedList: "a01,a02,a03,a04,a05,a06,a07,a08,a09,a10,a11,a12,a13,a14,a15,a16,a17,a18,a19,a20", orderBy: "name asc"},
		"page size too big": {pageSize: 1000, expectedError: "must not exceed"},
		"negative offset":   {pageSize: 5, offset: -29, expectedError: "negative"},
		"negative pageSize": {pageSize: -5, offset: 29, expectedError: "negative"},
		"bad offset":        {pageSize: 10, offset: 41},
	}
	s.generateApplications(30)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			apps, err := s.client.ListApplications(s.ProjectID(footen),
				&catalogv3.ListApplicationsRequest{PageSize: testCase.pageSize, Offset: testCase.offset, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListApplications(apps, err, testCase.wantedList, 34, testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestGetApplicationVersions() {
	applications, err := s.client.GetApplicationVersions(s.ProjectID(barten), &catalogv3.GetApplicationVersionsRequest{
		ApplicationName: "bar",
	})
	s.validateResponse(err, applications)
	s.Equal(2, len(applications.Application))
	for _, a := range applications.Application {
		s.Equal("bar", a.Name)
	}
}

func (s *NorthBoundTestSuite) TestGetApplication() {
	resp, err := s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "bar", "v0.2.0", "Application bar", "This is application bar",
		3, "p1", "bar-chart", "v0.2.0", barreg)
	s.Less(s.startTime, resp.Application.CreateTime.AsTime())
	s.Equal(catalogv3.Kind_KIND_NORMAL, resp.Application.Kind)

	// Try one that does not exist
	_, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "not-present", Version: "v0.1.0",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "application not-present:v0.1.0 not found"))

	// Try version that does not exist
	_, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.9.9",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "application bar:v0.9.9 not found"))

	// Try publisher that does not exist
	_, err = s.client.GetApplication(s.ProjectID("non-existent"), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.0",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "application bar:v0.2.0 not found"))
}

func (s *NorthBoundTestSuite) TestGetApplicationReferenceCount() {
	resp, err := s.client.GetApplicationReferenceCount(s.ProjectID(footen), &catalogv3.GetApplicationReferenceCountRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.Equal(uint32(2), resp.ReferenceCount)

	resp, err = s.client.GetApplicationReferenceCount(s.ProjectID(footen), &catalogv3.GetApplicationReferenceCountRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.Equal(uint32(3), resp.ReferenceCount)

	// Try one that does not exist - should return NotFound
	resp, err = s.client.GetApplicationReferenceCount(s.ProjectID(footen), &catalogv3.GetApplicationReferenceCountRequest{
		ApplicationName: "not-present", Version: "v0.1.0",
	})
	s.validateNotFound(err, resp)

	// Try one that does not exist
	_, err = s.client.GetApplicationReferenceCount(s.ProjectID(footen), &catalogv3.GetApplicationReferenceCountRequest{
		ApplicationName: "not-present", Version: "",
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "application invalid: incomplete request"))
}

func (s *NorthBoundTestSuite) TestUpdateApplication() {
	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.NoError(err)

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			Name:               "foo",
			Version:            "v0.1.0",
			Kind:               catalogv3.Kind_KIND_ADDON,
			DisplayName:        "new display name", // Changed
			Description:        "new description",  // Changed
			ChartName:          "chart-changed",    // Changed
			ChartVersion:       "v0.1.0",
			DefaultProfileName: "p2", // Changed
			IgnoredResources: []*catalogv3.ResourceReference{ // Changed
				{Name: "foo", Kind: "ConfigMap"},
			},
			Profiles: resp.Application.Profiles,
		},
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "foo", "v0.1.0", "new display name", "new description",
		2, "p2", "chart-changed", "v0.1.0", fooreg)
	s.Len(resp.Application.IgnoredResources, 1)
	s.Equal(catalogv3.Kind_KIND_ADDON, resp.Application.Kind)

	// Try one that does not exist
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "not-present",
		Version:         "v0.2.0",
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "not-present",
			Version:          "v0.2.0",
			ChartName:        "chart-changed",
			ChartVersion:     "v9.9.9",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "application not-present:v0.2.0 not found"))

	// Delete both profiles - but both are being used in Deployment profiles - must delete first
	s.deleteDeploymentProfiles(footen, "ca-gigi", "v0.2.1", "cp-1", "cp-2")
	s.deleteDeploymentProfiles(footen, "ca-gigi", "v0.3.4", "cp-1", "cp-2")
	s.deleteDeploymentProfiles(footen, "ca-fifi", "v0.2.0", "cp-1", "cp-2")

	// Try to update after profiles have been dropped
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "foo",
			Version:          "v0.1.0",
			DisplayName:      "new display name",
			Description:      "after profiles dropped", // Changed
			ChartName:        "chart-changed",
			ChartVersion:     "v0.1.0",
		},
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "foo", "v0.1.0", "new display name", "after profiles dropped",
		0, "", "chart-changed", "v0.1.0", fooreg)
}

func (s *NorthBoundTestSuite) TestUpdateApplicationIgnoredResourcesChanged() {
	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.NoError(err)

	// Change some aspects of the application
	app := resp.Application
	app.DisplayName = "new display name"
	app.Description = "new description"
	app.ChartName = "chart-changed"
	app.DefaultProfileName = "p2"
	app.IgnoredResources = []*catalogv3.ResourceReference{
		{Name: "foo", Kind: "ConfigMap", Namespace: "ns"},
	}

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application:     app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "foo", "v0.1.0", "new display name", "new description",
		2, "p2", "chart-changed", "v0.1.0", fooreg)
	s.Len(resp.Application.IgnoredResources, 1)
	s.Equal("ns", resp.Application.IgnoredResources[0].Namespace)

	// Update just the ignored resources
	app.IgnoredResources = []*catalogv3.ResourceReference{
		{Name: "foo", Kind: "ConfigMap"}, {Name: "bar", Kind: "SomeKind", Namespace: "ns"},
	}
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application:     app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "foo", "v0.1.0", "new display name", "new description",
		2, "p2", "chart-changed", "v0.1.0", fooreg)
	s.Len(resp.Application.IgnoredResources, 2)
	s.True(resp.Application.IgnoredResources[0].Namespace == "ns" || resp.Application.IgnoredResources[1].Namespace == "ns")
}

func (s *NorthBoundTestSuite) TestUpdateApplicationIgnoredResourcesDidNotChange() {
	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.NoError(err)

	// Change some aspects of the application
	app := resp.Application
	app.DisplayName = "new display name"
	app.Description = "new description"
	app.ChartName = "chart-changed"
	app.DefaultProfileName = "p2"
	app.IgnoredResources = []*catalogv3.ResourceReference{
		{Name: "foo", Kind: "ConfigMap"},
	}

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application:     app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "foo", "v0.1.0", "new display name", "new description",
		2, "p2", "chart-changed", "v0.1.0", fooreg)
	s.Len(resp.Application.IgnoredResources, 1)

	// Do the same update again, should be a no-op
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application:     app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "foo", "v0.1.0", "new display name", "new description",
		2, "p2", "chart-changed", "v0.1.0", fooreg)
	s.Len(resp.Application.IgnoredResources, 1)
}

func (s *NorthBoundTestSuite) TestUpdateApplicationInDeployedState() {
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

	// Attempt to update an app that is part of the deployed deployment package; it should fail
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			Name:               "foo",
			Version:            "v0.1.0",
			DisplayName:        "new display name",
			Description:        "new description",
			ChartName:          "chart-changed",
			ChartVersion:       "v0.1.0",
			DefaultProfileName: "p2",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition,
		"application foo:v0.1.0 failed precondition: cannot update application that is part of 1 packages; please create a new version instead"))
}

func (s *NorthBoundTestSuite) TestUpdateApplicationKindInDeployedState() {
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

	// Update kind of an app that is part of the deployed deployment package; it should succeed
	aresp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0",
	})
	s.validateResponse(err, aresp)
	s.Equal(catalogv3.Kind_KIND_NORMAL, aresp.Application.Kind)

	aresp.Application.Kind = catalogv3.Kind_KIND_EXTENSION
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo", Version: "v0.1.0", Application: aresp.Application,
	})
	s.validateResponse(err, none)
	s.Equal(catalogv3.Kind_KIND_EXTENSION, aresp.Application.Kind)
}

func (s *NorthBoundTestSuite) TestUpdateApplicationDeletingUsedProfile() {
	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)

	// Remove a profile that has been named in a deployment profile
	app := resp.Application
	app.Profiles = removeProfile(app.Profiles, "p1")

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
		Application:     app,
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition,
		"failed precondition: profile p1 cannot be deleted; it is in use by another deployment profile"))
}

func (s *NorthBoundTestSuite) TestUpdateApplicationDeletingUnUsedProfile() {
	resp, err := s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)

	// Remove a profile that has been named in a deployment profile
	app := resp.Application
	app.Profiles = removeProfile(app.Profiles, "p4")

	_, err = s.client.UpdateApplication(s.ProjectID(barten), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
		Application:     app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)
	s.validateApp(resp.Application, "bar", "v0.2.1", app.DisplayName, app.Description, 3,
		"p1", app.ChartName, app.ChartVersion, app.HelmRegistryName)
}

func (s *NorthBoundTestSuite) TestUpdateApplicationWithDeploymentRequirements() {
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "test-application",
			DisplayName:        "Test application",
			Description:        "This is a Test",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "profile-1",
			Profiles: []*catalogv3.Profile{
				{
					Name:        "profile-1",
					DisplayName: "Profile 1",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-fifi", Version: "v0.2.0", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
				{
					Name:        "profile-2",
					DisplayName: "Profile 2",
					ChartValues: "key1b: value1b\nkey2b: value2b\n",
				},
			},
		},
	})
	s.validateResponse(err, created)
	s.validateApp(created.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Equal(fooregalt, created.Application.ImageRegistryName)
	s.Equal(1, len(created.Application.Profiles[0].DeploymentRequirement)+len(created.Application.Profiles[1].DeploymentRequirement))

	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Equal(fooregalt, resp.Application.ImageRegistryName)
	s.Equal(1, len(resp.Application.Profiles[0].DeploymentRequirement)+len(resp.Application.Profiles[1].DeploymentRequirement))

	// Tweak the profile slightly.
	updatedChartValues := fmt.Sprintf("%s\nnewField: true\n", resp.Application.Profiles[0].ChartValues)
	resp.Application.Profiles[0].ChartValues = updatedChartValues

	// Now update the app...
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "test-application", Version: "0.1.0", Application: resp.Application,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		2, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Equal(fooregalt, resp.Application.ImageRegistryName)
	s.Equal(updatedChartValues, resp.Application.Profiles[0].ChartValues)
	s.Equal(1, len(resp.Application.Profiles[0].DeploymentRequirement)+len(resp.Application.Profiles[1].DeploymentRequirement))
}

func (s *NorthBoundTestSuite) TestUpdateApplicationWithDeploymentRequirementVersionChange() {
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Name:               "test-application",
			DisplayName:        "Test application",
			Description:        "This is a Test",
			Version:            "0.1.0",
			ChartName:          "test/chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "profile-1",
			Profiles: []*catalogv3.Profile{
				{
					Name:        "profile-1",
					DisplayName: "Profile 1",
					DeploymentRequirement: []*catalogv3.DeploymentRequirement{
						{Name: "ca-gigi", Version: "v0.2.1", DeploymentProfileName: "cp-2"},
					},
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.validateResponse(err, created)

	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		1, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Equal(fooregalt, resp.Application.ImageRegistryName)
	s.Equal("v0.2.1", resp.Application.Profiles[0].DeploymentRequirement[0].Version)

	// Now tweak the version of the deployment requirement and update the app...
	resp.Application.Profiles[0].DeploymentRequirement[0].Version = "v0.3.4"
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "test-application", Version: "0.1.0", Application: resp.Application,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.validateApp(resp.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		1, "profile-1", "test/chart", "0.1.0", fooreg)
	s.Equal(fooregalt, resp.Application.ImageRegistryName)
	s.Equal("v0.3.4", resp.Application.Profiles[0].DeploymentRequirement[0].Version)
}

func (s *NorthBoundTestSuite) TestApplicationCreateNewProfile() {
	resp, err := s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)

	// Remove a profile that has been named in a deployment profile
	app := resp.Application
	app.Profiles = append(app.Profiles, &catalogv3.Profile{
		Name:        "newp",
		DisplayName: "New Profile",
		Description: "This is a new profile",
		ChartValues: `foo: "bar"\nbar: "foo""\n`,
	})

	_, err = s.client.UpdateApplication(s.ProjectID(barten), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
		Application:     app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)
	s.validateApp(resp.Application, "bar", "v0.2.1", app.DisplayName, app.Description, 5,
		"p1", app.ChartName, app.ChartVersion, app.HelmRegistryName)
}

func (s *NorthBoundTestSuite) TestApplicationCreateNewProfileWhenDeployed() {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.NoError(err)

	resp.DeploymentPackage.IsDeployed = true

	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1", DeploymentPackage: resp.DeploymentPackage,
	})
	s.NoError(err)

	s.TestApplicationCreateNewProfile()
}

func (s *NorthBoundTestSuite) TestApplicationCreateNewProfileDuplicateName() {
	resp, err := s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)

	// Create a duplicate profile
	app := resp.Application
	app.Profiles = append(app.Profiles, &catalogv3.Profile{
		Name:        "p1",
		DisplayName: "Duplicate Profile",
		Description: "This is a duplicate profile",
		ChartValues: `foo: "bar"\nbar: "foo""\n`,
	})

	_, err = s.client.UpdateApplication(s.ProjectID(barten), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
		Application:     app,
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "profile p1 already exists"))
}

func (s *NorthBoundTestSuite) TestApplicationCreateNewProfileDuplicateDisplayName() {
	resp, err := s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)

	// Add a new profile
	app := resp.Application
	app.Profiles = append(app.Profiles, &catalogv3.Profile{
		Name:        "newp",
		DisplayName: "Profile 1 for bar",
		ChartValues: `foo: "bar"\nbar: "foo""\n`,
	})

	_, err = s.client.UpdateApplication(s.ProjectID(barten), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
		Application:     app,
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "profile newp already exists: profile newp display name Profile 1 for bar is not unique"))
}

func (s *NorthBoundTestSuite) TestApplicationUpdateProfileDuplicateDisplayName() {
	resp, err := s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.1",
	})
	s.NoError(err)

	// Update p2 profile to have a duplicate display name with profile p1
	app := resp.Application
	for _, p := range app.Profiles {
		if p.Name == "p2" {
			p.DisplayName = "Profile 1 for bar"
		}
	}

	_, err = s.client.UpdateApplication(s.ProjectID(barten), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "bar",
		Version:         "v0.2.1",
		Application:     app,
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "profile p2 already exists: profile p2 display name Profile 1 for bar is not unique"))
}

func (s *NorthBoundTestSuite) TestApplicationUpdateProfileFromNoProfile() {
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "app",
			Version:          "v1",
			ChartName:        "chart",
			ChartVersion:     "0.1.0",
		},
	})
	s.validateResponse(err, created)
	s.Len(created.Application.Profiles, 0)
	s.Equal("", created.Application.DefaultProfileName)

	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "app", Version: "v1",
	})
	s.validateResponse(err, resp)
	s.Len(resp.Application.Profiles, 0)
	s.Equal("", resp.Application.DefaultProfileName)

	// Add new profiles and set the default profile
	app := created.Application
	app.Profiles = []*catalogv3.Profile{{Name: "p1"}}
	app.DefaultProfileName = "p1"
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "app", Version: "v1", Application: app,
	})
	s.NoError(err)

	resp, err = s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "app", Version: "v1",
	})
	s.validateResponse(err, resp)
	s.Len(resp.Application.Profiles, 1)
	s.Equal("p1", resp.Application.DefaultProfileName)
}

func removeProfile(profiles []*catalogv3.Profile, name string) []*catalogv3.Profile {
	for i, p := range profiles {
		if p.Name == name {
			return append(profiles[:i], profiles[i+1:]...)
		}
	}
	return profiles
}

func (s *NorthBoundTestSuite) TestUpdateApplicationWithIllegalInputs() {
	// Try one that does not exist
	_, err := s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName: "not-present",
			Name:             "foo",
			Version:          "v0.1.0",
			ChartName:        "chart-changed",
			ChartVersion:     "v9.9.9",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "application foo:v0.1.0 invalid: helm registry not-present not found"))

	// Try with mismatched name
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "bar",
			Version:          "v0.1.0",
			ChartName:        "chart-changed",
			ChartVersion:     "v0.2.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "application invalid: name cannot be changed foo != bar"))

	// Try with mismatched version
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "foo",
		Version:         "v0.1.0",
		Application: &catalogv3.Application{
			HelmRegistryName: fooreg,
			Name:             "foo",
			Version:          "v0.2.0",
			ChartName:        "chart-changed",
			ChartVersion:     "v0.2.0",
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "application invalid: version cannot be changed v0.1.0 != v0.2.0"))
}

func (s *NorthBoundTestSuite) TestDeleteApplication() {
	// Try deleting application bar:v0.2.0
	// Because it is being referred to by ca-fifi, this will result in an error
	_, err := s.client.DeleteApplication(s.ProjectID(barten), &catalogv3.DeleteApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.0",
	})
	s.ErrorIs(err, status.Errorf(codes.FailedPrecondition,
		"application bar:v0.2.0 failed precondition: cannot delete application that is part of one or more deployment-package"))

	// Use UpdateDeploymentPackage to remove the reference on bar:v0.2.0 from ca-fifi
	dpresp, err := s.client.GetDeploymentPackage(s.ProjectID(barten), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-fifi", Version: "v0.2.0",
	})
	s.NoError(err)

	apprefs := dpresp.DeploymentPackage.ApplicationReferences
	for i, ref := range apprefs {
		if ref.Name == "bar" {
			dpresp.DeploymentPackage.ApplicationReferences = append(apprefs[:i], apprefs[i+1:]...)
			break
		}
	}
	profiles := dpresp.DeploymentPackage.Profiles
	for _, profile := range profiles {
		delete(profile.ApplicationProfiles, "bar")
	}
	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(barten), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-fifi", Version: "v0.2.0", DeploymentPackage: dpresp.DeploymentPackage,
	})
	s.NoError(err)

	deleted, err := s.client.DeleteApplication(s.ProjectID(barten), &catalogv3.DeleteApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.0",
	})
	s.validateResponse(err, deleted)

	// Make sure the app is no longer there
	_, err = s.client.GetApplication(s.ProjectID(barten), &catalogv3.GetApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.0",
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "application bar:v0.2.0 not found"))

	// Make sure that the ca-fifi does not have an app reference for this
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(barten), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-fifi", Version: "v0.2.0",
	})
	s.validateResponse(err, resp)
	s.Len(resp.DeploymentPackage.ApplicationReferences, 2)

	// Make sure that the ca-gigi is unaffected as it has a different version of the app
	resp, err = s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)
	s.Len(resp.DeploymentPackage.ApplicationReferences, 2)

	// Try deleting it again, i.e. non-existent - should return a NotFound error
	deleted, err = s.client.DeleteApplication(s.ProjectID(barten), &catalogv3.DeleteApplicationRequest{
		ApplicationName: "bar", Version: "v0.2.0",
	})
	s.validateNotFound(err, deleted)
}

func (s *NorthBoundTestSuite) TestAddProfileWithNoDefaultProfile() {
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:  fooreg,
			ImageRegistryName: fooregalt,
			Name:              "test-application",
			DisplayName:       "Test application",
			Description:       "This is a Test",
			Version:           "0.1.0",
			ChartName:         "test-chart",
			ChartVersion:      "0.1.0",
		},
	})
	s.validateResponse(err, created)
	s.validateApp(created.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		0, "", "test-chart", "0.1.0", fooreg)

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
		Application: &catalogv3.Application{
			Name:               "test-application",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "default",
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Profiles: []*catalogv3.Profile{
				{
					Name:        "default",
					DisplayName: "Default Profile",
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.NoError(err)
}

func (s *NorthBoundTestSuite) TestApplicationAuthErrors() {
	var err error
	server := s.newMockOPAServer()

	app := &catalogv3.Application{
		Name:         "app",
		Version:      "6.7",
		ChartName:    "c",
		ChartVersion: "2.4",
	}

	_, err = server.CreateApplication(s.ServerProjectID(footen), &catalogv3.CreateApplicationRequest{Application: app})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.UpdateApplication(s.ServerProjectID(footen), &catalogv3.UpdateApplicationRequest{ApplicationName: "app", Version: app.Version, Application: app})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.DeleteApplication(s.ServerProjectID(footen), &catalogv3.DeleteApplicationRequest{ApplicationName: "app", Version: "1.0"})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.GetApplication(s.ServerProjectID(footen), &catalogv3.GetApplicationRequest{ApplicationName: "app", Version: "1.0"})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.GetApplicationReferenceCount(s.ServerProjectID(footen), &catalogv3.GetApplicationReferenceCountRequest{ApplicationName: "app", Version: "1.0"})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.GetApplicationVersions(s.ServerProjectID(footen), &catalogv3.GetApplicationVersionsRequest{ApplicationName: "app"})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.ListApplications(s.ServerProjectID(footen), &catalogv3.ListApplicationsRequest{})
	s.ErrorIs(err, expectedAuthError)
}

func (s *NorthBoundDBErrTestSuite) TestApplicationListInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.ListApplicationsRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.ListApplications(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestApplicationGetInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.GetApplicationRequest
	}{
		"nil request":       {req: nil},
		"empty app name":    {req: &catalogv3.GetApplicationRequest{ApplicationName: ""}},
		"empty app version": {req: &catalogv3.GetApplicationRequest{ApplicationName: "app", Version: ""}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.GetApplication(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestApplicationUpdateInvalidArgs() {
	app := &catalogv3.Application{
		Name:         "app!",
		Version:      "6.7",
		ChartName:    "c",
		ChartVersion: "2.4",
	}
	tests := map[string]struct {
		req *catalogv3.UpdateApplicationRequest
	}{
		"nil request":        {req: nil},
		"nil application":    {req: &catalogv3.UpdateApplicationRequest{Application: nil}},
		"empty app name":     {req: &catalogv3.UpdateApplicationRequest{Application: app, ApplicationName: ""}},
		"empty app version":  {req: &catalogv3.UpdateApplicationRequest{Application: app, ApplicationName: "app", Version: ""}},
		"app validate error": {req: &catalogv3.UpdateApplicationRequest{Application: app, ApplicationName: "app", Version: "1.0"}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.UpdateApplication(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundTestSuite) TestApplicationEvents() {
	s.T().Skip()
	ctx, cancel := context.WithCancel(s.ProjectID(footen))
	stream, err := s.client.WatchApplications(ctx, &catalogv3.WatchApplicationsRequest{NoReplay: true})
	s.NoError(err)

	app := s.createApp(footen, fooreg, "newapp", "0.1.1", 2)

	resp, err := stream.Recv()
	s.NoError(err)
	s.Equal(CreatedEvent, EventType(resp.Event.Type))
	s.validateApp(resp.Application, app.Name, app.Version, app.DisplayName, app.Description, len(app.Profiles),
		app.DefaultProfileName, app.ChartName, app.ChartVersion, app.HelmRegistryName)

	app.DisplayName = "New App"
	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: app.Name, Version: app.Version, Application: app,
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(UpdatedEvent, EventType(resp.Event.Type))
	s.validateApp(resp.Application, app.Name, app.Version, app.DisplayName, app.Description, len(app.Profiles),
		app.DefaultProfileName, app.ChartName, app.ChartVersion, app.HelmRegistryName)

	_, err = s.client.DeleteApplication(s.ProjectID(footen), &catalogv3.DeleteApplicationRequest{
		ApplicationName: app.Name, Version: app.Version,
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(DeletedEvent, EventType(resp.Event.Type))
	s.validateApp(resp.Application, app.Name, app.Version, "", "", 0, "", "", "", "")

	// Make sure we get an error back for a Recv() on a closed channel
	cancel()
	s.createApp(footen, fooreg, "aaa", "1.0", 1)
	resp, err = stream.Recv()
	s.Error(err)
	s.Nil(resp)
}

func (s *NorthBoundTestSuite) TestApplicationsEventsWithReplay() {
	s.T().Skip()
	ctx, cancel := context.WithCancel(s.ProjectID(footen))
	stream, err := s.client.WatchApplications(ctx, &catalogv3.WatchApplicationsRequest{})
	s.NoError(err)

	existing := "foo bar goo"
	for i := 0; i < 4; i++ {
		resp, err := stream.Recv()
		s.NoError(err)
		s.Equal(ReplayedEvent, EventType(resp.Event.Type))
		s.True(strings.Contains(existing, resp.Application.Name), "unexpected: %s", resp.Application.Name)
	}

	app := s.createApp(footen, fooreg, "newapp", "0.1.1", 2)

	resp, err := stream.Recv()
	s.NoError(err)
	s.Equal(CreatedEvent, EventType(resp.Event.Type))
	s.validateApp(resp.Application, app.Name, app.Version, app.DisplayName, app.Description, len(app.Profiles),
		app.DefaultProfileName, app.ChartName, app.ChartVersion, app.HelmRegistryName)

	// Make sure we get an error back for a Recv() on a closed channel
	cancel()
	s.createApp(footen, fooreg, "aaa", "1.0", 1)
	resp, err = stream.Recv()
	s.Error(err)
	s.Nil(resp)
}

func (s *NorthBoundDBErrTestSuite) TestApplicationWatchInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.WatchApplicationsRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			err := s.server.WatchApplications(testCase.req, nil)
			s.validateInvalidArgumentError(err, nil)
		})
	}
}

type catalogServiceWatchApplicationsServer struct {
	testServerStream
	sendError bool
}

func (c *catalogServiceWatchApplicationsServer) Send(*catalogv3.WatchApplicationsResponse) error {
	if c.sendError {
		return errors.New("no send")
	}
	return nil
}

func (s *NorthBoundDBErrTestSuite) TestWatchApplicationDatabaseErrors() {
	s.T().Skip("FIXME: Hard to troubleshoot")
	saveBase64Factory := Base64Factory
	defer func() { Base64Factory = saveBase64Factory }()
	Base64Factory = newBase64Noop

	server := &catalogServiceWatchApplicationsServer{sendError: false}
	req := &catalogv3.WatchApplicationsRequest{}

	// Test unable to start a transaction
	err := s.server.WatchApplications(req, server)
	s.validateDBError(err, nil)

	// Test publisher existence query failure
	s.mock.ExpectBegin()
	err = s.server.WatchApplications(req, server)
	s.validateDBError(err, nil)

	// parameter template error
	s.mock.ExpectBegin()
	s.addMockedEmptyQueryRows(5)
	err = s.server.WatchApplications(req, server)
	s.validateDBError(err, nil)

	// transaction commit error
	s.mock.ExpectBegin()
	s.addMockedEmptyQueryRows(7)
	err = s.server.WatchApplications(req, server)
	s.validateDBError(err, nil)

	// send failure in replay
	server.sendError = true
	s.mock.ExpectBegin()
	s.addMockedEmptyQueryRows(12)
	err = s.server.WatchApplications(req, server)
	s.Contains(err.Error(), "no send")

	// send failure for event
	ch := make(chan *catalogv3.WatchApplicationsResponse)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.server.watchApplicationEvents(server, ch)
		wg.Done()
	}()
	ch <- nil
	wg.Wait()
	s.Contains(err.Error(), "no send")
	close(ch)

	// end of channel
	ch = make(chan *catalogv3.WatchApplicationsResponse)
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.server.watchApplicationEvents(server, ch)
		wg.Done()
	}()
	close(ch)
	wg.Wait()
	s.NoError(err)

	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *NorthBoundTestSuite) checkParameterTemplates(expectedTemplates []*catalogv3.ParameterTemplate) {
	queried, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, queried)
	s.Len(queried.Application.Profiles, 1)

	existingParameterTemplates := make(map[string]*catalogv3.ParameterTemplate)
	for _, existingTemplate := range queried.Application.Profiles[0].ParameterTemplates {
		existingParameterTemplates[existingTemplate.Name] = existingTemplate
	}

	s.Len(queried.Application.Profiles[0].ParameterTemplates, len(expectedTemplates))

	for _, expectedTemplate := range expectedTemplates {
		pt, ok := existingParameterTemplates[expectedTemplate.Name]
		s.True(ok)
		s.Equal(expectedTemplate.Default, pt.Default)
		s.Equal(expectedTemplate.DisplayName, pt.DisplayName)
		s.Equal(expectedTemplate.Type, pt.Type)
		s.Len(pt.SuggestedValues, len(expectedTemplate.SuggestedValues))
		for i, expectedSuggestedValue := range expectedTemplate.SuggestedValues {
			s.Equal(expectedSuggestedValue, pt.SuggestedValues[i])
		}
	}
}

func (s *NorthBoundTestSuite) TestApplicationParameterTemplate() {
	parameterTemplate1 := &catalogv3.ParameterTemplate{
		Name:            "pt1",
		DisplayName:     "P T 1",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
	}
	parameterTemplate1mod := &catalogv3.ParameterTemplate{
		Name:            "pt1",
		DisplayName:     "P T 1mod",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
	}
	parameterTemplate1dup := &catalogv3.ParameterTemplate{
		Name:            "pt1",
		DisplayName:     "P T 1 dup",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
	}
	parameterTemplate1modSV := &catalogv3.ParameterTemplate{
		Name:            "pt1",
		DisplayName:     "P T 1mod",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2.1"},
	}
	parameterTemplate2 := &catalogv3.ParameterTemplate{
		Name:            "pt2",
		DisplayName:     "P T 2",
		Default:         "1",
		Type:            "number",
		SuggestedValues: []string{"1", "2", "3", "4", "5"},
	}
	parameterTemplate3 := &catalogv3.ParameterTemplate{
		Name:            "pt3",
		DisplayName:     "P T 2",
		Default:         "1",
		Type:            "number",
		SuggestedValues: []string{"1", "2", "3", "4", "5"},
	}
	parameterTemplate4 := &catalogv3.ParameterTemplate{
		Name:            `prometheus\.io/scrape`,
		DisplayName:     "P T 2",
		Default:         "1",
		Type:            "number",
		SuggestedValues: []string{"1", "2", "3", "4", "5"},
	}
	veryLongName := strings.Repeat("nam.", 1000)
	parameterTemplate5 := &catalogv3.ParameterTemplate{
		Name:            veryLongName,
		DisplayName:     "P T 2",
		Default:         "1",
		Type:            "number",
		SuggestedValues: []string{"1", "2", "3", "4", "5"},
	}
	veryLongData := strings.Repeat("Data", 1000)
	parameterTemplate6 := &catalogv3.ParameterTemplate{
		Name:            veryLongName,
		DisplayName:     "P T 2",
		Default:         veryLongData,
		Type:            "number",
		SuggestedValues: []string{"1", "2", "3", "4", "5"},
	}
	parameterTemplate7 := &catalogv3.ParameterTemplate{
		Name:            "badDisplayName",
		DisplayName:     " P T 2",
		Default:         "default",
		Type:            "number",
		SuggestedValues: []string{"1", "2", "3", "4", "5"},
	}
	parameterTemplate8 := &catalogv3.ParameterTemplate{
		Name:            "name-withSome.really\\very_special/characters",
		DisplayName:     "P T 1",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
	}
	parameterTemplate9a := &catalogv3.ParameterTemplate{
		Name:            "mandatory-thing-with-default",
		DisplayName:     "Mandatory With Default",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
		Mandatory:       true,
	}
	parameterTemplate9b := &catalogv3.ParameterTemplate{
		Name:            "secret-thing-with-default",
		DisplayName:     "Secret With Default",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
		Secret:          true,
	}

	// Create the initial application
	parameterTemplates := []*catalogv3.ParameterTemplate{parameterTemplate1}
	baseApplication := catalogv3.Application{
		HelmRegistryName:   fooreg,
		ImageRegistryName:  fooregalt,
		Name:               "test-application",
		DisplayName:        "Test application",
		Description:        "This is a Test",
		Version:            "0.1.0",
		ChartName:          "test-chart",
		ChartVersion:       "0.1.0",
		DefaultProfileName: "profile-1",
		Profiles: []*catalogv3.Profile{
			{
				Name:               "profile-1",
				DisplayName:        "Profile 1",
				ChartValues:        "key1a: value1a\nkey2a: value2a\n",
				ParameterTemplates: parameterTemplates,
			},
		},
	}
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &baseApplication,
	})
	s.validateResponse(err, created)
	s.checkParameterTemplates(parameterTemplates)

	testCases := []struct {
		name               string
		parameterTemplates []*catalogv3.ParameterTemplate
		expectedError      string
	}{
		{name: "Update existing template", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1mod}},
		{name: "Update existing suggested value", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1modSV}},
		{name: "Adding template", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1, parameterTemplate2}},
		{name: "Removing template", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate2}},
		{name: "Removing and adding template", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate3}},
		{name: "Adding template with dotted path", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1, parameterTemplate4}},
		{name: "Adding template with very long name", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1, parameterTemplate5}},
		{name: "Adding template with very long data", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1, parameterTemplate6}},
		{name: "Adding duplicates", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate1, parameterTemplate1dup}, expectedError: "duplicate parameter template"},
		{name: "Bad display name", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate7}, expectedError: "display name cannot contain leading or trailing spaces"},
		{name: "Adding template with special characters", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate8}},
		{name: "Mandatory with default", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate9a}, expectedError: "mandatory or secret parameter template mandatory-thing-with-default should have no default"},
		{name: "Secret with default", parameterTemplates: []*catalogv3.ParameterTemplate{parameterTemplate9b}, expectedError: "mandatory or secret parameter template secret-thing-with-default should have no default"},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(_ *testing.T) {
			baseApplication.Profiles[0].ParameterTemplates = testCase.parameterTemplates

			_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
				ApplicationName: baseApplication.Name, Version: baseApplication.Version,
				Application: &baseApplication,
			})
			if testCase.expectedError == "" {
				s.NoError(err)
				s.checkParameterTemplates(testCase.parameterTemplates)
			} else {
				s.Error(err)
				s.Contains(err.Error(), testCase.expectedError)
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestDupApplicationParameterTemplateCreate() {
	parameterTemplate1 := &catalogv3.ParameterTemplate{
		Name:            "pt1",
		DisplayName:     "P T 1",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
	}
	parameterTemplate1dup := &catalogv3.ParameterTemplate{
		Name:            "pt1",
		DisplayName:     "P T 1 dup",
		Default:         "default",
		Type:            "string",
		SuggestedValues: []string{"v1", "v2"},
	}

	// Try to create the application
	parameterTemplates := []*catalogv3.ParameterTemplate{parameterTemplate1, parameterTemplate1dup}
	baseApplication := catalogv3.Application{
		HelmRegistryName:   fooreg,
		ImageRegistryName:  fooregalt,
		Name:               "test-application",
		DisplayName:        "Test application",
		Description:        "This is a Test",
		Version:            "0.1.0",
		ChartName:          "test-chart",
		ChartVersion:       "0.1.0",
		DefaultProfileName: "profile-1",
		Profiles: []*catalogv3.Profile{
			{
				Name:               "profile-1",
				DisplayName:        "Profile 1",
				ChartValues:        "key1a: value1a\nkey2a: value2a\n",
				ParameterTemplates: parameterTemplates,
			},
		},
	}
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &baseApplication,
	})
	s.Error(err)
	s.Nil(created)
	s.Contains(err.Error(), "application profile-1 invalid: duplicate parameter template pt1")
}

func TestEmptyAppsDBQuery(t *testing.T) {
	s := &NorthBoundTestSuite{populateDB: false}
	s.SetT(t)
	s.SetupTest()

	apps, err := s.client.ListApplications(s.ProjectID(footen), &catalogv3.ListApplicationsRequest{})
	s.validateResponse(err, apps)
	s.Equal(0, len(apps.Applications))

	publisher, err := s.client.GetApplication(s.ProjectID("nobody"), &catalogv3.GetApplicationRequest{
		Version:         "1.0",
		ApplicationName: "none",
	})
	s.Nil(publisher)
	s.Error(err)
	s.Contains(err.Error(), "not found")
}

func (s *NorthBoundTestSuite) TestAddProfile() {
	var err error
	created, err := s.client.CreateApplication(s.ProjectID(footen), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			HelmRegistryName:  fooreg,
			ImageRegistryName: fooregalt,
			Name:              "test-application",
			DisplayName:       "Test application",
			Description:       "This is a Test",
			Version:           "0.1.0",
			ChartName:         "test-chart",
			ChartVersion:      "0.1.0",
		},
	})
	s.validateResponse(err, created)
	s.validateApp(created.Application, "test-application", "0.1.0", "Test application", "This is a Test",
		0, "", "test-chart", "0.1.0", fooreg)

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
		Application: &catalogv3.Application{
			Name:               "test-application",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "default",
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Profiles: []*catalogv3.Profile{
				{
					Name:        "default",
					DisplayName: "Default Profile",
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
			},
		},
	})
	s.NoError(err)

	_, err = s.client.UpdateApplication(s.ProjectID(footen), &catalogv3.UpdateApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
		Application: &catalogv3.Application{
			Name:               "test-application",
			Version:            "0.1.0",
			ChartName:          "test-chart",
			ChartVersion:       "0.1.0",
			DefaultProfileName: "default",
			HelmRegistryName:   fooreg,
			ImageRegistryName:  fooregalt,
			Profiles: []*catalogv3.Profile{
				{
					Name:        "default",
					DisplayName: "Default Profile",
					ChartValues: "key1a: value1a\nkey2a: value2a\n",
				},
				{
					Name:        "newone",
					DisplayName: "New One",
					ChartValues: "key2b: value2b\nkey2b: value2b\n",
				},
			},
		},
	})
	s.NoError(err)

	resp, err := s.client.GetApplication(s.ProjectID(footen), &catalogv3.GetApplicationRequest{
		ApplicationName: "test-application",
		Version:         "0.1.0",
	})
	s.validateResponse(err, resp)
	s.Len(resp.Application.Profiles, 2)
}
