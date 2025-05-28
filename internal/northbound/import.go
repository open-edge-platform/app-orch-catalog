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

	"github.com/open-edge-platform/app-orch-catalog/internal/helm"
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

	if err = g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	helm, err := helm.FetchHelmChartOCI(req.Url, req.Username, req.AuthToken)
	if err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithMessage(err.Error()))
	}

	_ = helm

	_ = projectUUID
	resp := &catalogv3.ImportResponse{}

	return resp, nil
}
