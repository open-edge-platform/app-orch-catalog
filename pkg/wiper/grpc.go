// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package wiper

import (
	"context"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"google.golang.org/grpc/metadata"
)

var log = dazl.GetPackageLogger()

// ProjectWiper is a tool for wiping data associated with a project
type ProjectWiper interface {
	Wipe(ctx context.Context, projectUUID string) []error
}

type grpcWiper struct {
	client catalogv3.CatalogServiceClient
}

// NewGRPCWiper creates a project wiper that uses gRPC to wipe data
func NewGRPCWiper(client catalogv3.CatalogServiceClient) ProjectWiper {
	return &grpcWiper{client: client}
}

// Annotates a copy of the given context with active project ID
func withActiveProjectID(ctx context.Context, projectUUID string) context.Context {
	if projectUUID == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "activeprojectid", projectUUID)
}

// Wipe deletes all entities (packages, apps, registries, and artifacts) for the given project.
func (w *grpcWiper) Wipe(ctx context.Context, projectUUID string) []error {
	var errors []error
	pctx := withActiveProjectID(ctx, projectUUID)

	errors = append(errors, w.preparePackagesForDeletion(pctx)...)

	errors = append(errors, w.wipePackages(pctx)...)
	errors = append(errors, w.wipeApplications(pctx)...)
	errors = append(errors, w.wipeArtifacts(pctx)...)
	errors = append(errors, w.wipeRegistries(pctx)...)
	return errors
}

// Sweeps through all packages, marking them as not deployed
func (w *grpcWiper) preparePackagesForDeletion(ctx context.Context) []error {
	var errors []error
	hasMorePages := true
	offset := int32(0)
	for hasMorePages {
		resp, err := w.client.ListDeploymentPackages(ctx, &catalogv3.ListDeploymentPackagesRequest{
			PageSize: maxPageSize,
			Offset:   offset,
		})
		if err != nil {
			return append(errors, err)
		}
		for _, app := range resp.DeploymentPackages {
			if err = w.preparePackageForDeletion(ctx, app.Name, app.Version); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.TotalElements > offset+int32(len(resp.DeploymentPackages))
		offset = offset + int32(len(resp.DeploymentPackages))
		log.Debugf("Total elements: %d, offset: %d, len: %d hasMorePages: %t", resp.TotalElements, offset, len(resp.DeploymentPackages), hasMorePages)
	}
	return errors
}

func (w *grpcWiper) preparePackageForDeletion(ctx context.Context, name string, version string) error {
	gresp, err := w.client.GetDeploymentPackage(ctx, &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: name, Version: version,
	})
	if err != nil {
		return err
	}

	// Update package to sever any dependencies from its point of view
	pkg := gresp.DeploymentPackage
	pkg.IsDeployed = false
	pkg.ApplicationReferences = []*catalogv3.ApplicationReference{}
	pkg.ApplicationDependencies = []*catalogv3.ApplicationDependency{}
	pkg.DefaultNamespaces = map[string]string{}
	pkg.DefaultProfileName = ""
	pkg.Profiles = []*catalogv3.DeploymentProfile{}

	if _, err = w.client.UpdateDeploymentPackage(ctx, &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: name, Version: version, DeploymentPackage: pkg,
	}); err != nil {
		return err
	}
	return nil
}

func (w *grpcWiper) wipePackages(ctx context.Context) []error {
	var errors []error
	hasMorePages := true
	for hasMorePages {
		resp, err := w.client.ListDeploymentPackages(ctx, &catalogv3.ListDeploymentPackagesRequest{PageSize: maxPageSize})
		if err != nil {
			return append(errors, err)
		}
		for _, pkg := range resp.DeploymentPackages {
			log.Debugf("Deleting deployment package %s:%s", pkg.Name, pkg.Version)
			if _, err = w.client.DeleteDeploymentPackage(ctx, &catalogv3.DeleteDeploymentPackageRequest{
				DeploymentPackageName: pkg.Name,
				Version:               pkg.Version,
			}); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.TotalElements > int32(len(resp.DeploymentPackages))
		log.Debugf("Total elements: %d, len: %d hasMorePages: %t", resp.TotalElements, len(resp.DeploymentPackages), hasMorePages)
	}
	return errors
}

func (w *grpcWiper) wipeApplications(ctx context.Context) []error {
	var errors []error
	hasMorePages := true
	for hasMorePages {
		resp, err := w.client.ListApplications(ctx, &catalogv3.ListApplicationsRequest{PageSize: maxPageSize})
		if err != nil {
			return append(errors, err)
		}
		for _, app := range resp.Applications {
			log.Debugf("Deleting application %s:%s", app.Name, app.Version)
			if _, err = w.client.DeleteApplication(ctx, &catalogv3.DeleteApplicationRequest{
				ApplicationName: app.Name,
				Version:         app.Version,
			}); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.TotalElements > int32(len(resp.Applications))
		log.Debugf("Total elements: %d, len: %d hasMorePages: %t", resp.TotalElements, len(resp.Applications), hasMorePages)
	}
	return errors
}

func (w *grpcWiper) wipeArtifacts(ctx context.Context) []error {
	var errors []error
	hasMorePages := true
	for hasMorePages {
		resp, err := w.client.ListArtifacts(ctx, &catalogv3.ListArtifactsRequest{PageSize: maxPageSize})
		if err != nil {
			return append(errors, err)
		}
		for _, artifact := range resp.Artifacts {
			log.Debugf("Deleting artifact %s", artifact.Name)
			if _, err = w.client.DeleteArtifact(ctx, &catalogv3.DeleteArtifactRequest{
				ArtifactName: artifact.Name,
			}); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.TotalElements > int32(len(resp.Artifacts))
		log.Debugf("Total elements: %d, len: %d hasMorePages: %t", resp.TotalElements, len(resp.Artifacts), hasMorePages)
	}
	return errors
}

func (w *grpcWiper) wipeRegistries(ctx context.Context) []error {
	var errors []error
	hasMorePages := true
	for hasMorePages {
		resp, err := w.client.ListRegistries(ctx, &catalogv3.ListRegistriesRequest{PageSize: maxPageSize})
		if err != nil {
			return append(errors, err)
		}
		for _, registry := range resp.Registries {
			log.Debugf("Deleting registry %s", registry.Name)
			if _, err = w.client.DeleteRegistry(ctx, &catalogv3.DeleteRegistryRequest{
				RegistryName: registry.Name,
			}); err != nil {
				errors = append(errors, err)
			}
		}
		hasMorePages = resp.TotalElements > int32(len(resp.Registries))
		log.Debugf("Total elements: %d, len: %d hasMorePages: %t", resp.TotalElements, len(resp.Registries), hasMorePages)
	}
	return errors
}
