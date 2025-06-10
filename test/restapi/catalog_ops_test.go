// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	// Third-party imports

	// Project-specific imports
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/stretchr/testify/assert"
)

type (
	// DeploymentPackage is the JSON representation of a deployment package.
	DeploymentPackage struct {
		ApplicationDependencies *[]ApplicationDependency `json:"applicationDependencies,omitempty"`
		ApplicationReferences   []ApplicationReference   `json:"applicationReferences"`
		Artifacts               []ArtifactReference      `json:"artifacts"`
		DefaultNamespaces       *map[string]string       `json:"defaultNamespaces,omitempty"`
		DefaultProfileName      string                   `json:"defaultProfileName,omitempty"`
		Description             string                   `json:"description,omitempty"`
		DisplayName             string                   `json:"displayName,omitempty"`
		Extensions              []APIExtension           `json:"extensions"`
		IsDeployed              bool                     `json:"isDeployed,omitempty"`
		IsVisible               bool                     `json:"isVisible,omitempty"`
		Name                    string                   `json:"name"`
		Profiles                []Profile                `json:"profiles,omitempty"`
		Version                 string                   `json:"version"`
		Kind                    string                   `json:"kind"`
	}

	// DeploymentPackages is the JSON representation of a list of deployment packages.
	DeploymentPackages struct {
		DeploymentPackages []DeploymentPackage `json:"DeploymentPackages"`
	}

	// DeploymentPackageGetResponse is the JSON representation of a get of an application.
	DeploymentPackageGetResponse struct {
		DeploymentPackage DeploymentPackage `json:"deploymentPackage"`
	}

	Profile struct {
		ChartValues string `json:"chartValues,omitempty"`
		Description string `json:"description,omitempty"`
		DisplayName string `json:"displayName,omitempty"`
		Name        string `json:"name"`
	}

	// Application is the JSON representation of an application.
	Application struct {
		ChartName          string    `json:"chartName"`
		ChartVersion       string    `json:"chartVersion"`
		DefaultProfileName string    `json:"defaultProfileName,omitempty"`
		Description        string    `json:"description,omitempty"`
		DisplayName        string    `json:"displayName,omitempty"`
		HelmRegistryName   string    `json:"helmRegistryName"`
		ImageRegistryName  string    `json:"imageRegistryName,omitempty"`
		Name               string    `json:"name"`
		Profiles           []Profile `json:"profiles,omitempty"`
		Version            string    `json:"version"`
		Kind               string    `json:"kind"`
	}

	// ApplicationGetResponse is the JSON representation of a get of an application.
	ApplicationGetResponse struct {
		Application Application `json:"application"`
	}

	ApplicationDependency struct{}
	ApplicationReference  struct{}
	ArtifactReference     struct{}
	Endpoint              struct {
		AuthType     string `json:"authType"`
		ExternalPath string `json:"externalPath"`
		InternalPath string `json:"internalPath"`
		Scheme       string `json:"scheme"`
		ServiceName  string `json:"serviceName"`
	}

	UIExtension struct{}

	APIExtension struct {
		Description string      `json:"description,omitempty"`
		DisplayName string      `json:"displayName,omitempty"`
		Endpoints   []Endpoint  `json:"endpoints,omitempty"`
		Name        string      `json:"name"`
		UiExtension UIExtension `json:"uiExtension,omitempty"` //nolint: revive,stylecheck
		Version     string      `json:"version"`
	}
	// Applications is the JSON representation of a list of applications.
	Applications struct {
		Applications []Application `json:"applications"`
	}

	Registry struct {
		AuthToken   string  `json:"authToken,omitempty"`
		Cacerts     string  `json:"cacerts,omitempty"`
		Description string  `json:"description,omitempty"`
		DisplayName string  `json:"displayName,omitempty"`
		Name        string  `json:"name"`
		RootURL     string  `json:"rootUrl"`
		SecretID    *string `json:"secretId,omitempty"`
		Type        string  `json:"type"`
		Username    string  `json:"username,omitempty"`
	}

	// RegistryGetResponse is the JSON representation of the result of a get of a registry
	RegistryGetResponse struct {
		Registry Registry `json:"registry"`
	}
)

func (s *TestSuite) GetApplication(name string, version string, mustExist bool) (*Application, error) {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", s.CatalogRESTServerUrl, applicationsEndpoint, name, version)

	req, err := http.NewRequest("GET", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound && !mustExist {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get application: %s", res.Status)
	}

	var appResponse ApplicationGetResponse
	err = json.NewDecoder(res.Body).Decode(&appResponse)
	if err != nil {
		return nil, err
	}

	return &appResponse.Application, nil
}

func (s *TestSuite) DeleteApplication(name string, version string, mustExist bool) error {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", s.CatalogRESTServerUrl, applicationsEndpoint, name, version)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound && !mustExist {
		return nil
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete application: %s", res.Status)
	}
	return nil
}

func (s *TestSuite) GetDeploymentPackage(name string, version string, mustExist bool) (*DeploymentPackage, error) {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", s.CatalogRESTServerUrl, deploymentPackagesEndpoint, name, version)

	req, err := http.NewRequest("GET", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound && !mustExist {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get deployment package: %s", res.Status)
	}

	var pkgResp DeploymentPackageGetResponse
	err = json.NewDecoder(res.Body).Decode(&pkgResp)
	if err != nil {
		return nil, err
	}

	return &pkgResp.DeploymentPackage, nil
}

func (s *TestSuite) DeleteDeploymentPackage(name string, version string, mustExist bool) error {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", s.CatalogRESTServerUrl, deploymentPackagesEndpoint, name, version)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound && !mustExist {
		return nil
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete deployment package: %s", res.Status)
	}

	return nil
}

func (s *TestSuite) GetRegistry(name string, mustExist bool) (*Registry, error) {
	requestURL := fmt.Sprintf("%s%s/%s", s.CatalogRESTServerUrl, registriesEndpoint, name)

	req, err := http.NewRequest("GET", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound && !mustExist {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get registry: %s", res.Status)
	}

	var regResp RegistryGetResponse
	err = json.NewDecoder(res.Body).Decode(&regResp)
	if err != nil {
		return nil, err
	}

	return &regResp.Registry, nil
}

func (s *TestSuite) DeleteRegistry(name string, mustExist bool) error {
	requestURL := fmt.Sprintf("%s%s/%s", s.CatalogRESTServerUrl, registriesEndpoint, name)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound && !mustExist {
		return nil
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete registry: %s", res.Status)
	}

	return nil
}

func (s *TestSuite) UploadTarball(pathName string) (*http.Response, error) {
	file, err := os.Open(pathName)
	assert.NoError(s.T(), err)
	defer file.Close()

	filename := filepath.Base(pathName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", filename)
	_, err = io.Copy(part, file)
	assert.NoError(s.T(), err)
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, uploadEndpoint), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
