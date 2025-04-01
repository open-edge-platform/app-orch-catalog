// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	nberrors "github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/validator"
	"io"
	"path"
	"regexp"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/artifact"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/registry"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	yaml "gopkg.in/yaml.v3"
)

// Structure to track multiple uploads for the same session
type uploadSession struct {
	sessionID   string
	projectUUID string
	lock        sync.RWMutex
	uploads     []*catalogv3.Upload
	files       map[string]*catalogv3.Upload
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
		files:       make(map[string]*catalogv3.Upload, 0),
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
	orderedSpecs, err := u.loadYamlSpecs()
	if err != nil {
		return err
	}

	for _, d := range orderedSpecs {
		switch d.SpecSchema {
		case upload.DeploymentPackageType:
			err = u.loadDeploymentPackage(ctx, tx, d)
		case upload.DeploymentPackageLegacyType:
			err = u.loadDeploymentPackage(ctx, tx, d)
		case upload.ApplicationType:
			err = u.loadApplication(ctx, tx, d)
		case upload.RegistryType:
			err = u.loadRegistry(ctx, tx, d)
		case upload.ArtifactType:
			err = u.loadArtifact(ctx, tx, d)
		default:
			return nberrors.NewInvalidArgument(nberrors.WithMessage("uploaded file %s: unhandled type %s", d.FileName, d.SpecSchema))
		}
		if err != nil {
			return nberrors.NewInvalidArgument(nberrors.WithError(err), nberrors.WithMessage("uploaded file %s: %v", d.FileName, err))
		}
	}

	return nil
}

// shouldValidateYAMLSchema determines if the schema checker should run on the given
// artifact. If the YAML can be unmarshaled or contains the schema directive,
// it should be validated. Values files containing {{ }} markers should
// not be validated.
func shouldValidateYAMLSchema(fileBytes []byte) bool {
	var raw interface{}
	err := yaml.Unmarshal(fileBytes, &raw)
	if err != nil {
		re, _ := regexp.Compile(`\$schema:\s*"https://schema.intel.com/catalog.orchestrator/0.1/schema`)

		if !re.MatchString(string(fileBytes)) {
			return false
		}
	}
	return true
}

// loadYamlSpecs loads the contents of the specified uploads and returns them as an ordered
// collection of YamlSpecs
func (u *uploadSession) loadYamlSpecs() (upload.YamlSpecs, error) {
	orderedSpecs := make(upload.YamlSpecs, 0)

	for _, file := range u.uploads {
		fileBytes := file.Artifact
		u.files[file.FileName] = file

		// Unmarshal our input YAML file into empty interface
		if shouldValidateYAMLSchema(fileBytes) {
			// Deal only with files that can be successfully unmarshalled; value.yaml files with templates can't be for example
			decoder := yaml.NewDecoder(bytes.NewBuffer(fileBytes))
			for {
				var d upload.YamlSpec
				if err := decoder.Decode(&d); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return nil, fmt.Errorf("document decode failed: %w", err)
				}
				d.FileName = file.FileName
				if d.SpecSchema != "" {
					// check that the uploaded YAML complies with the schema
					v, err := validator.NewValidator()
					if err != nil {
						return nil, err
					}

					err = v.Validate(file.Artifact)
					if err != nil {
						log.Infof("YAML validation failed for %s:%s", file.FileName, err)
						return nil, nberrors.NewInvalidArgument(
							nberrors.WithMessage("uploaded file %s is invalid YAML: %+v", file.FileName, err),
							nberrors.WithError(err))
					}
					orderedSpecs = append(orderedSpecs, d)
				}
			}
		}
	}

	sort.Sort(orderedSpecs)
	return orderedSpecs, nil
}

func valueOrDefault(val string, def string) string {
	if val == "" {
		return def
	}
	return val
}

func (u *uploadSession) loadRegistry(ctx context.Context, tx *generated.Tx, d upload.YamlSpec) error {
	reg := &catalogv3.Registry{
		Name:         d.Name,
		DisplayName:  d.DisplayName,
		Description:  d.Description,
		RootUrl:      d.RootURL,
		InventoryUrl: d.InventoryURL,
		Username:     d.UserName,
		AuthToken:    d.AuthToken,
		Type:         valueOrDefault(d.GetRegistryType(), helmType),
		ApiType:      d.APIType,
		Cacerts:      d.CACerts,
	}

	_, err := tx.Registry.Query().Where(registry.ProjectUUID(u.projectUUID), registry.Name(reg.Name)).First(ctx)
	if err != nil {
		_, err = u.g.createRegistry(ctx, tx, u.projectUUID, reg, u.registryEvents)
		return err
	}
	return u.g.updateRegistry(ctx, tx, u.projectUUID, reg, u.registryEvents)
}

func (u *uploadSession) loadArtifact(ctx context.Context, tx *generated.Tx, d upload.YamlSpec) error {
	artifactBinary, err := b64.StdEncoding.DecodeString(d.Artifact)
	if err != nil {
		return nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.ArtifactType),
			nberrors.WithMessage("error decoding base64 from file %s %v", d.FileName, err))
	}
	if malware.DefaultScanner != nil {
		// Note: Artifact Size is already limited by protobuf validator
		okay, res, err := malware.DefaultScanner.ScanBytes(artifactBinary)
		if err != nil {
			if malware.DefaultScanner.IsPermissive() {
				log.Warn("Malware scanner is not available. Skipping scan due to permissive mode.")
			} else {
				log.Warn("Malware scanner returned error %s", err)
				return nberrors.NewUnavailable(
					nberrors.WithResourceType(nberrors.ArtifactType),
					nberrors.WithMessage("malware scanner configured but not available"))
			}
		} else if !okay {
			return nberrors.NewInvalidArgument(
				nberrors.WithResourceType(nberrors.ArtifactType),
				nberrors.WithMessage("malware detected: %s", res))
		}
	}

	art := &catalogv3.Artifact{
		Name:        d.Name,
		DisplayName: d.DisplayName,
		Description: d.Description,
		MimeType:    d.MimeType,
		Artifact:    artifactBinary,
	}

	_, err = tx.Artifact.Query().Where(artifact.ProjectUUID(u.projectUUID), artifact.Name(art.Name)).First(ctx)
	if err != nil {
		_, err = u.g.createArtifact(ctx, tx, u.projectUUID, art, u.artifactEvents)
		return err
	}
	return u.g.updateArtifact(ctx, tx, u.projectUUID, art, u.artifactEvents)
}

func (u *uploadSession) loadApplication(ctx context.Context, tx *generated.Tx, d upload.YamlSpec) error {
	app := &catalogv3.Application{
		Name:               d.Name,
		Version:            d.Version,
		Kind:               kindFromDB(d.Kind),
		DisplayName:        d.DisplayName,
		Description:        d.Description,
		ChartName:          d.ChartName,
		ChartVersion:       d.ChartVersion,
		HelmRegistryName:   d.GetHelmRegistry(),
		ImageRegistryName:  d.ImageRegistry,
		DefaultProfileName: d.DefaultProfile,
		Profiles:           make([]*catalogv3.Profile, 0, len(d.Profiles)),
		IgnoredResources:   make([]*catalogv3.ResourceReference, 0, len(d.IgnoredResources)),
	}

	if len(d.Profiles) > 0 {
		for _, p := range d.Profiles {
			prof, err := u.loadProfile(d.FileName, p)
			if err != nil {
				return err
			}
			app.Profiles = append(app.Profiles, prof)
		}
		if app.DefaultProfileName == "" {
			app.DefaultProfileName = d.Profiles[0].Name
		}
	}

	if len(d.IgnoredResources) > 0 {
		for _, r := range d.IgnoredResources {
			app.IgnoredResources = append(app.IgnoredResources, &catalogv3.ResourceReference{
				Name:      r.Name,
				Kind:      r.Kind,
				Namespace: r.Namespace,
			})
		}
	}

	_, err := tx.Application.Query().Where(application.ProjectUUID(u.projectUUID), application.Name(app.Name), application.Version(app.Version)).First(ctx)
	if err != nil {
		_, err = u.g.createApplication(ctx, tx, u.projectUUID, app, u.applicationEvents)
		return err
	}
	return u.g.updateApplication(ctx, tx, u.projectUUID, app, u.applicationEvents)
}

func (u *uploadSession) loadProfile(appFileName string, p upload.Profile) (*catalogv3.Profile, error) {
	upload, ok := u.files[p.ValuesFileName]
	if !ok {
		upload, ok = u.files[fmt.Sprintf("%s/%s", path.Dir(appFileName), p.ValuesFileName)]
		if !ok {
			return nil, nberrors.NewInvalidArgument(
				nberrors.WithMessage("chart values file %s not found in uploads", p.ValuesFileName))
		}
	}

	requirements := make([]*catalogv3.DeploymentRequirement, 0)
	for _, dr := range p.DeploymentRequirements {
		newRequirement := &catalogv3.DeploymentRequirement{
			Name:                  dr.Name,
			Version:               dr.Version,
			DeploymentProfileName: dr.DeploymentProfile,
		}
		requirements = append(requirements, newRequirement)
	}

	parameterTemplates := make([]*catalogv3.ParameterTemplate, 0)
	for _, pt := range p.ParameterTemplates {
		newParameterTemplate := &catalogv3.ParameterTemplate{
			Name:            pt.Name,
			DisplayName:     pt.DisplayName,
			Default:         pt.Default,
			Type:            pt.Type,
			Validator:       pt.Validator,
			SuggestedValues: pt.SuggestedValues,
			Secret:          pt.Secret,
			Mandatory:       pt.Mandatory,
		}
		parameterTemplates = append(parameterTemplates, newParameterTemplate)
	}

	yamlString := string(upload.Artifact)
	return &catalogv3.Profile{
		Name:                  p.Name,
		DisplayName:           p.DisplayName,
		Description:           p.Description,
		ChartValues:           yamlString,
		DeploymentRequirement: requirements,
		ParameterTemplates:    parameterTemplates,
	}, nil
}

func (u *uploadSession) loadDeploymentPackage(ctx context.Context, tx *generated.Tx, d upload.YamlSpec) error {
	pkg := &catalogv3.DeploymentPackage{
		Name:                    d.Name,
		DisplayName:             d.DisplayName,
		Description:             d.Description,
		Version:                 d.Version,
		Kind:                    kindFromDB(d.Kind),
		DefaultProfileName:      d.DefaultProfile,
		Profiles:                make([]*catalogv3.DeploymentProfile, 0, len(d.DeploymentProfiles)),
		ApplicationReferences:   make([]*catalogv3.ApplicationReference, 0, len(d.Applications)),
		ApplicationDependencies: make([]*catalogv3.ApplicationDependency, 0, len(d.ApplicationDependencies)),
		Extensions:              make([]*catalogv3.APIExtension, 0, len(d.Extensions)),
		Artifacts:               make([]*catalogv3.ArtifactReference, 0, len(d.ArtifactReferences)),
		DefaultNamespaces:       make(map[string]string, len(d.DefaultNamespaces)),
		Namespaces:              make([]*catalogv3.Namespace, 0, len(d.Namespaces)),
	}

	for _, a := range d.Applications {
		pkg.ApplicationReferences = append(pkg.ApplicationReferences,
			&catalogv3.ApplicationReference{
				Name:    a.Name,
				Version: a.Version,
			})
	}

	for _, e := range d.Extensions {
		var endpoints []*catalogv3.Endpoint
		for _, ep := range e.Endpoints {
			endpoints = append(endpoints,
				&catalogv3.Endpoint{
					ServiceName:  ep.ServiceName,
					ExternalPath: ep.ExternalPath,
					InternalPath: ep.InternalPath,
					Scheme:       ep.Scheme,
					AuthType:     ep.AuthType,
					AppName:      ep.AppName,
				})
		}
		var uiExtension *catalogv3.UIExtension
		if e.UIExtension != nil {
			uiExtension =
				&catalogv3.UIExtension{
					ServiceName: e.UIExtension.ServiceName,
					Label:       *e.UIExtension.Label,
					Description: e.UIExtension.Description,
					FileName:    e.UIExtension.FileName,
					AppName:     e.UIExtension.AppName,
					ModuleName:  e.UIExtension.ModuleName,
				}
		}
		pkg.Extensions = append(pkg.Extensions,
			&catalogv3.APIExtension{
				Name:        e.Name,
				Version:     e.Version,
				DisplayName: e.DisplayName,
				Description: e.Description,
				Endpoints:   endpoints,
				UiExtension: uiExtension,
			})
	}

	for _, a := range d.GetArtifacts() {
		pkg.Artifacts = append(pkg.Artifacts,
			&catalogv3.ArtifactReference{
				Name:    a.Name,
				Purpose: a.Purpose,
			})
	}

	if len(d.ApplicationDependencies) > 0 {
		for _, a := range d.ApplicationDependencies {
			pkg.ApplicationDependencies = append(pkg.ApplicationDependencies,
				&catalogv3.ApplicationDependency{
					Name:     a.Name,
					Requires: a.Requires,
				})
		}
	}

	if len(d.DefaultNamespaces) != 0 {
		for k, v := range d.DefaultNamespaces {
			pkg.DefaultNamespaces[k] = v
		}
	}

	if len(d.Namespaces) != 0 {
		for _, ns := range d.Namespaces {
			pkg.Namespaces = append(pkg.Namespaces,
				&catalogv3.Namespace{
					Name:        ns.Name,
					Labels:      ns.Labels,
					Annotations: ns.Annotations,
				})
		}
	}

	if len(d.DeploymentProfiles) > 0 {
		for _, profile := range d.DeploymentProfiles {
			pkg.Profiles = append(pkg.Profiles, u.deploymentProfile(profile))
		}
		if pkg.DefaultProfileName == "" {
			pkg.DefaultProfileName = d.DeploymentProfiles[0].Name
		}
	}

	pkg.IsVisible = d.IsVisible
	pkg.IsDeployed = d.IsDeployed
	pkg.ForbidsMultipleDeployments = d.ForbidsMultipleDeployments

	_, err := tx.DeploymentPackage.Query().Where(deploymentpackage.ProjectUUID(u.projectUUID),
		deploymentpackage.Name(pkg.Name), deploymentpackage.Version(pkg.Version)).First(ctx)
	if err != nil {
		_, err = u.g.createDeploymentPackage(ctx, tx, u.projectUUID, pkg, u.deploymentPackageEvents)
		return err
	}
	return u.g.updateDeploymentPackage(ctx, tx, u.projectUUID, pkg, u.deploymentPackageEvents)
}

func (u *uploadSession) deploymentProfile(deploymentProfile upload.DeploymentProfile) *catalogv3.DeploymentProfile {
	prof := &catalogv3.DeploymentProfile{
		Name:                deploymentProfile.Name,
		DisplayName:         deploymentProfile.DisplayName,
		Description:         deploymentProfile.Description,
		ApplicationProfiles: make(map[string]string, len(deploymentProfile.ApplicationProfiles)),
	}

	for _, p := range deploymentProfile.ApplicationProfiles {
		prof.ApplicationProfiles[p.ApplicationName] = p.ProfileName
	}
	return prof
}
