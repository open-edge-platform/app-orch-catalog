// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/methods"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
)

func (s *TestSuite) TestListBootStrapExtensions() {
	res, err := methods.MakeHTTPRequest("GET", types.ApplicationsEndpoint, nil)
	s.Require().NoError(err)
	defer res.Body.Close()
	s.assertStatus(res, http.StatusOK)

	body := s.readResponseBody(res)
	var result struct {
		Applications []types.Application `json:"applications"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), len(types.GetApplications()), len(result.Applications), "Mismatch in the number of applications")
	log.Printf("Extensions:")
	for _, app := range result.Applications {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, Version: %s, Kind: %s, ChartName: %s, ChartVersion: %s, HelmRegistryName: %s",
			app.Name, app.DisplayName, app.Description, app.Version, app.Kind, app.ChartName, app.ChartVersion, app.HelmRegistryName)
	}
}

func (s *TestSuite) TestListBootStrapDeploymentPackages() {
	res, err := methods.MakeHTTPRequestWithQuery("GET", types.DeploymentPackagesEndpoint, map[string]string{
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
	log.Printf("Deployment Packages:")
	for _, pkg := range result.DeploymentPackages {
		log.Printf("Name: %s, Description: %s, Version: %s, Kind: %s", pkg.Name, pkg.Description, pkg.Version, pkg.Kind)
	}
}

func (s *TestSuite) TestListBootStrapRegistries() {
	res, err := methods.MakeHTTPRequestWithQuery("GET", types.RegistriesEndpoint, map[string]string{
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
	log.Printf("Registries:")
	for _, registry := range result.Registries {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, RootURL: %s, Type: %s",
			registry.Name, registry.DisplayName, registry.Description, registry.RootURL, registry.Type)
	}
}

func (s *TestSuite) TestVerifyBootstrappedRegistriesExist() {
	for _, registry := range s.catalogClient.GetRegistries() {
		res, err := methods.MakeHTTPRequest("GET", fmt.Sprintf("%s/%s", types.RegistriesEndpoint, registry.Name), nil)
		s.Require().NoError(err)
		defer res.Body.Close()
		s.assertStatus(res, http.StatusOK)

		body := s.readResponseBody(res)
		var result struct {
			Registry types.Registry `json:"registry"`
		}
		s.unmarshalJSON(body, &result)

		assert.Equal(s.T(), registry.Name, result.Registry.Name, "Mismatch in 'Name' for registry: %s", registry.Name)
		assert.Equal(s.T(), registry.DisplayName, result.Registry.DisplayName, "Mismatch in 'DisplayName' for registry: %s", registry.Name)
		assert.Equal(s.T(), registry.RootURL, result.Registry.RootURL, "Mismatch in 'RootURL' for registry: %s", registry.Name)
		assert.Equal(s.T(), registry.Type, result.Registry.Type, "Mismatch in 'Type' for registry: %s", registry.Name)
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
