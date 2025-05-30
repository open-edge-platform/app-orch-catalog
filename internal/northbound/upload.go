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
	"sync"

	nberrors "github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"

	"github.com/google/uuid"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/artifact"
	"github.com/open-edge-platform/app-orch-catalog/internal/yamlreader"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
)

const (
	MaxExtractedFileSize = 10 * 1024 * 1024 // to limit the size of extracted files and mitigate decompression bomb lint message
)

// Structure to track multiple uploads for the same session
type uploadSession struct {
	yamlreader.YamlReader

	sessionID   string
	projectUUID string
	lock        sync.RWMutex
	uploads     []*catalogv3.Upload
	g           *Server

	registryEvents          *RegistryEvents
	artifactEvents          *ArtifactEvents
	applicationEvents       *ApplicationEvents
	deploymentPackageEvents *DeploymentPackageEvents
}

// Registers a new session and returns it.
func (g *Server) newSession(projectUUID string) *uploadSession {
	session := &uploadSession{
		sessionID:   uuid.NewString(),
		projectUUID: projectUUID,
		uploads:     make([]*catalogv3.Upload, 0, 1),
		g:           g,

		registryEvents:          &RegistryEvents{},
		artifactEvents:          &ArtifactEvents{},
		applicationEvents:       &ApplicationEvents{},
		deploymentPackageEvents: &DeploymentPackageEvents{},
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	g.uploadSessions[session.sessionID] = session
	return session
}

// Registers a new session and returns it.
func (g *Server) getSession(sessionID string, projectUUID string) (*uploadSession, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	session, ok := g.uploadSessions[sessionID]
	if !ok {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.UploadSession),
			nberrors.WithResourceName(sessionID),
			nberrors.WithMessage(`session not found`))
	}
	if projectUUID != session.projectUUID {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.UploadSession),
			nberrors.WithResourceName(sessionID),
			nberrors.WithMessage(`session not found for project`))
	}
	return session, nil
}

// UploadCatalogEntities allows upload of YAML files containing catalog entity descriptions or TAR file containing such YAML files through gRPC
func (g *Server) UploadCatalogEntities(ctx context.Context, req *catalogv3.UploadCatalogEntitiesRequest) (*catalogv3.UploadCatalogEntitiesResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}

	if req == nil || req.Upload == nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithMessage("incomplete request"))
	} else if err = req.Upload.Validate(); err != nil {
		return nil, nberrors.NewInvalidArgument(
			nberrors.WithMessage(err.Error()))
	}

	if err = g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	var session *uploadSession

	// Is this is first request? If so, create a new session
	if req.SessionId == "" {
		session = g.newSession(projectUUID)
	} else {
		// Otherwise, look up the specified session by its ID
		session, err = g.getSession(req.SessionId, projectUUID)
		if err != nil {
			return nil, err
		}
	}

	session.lock.Lock()
	defer session.lock.Unlock()

	if malware.DefaultScanner != nil {
		okay, res, err := malware.DefaultScanner.ScanBytes(req.Upload.Artifact)
		if err != nil {
			if malware.DefaultScanner.IsPermissive() {
				log.Warn("Malware scanner is not available. Skipping scan due to permissive mode.")
			} else {
				log.Warn("Malware scanner returned error %s", err)
				return nil, nberrors.NewUnavailable(
					nberrors.WithMessage("malware scanner configured but not available"))
			}
		} else if !okay {
			return nil, nberrors.NewInvalidArgument(
				nberrors.WithMessage("malware detected: %s", res))
		}
	}

	// Register the upload under the session file system
	session.uploads = append(session.uploads, req.Upload)
	resp := &catalogv3.UploadCatalogEntitiesResponse{SessionId: session.sessionID, ErrorMessages: nil}

	// If this is a last upload, process all uploaded entities in a single transaction
	if req.LastUpload {
		tx, err := g.startTransaction(ctx)
		if err != nil {
			return nil, err
		}

		if err := session.processUploadSession(ctx, tx); err != nil {
			g.rollbackTransaction(tx)
			return nil, err
		}

		err = g.commitTransaction(tx)
		if err != nil {
			return nil, err
		}

		session.registryEvents.sendToAll(g.listeners)
		session.artifactEvents.sendToAll(g.listeners)
		session.applicationEvents.sendToAll(g.listeners)
		session.deploymentPackageEvents.sendToAll(g.listeners)
	}

	return resp, nil
}

func (u *uploadSession) processUploadSession(ctx context.Context, tx *generated.Tx) error {
	// Turn the uploaded file list into a FileSet, so we can pass it to ExpandFileSets
	uploadedFiles := make(yamlreader.FileSet, 0)
	for _, upload := range u.uploads {
		uploadedFiles[upload.FileName] = upload.Artifact
	}

	// Turn the uploads into independent filesets. Each tarball will be a fileset, and
	// any raw files will be collected into a fileset.
	fileSets, err := u.ExpandFileSet(uploadedFiles)
	if err != nil {
		return err
	}

	// fileSets now contains a set of fileSets, each one that is either a separate tarball
	// that was uploaded, or is the set of raw yaml files that were uploaded. Process each
	// fileSet independently.

	for _, fileSet := range fileSets {
		err := u.ProcessFiles(ctx, fileSet, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *uploadSession) ProcessFiles(ctx context.Context, files yamlreader.FileSet, tx *generated.Tx) error {
	// Process each FileSet, load its yaml, sort, and add to the catalog.

	orderedSpecs, err := u.LoadYamlSpecs(files)
	if err != nil {
		return err
	}

	for _, d := range orderedSpecs {
		switch d.SpecSchema {
		case upload.DeploymentPackageType:
			var dp *catalogv3.DeploymentPackage
			dp, err = u.ReadDeploymentPackage(d)
			if err == nil {
				err = u.loadDeploymentPackage(ctx, tx, dp)
			}
		case upload.DeploymentPackageLegacyType:
			var dp *catalogv3.DeploymentPackage
			dp, err = u.ReadDeploymentPackage(d)
			if err == nil {
				err = u.loadDeploymentPackage(ctx, tx, dp)
			}
		case upload.ApplicationType:
			var app *catalogv3.Application
			app, err = u.ReadApplication(d, files) // application uses the FileSet to lookup profiles
			if err == nil {
				err = u.loadApplication(ctx, tx, app)
			}
		case upload.RegistryType:
			var reg *catalogv3.Registry
			reg, err = u.ReadRegistry(d)
			if err == nil {
				err = u.loadRegistry(ctx, tx, reg)
			}
		case upload.ArtifactType:
			var art *catalogv3.Artifact
			art, err = u.ReadArtifact(d)
			if err == nil {
				err = u.loadArtifact(ctx, tx, art)
			}
		default:
			return nberrors.NewInvalidArgument(nberrors.WithMessage("uploaded file %s: unhandled type %s", d.FileName, d.SpecSchema))
		}
		if err != nil {
			return nberrors.NewInvalidArgument(nberrors.WithError(err), nberrors.WithMessage("uploaded file %s: %v", d.FileName, err))
		}
	}

	return nil
}

func (u *uploadSession) loadRegistry(ctx context.Context, tx *generated.Tx, reg *catalogv3.Registry) error {
	return u.g.createOrUpdateRegistry(ctx, tx, u.projectUUID, reg, u.registryEvents)
}

func (u *uploadSession) loadArtifact(ctx context.Context, tx *generated.Tx, art *catalogv3.Artifact) error {
	_, err := tx.Artifact.Query().Where(artifact.ProjectUUID(u.projectUUID), artifact.Name(art.Name)).First(ctx)
	if err != nil {
		_, err = u.g.createArtifact(ctx, tx, u.projectUUID, art, u.artifactEvents)
		return err
	}
	return u.g.updateArtifact(ctx, tx, u.projectUUID, art, u.artifactEvents)
}

func (u *uploadSession) loadApplication(ctx context.Context, tx *generated.Tx, app *catalogv3.Application) error {
	return u.g.createOrUpdateApplication(ctx, tx, u.projectUUID, app, u.applicationEvents)
}

func (u *uploadSession) loadDeploymentPackage(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage) error {
	return u.g.createOrUpdateDeploymentPackage(ctx, tx, u.projectUUID, pkg, u.deploymentPackageEvents)
}
