// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package methods

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/auth"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type CatalogClient struct {
	CatalogRESTServerUrl string
	Token                string
	ProjectID            string
	OrchDomain           string
}

type ShortRegistry struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	RootURL     string `json:"rootUrl"`
	Type        string `json:"type"`
}

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

func (c *CatalogClient) GetApplication(name, version string, mustExist bool) (*types.Application, error) {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", c.CatalogRESTServerUrl, types.ApplicationsEndpoint, name, version)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

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

	var appResponse types.ApplicationGetResponse
	err = json.NewDecoder(res.Body).Decode(&appResponse)
	if err != nil {
		return nil, err
	}

	return &appResponse.Application, nil
}

func (c *CatalogClient) DeleteApplication(name, version string, mustExist bool) error {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", c.CatalogRESTServerUrl, types.ApplicationsEndpoint, name, version)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return err
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

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

func (c *CatalogClient) GetDeploymentPackage(name, version string, mustExist bool) (*types.DeploymentPackage, error) {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", c.CatalogRESTServerUrl, types.DeploymentPackagesEndpoint, name, version)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

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

	var pkgResp types.DeploymentPackageGetResponse
	err = json.NewDecoder(res.Body).Decode(&pkgResp)
	if err != nil {
		return nil, err
	}

	return &pkgResp.DeploymentPackage, nil
}

func (c *CatalogClient) DeleteDeploymentPackage(name, version string, mustExist bool) error {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", c.CatalogRESTServerUrl, types.DeploymentPackagesEndpoint, name, version)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return err
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

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

func (c *CatalogClient) GetRegistry(name string, mustExist bool) (*types.Registry, error) {
	requestURL := fmt.Sprintf("%s%s/%s", c.CatalogRESTServerUrl, types.RegistriesEndpoint, name)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

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

	var regResp types.RegistryGetResponse
	err = json.NewDecoder(res.Body).Decode(&regResp)
	if err != nil {
		return nil, err
	}

	return &regResp.Registry, nil
}

func (c *CatalogClient) DeleteRegistry(name string, mustExist bool) error {
	requestURL := fmt.Sprintf("%s%s/%s", c.CatalogRESTServerUrl, types.RegistriesEndpoint, name)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return err
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

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

func (c *CatalogClient) Delete(url string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer res.Body.Close()

	if res.Status != "200 OK" {
		return fmt.Errorf("unexpected response status: %s for DELETE on url %s", res.Status, url)
	}

	return nil
}

func (c *CatalogClient) UploadTarball(pathName string) (*http.Response, error) {
	file, err := os.Open(pathName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", pathName, err)
	}
	defer file.Close()

	filename := filepath.Base(pathName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file for %s: %w", filename, err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content for %s: %w", filename, err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.CatalogRESTServerUrl, types.UploadEndpoint), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for uploading tarball: %w", err)
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

	return http.DefaultClient.Do(req)
}

func (c *CatalogClient) ExportDeploymentPackage(name string, version string) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s/download", c.CatalogRESTServerUrl, types.DeploymentPackagesEndpoint, name, version)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request for exporting Deployment Package: %w", err)
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request for exporting Deployment Package: %w", err)
	}

	// Note: res.Body is intentionally not closed here, as the caller is expected to handle it.

	return res, nil
}

func (c *CatalogClient) ImportHelmChart(importRequest *ImportRequest) (int, string, error) {
	params := url.Values{}
	params.Add("url", importRequest.URL)

	requestURL := fmt.Sprintf("%s%s?%s", c.CatalogRESTServerUrl, types.ImportEndpoint, params.Encode())

	req, err := http.NewRequest("POST", requestURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create HTTP request for importing Helm chart: %w", err)
	}
	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to execute HTTP request for importing Helm chart: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return res.StatusCode, string(body), nil
}

func (c *CatalogClient) GetRegistries() []types.Registry {
	dockerURL := fmt.Sprintf("https://registry-oci.%s/", c.OrchDomain)
	helmURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", c.OrchDomain)

	regs := []types.Registry{}
	for _, ra := range []ShortRegistry{
		{"akri-helm-registry", "akri-helm-registry", "Public registry for akri chart", "https://project-akri.github.io/akri/", "HELM"},
		{"bitnami-helm-oci", "bitnami-helm-oci", "Bitnami helm registry", "oci://registry-1.docker.io/bitnamicharts", "HELM"},
		{"fluent-bit", "fluent-bit", "Public registry for fluent bit chart", "https://fluent.github.io/helm-charts", "HELM"},
		{"gatekeeper", "gatekeeper", "Public registry for gatekeeper chart", "https://open-policy-agent.github.io/gatekeeper/charts", "HELM"},
		{"harbor-docker-oci", "harbor oci docker", "Harbor OCI docker images registry", dockerURL, "IMAGE"},
		{"harbor-helm-oci", "harbor oci helm", "Harbor OCI helm charts registry", helmURL, "HELM"},
		{"intel-github-io", "intel-github-io", "Intel Public registry with device operator & plugins", "https://intel.github.io/helm-charts", "HELM"},
		{"intel-rs-helm", "intel-rs-helm", "Repo on registry registry-rs.edgeorchestration.intel.com", "oci://rs-proxy.orch-platform.svc.cluster.local:8443", "HELM"},
		{"intel-rs-images", "intel-rs-image", "Repo on registry registry-rs.edgeorchestration.intel.com", "oci://registry-rs.edgeorchestration.intel.com", "IMAGE"},
		{"jetstack", "jetstack", "Public registry for cert manager chart", "https://charts.jetstack.io", "HELM"},
	} {
		regs = append(regs, types.Registry{
			Name:        ra.Name,
			DisplayName: ra.DisplayName,
			Description: ra.Description,
			RootURL:     ra.RootURL,
			Type:        ra.Type,
		})
	}

	return regs
}

func MakeHTTPRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s%s", s.catalogClient.CatalogRESTServerUrl, endpoint)
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	auth.AddRestAuthHeader(req, s.catalogClient.Token, s.catalogClient.ProjectID)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	return res, nil
}

func MakeHTTPRequestWithQuery(method, endpoint string, queryParams map[string]string) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s%s", s.catalogClient.CatalogRESTServerUrl, endpoint)
	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	auth.AddRestAuthHeader(req, s.catalogClient.Token, s.catalogClient.ProjectID)
	query := req.URL.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	return res, nil
}
