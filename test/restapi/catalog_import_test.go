// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"net/http"

	"github.com/open-edge-platform/app-orch-catalog/test/utils/methods"
)

func (s *TestSuite) TestImportHelmChart() {
	importRequest := &methods.ImportRequest{
		URL: "oci://ghcr.io/open-edge-platform/geti/helm/impt:2.9.0",
	}

	status, body, err := s.catalogClient.ImportHelmChart(importRequest)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, status, "Expected status code 200 for successful import")

	if http.StatusOK != status {
		// for debugging purposes, log the response body. It will have error messages from the importer.
		s.T().Logf("Response body: %s\n", body)
	}

	app, err := s.catalogClient.GetApplication("impt", "2.9.0", true)
	s.Require().NoError(err, "Expected to retrieve application after import")

	s.Equal("impt", app.Name, "Expected application name to be 'impt'")
	s.Equal("2.9.0", app.Version, "Expected application version to be '2.9.0'")
	s.Equal("impt-registry", app.HelmRegistryName, "Expected application registry to be 'impt-registry'")
	s.Equal("impt", app.ChartName, "Expected application chart name to be 'impt'")
	s.Equal("2.9.0", app.ChartVersion, "Expected application chart version to be '2.9.0'")
	s.Equal("default", app.DefaultProfileName, "Expected application default profile name to be 'default'")

	pkg, err := s.catalogClient.GetDeploymentPackage("impt", "2.9.0", true)
	s.Require().NoError(err, "Expected to retrieve deployment package after import")

	s.Equal("impt", pkg.Name, "Expected deployment package name to be 'impt'")
	s.Equal("2.9.0", pkg.Version, "Expected deployment package version to be '2.9.0'")
	s.Equal("default", pkg.DefaultProfileName, "Expected deployment package default profile name to be 'default'")

	reg, err := s.catalogClient.GetRegistry("impt-registry", true)
	s.Require().NoError(err, "Expected to retrieve registry")

	s.Equal("impt-registry", reg.Name, "Expected registry name to be 'impt-registry'")
	s.Equal("oci://ghcr.io/open-edge-platform/geti/helm", reg.RootURL, "Expected registry URL to match the input")

	/* cleanup -- these should all exist at this point */

	s.NoError(s.catalogClient.DeleteDeploymentPackage("impt", "2.9.0", true), "Expected to delete deployment package")
	s.NoError(s.catalogClient.DeleteApplication("impt", "2.9.0", true), "Expected to delete application")
	s.NoError(s.catalogClient.DeleteRegistry("impt-registry", true), "Expected to delete registry")
}

func (s *TestSuite) TestImportHelmChartBadURL() {
	importRequest := &methods.ImportRequest{
		URL: "oci://ghcr.invalid/open-edge-platform/geti/helm/impt:2.9.0",
	}

	status, _, err := s.catalogClient.ImportHelmChart(importRequest)
	s.Require().NoError(err, "Expected no error when importing Helm chart with bad URL")
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	// TODO: test type of error returned
}

func (s *TestSuite) TestImportHelmChartNotAURL() {
	importRequest := &methods.ImportRequest{
		URL: "this is not a url",
	}

	status, body, err := s.catalogClient.ImportHelmChart(importRequest)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	s.Contains(body, "Scheme is not oci", "Expected error message to contain 'scheme is not oci'")
}

func (s *TestSuite) TestImportHelmChartBadObject() {
	importRequest := &methods.ImportRequest{
		URL: "oci://registry-rs.edgeorchestration.intel.com/edge-orch/en/file/cluster-extension-manifest:v1.1.2",
	}

	status, _, err := s.catalogClient.ImportHelmChart(importRequest)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	// TODO: test type of error returned
}
