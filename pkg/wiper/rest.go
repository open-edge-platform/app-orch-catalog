// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package wiper

import (
	"context"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound"
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"net/http"
)

type restWiper struct {
	client     restapi.ClientWithResponses
	reqEditors []restapi.RequestEditorFn
}

// NewRESTWiper creates a project wiper that uses REST API to wipe data
func NewRESTWiper(client restapi.ClientWithResponses, reqEditors ...restapi.RequestEditorFn) ProjectWiper {
	return &restWiper{client: client, reqEditors: reqEditors}
}

// Wipe deletes all entities (packages, apps, registries, and artifacts) for the given project.
func (w *restWiper) Wipe(ctx context.Context, projectUUID string) []error {
	var errors []error
	pctx := withActiveProjectID(ctx, projectUUID)

	errors = append(errors, w.preparePackagesForDeletion(pctx)...)
	errors = append(errors, w.prepareApplicationsForDeletion(pctx)...)

	errors = append(errors, w.wipePackages(pctx)...)
	errors = append(errors, w.wipeApplications(pctx)...)
	errors = append(errors, w.wipeArtifacts(pctx)...)
	errors = append(errors, w.wipeRegistries(pctx)...)
	return errors
}

var (
	maxPageSize = int32(northbound.MaxPageSize)
	notDeployed = false
)

// Sweeps through all packages, marking them as not deployed
func (w *restWiper) preparePackagesForDeletion(ctx context.Context) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListDeploymentPackagesWithResponse(ctx, &restapi.CatalogServiceListDeploymentPackagesParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, pkg := range resp.JSON200.DeploymentPackages {
		if err = w.preparePackageForDeletion(ctx, pkg.Name, pkg.Version); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) preparePackageForDeletion(ctx context.Context, name string, version string) error {
	gresp, err := w.client.CatalogServiceGetDeploymentPackageWithResponse(ctx, name, version, w.reqEditors...)
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

	if _, err = w.client.CatalogServiceUpdateDeploymentPackageWithResponse(ctx, name, version, pkg, w.reqEditors...); err != nil {
		return err
	}
	return nil
}

// Sweeps through all applications, severing their dependencies on any deployment packages
func (w *restWiper) prepareApplicationsForDeletion(ctx context.Context) []error {
	var errors []error
	offset := int32(0)
	hasMorePages := true
	for hasMorePages {
		resp, err := w.client.CatalogServiceListApplicationsWithResponse(ctx, &restapi.CatalogServiceListApplicationsParams{PageSize: &maxPageSize, Offset: &offset}, w.reqEditors...)
		if resp.StatusCode() != http.StatusOK {
			return nil
		}

		if err != nil {
			return append(errors, err)
		}
		for _, app := range resp.JSON200.Applications {
			if err = w.prepareApplicationForDeletion(ctx, app.Name, app.Version); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.JSON200.TotalElements > offset+int32(len(resp.JSON200.Applications))
		offset = offset + int32(len(resp.JSON200.Applications))
	}
	return errors
}

func (w *restWiper) prepareApplicationForDeletion(ctx context.Context, name string, version string) error {
	gresp, err := w.client.CatalogServiceGetApplicationWithResponse(ctx, name, version, w.reqEditors...)
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

	if _, err = w.client.CatalogServiceUpdateApplicationWithResponse(ctx, name, version, app, w.reqEditors...); err != nil {
		return err
	}
	return nil
}

func (w *restWiper) wipePackages(ctx context.Context) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListDeploymentPackagesWithResponse(ctx, &restapi.CatalogServiceListDeploymentPackagesParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, pkg := range resp.JSON200.DeploymentPackages {
		if _, err = w.client.CatalogServiceDeleteDeploymentPackageWithResponse(ctx, pkg.Name, pkg.Version, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) wipeApplications(ctx context.Context) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListApplicationsWithResponse(ctx, &restapi.CatalogServiceListApplicationsParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, app := range resp.JSON200.Applications {
		if _, err = w.client.CatalogServiceDeleteApplicationWithResponse(ctx, app.Name, app.Version, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) wipeArtifacts(ctx context.Context) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListArtifactsWithResponse(ctx, &restapi.CatalogServiceListArtifactsParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, artifact := range resp.JSON200.Artifacts {
		if _, err = w.client.CatalogServiceDeleteArtifactWithResponse(ctx, artifact.Name, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

func (w *restWiper) wipeRegistries(ctx context.Context) []error {
	var errors []error
	resp, err := w.client.CatalogServiceListRegistriesWithResponse(ctx, &restapi.CatalogServiceListRegistriesParams{PageSize: &maxPageSize}, w.reqEditors...)
	if err != nil {
		return append(errors, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil
	}

	for _, registry := range resp.JSON200.Registries {
		if _, err = w.client.CatalogServiceDeleteRegistryWithResponse(ctx, registry.Name, w.reqEditors...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
