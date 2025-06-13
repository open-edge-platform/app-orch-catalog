// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	// Third-party imports
	"github.com/stretchr/testify/assert"
	// Project-specific imports
)

// processResponse validates response status and returns body
func (s *TestSuite) processResponse(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	return body, err
}

// unmarshalJSON unmarshals response body into provided result struct
func (s *TestSuite) unmarshalJSON(body []byte, result interface{}) error {
	err := json.Unmarshal(body, result)
	assert.NoError(s.T(), err)
	return err
}

func (s *TestSuite) TestListBootStrapExtensions() {
	res, err := s.catalogClient.MakeAuthenticatedRequest("GET", types.ApplicationsEndpoint, nil, nil)
	assert.NoError(s.T(), err)

	body, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	var result struct {
		Applications []types.Application `json:"applications"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), len(types.GetApplications()), len(result.Applications), "Mismatch in the number of applications")
	// Log application details for debugging purposes
	log.Printf("Extensions:")
	for _, app := range result.Applications {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, Version: %s, Kind: %s, ChartName: %s, ChartVersion: %s, HelmRegistryName: %s",
			app.Name, app.DisplayName, app.Description, app.Version, app.Kind, app.ChartName, app.ChartVersion, app.HelmRegistryName)
	}
}

func (s *TestSuite) TestListBootStrapDeploymentPackages() {
	queryParams := map[string]string{
		"orderBy":  "name",
		"pageSize": "10",
		"offset":   "0",
		"kinds":    "KIND_EXTENSION",
	}

	res, err := s.catalogClient.MakeAuthenticatedRequest("GET", types.DeploymentPackagesEndpoint, nil, queryParams)
	assert.NoError(s.T(), err)

	body, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	var result struct {
		DeploymentPackages []types.DeploymentPackage `json:"deploymentPackages"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), len(types.GetDeploymentPackages()), len(result.DeploymentPackages), "Mismatch in the number of deployment packages")

	// Log deployment package details for debugging purposes
	log.Printf("Deployment Packages:")
	for _, pkg := range result.DeploymentPackages {
		log.Printf("Name: %s, Description: %s, Version: %s, Kind: %s",
			pkg.Name, pkg.Description, pkg.Version, pkg.Kind)
	}
}

func (s *TestSuite) TestListBootStrapRegistries() {
	queryParams := map[string]string{
		"orderBy":           "name",
		"pageSize":          "10",
		"offset":            "0",
		"showSensitiveInfo": "true",
	}

	res, err := s.catalogClient.MakeAuthenticatedRequest("GET", types.RegistriesEndpoint, nil, queryParams)
	assert.NoError(s.T(), err)
	defer res.Body.Close()

	if res.Status != "200 OK" {
		s.Equal("200 OK", res.Status)
		return // Everything else is going to fail...
	}

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)

	var result struct {
		Registries []types.Registry `json:"registries"`
	}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Assert that the size of the result.Registries matches the size of getRegistries
	assert.Equal(s.T(), len(s.catalogClient.GetRegistries()), len(result.Registries), "Mismatch in the number of registries")
	// Log registry details for debugging purposes
	log.Printf("Registries:")
	for _, registry := range result.Registries {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, RootURL: %s, Type: %s",
			registry.Name, registry.DisplayName, registry.Description, registry.RootURL, registry.Type)
	}
}

func (s *TestSuite) TestVerifyBootstrappedRegistriesExist() {
	for _, registry := range s.catalogClient.GetRegistries() {
		endpoint := fmt.Sprintf("%s/%s", types.RegistriesEndpoint, registry.Name)
		res, err := s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, nil)
		assert.NoError(s.T(), err)

		if res.Status != "200 OK" {
			assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for Registry: %s", registry.Name)
			res.Body.Close()
			continue
		}

		body, err := s.processResponse(res)
		assert.NoError(s.T(), err)

		var result struct {
			Registry types.Registry `json:"registry"`
		}
		s.unmarshalJSON(body, &result)

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

func (s *TestSuite) TestVerifyBootstrappedExtensionsExist() {
	for _, app := range types.GetApplications() {
		endpoint := fmt.Sprintf("%s/%s/versions", types.ApplicationsEndpoint, app.Name)
		res, err := s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, nil)
		assert.NoError(s.T(), err)

		if res.Status != "200 OK" {
			assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for application: %s - %s", app.Name, endpoint)
			res.Body.Close()
			continue
		}

		body, err := s.processResponse(res)
		assert.NoError(s.T(), err)

		var result struct {
			Application []types.Application `json:"application"`
		}
		s.unmarshalJSON(body, &result)

		s.True(len(result.Application) > 0, "Expected at least one application for %s", app.Name)

		if len(result.Application) > 0 {
			gotApp := result.Application[0]

			switch {
			case app.Name != gotApp.Name:
				assert.Equalf(s.T(), app.Name, gotApp.Name, "Mismatch in 'Name' for application: %s", app.Name)
			case app.DisplayName != gotApp.DisplayName:
				assert.Equalf(s.T(), app.DisplayName, gotApp.DisplayName, "Mismatch in 'DisplayName' for application: %s", app.Name)
			case app.ChartName != gotApp.ChartName:
				assert.Equalf(s.T(), app.ChartName, gotApp.ChartName, "Mismatch in 'ChartName' for application: %s", app.Name)
			case app.Kind != gotApp.Kind:
				assert.Equalf(s.T(), app.Kind, gotApp.Kind, "Mismatch in 'Kind' for application: %s", app.Name)
			case app.HelmRegistryName != gotApp.HelmRegistryName:
				assert.Equalf(s.T(), app.HelmRegistryName, gotApp.HelmRegistryName, "Mismatch in 'HelmRegistryName' for application: %s", app.Name)
			}
			//assert.Equal(s.T(), app.Description, result.Application.Description)
		}
	}
}

func (s *TestSuite) TestVerifyBootstrappedDeploymentPackagesExist() {
	for _, pkg := range types.GetDeploymentPackages() {
		endpoint := fmt.Sprintf("%s/%s/versions", types.DeploymentPackagesEndpoint, pkg.Name)
		res, err := s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, nil)
		assert.NoError(s.T(), err)

		if res.Status != "200 OK" {
			assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for Package: %s", pkg.Name)
			res.Body.Close()
			continue // Everything else is going to fail...
		}

		body, err := s.processResponse(res)
		assert.NoError(s.T(), err)

		var result struct {
			DeploymentPackage []types.DeploymentPackage `json:"deploymentPackages"`
		}
		s.unmarshalJSON(body, &result)

		s.True(len(result.DeploymentPackage) > 0, "Expected at least one deployment package for %s", pkg.Name)
		if len(result.DeploymentPackage) > 0 {
			gotPkg := result.DeploymentPackage[0]

			switch {
			case pkg.Name != gotPkg.Name:
				assert.Equalf(s.T(), pkg.Name, gotPkg.Name, "Mismatch in 'Name' for deployment package: %s", pkg.Name)
			case pkg.Kind != gotPkg.Kind:
				assert.Equalf(s.T(), pkg.Kind, gotPkg.Kind, "Mismatch in 'Kind' for deployment package: %s", pkg.Name)
			}
		}
	}
}

func (s *TestSuite) Delete(url string) {
	endpoint := strings.TrimPrefix(url, s.catalogClient.CatalogRESTServerUrl)
	res, err := s.catalogClient.MakeAuthenticatedRequest("DELETE", endpoint, nil, nil)
	assert.NoError(s.T(), err)

	defer res.Body.Close()
	if res.Status != "200 OK" {
		assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for delete on url %s", url)
	}
}

func (s *TestSuite) TestUploadTarball() {
	ctx := context.TODO()
	res, err := s.catalogClient.UploadTarball(ctx, types.WordpressTarballPathName)
	assert.NoError(s.T(), err, "Expected to upload tarball without error")
	assert.Equal(s.T(), http.StatusOK, res.StatusCode, "Expected HTTP status code 200 OK for upload")

	defer res.Body.Close()
	if res.Status != "200 OK" {
		// print response message if something has gone wrong, for debugging
		bodyBytes, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)
		log.Printf("Response Body: %s", string(bodyBytes))
	}

	// Make sure the wordpress DP was created
	endpoint := fmt.Sprintf("%s/test-wordpress/versions/0.1.1", types.DeploymentPackagesEndpoint)
	queryParams := map[string]string{
		"orderBy":  "name",
		"pageSize": "10",
		"offset":   "0",
	}

	res, err = s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, queryParams)
	assert.NoError(s.T(), err)

	body, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	var result struct {
		DeploymentPackage types.DeploymentPackage `json:"deploymentPackage"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), types.WordpressName, result.DeploymentPackage.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), types.WordpressVersion, result.DeploymentPackage.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after upload")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after upload")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after upload")
}

func (s *TestSuite) TestUploadSeparateFiles() {
	ctx := context.TODO()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	pathNames := []string{"../testdata/wordpress/app-wordpress-0.1.1.yaml",
		"../testdata/wordpress/dp-wordpress-0.1.1.yaml",
		"../testdata/wordpress/registry-bitnami.yaml",
		"../testdata/wordpress/values-wordpress-0.1.1.yaml",
	}

	for _, pathName := range pathNames {
		file, err := os.Open(pathName)
		assert.NoError(s.T(), err)
		defer file.Close()

		fileName := pathName[strings.LastIndex(pathName, "/")+1:]

		part, _ := writer.CreateFormFile("files", fileName)
		_, err = io.Copy(part, file)
		assert.NoError(s.T(), err)
	}

	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	res, err := s.catalogClient.MakeAuthenticatedRequest("POST", types.UploadEndpoint, body, nil, headers)
	assert.NoError(s.T(), err)

	defer res.Body.Close()
	assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for upload")
	if res.Status != "200 OK" {
		// print response message if something has gone wrong, for debugging
		bodyBytes, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)
		log.Printf("Response Body: %s", string(bodyBytes))
	}

	// Make sure the wordpress DP was created
	endpoint := fmt.Sprintf("%s/test-wordpress/versions/0.1.1", types.DeploymentPackagesEndpoint)
	queryParams := map[string]string{
		"orderBy":  "name",
		"pageSize": "10",
		"offset":   "0",
	}

	res, err = s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, queryParams)
	assert.NoError(s.T(), err)

	resBody, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	var result struct {
		DeploymentPackage types.DeploymentPackage `json:"deploymentPackage"`
	}
	s.unmarshalJSON(resBody, &result)

	assert.Equal(s.T(), "test-wordpress", result.DeploymentPackage.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), "0.1.1", result.DeploymentPackage.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after upload")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after upload")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after upload")
}

func (s *TestSuite) TestGetCharts() {
	res, err := s.catalogClient.MakeAuthenticatedRequest("GET", "/catalog.orchestrator.apis/charts", nil, map[string]string{"registry": "harbor-helm-oci"})
	assert.NoError(s.T(), err)

	body, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	// On a fresh orchestrator there should be no charts in the registry
	assert.Equal(s.T(), "null", string(body), "Expected the response body to be empty")
}
