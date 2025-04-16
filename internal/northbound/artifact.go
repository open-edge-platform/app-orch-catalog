// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/artifact"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/predicate"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
	"image/jpeg"
	"image/png"
	"strings"
)

// ggg


// CreateArtifact creates an artifact through gRPC
func (g *Server) CreateArtifact(ctx context.Context, req *catalogv3.CreateArtifactRequest) (*catalogv3.CreateArtifactResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Artifact == nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("incomplete request"))
	} else if err := req.Artifact.Validate(); err != nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithResourceName(req.Artifact.Name),
			errors.WithMessage(err.Error()))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &ArtifactEvents{}
	created, err := g.createArtifact(ctx, tx, projectUUID, req.Artifact, events)
	if err != nil {
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "created", "artifact", projectUUID, req.Artifact.Name)
	events.sendToAll(g.listeners)

	return &catalogv3.CreateArtifactResponse{
		Artifact: &catalogv3.Artifact{
			Name:        created.Name,
			DisplayName: created.DisplayName,
			Description: created.Description,
			MimeType:    created.MimeType,
			Artifact:    created.Artifact,
			CreateTime:  timestamppb.New(created.CreateTime),
		},
	}, nil
}

func (g *Server) createArtifact(ctx context.Context, tx *generated.Tx, projectUUID string, art *catalogv3.Artifact, events *ArtifactEvents) (*generated.Artifact, error) {
	displayName, ok := validateDisplayName(art.Name, art.DisplayName)
	if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}
	art.DisplayName = displayName

	if err := validateArtifactData(art.Name, art.MimeType, art.Artifact); err != nil {
		return nil, err
	}

	if art.Artifact != nil {
		// Note: Artifact Size is already limited by protobuf validator
		if malware.DefaultScanner != nil {
			okay, res, err := malware.DefaultScanner.ScanBytes(art.Artifact)
			if err != nil {
				if malware.DefaultScanner.IsPermissive() {
					log.Warn("Malware scanner is not available. Skipping scan due to permissive mode.")
				} else {
					log.Warn("Malware scanner returned error %s", err)
					return nil, errors.NewUnavailable(
						errors.WithResourceType(errors.ArtifactType),
						errors.WithMessage("malware scanner configured but not available"))
				}
			} else if !okay {
				return nil, errors.NewInvalidArgument(
					errors.WithResourceType(errors.ArtifactType),
					errors.WithMessage("malware detected: %s", res))
			}
		}
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkArtifactUniqueness(ctx, tx, projectUUID, art); err != nil {
		return nil, err
	}

	created, err := tx.Artifact.Create().
		SetProjectUUID(projectUUID).
		SetName(art.Name).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(art.Description).
		SetMimeType(art.MimeType).
		SetArtifact(art.Artifact).
		Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(art.Name),
				errors.WithMessage("artifact %s already exists", art.Name))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}
	events.append(CreatedEvent, projectUUID, art)
	return created, nil
}

// Returns an error if the artifact display name is not unique
func (g *Server) checkArtifactUniqueness(ctx context.Context, tx *generated.Tx, projectUUID string, a *catalogv3.Artifact) error {
	if len(a.DisplayName) > 0 {
		existing, err := tx.Artifact.Query().
			Where(
				artifact.ProjectUUID(projectUUID),
				artifact.DisplayNameLc(strings.ToLower(a.DisplayName)),
				artifact.Not(artifact.Name(a.Name))).
			Count(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
		if existing > 0 {
			return errors.NewAlreadyExists(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(a.Name),
				errors.WithMessage("display name %s is not unique", a.DisplayName))
		}
	}
	return nil
}

// ListArtifacts gets a list of all artifacts through gRPC
func (g *Server) ListArtifacts(ctx context.Context, req *catalogv3.ListArtifactsRequest) (*catalogv3.ListArtifactsResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	orderBys, err := parseOrderBy(req.OrderBy, errors.ArtifactType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	filters, err := parseFilter(req.Filter, errors.ArtifactType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	artifacts, _, totalElements, err := g.getArtifacts(ctx, tx, projectUUID, orderBys, filters, req.PageSize, req.Offset)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	return &catalogv3.ListArtifactsResponse{Artifacts: artifacts, TotalElements: totalElements}, nil
}

var artifactColumns = map[string]string{
	"project_uuid": "",
	"name":         "name",
	"displayName":  "display_name",
	"description":  "description",
	"mimeType":     "mime_type",
	"createTime":   "create_time",
	"updateTime":   "update_time",
}

func (g *Server) getArtifacts(ctx context.Context, tx *generated.Tx,
	projectUUID string, orderBys []*orderBy, filters []*filter,
	pageSize int32, offset int32) ([]*catalogv3.Artifact, []string, int32, error) {
	var err error
	var artifactsDB []*generated.Artifact
	var orderOptions []artifact.OrderOption
	artifactsQuery := tx.Artifact.Query()

	options, err := orderByOptions(orderBys, artifactColumns, errors.ArtifactType)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, pred := range options {
		orderOptions = append(orderOptions, pred)
	}
	artifactsQuery = artifactsQuery.Order(orderOptions...)

	filterPreds, err := filterPredicates(filters, artifactColumns, errors.ArtifactType)
	if err != nil {
		return nil, nil, 0, err
	}
	var artifactPreds []predicate.Artifact
	for _, pred := range filterPreds {
		artifactPreds = append(artifactPreds, pred)
	}
	artifactsQuery = artifactsQuery.Where(artifact.Or(artifactPreds...))

	if projectUUID == AdminProjectID {
		artifactsDB, err = artifactsQuery.All(ctx)
	} else {
		artifactsDB, err = artifactsQuery.Where(artifact.ProjectUUID(projectUUID)).All(ctx)
	}
	if err != nil {
		return nil, nil, 0, errors.NewDBError(errors.WithError(err))
	}

	startIndex, endIndex, totalElements, err := computePageRange(pageSize, offset, len(artifactsDB))
	if err != nil {
		return nil, nil, 0, err
	}

	artifacts := make([]*catalogv3.Artifact, 0, len(artifactsDB))
	projectUUIDs := make([]string, 0, len(artifactsDB))

	if len(artifactsDB) == 0 {
		return []*catalogv3.Artifact{}, []string{}, 0, nil
	}

	for i := startIndex; i <= endIndex; i++ {
		artifactDB := artifactsDB[i]
		artifacts = append(artifacts, &catalogv3.Artifact{
			Name:        artifactDB.Name,
			DisplayName: artifactDB.DisplayName,
			Description: artifactDB.Description,
			MimeType:    artifactDB.MimeType,
			Artifact:    artifactDB.Artifact,
			CreateTime:  timestamppb.New(artifactDB.CreateTime),
			UpdateTime:  timestamppb.New(artifactDB.UpdateTime),
		})
		projectUUIDs = append(projectUUIDs, artifactDB.ProjectUUID)
	}
	return artifacts, projectUUIDs, totalElements, nil
}

// GetArtifact gets a single artifact through gRPC
func (g *Server) GetArtifact(ctx context.Context, req *catalogv3.GetArtifactRequest) (*catalogv3.GetArtifactResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.ArtifactName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	artifactDb, err := tx.Artifact.Query().
		Where(artifact.ProjectUUID(projectUUID), artifact.Name(req.GetArtifactName())).First(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		if generated.IsNotFound(err) {
			return nil, errors.NewNotFound(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(req.ArtifactName))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	return &catalogv3.GetArtifactResponse{
		Artifact: &catalogv3.Artifact{
			Name:        artifactDb.Name,
			DisplayName: artifactDb.DisplayName,
			Description: artifactDb.Description,
			MimeType:    artifactDb.MimeType,
			Artifact:    artifactDb.Artifact,
			CreateTime:  timestamppb.New(artifactDb.CreateTime),
			UpdateTime:  timestamppb.New(artifactDb.UpdateTime),
		},
	}, nil
}

// UpdateArtifact updates an artifact through gRPC
func (g *Server) UpdateArtifact(ctx context.Context, req *catalogv3.UpdateArtifactRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Artifact == nil || req.ArtifactName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("incomplete request"))
	} else if err := req.Artifact.Validate(); err != nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage(err.Error()))
	} else if req.ArtifactName != req.Artifact.Name {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("name cannot be changed %s != %s", req.ArtifactName, req.Artifact.Name))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &ArtifactEvents{}
	if err = g.updateArtifact(ctx, tx, projectUUID, req.Artifact, events); err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "updated", "artifact", projectUUID, req.Artifact.Name)
	events.sendToAll(g.listeners)

	return &emptypb.Empty{}, nil
}

func (g *Server) updateArtifact(ctx context.Context, tx *generated.Tx, projectUUID string, art *catalogv3.Artifact, events *ArtifactEvents) error {
	displayName, ok := validateDisplayName(art.Name, art.DisplayName)
	if !ok {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}
	art.DisplayName = displayName

	if err := validateArtifactData(art.Name, art.MimeType, art.Artifact); err != nil {
		return err
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkArtifactUniqueness(ctx, tx, projectUUID, art); err != nil {
		return err
	}

	updateCount, err := tx.Artifact.Update().
		Where(artifact.ProjectUUID(projectUUID), artifact.Name(art.Name)).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(art.Description).
		SetMimeType(art.MimeType).
		SetArtifact(art.Artifact).
		Save(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return errors.NewDBError(errors.WithError(err))
	} else if updateCount == 0 {
		g.rollbackTransaction(tx)
		return errors.NewNotFound(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithResourceName(art.Name))
	}
	events.append(UpdatedEvent, projectUUID, art)
	return nil
}

// DeleteArtifact deletes an artifact through gRPC
func (g *Server) DeleteArtifact(ctx context.Context, req *catalogv3.DeleteArtifactRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.ArtifactName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &ArtifactEvents{}
	deleteCount, err := tx.Artifact.Delete().
		Where(artifact.ProjectUUID(projectUUID), artifact.Name(req.ArtifactName)).Exec(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(req.ArtifactName),
				errors.WithMessage("%s %s is in use and cannot be deleted", errors.ArtifactType, req.ArtifactName))
		}
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	} else if deleteCount == 0 {
		g.rollbackTransaction(tx)
		return nil, errors.NewNotFound(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithResourceName(req.ArtifactName))
	}
	if _, err = g.checkDeleteResult(ctx, tx, err, fmt.Sprintf("artifact %s", req.ArtifactName), projectUUID); err != nil {
		return nil, err
	}
	events.append(DeletedEvent, projectUUID, &catalogv3.Artifact{Name: req.ArtifactName})
	events.sendToAll(g.listeners)

	return &emptypb.Empty{}, nil
}

func validateArtifactData(name string, mime string, payload []byte) error {
	switch mime {
	case "text/plain":
		if err := validateTextPlain(payload); err != nil {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(name),
				errors.WithMessage("artifact data is not valid Plain Text"))
		}
	case "application/json":
		if err := validateApplicationJSON(payload); err != nil {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(name),
				errors.WithMessage("artifact data is not valid JSON"))
		}
	case "application/yaml":
		if err := validateApplicationYAML(payload); err != nil {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(name),
				errors.WithMessage("artifact data is not valid YAML"))
		}
	case "image/png":
		if err := validateImagePng(payload); err != nil {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(name),
				errors.WithMessage("artifact data is not valid PNG"))
		}
	case "image/jpeg":
		if err := validateImageJpeg(payload); err != nil {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ArtifactType),
				errors.WithResourceName(name),
				errors.WithMessage("artifact data is not valid JPEG"))
		}
	default:
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithResourceName(name),
			errors.WithMessage("artifact contents do not match mime type %s", mime))
	}

	return nil
}

func validateTextPlain(data []byte) error {
	validUtf8 := strings.ToValidUTF8(string(data), "")
	if len(validUtf8) != len(data) {
		return fmt.Errorf("some chars are not valid UTF-8")
	}
	return nil
}

func validateApplicationJSON(data []byte) error {
	if notPlainTextErr := validateTextPlain(data); notPlainTextErr != nil {
		return notPlainTextErr
	}
	var something map[string]any
	return json.Unmarshal(data, &something)
}

func validateApplicationYAML(data []byte) error {
	if notPlainTextErr := validateTextPlain(data); notPlainTextErr != nil {
		return notPlainTextErr
	}
	var something map[string]any
	return yaml.Unmarshal(data, &something)
}

func validateImagePng(data []byte) error {
	_, err := png.Decode(bytes.NewReader(data))
	return err
}

func validateImageJpeg(data []byte) error {
	_, err := jpeg.Decode(bytes.NewReader(data))
	return err
}

// WatchArtifacts watches inventory of artifacts for changes.
func (g *Server) WatchArtifacts(req *catalogv3.WatchArtifactsRequest, server catalogv3.CatalogService_WatchArtifactsServer) error {
	if server == nil {
		return errors.NewInvalidArgument(
			errors.WithMessage("incomplete request"))
	}
	projectUUID, err := GetActiveProjectIDAllowAdmin(server.Context(), req.ProjectId)
	if err != nil {
		return err
	}
	if req == nil {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.ArtifactType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(server.Context(), req); err != nil {
		return err
	}

	ch := make(chan *catalogv3.WatchArtifactsResponse)

	// If replay requested
	if !req.NoReplay {
		// Get list of artifacts
		ctx := server.Context()
		tx, err := g.startTransaction(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}

		artifacts, projectUUIDs, _, err := g.getArtifacts(ctx, tx, projectUUID, nil, nil, 0, 0)
		if err != nil {
			g.rollbackTransaction(tx)
			return err
		}

		events := &ArtifactEvents{}
		for i, art := range artifacts {
			events.append(ReplayedEvent, projectUUIDs[i], art)
		}

		// Send each replay event to the stream
		for _, e := range events.queue {
			if err = server.Send(e); err != nil {
				g.rollbackTransaction(tx)
				return err
			}
		}

		// Register the stream, so it can start receiving updates
		g.listeners.addArtifactListener(ch, req)

		err = g.commitTransaction(tx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	} else {
		// Register the stream, so it can start receiving updates
		g.listeners.addArtifactListener(ch, req)
	}
	defer g.listeners.deleteArtifactListener(ch)
	return g.watchArtifactEvents(server, ch)
}

func (g *Server) watchArtifactEvents(server catalogv3.CatalogService_WatchArtifactsServer, ch chan *catalogv3.WatchArtifactsResponse) error {
	for e := range ch {
		if err := server.Send(e); err != nil {
			return err
		}
	}
	return nil
}
