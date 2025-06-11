// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	// Third-party imports

	// Project-specific imports
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
)

const importEndpoint = "/catalog.orchestrator.apis/v3/import"

type ImportRequest struct {
	URL                       string `json:"url"`
	Username                  string `json:"username,omitempty"`
	AuthToken                 string `json:"auth_token,omitempty"`
	ChartValues               string `json:"chart_values,omitempty"`
	IncludeAuth               bool   `json:"include_auth,omitempty"`
	GenerateDefaultValues     bool   `json:"generate_default_values,omitempty"`
	GenerateDefaultParameters bool   `json:"generate_default_parameters,omitempty"`
	Namespace                 string `json:"namespace,omitempty"`
}

func (s *TestSuite) ImportHelmChart(importRequest *ImportRequest) (int, string) {
	params := url.Values{}
	params.Add("url", importRequest.URL)

	requestURL := fmt.Sprintf("%s%s?%s", s.CatalogRESTServerUrl, importEndpoint, params.Encode())

	req, err := http.NewRequest("POST", requestURL, nil)
	s.Require().NoError(err, "Expected to create HTTP request for importing Helm chart")
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	log.Printf("Importing Helm chart with request URL: %s", requestURL)

	res, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	s.Require().NoError(err)

	return res.StatusCode, string(body)
}

func (s *TestSuite) TestImportHelmChart() {
	importRequest := &ImportRequest{
		URL: "oci://ghcr.io/open-edge-platform/geti/helm/impt:2.9.0",
	}

	status, body := s.ImportHelmChart(importRequest)
	s.Equal(http.StatusOK, status, "Expected status code 200 for successful import")

	_ = body

	app, err := s.GetApplication("impt", "2.9.0", true)
	s.Require().NoError(err, "Expected to retrieve application after import")

	s.Equal("impt", app.Name, "Expected application name to be 'impt'")
	s.Equal("2.9.0", app.Version, "Expected application version to be '2.9.0'")
	s.Equal("impt-registry", app.HelmRegistryName, "Expected application registry to be 'impt-registry'")
	s.Equal("impt", app.ChartName, "Expected application chart name to be 'impt'")
	s.Equal("2.9.0", app.ChartVersion, "Expected application chart version to be '2.9.0'")
	s.Equal("default", app.DefaultProfileName, "Expected application default profile name to be 'default'")

	pkg, err := s.GetDeploymentPackage("impt", "2.9.0", true)
	s.Require().NoError(err, "Expected to retrieve deployment package after import")

	s.Equal("impt", pkg.Name, "Expected deployment package name to be 'impt'")
	s.Equal("2.9.0", pkg.Version, "Expected deployment package version to be '2.9.0'")
	s.Equal("default", pkg.DefaultProfileName, "Expected deployment package default profile name to be 'default'")

	reg, err := s.GetRegistry("impt-registry", true)
	s.Require().NoError(err, "Expected to retrieve registry")

	s.Equal("impt-registry", reg.Name, "Expected registry name to be 'impt-registry'")
	s.Equal("oci://ghcr.io/open-edge-platform/geti/helm", reg.RootURL, "Expected registry URL to match the input")

	/* cleanup -- these should all exist at this point */

	s.NoError(s.DeleteDeploymentPackage("impt", "2.9.0", true), "Expected to delete registry")
	s.NoError(s.DeleteApplication("impt", "2.9.0", true), "Expected to delete registry")
	s.NoError(s.DeleteRegistry("impt-registry", true), "Expected to delete registry")
}

func (s *TestSuite) TestImportHelmChartBadURL() {
	importRequest := &ImportRequest{
		URL: "oci://ghcr.invalid/open-edge-platform/geti/helm/impt:2.9.0",
	}

	status, body := s.ImportHelmChart(importRequest)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	s.Contains(body, "failed to resolve", "Expected error message to contain 'failed to resolve'")
}

func (s *TestSuite) TestImportHelmChartNotAURL() {
	importRequest := &ImportRequest{
		URL: "this is not a url",
	}

	status, body := s.ImportHelmChart(importRequest)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	s.Contains(body, "Scheme is not oci", "Expected error message to contain 'scheme is not oci'")
}

func (s *TestSuite) TestImportHelmChartBadObject() {
	importRequest := &ImportRequest{
		URL: "oci://registry-rs.edgeorchestration.intel.com/edge-orch/en/file/cluster-extension-manifest:v1.1.2",
	}

	status, body := s.ImportHelmChart(importRequest)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	s.Contains(body, "Failed to create gzip reader", "Expected error message to contain 'failed to resolve'")
}
