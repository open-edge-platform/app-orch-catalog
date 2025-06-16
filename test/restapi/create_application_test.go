// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"context"
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"net/http"
)

func (s *TestSuite) TestCreateApplicationValidParams() {
	ctx := context.TODO()

	// Create a new application
	app := &restapi.Application{
		Name:             "httpbin",
		Version:          "0.1.8",
		DisplayName:      types.GetPointerString("HttpBin Go"),
		Description:      types.GetPointerString("Helm chart to install httpbingo.org on Kubernetes."),
		ChartName:        "httpbin",
		ChartVersion:     "0.1.8",
		HelmRegistryName: "harbor-helm-oci",
	}

	err := s.catalogClient.DeleteApplication(ctx, app.Name, app.Version, false)
	s.Require().NoError(err, "Expected to delete application successfully")

	createdApp, status, err := s.catalogClient.CreateApplication(ctx, app)
	s.Require().NoError(err, "Expected to create application successfully")
	s.Require().Equal(http.StatusOK, status, "Expected status code 200 for application creation")
	s.Require().NotNil(createdApp, "Expected created application to be non-nil")
	s.Require().Equal(app.Name, createdApp.Name, "Expected application name to match")
	s.Require().Equal(app.Version, createdApp.Version, "Expected application version to match")
	s.Require().Equal(*app.DisplayName, *createdApp.DisplayName, "Expected application display name to match")
	s.Require().Equal(*app.Description, *createdApp.Description, "Expected application description to match")
	s.Require().Equal(app.ChartName, createdApp.ChartName, "Expected application chart name to match")
	s.Require().Equal(app.ChartVersion, createdApp.ChartVersion, "Expected application chart version to match")
	s.Require().Equal(app.HelmRegistryName, createdApp.HelmRegistryName, "Expected application helm registry name to match")

	err = s.catalogClient.DeleteApplication(ctx, app.Name, app.Version, true)
	s.Require().NoError(err, "Expected to delete application successfully")

}

func (s *TestSuite) TestCreateApplicationInvalidParams() {
	ctx := context.TODO()

	testCases := []struct {
		name          string
		application   *restapi.Application
		expectedCode  int
		errorExpected bool
	}{
		{
			name: "Empty application name",
			application: &restapi.Application{
				Name:             "",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		{
			name: "Empty version",
			application: &restapi.Application{
				Name:             "test-app",
				Version:          "",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		{
			name: "Empty chart name",
			application: &restapi.Application{
				Name:             "test-app",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		{
			name: "Empty chart version",
			application: &restapi.Application{
				Name:             "test-app",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		{
			name: "Empty helm registry name",
			application: &restapi.Application{
				Name:             "test-app",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Try to create application with invalid parameters
			createdApp, status, err := s.catalogClient.CreateApplication(ctx, tc.application)

			if tc.errorExpected {
				s.T().Log(tc.expectedCode, ":", status, ":", err.Error())
				s.Require().Error(err, "Expected error for invalid application parameters")
				s.Require().Equal(tc.expectedCode, status, "Expected correct status code")
				s.Require().Nil(createdApp, "Expected no application to be created")
			} else {
				s.Require().NoError(err, "Did not expect error for test case")
				s.Require().Equal(tc.expectedCode, status, "Expected correct status code")
				s.Require().Nil(createdApp, "Expected no application to be created")
			}
		})
	}
}
