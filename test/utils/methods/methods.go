// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package methods

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
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
	OrchDomain           string
	Client               *restClient.ClientWithResponses
	CatalogRESTServerUrl string
	Token                string
	ProjectID            string
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

func createCatalogClient(restServerURL, token, projectID string) (*restClient.ClientWithResponses, error) {
	catalogClient, err := restClient.NewClientWithResponses(restServerURL, restClient.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		auth.AddRestAuthHeader(req, token, projectID)
		return nil
	}))
	if err != nil {
		return nil, err
	}

	return catalogClient, err
}

func NewCatalogClient(catalogRESTServerUrl, token, projectID, orchDomain string) *CatalogClient {
	client, err := createCatalogClient(catalogRESTServerUrl, token, projectID)
	if err != nil {
		fmt.Printf("Failed to create catalog client: %v\n", err)
		return nil
	}
	return &CatalogClient{
		OrchDomain: orchDomain,
		Client:     client,
		Token:      token,
		ProjectID:  projectID,
	}
}

func (c *CatalogClient) GetApplication(ctx context.Context, name, version string, mustExist bool) (*types.Application, error) {
	res, err := c.Client.CatalogServiceGetApplication(ctx, name, version)
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

func (c *CatalogClient) DeleteApplication(ctx context.Context, name, version string, mustExist bool) error {
	res, err := c.Client.CatalogServiceDeleteApplication(ctx, name, version)
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

func (c *CatalogClient) GetDeploymentPackage(ctx context.Context, name, version string, mustExist bool) (*types.DeploymentPackage, error) {
	res, err := c.Client.CatalogServiceGetDeploymentPackage(ctx, name, version)
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

func (c *CatalogClient) DeleteDeploymentPackage(ctx context.Context, name, version string, mustExist bool) error {
	res, err := c.Client.CatalogServiceDeleteDeploymentPackage(ctx, name, version)
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

func (c *CatalogClient) GetRegistry(ctx context.Context, name string, mustExist bool) (*types.Registry, error) {
	res, err := c.Client.CatalogServiceGetRegistry(ctx, name, &restClient.CatalogServiceGetRegistryParams{})
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

func (c *CatalogClient) DeleteRegistry(ctx context.Context, name string, mustExist bool) error {
	res, err := c.Client.CatalogServiceDeleteRegistry(ctx, name)
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

func (c *CatalogClient) UploadTarball(ctx context.Context, pathName string) (*http.Response, error) {
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

func (c *CatalogClient) ImportHelmChart(ctx context.Context, importRequest *ImportRequest) (int, string, error) {
	params := url.Values{}
	params.Add("url", importRequest.URL)
	res, err := c.Client.CatalogServiceImport(ctx, &restClient.CatalogServiceImportParams{
		Url:                       &importRequest.URL,
		Username:                  &importRequest.Username,
		AuthToken:                 &importRequest.AuthToken,
		ChartValues:               &importRequest.ChartValues,
		IncludeAuth:               &importRequest.IncludeAuth,
		GenerateDefaultValues:     &importRequest.GenerateDefaultValues,
		GenerateDefaultParameters: &importRequest.GenerateDefaultParameters,
		Namespace:                 &importRequest.Namespace,
	})
	if err != nil {
		return 0, "", fmt.Errorf("failed to import Helm chart: %w", err)
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

// MakeAuthenticatedRequest creates and sends an HTTP request with auth headers
func (c *CatalogClient) MakeAuthenticatedRequest(method, endpoint string, requestBody io.Reader, queryParams map[string]string, headers ...map[string]string) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s%s", c.CatalogRESTServerUrl, endpoint)
	req, err := http.NewRequest(method, requestURL, requestBody)
	if err != nil {
		return nil, err
	}

	auth.AddRestAuthHeader(req, c.Token, c.ProjectID)

	// Add custom headers if provided
	if len(headers) > 0 && headers[0] != nil {
		for key, value := range headers[0] {
			req.Header.Set(key, value)
		}
	}

	// Add query parameters if provided
	if len(queryParams) > 0 {
		query := req.URL.Query()
		for key, value := range queryParams {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
	}

	return http.DefaultClient.Do(req)
}
