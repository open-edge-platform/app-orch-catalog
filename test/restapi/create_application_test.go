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

}
