// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
)

func (s *TestSuite) TestListBootStrapExtensions() {
	res, err := s.catalogClient.MakeHTTPRequest("GET", types.ApplicationsEndpoint, nil)
	s.Require().NoError(err)
	defer res.Body.Close()
	s.assertStatus(res, http.StatusOK)

	body := s.readResponseBody(res)
	var result struct {
		Applications []types.Application `json:"applications"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), len(types.GetApplications()), len(result.Applications), "Mismatch in the number of applications")
	s.T().Log("Extensions:")
	for _, app := range result.Applications {
		s.T().Logf("Name: %s, DisplayName: %s, Description: %s, Version: %s, Kind: %s, ChartName: %s, ChartVersion: %s, HelmRegistryName: %s",
			app.Name, app.DisplayName, app.Description, app.Version, app.Kind, app.ChartName, app.ChartVersion, app.HelmRegistryName)
	}
}

func (s *TestSuite) TestListBootStrapDeploymentPackages() {
	res, err := s.catalogClient.MakeHTTPRequestWithQuery("GET", types.DeploymentPackagesEndpoint, map[string]string{
		"orderBy": "name", "pageSize": "10", "offset": "0", "kinds": "KIND_EXTENSION",
	})
	s.Require().NoError(err)
	defer res.Body.Close()
	s.assertStatus(res, http.StatusOK)

	body := s.readResponseBody(res)
	var result struct {
		DeploymentPackages []types.DeploymentPackage `json:"deploymentPackages"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), len(types.GetDeploymentPackages()), len(result.DeploymentPackages), "Mismatch in the number of deployment packages")
	s.T().Log("Deployment Packages:")
	for _, pkg := range result.DeploymentPackages {
		s.T().Logf("Name: %s, Description: %s, Version: %s, Kind: %s", pkg.Name, pkg.Description, pkg.Version, pkg.Kind)
	}
}

func (s *TestSuite) TestListBootStrapRegistries() {
	res, err := s.catalogClient.MakeHTTPRequestWithQuery("GET", types.RegistriesEndpoint, map[string]string{
		"orderBy": "name", "pageSize": "10", "offset": "0", "showSensitiveInfo": "true",
	})
	s.Require().NoError(err)
	defer res.Body.Close()
	s.assertStatus(res, http.StatusOK)

	body := s.readResponseBody(res)
	var result struct {
		Registries []types.Registry `json:"registries"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), len(s.catalogClient.GetRegistries()), len(result.Registries), "Mismatch in the number of registries")
	s.T().Log("Registries:")
	for _, registry := range result.Registries {
		s.T().Logf("Name: %s, DisplayName: %s, Description: %s, RootURL: %s, Type: %s",
			registry.Name, registry.DisplayName, registry.Description, registry.RootURL, registry.Type)
	}
}

func (s *TestSuite) TestVerifyBootstrappedRegistriesExist() {
	for _, registry := range s.catalogClient.GetRegistries() {
		requestURL := fmt.Sprintf("%s%s/%s", s.catalogClient.CatalogRESTServerUrl, types.RegistriesEndpoint, registry.Name)
		res, err := s.catalogClient.MakeHTTPRequest("GET", requestURL, nil)
		s.Require().NoError(err)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			Registry types.Registry `json:"registry"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		switch {
		case registry.Name != result.Registry.Name:
			assert.Equal(s.T(), registry.Name, result.Registry.Name, "Mismatch in 'Name' for registry: %s", registry.Name)
		case registry.DisplayName != result.Registry.DisplayName:
			assert.Equal(s.T(), registry.DisplayName, result.Registry.DisplayName, "Mismatch in 'DisplayName' for registry: %s", registry.Name)
		case registry.RootURL != result.Registry.RootURL:
			oldDockerURL := fmt.Sprintf("https://registry-oci.%s/", s.orchDomain)
			newDockerURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", s.orchDomain)
			// Docker Registry URL was changed recently. Avoid throwing errors in a development environment that's using the new URL
			// TODO: remove this special case when component-tests are moved forward
			if registry.RootURL != oldDockerURL || result.Registry.RootURL != newDockerURL {
				assert.Equal(s.T(), registry.RootURL, result.Registry.RootURL, "Mismatch in 'RootURL' for registry: %s", registry.Name)
			}
		case registry.Type != result.Registry.Type:
			assert.Equal(s.T(), registry.Type, result.Registry.Type, "Mismatch in 'Type' for registry: %s", registry.Name)
		}
		// assert.Equal(s.T(), registry.Description, result.Registry.Description)
	}
}

func (s *TestSuite) assertStatus(res *http.Response, expectedStatus int) {
	assert.Equal(s.T(), expectedStatus, res.StatusCode, "Unexpected response status")
}

func (s *TestSuite) readResponseBody(res *http.Response) []byte {
	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	return body
}

func (s *TestSuite) unmarshalJSON(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	assert.NoError(s.T(), err)
}
