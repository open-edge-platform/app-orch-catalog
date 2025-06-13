// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	restclient "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"github.com/open-edge-platform/app-orch-deployment/app-service-proxy/test/headertest"
	"github.com/open-edge-platform/app-orch-deployment/test-common-utils/pkg/git"
)

// TODO: update the one in ADM repo to support registry as arg,  then remove this duplicate function
func UploadHttpbinHelm(path, registry, harborPwd string) error {
	chartPath := path + "/helm"                      // Path to your chart directory
	repo := "catalog-apps-sample-org-sample-project" // Repository name in your OCI registry
	username := "sample-project-edge-mgr"

	// 1. Login to the OCI registry
	regRef := fmt.Sprintf("https://%s/%s", registry, repo)
	loginCmd := exec.Command(
		"helm", "registry", "login",
		"-u", username,
		"--password", harborPwd,
		regRef,
	)
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	fmt.Println("Logging in to OCI registry...")
	if err := loginCmd.Run(); err != nil {
		fmt.Printf("Failed to login to OCI registry: %v\n", err)
		os.Exit(1)
	}

	// 2. Package the Helm chart
	version := "0.1.8"
	pkgCmd := exec.Command("helm", "package", chartPath, "--version", version)
	pkgCmd.Stdout = os.Stdout
	pkgCmd.Stderr = os.Stderr
	fmt.Println("Packaging chart...")
	if err := pkgCmd.Run(); err != nil {
		fmt.Printf("Failed to package chart: %v\n", err)
		os.Exit(1)
	}

	chartName := "httpbin"
	chartTGZ := fmt.Sprintf("%s-%s.tgz", chartName, version)

	// 3. Push the chart to OCI registry
	ociRef := fmt.Sprintf("oci://%s/%s", registry, repo)
	fmt.Println("Pushing chart to OCI registry...")
	retries := 40
	var err error
	for i := 0; i < retries; i++ {
		pushCmd := exec.Command("helm", "push", chartTGZ, ociRef)
		if err = pushCmd.Run(); err != nil {
			fmt.Printf("retry count %d\n", i)
			time.Sleep(1 * time.Second)
		}
	}

	if err != nil {
		fmt.Printf("Failed to push chart: %v\n", err)
		os.Exit(1)
	}
	// Optional: Cleanup the packaged file
	os.Remove(filepath.Join(".", chartTGZ))
	fmt.Println("Done!")
	return nil
}

func (s *TestSuite) UploadChartToHarbor() (string, string, string) {
	httpbinPath, err := git.CloneHttpbin()
	if err != nil {
		s.T().Fatalf("error: %v", err)
	}
	defer os.RemoveAll(filepath.Dir(filepath.Dir(httpbinPath))) // Clean up the temporary directory after upload
	harborHostname := fmt.Sprintf("registry-oci.%s", s.orchDomain)
	harborUrl := fmt.Sprintf("https://%s", harborHostname)
	secret, _ := headertest.GetCliSecretHarbor(harborUrl, s.token)
	err = UploadHttpbinHelm(httpbinPath, harborHostname, secret)
	if err != nil {
		s.T().Fatalf("error: %v", err)
	}
	return fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project/httpbin", s.orchDomain),
		"sample-project-edge-mgr",
		secret
}

func (s *TestSuite) TestImportHelmChart() {
	ociURL, ociUsername, ociPassword := s.UploadChartToHarbor()

	ctx := context.TODO()
	importRequest := &restclient.CatalogServiceImportParams{
		Url:       types.GetPointerString(ociURL),
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

	app, status, err := s.catalogClient.GetApplication(ctx, "httpbin", "0.1.8")
	s.Require().NoError(err, "Expected to retrieve application after import")
	s.Require().Equal(http.StatusOK, status, "Expected status code 200 for application retrieval")

	s.Equal("httpbin", app.Name, "Expected application name to be 'httpbin'")
	s.Equal("0.1.8", app.Version, "Expected application version to be '0.1.8'")
	s.Equal("httpbin-registry", app.HelmRegistryName, "Expected application registry to be 'httpbin-registry'")
	s.Equal("httpbin", app.ChartName, "Expected application chart name to be 'httpbin'")
	s.Equal("0.1.8", app.ChartVersion, "Expected application chart version to be '0.1.8'")
	s.Equal("default", *app.DefaultProfileName, "Expected application default profile name to be 'default'")

	pkg, status, err := s.catalogClient.GetDeploymentPackage(ctx, "httpbin", "0.1.8")
	s.Require().NoError(err, "Expected to retrieve deployment package after import")

	s.Equal("httpbin", pkg.Name, "Expected deployment package name to be 'httpbin'")
	s.Equal("0.1.8", pkg.Version, "Expected deployment package version to be '0.1.8'")
	s.Equal("default", *pkg.DefaultProfileName, "Expected deployment package default profile name to be 'default'")

	reg, status, err := s.catalogClient.GetRegistry(ctx, "httpbin-registry")
	s.Require().NoError(err, "Expected to retrieve registry")
	s.Require().Equal(http.StatusOK, status, "Expected status code 200 for registry retrieval")

	s.Equal("httpbin-registry", reg.Name, "Expected registry name to be 'httpbin-registry'")
	// s.Equal("oci://ghcr.io/open-edge-platform/geti/helm", reg.RootUrl, "Expected registry URL to match the input")

	/* cleanup -- these should all exist at this point */

	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, "httpbin", "0.1.8", true), "Expected to delete deployment package")
	s.NoError(s.catalogClient.DeleteApplication(ctx, "httpbin", "0.1.8", true), "Expected to delete application")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, "httpbin-registry", true), "Expected to delete registry")
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
