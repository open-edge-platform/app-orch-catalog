// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package methods

import (
	"bytes"
	"context"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient/utilities"

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
	UtilitiesClient      *utilities.ClientWithResponses
	CatalogRESTServerUrl string
	Token                string
	ProjectID            string
}

// Client Creation Functions
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

func createUtilitiesClient(restServerURL, token, projectID string) (*utilities.ClientWithResponses, error) {
	utilitiesClient, err := utilities.NewClientWithResponses(restServerURL, utilities.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		auth.AddRestAuthHeader(req, token, projectID)
		return nil
	}))
	if err != nil {
		return nil, err
	}

	return utilitiesClient, err
}

func NewCatalogClient(catalogRESTServerUrl, token, projectID, orchDomain string) *CatalogClient {
	client, err := createCatalogClient(catalogRESTServerUrl, token, projectID)
	if err != nil {
		fmt.Printf("Failed to create catalog client: %v\n", err)
		return nil
	}
	utilitiesClient, err := createUtilitiesClient(catalogRESTServerUrl, token, projectID)
	if err != nil {
		fmt.Printf("Failed to create utilities client: %v\n", err)
		return nil
	}
	return &CatalogClient{
		OrchDomain:           orchDomain,
		Client:               client,
		UtilitiesClient:      utilitiesClient,
		Token:                token,
		ProjectID:            projectID,
		CatalogRESTServerUrl: catalogRESTServerUrl,
	}
}

// Application-related Functions
func (c *CatalogClient) GetApplicationList(ctx context.Context) ([]restClient.Application, int, error) {
	resp, err := c.Client.CatalogServiceListApplicationsWithResponse(ctx, &restClient.CatalogServiceListApplicationsParams{})
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to list applications: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to list applications: response is nil")
	}

	return resp.JSON200.Applications, resp.StatusCode(), nil
}

func (c *CatalogClient) GetApplication(ctx context.Context, name, version string) (*restClient.Application, int, error) {
	resp, err := c.Client.CatalogServiceGetApplicationWithResponse(ctx, name, version)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to get application: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to get application: response is nil")
	}
	return &resp.JSON200.Application, resp.StatusCode(), nil
}

func (c *CatalogClient) GetApplicationVersions(ctx context.Context, name string) ([]restClient.Application, int, error) {
	resp, err := c.Client.CatalogServiceGetApplicationVersionsWithResponse(ctx, name)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to get application: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to get application: response is nil")
	}
	return resp.JSON200.Application, resp.StatusCode(), nil
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

// Deployment Package-related Functions
func (c *CatalogClient) GetDeploymentPackage(ctx context.Context, name, version string) (*restClient.DeploymentPackage, int, error) {
	resp, err := c.Client.CatalogServiceGetDeploymentPackageWithResponse(ctx, name, version)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to get deployment package: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to get deployment package: response is nil")
	}

	return &resp.JSON200.DeploymentPackage, resp.StatusCode(), nil
}

func (c *CatalogClient) GetDeploymentPackageVersions(ctx context.Context, name string) ([]restClient.DeploymentPackage, int, error) {
	resp, err := c.Client.CatalogServiceGetDeploymentPackageVersionsWithResponse(ctx, name)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to get deployment package: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to get deployment package: response is nil")
	}

	return resp.JSON200.DeploymentPackages, resp.StatusCode(), nil
}

func (c *CatalogClient) ListDeploymentPackages(ctx context.Context, params *restClient.CatalogServiceListDeploymentPackagesParams) ([]restClient.DeploymentPackage, int, error) {
	resp, err := c.Client.CatalogServiceListDeploymentPackagesWithResponse(ctx, params)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to list deployment packages: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to list deployment packages: response is nil")
	}

	return resp.JSON200.DeploymentPackages, resp.StatusCode(), nil
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

func (c *CatalogClient) ExportDeploymentPackage(ctx context.Context, name string, version string) (*utilities.CatalogServiceDownloadDeploymentPackageResponse, int, error) {
	resp, err := c.UtilitiesClient.CatalogServiceDownloadDeploymentPackageWithResponse(ctx, name, version)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to export deployment package: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to export deployment package: response is nil")
	}

	return resp, resp.StatusCode(), nil
}

// Registry-related Functions
func (c *CatalogClient) GetRegistry(ctx context.Context, name string) (*restClient.Registry, int, error) {
	resp, err := c.Client.CatalogServiceGetRegistryWithResponse(ctx, name, &restClient.CatalogServiceGetRegistryParams{})
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to get registry: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to get registry: response is nil")
	}

	return &resp.JSON200.Registry, resp.StatusCode(), nil
}

func (c *CatalogClient) ListRegistries(ctx context.Context, params *restClient.CatalogServiceListRegistriesParams) ([]restClient.Registry, int, error) {
	resp, err := c.Client.CatalogServiceListRegistriesWithResponse(ctx, params)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to list registries: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to list registries: response is nil")
	}

	return resp.JSON200.Registries, resp.StatusCode(), nil
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

func (c *CatalogClient) GetRegistries() []restClient.Registry {
	return types.GetRegistryDefinitions(c.OrchDomain)
}

// createMultipartBody creates a multipart form body from a file path
// Returns io.Reader (request body), content-type string, and error
func createMultipartBody(pathName string) (io.Reader, string, error) {
	file, err := os.Open(pathName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file %s: %w", pathName, err)
	}
	defer file.Close()

	filename := filepath.Base(pathName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create form file for %s: %w", filename, err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, "", fmt.Errorf("failed to copy file content for %s: %w", filename, err)
	}

	if err = writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

// Import/Upload-related Functions
func (c *CatalogClient) UploadTarball(ctx context.Context, pathName string) (*utilities.CatalogServiceBulkCatalogUploadResponse, int, error) {
	body, contentType, err := createMultipartBody(pathName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create multipart body: %w", err)
	}

	resp, err := c.UtilitiesClient.CatalogServiceBulkCatalogUploadWithBodyWithResponse(ctx, contentType, body)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return resp, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return resp, resp.StatusCode(), fmt.Errorf("failed to upload tarball: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to upload tarball: response is nil")
	}

	return resp, resp.StatusCode(), nil
}

func (c *CatalogClient) GetCharts(ctx context.Context, params *utilities.CatalogServiceGetRegistryChartsParams) (*utilities.CatalogServiceGetRegistryChartsResponse, int, error) {
	resp, err := c.UtilitiesClient.CatalogServiceGetRegistryChartsWithResponse(ctx, params)
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return resp, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return resp, resp.StatusCode(), fmt.Errorf("failed to get charts: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to get charts: response is nil")
	}

	return resp, resp.StatusCode(), nil
}

func (c *CatalogClient) ImportHelmChart(ctx context.Context, importRequest *restClient.CatalogServiceImportParams) (int, string, error) {
	params := url.Values{}
	params.Add("url", *importRequest.Url)
	res, err := c.Client.CatalogServiceImport(ctx, importRequest)
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
