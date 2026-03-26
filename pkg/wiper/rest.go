// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package wiper

import (
	"context"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"net/http"
)

type restWiper struct {
	client     restClient.ClientWithResponses
	reqEditors []restClient.RequestEditorFn
}

// NewRESTWiper creates a project wiper that uses REST API to wipe data
func NewRESTWiper(client restClient.ClientWithResponses, reqEditors ...restClient.RequestEditorFn) ProjectWiper {
	return &restWiper{client: client, reqEditors: reqEditors}
}

// Wipe deletes all entities (packages, apps, registries, and artifacts) for the given project.
func (w *restWiper) Wipe(ctx context.Context, projectUUID string) []error {
	var errors []error
	pctx := withActiveProjectID(ctx, projectUUID)

	errors = append(errors, w.preparePackagesForDeletion(pctx, projectUUID)...)
	errors = append(errors, w.prepareApplicationsForDeletion(pctx, projectUUID)...)

	errors = append(errors, w.wipePackages(pctx, projectUUID)...)
	errors = append(errors, w.wipeApplications(pctx, projectUUID)...)
	errors = append(errors, w.wipeArtifacts(pctx, projectUUID)...)
	errors = append(errors, w.wipeRegistries(pctx, projectUUID)...)
	return errors
}

var (
	maxPageSize = int32(northbound.MaxPageSize)
	notDeployed = false
)

// Sweeps through all packages, marking them as not deployed
func (w *restWiper) preparePackagesForDeletion(ctx context.Context, projectID string) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListDeploymentPackagesWithResponse(ctx, projectID, &restClient.CatalogServiceListDeploymentPackagesParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, pkg := range resp.JSON200.DeploymentPackages {
		if err = w.preparePackageForDeletion(ctx, projectID, pkg.Name, pkg.Version); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) preparePackageForDeletion(ctx context.Context, projectID string, name string, version string) error {
	gresp, err := w.client.CatalogServiceGetDeploymentPackageWithResponse(ctx, projectID, name, version, w.reqEditors...)
	if err != nil {
		return err
	}
	if gresp.StatusCode() != http.StatusOK {
		return nil
	}

	// Update package to sever any dependencies from its point of view
	pkg := gresp.JSON200.DeploymentPackage
	pkg.IsDeployed = &notDeployed
	pkg.Profiles = nil
	pkg.ApplicationReferences = nil
	pkg.ApplicationDependencies = nil
	pkg.DefaultNamespaces = nil
	pkg.DefaultProfileName = nil

	if _, err = w.client.CatalogServiceUpdateDeploymentPackageWithResponse(ctx, projectID, name, version, pkg, w.reqEditors...); err != nil {
		return err
	}
	return nil
}

// Sweeps through all applications, severing their dependencies on any deployment packages
func (w *restWiper) prepareApplicationsForDeletion(ctx context.Context, projectID string) []error {
	var errors []error
	offset := int32(0)
	hasMorePages := true
	for hasMorePages {
		resp, err := w.client.CatalogServiceListApplicationsWithResponse(ctx, projectID, &restClient.CatalogServiceListApplicationsParams{PageSize: &maxPageSize, Offset: &offset}, w.reqEditors...)
		if err != nil {
			return append(errors, err)
		}
		if resp == nil || resp.StatusCode() != http.StatusOK {
			return nil
		}
		for _, app := range resp.JSON200.Applications {
			if err = w.prepareApplicationForDeletion(ctx, projectID, app.Name, app.Version); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.JSON200.TotalElements > offset+int32(len(resp.JSON200.Applications))
		offset = offset + int32(len(resp.JSON200.Applications))
	}
	return errors
}

func (w *restWiper) prepareApplicationForDeletion(ctx context.Context, projectID string, name string, version string) error {
	gresp, err := w.client.CatalogServiceGetApplicationWithResponse(ctx, projectID, name, version, w.reqEditors...)
	if err != nil {
		return err
	}

	// Update app to remove any profiles that might have dependencies on packages
	if gresp.StatusCode() != http.StatusOK {
		return nil
	}

	app := gresp.JSON200.Application
	app.Profiles = nil
	app.DefaultProfileName = nil

	if _, err = w.client.CatalogServiceUpdateApplicationWithResponse(ctx, projectID, name, version, app, w.reqEditors...); err != nil {
		return err
	}
	return nil
}

func (w *restWiper) wipePackages(ctx context.Context, projectID string) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListDeploymentPackagesWithResponse(ctx, projectID, &restClient.CatalogServiceListDeploymentPackagesParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, pkg := range resp.JSON200.DeploymentPackages {
		if _, err = w.client.CatalogServiceDeleteDeploymentPackageWithResponse(ctx, projectID, pkg.Name, pkg.Version, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) wipeApplications(ctx context.Context, projectID string) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListApplicationsWithResponse(ctx, projectID, &restClient.CatalogServiceListApplicationsParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, app := range resp.JSON200.Applications {
		if _, err = w.client.CatalogServiceDeleteApplicationWithResponse(ctx, projectID, app.Name, app.Version, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) wipeArtifacts(ctx context.Context, projectID string) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListArtifactsWithResponse(ctx, projectID, &restClient.CatalogServiceListArtifactsParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, artifact := range resp.JSON200.Artifacts {
		if _, err = w.client.CatalogServiceDeleteArtifactWithResponse(ctx, projectID, artifact.Name, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) wipeRegistries(ctx context.Context, projectID string) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListRegistriesWithResponse(ctx, projectID, &restClient.CatalogServiceListRegistriesParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, registry := range resp.JSON200.Registries {
		if _, err = w.client.CatalogServiceDeleteRegistryWithResponse(ctx, projectID, registry.Name, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
