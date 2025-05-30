// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

/* The Catalog Service supports Uploads via the Upload endpoint.
 *
 * An upload consists of a set of files. These may either be individual YAML files or
 * they may be tarballs that contain individual YAML files. If a tarball is provided, then
 * it will be extracted and the individual YAML files will be processed.
 *
 * Each tarball is loaded independently, and the tarballs are loaded independently of raw
 * yaml files that may be present.
 */

import (
	"context"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/registry"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
)

func (g *Server) createOrUpdateRegistry(ctx context.Context, tx *generated.Tx, projectUUID string, reg *catalogv3.Registry, registryEvents *RegistryEvents) error {
	_, err := tx.Registry.Query().Where(registry.ProjectUUID(projectUUID), registry.Name(reg.Name)).First(ctx)
	if err != nil {
		_, err = g.createRegistry(ctx, tx, projectUUID, reg, registryEvents)
	} else {
		err = g.updateRegistry(ctx, tx, projectUUID, reg, registryEvents)
	}
	return err
}

func (g *Server) createOrUpdateApplication(ctx context.Context, tx *generated.Tx, projectUUID string, app *catalogv3.Application, appEvents *ApplicationEvents) error {
	_, err := tx.Application.Query().Where(application.ProjectUUID(projectUUID), application.Name(app.Name), application.Version(app.Version)).First(ctx)
	if err != nil {
		_, err = g.createApplication(ctx, tx, projectUUID, app, appEvents)
	} else {
		err = g.updateApplication(ctx, tx, projectUUID, app, appEvents)
	}
	return err
}

func (g *Server) createOrUpdateDeploymentPackage(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage, dpEvents *DeploymentPackageEvents) error {
	_, err := tx.DeploymentPackage.Query().Where(deploymentpackage.ProjectUUID(projectUUID),
		deploymentpackage.Name(pkg.Name), deploymentpackage.Version(pkg.Version)).First(ctx)
	if err != nil {
		_, err = g.createDeploymentPackage(ctx, tx, projectUUID, pkg, dpEvents)
	} else {
		err = g.updateDeploymentPackage(ctx, tx, projectUUID, pkg, dpEvents)
	}
	return err
}
