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
	//var invalidKind restapi.ApplicationKind = "INVALID_KIND" // Only if this conversion is allowed

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
				Name:             "test-app-empty-version",
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
				Name:             "test-app-empty-chart-name",
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
				Name:             "test-app-empty-chart-version",
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
				Name:             "test-app-empty-registry",
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
		// Name validation test cases
		{
			name: "Name too long (exceeds 26 chars)",
			application: &restapi.Application{
				Name:             "this-name-is-way-too-long-for-app",
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
			name: "Name starting with hyphen",
			application: &restapi.Application{
				Name:             "-invalid-name",
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
			name: "Name with invalid characters",
			application: &restapi.Application{
				Name:             "invalid_name$",
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
		// Version validation test cases
		{
			name: "Version too long (exceeds 20 chars)",
			application: &restapi.Application{
				Name:             "test-app-long-version",
				Version:          "1.0.0-beta.1234567890.1",
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
			name: "Version with invalid characters",
			application: &restapi.Application{
				Name:             "test-app-invalid-version",
				Version:          "1.0.0_beta1",
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
			name: "Version starting with hyphen",
			application: &restapi.Application{
				Name:             "test-app-invalid-version",
				Version:          "-1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		// DisplayName validation test cases
		{
			name: "DisplayName too long (exceeds 40 chars)",
			application: &restapi.Application{
				Name:             "test-app-long-display-name",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("This display name is too long for the application requirements"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		// Description validation test cases
		{
			name: "Description too long (exceeds 1000 chars)",
			application: &restapi.Application{
				Name:             "test-app-long-description",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString(string(make([]byte, 1001))), // 1001 character string
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		// ChartName validation test cases
		{
			name: "ChartName too long (exceeds 200 chars)",
			application: &restapi.Application{
				Name:             "test-app-long-chart-name",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        string(make([]byte, 201)), // 201 character string
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		{
			name: "ChartName with invalid characters",
			application: &restapi.Application{
				Name:             "test-app-invalid-chart-name",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "Invalid_Chart_Name",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		// ChartVersion validation test cases
		{
			name: "ChartVersion too long (exceeds 53 chars)",
			application: &restapi.Application{
				Name:             "test-app-long-chart-version",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0-beta.1234567890.1234567890.1234567890.1234567890.123",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		{
			name: "ChartVersion with invalid characters",
			application: &restapi.Application{
				Name:             "test-app-invalid",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0_beta1",
				HelmRegistryName: "harbor-helm-oci",
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},
		// Kind validation test case
		// TODO: Looks like even if we provide an invalid kind, the API does not return an error.
		/*{
			name: "Invalid kind value",
			application: &restapi.Application{
				Name:             "test-app-invalid-kind",
				Version:          "1.0.0",
				DisplayName:      types.GetPointerString("Test App"),
				Description:      types.GetPointerString("Test Description"),
				ChartName:        "test-chart",
				ChartVersion:     "1.0.0",
				HelmRegistryName: "harbor-helm-oci",
				Kind:             &invalidKind,
			},
			expectedCode:  http.StatusBadRequest,
			errorExpected: true,
		},*/
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Try to create application with invalid parameters
			createdApp, status, err := s.catalogClient.CreateApplication(ctx, tc.application)

			if tc.errorExpected {
				s.T().Log("status code:", status, "error:", err)
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
