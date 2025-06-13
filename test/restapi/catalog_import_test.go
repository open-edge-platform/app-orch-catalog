// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	restclient "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"github.com/open-edge-platform/app-orch-deployment/test-common-utils/pkg/git"
	"github.com/open-edge-platform/app-orch-deployment/test-common-utils/pkg/loader"
)

func (s *TestSuite) UploadChartToHarbor() string {
	httpbinPath, err := git.CloneHttpbin()
	if err != nil {
		s.T().Fatalf("error: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(filepath.Dir(httpbinPath))) // Clean up the temporary directory after upload
	secret, _ := GetCliSecretHarbor(fmt.Sprintf("https://registry-oci.%s", s.orchDomain), s.Token)
	err = loader.UploadHttpbinHelm(httpbinPath, secret)
	if err != nil {
		s.T().Fatalf("error: %v", err)
	}
	return fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project/httpbin", s.orchDomain),
		"sample-project-edge-mgr",
		secret
}

func (s *TestSuite) TestImportHelmChart() {
	ociURL, ociUsername, ociPassword := s.UploadCharToHarbor()

	ctx := context.TODO()
	importRequest := &restclient.CatalogServiceImportParams{
		Url:       types.GetPointerString(ocuURL),
		Username:  types.GetPointerString(ociUsername),
		AuthToken: types.GetPointerString(ociPassword),
	}

	status, body, err := s.catalogClient.ImportHelmChart(ctx, importRequest)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, status, "Expected status code 200 for successful import")

	if http.StatusOK != status {
		// for debugging purposes, log the response body. It will have error messages from the importer.
		s.T().Logf("Response body: %s\n", body)
	}

	app, status, err := s.catalogClient.GetApplication(ctx, "impt", "2.9.0")
	s.Require().NoError(err, "Expected to retrieve application after import")
	s.Require().Equal(http.StatusOK, status, "Expected status code 200 for application retrieval")

	s.Equal("impt", app.Name, "Expected application name to be 'impt'")
	s.Equal("2.9.0", app.Version, "Expected application version to be '2.9.0'")
	s.Equal("impt-registry", app.HelmRegistryName, "Expected application registry to be 'impt-registry'")
	s.Equal("impt", app.ChartName, "Expected application chart name to be 'impt'")
	s.Equal("2.9.0", app.ChartVersion, "Expected application chart version to be '2.9.0'")
	s.Equal("default", *app.DefaultProfileName, "Expected application default profile name to be 'default'")

	pkg, status, err := s.catalogClient.GetDeploymentPackage(ctx, "impt", "2.9.0")
	s.Require().NoError(err, "Expected to retrieve deployment package after import")

	s.Equal("impt", pkg.Name, "Expected deployment package name to be 'impt'")
	s.Equal("2.9.0", pkg.Version, "Expected deployment package version to be '2.9.0'")
	s.Equal("default", *pkg.DefaultProfileName, "Expected deployment package default profile name to be 'default'")

	reg, status, err := s.catalogClient.GetRegistry(ctx, "impt-registry")
	s.Require().NoError(err, "Expected to retrieve registry")
	s.Require().Equal(http.StatusOK, status, "Expected status code 200 for registry retrieval")

	s.Equal("impt-registry", reg.Name, "Expected registry name to be 'impt-registry'")
	s.Equal("oci://ghcr.io/open-edge-platform/geti/helm", reg.RootUrl, "Expected registry URL to match the input")

	/* cleanup -- these should all exist at this point */

	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, "impt", "2.9.0", true), "Expected to delete deployment package")
	s.NoError(s.catalogClient.DeleteApplication(ctx, "impt", "2.9.0", true), "Expected to delete application")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, "impt-registry", true), "Expected to delete registry")
}

func (s *TestSuite) TestImportHelmChartBadURL() {

	ctx := context.TODO()
	importRequest := &restclient.CatalogServiceImportParams{
		Url: types.GetPointerString("oci://ghcr.invalid/open-edge-platform/geti/helm/impt:2.9.0"),
	}

	status, _, err := s.catalogClient.ImportHelmChart(ctx, importRequest)
	s.Require().NoError(err, "Expected no error when importing Helm chart with bad URL")
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	// TODO: test type of error returned
}

func (s *TestSuite) TestImportHelmChartNotAURL() {

	ctx := context.TODO()

	importRequest := &restclient.CatalogServiceImportParams{
		Url: types.GetPointerString("this is not a url"),
	}

	status, body, err := s.catalogClient.ImportHelmChart(ctx, importRequest)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	s.Contains(body, "Scheme is not oci", "Expected error message to contain 'scheme is not oci'")
}

func (s *TestSuite) TestImportHelmChartBadObject() {
	ctx := context.TODO()

	importRequest := &restclient.CatalogServiceImportParams{
		Url: types.GetPointerString("oci://registry-rs.edgeorchestration.intel.com/edge-orch/en/file/cluster-extension-manifest:v1.1.2"),
	}

	status, _, err := s.catalogClient.ImportHelmChart(ctx, importRequest)
	s.NoError(err)
	s.Equal(http.StatusBadRequest, status, "Expected status code 400 for invalid Helm chart URL")
	// TODO: test type of error returned
}
