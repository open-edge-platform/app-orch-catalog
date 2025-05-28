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

	"github.com/open-edge-platform/app-orch-catalog/internal/dp"
	"github.com/open-edge-platform/app-orch-catalog/internal/helm"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	nberrors "github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
)

// UploadCatalogEntities allows upload of YAML files containing catalog entity descriptions or TAR file containing such YAML files through gRPC
func (g *Server) Import(ctx context.Context, req *catalogv3.ImportRequest) (*catalogv3.ImportResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.Url == "" {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithMessage("incomplete request"))
	}

	//if err = g.authCheckAllowed(ctx, req); err != nil {
	//	return nil, err
	//}

	helm, err := helm.FetchHelmChartOCI(req.Url, req.Username, req.AuthToken)
	if err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithMessage(err.Error()))
	}

	_, pkg, app, reg, err := dp.GenerateDeploymentPackageResources(helm, req.ChartValues, req.Namespace, req.IncludeAuth)

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	registryEvents := &RegistryEvents{}
	_, err = g.createRegistry(ctx, tx, projectUUID, reg, registryEvents)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	appEvents := &ApplicationEvents{}
	_, err = g.createApplication(ctx, tx, projectUUID, app, appEvents)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	dpEvents := &DeploymentPackageEvents{}
	_, err = g.createDeploymentPackage(ctx, tx, projectUUID, pkg, dpEvents)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "created", "registry", projectUUID, reg.Name)
	logActivity(ctx, "created", "application", projectUUID, app.Name, app.Version)
	logActivity(ctx, "created", "deployment-package", projectUUID, pkg.Name, pkg.Version)

	registryEvents.sendToAll(g.listeners)
	appEvents.sendToAll(g.listeners)
	dpEvents.sendToAll(g.listeners)

	resp := &catalogv3.ImportResponse{}
	return resp, nil
}
