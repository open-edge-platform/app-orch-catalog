package methods

import (
	"context"
	"fmt"
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"net/http"
)

func (c *CatalogClient) CreateApplication(ctx context.Context, application *restapi.Application) (*restapi.Application, int, error) {
	resp, err := c.Client.CatalogServiceCreateApplicationWithResponse(ctx, restapi.CatalogServiceCreateApplicationJSONRequestBody{
		ChartName:          application.ChartName,
		ChartVersion:       application.ChartVersion,
		DefaultProfileName: application.DefaultProfileName,
		Description:        application.Description,
		DisplayName:        application.DisplayName,
		HelmRegistryName:   application.HelmRegistryName,
		Name:               application.Name,
		Profiles:           application.Profiles,
		Version:            application.Version,
		IgnoredResources:   application.IgnoredResources,
		Kind:               application.Kind,
	})
	if err != nil || resp == nil || resp.StatusCode() != 200 {
		if err != nil {
			if resp != nil {
				return nil, resp.StatusCode(), fmt.Errorf("%v", err)
			}
			return nil, 0, fmt.Errorf("%v", err)
		}
		if resp != nil {
			return nil, resp.StatusCode(), fmt.Errorf("failed to create application: %v", string(resp.Body))
		}
		return nil, 0, fmt.Errorf("failed to create application: response is nil")
	}

	return &resp.JSON200.Application, resp.StatusCode(), nil
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
