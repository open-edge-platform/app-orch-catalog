// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/ignoredresource"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/predicate"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/profile"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/registry"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	kindNormal    = "normal"
	kindAddon     = "addon"
	kindExtension = "extension"
)

func kindToDB(kind catalogv3.Kind) string {
	switch kind {
	case catalogv3.Kind_KIND_NORMAL:
		return kindNormal
	case catalogv3.Kind_KIND_ADDON:
		return kindAddon
	case catalogv3.Kind_KIND_EXTENSION:
		return kindExtension
	}
	return kindNormal
}

func kindFromDB(kind string) catalogv3.Kind {
	switch kind {
	case kindNormal:
		return catalogv3.Kind_KIND_NORMAL
	case kindAddon:
		return catalogv3.Kind_KIND_ADDON
	case kindExtension:
		return catalogv3.Kind_KIND_EXTENSION
	}
	return catalogv3.Kind_KIND_NORMAL
}

func isSameKind(kind catalogv3.Kind, kindDB string) bool {
	return (kind == catalogv3.Kind_KIND_NORMAL && kindDB == kindNormal) ||
		(kind == catalogv3.Kind_KIND_EXTENSION && kindDB == kindExtension) ||
		(kind == catalogv3.Kind_KIND_ADDON && kindDB == kindAddon) ||
		(kind == catalogv3.Kind_KIND_UNSPECIFIED && kindDB == "")
}

// CreateApplication creates an Application from gRPC request
func (g *Server) CreateApplication(ctx context.Context, req *catalogv3.CreateApplicationRequest) (*catalogv3.CreateApplicationResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Application == nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	} else if err := req.Application.Validate(); err != nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage(err.Error()))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &ApplicationEvents{}
	created, err := g.createApplication(ctx, tx, projectUUID, req.Application, events)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "created", "application", projectUUID, req.Application.Name, req.Application.Version)
	events.sendToAll(g.listeners)

	return &catalogv3.CreateApplicationResponse{
		Application: &catalogv3.Application{
			Name:               created.Name,
			DisplayName:        created.DisplayName,
			Description:        created.Description,
			HelmRegistryName:   req.Application.HelmRegistryName,
			ImageRegistryName:  req.Application.ImageRegistryName,
			Version:            created.Version,
			ChartName:          created.ChartName,
			Profiles:           req.Application.Profiles,
			IgnoredResources:   req.Application.IgnoredResources,
			ChartVersion:       created.ChartVersion,
			DefaultProfileName: req.Application.DefaultProfileName,
			Kind:               kindFromDB(created.Kind),
			CreateTime:         timestamppb.New(created.CreateTime),
		},
	}, nil
}

func (g *Server) createApplication(ctx context.Context, tx *generated.Tx, projectUUID string, app *catalogv3.Application, events *ApplicationEvents) (*generated.Application, error) {
	if len(app.Profiles) > 0 && app.DefaultProfileName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("default profile name must be specified"))
	}

	displayName, ok := validateDisplayName(app.Name, app.DisplayName)
	if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkApplicationUniqueness(ctx, tx, projectUUID, app); err != nil {
		return nil, err
	}

	helmRegistry, ok, err := g.getRegistry(ctx, tx, projectUUID, app.HelmRegistryName, helmType)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("helm registry %s not found", app.HelmRegistryName))
	}

	stmt := tx.Application.Create().
		SetProjectUUID(projectUUID).
		SetRegistryFkID(helmRegistry.ID).
		SetName(app.Name).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(app.Description).
		SetVersion(app.Version).
		SetChartName(app.ChartName).
		SetChartVersion(app.ChartVersion).
		SetKind(kindToDB(app.Kind))

	// If image registry has been specified, apply it as well.
	if len(app.ImageRegistryName) > 0 {
		imageRegistry, ok, err := g.getRegistry(ctx, tx, projectUUID, app.ImageRegistryName, imageType)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithMessage("image registry %s not found", app.ImageRegistryName))
		}

		stmt.SetImageRegistryFkID(imageRegistry.ID)
	}

	created, err := stmt.Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithResourceName(app.Name),
				errors.WithMessage("deployment application %s already exists", app.Name))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}

	// Create any profiles and record a default profile
	if err = g.createProfiles(ctx, tx, projectUUID, app.Profiles, created); err != nil {
		return nil, err
	}
	if ok, err = g.updateDefaultProfile(ctx, tx, app.DefaultProfileName, created); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(app.Name),
			errors.WithResourceVersion(app.Version),
			errors.WithMessage("could not update default profile %s", app.DefaultProfileName))
	}

	// Create any ignored resources for this application
	if err = g.createIgnoredResources(ctx, tx, app, created); err != nil {
		return nil, err
	}
	events.append(CreatedEvent, projectUUID, app)
	return created, nil
}

// Returns an error if the application display name is not unique
func (g *Server) checkApplicationUniqueness(ctx context.Context, tx *generated.Tx, projectUUID string, a *catalogv3.Application) error {
	if len(a.DisplayName) > 0 {
		exists, err := tx.Application.Query().
			Where(
				application.ProjectUUID(projectUUID),
				application.DisplayNameLc(strings.ToLower(a.DisplayName)),
				application.Not(application.Name(a.Name))).
			Exist(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		} else if exists {
			return errors.NewAlreadyExists(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithResourceName(a.Name),
				errors.WithResourceVersion(a.Version),
				errors.WithMessage("display name already exists"))
		}

	}
	return nil
}

// Create any profiles and record a default profile
func (g *Server) createProfiles(ctx context.Context, tx *generated.Tx, projectUUID string, profiles []*catalogv3.Profile, appDB *generated.Application) error {
	for _, profile := range profiles {
		_, err := g.injectProfile(ctx, tx, projectUUID, profile, appDB)
		if err != nil {
			return err
		}
	}
	return nil
}

// Updates the default profile
func (g *Server) updateDefaultProfile(ctx context.Context, tx *generated.Tx, name string, appDB *generated.Application) (bool, error) {
	if name == "" {
		// If the default profile is not specified, clear it from the app
		_, err := tx.Application.Update().Where(application.ID(appDB.ID)).ClearDefaultProfile().Save(ctx)
		if err != nil {
			return false, errors.NewDBError(errors.WithError(err))
		}
	} else {
		// If the default profile name has been specified, find one in the database.
		defaultProfile, err := appDB.QueryProfiles().Where(profile.HasApplicationFkWith(application.ID(appDB.ID)), profile.Name(name)).First(ctx)
		if err != nil {
			if generated.IsNotFound(err) {
				return false, nil
			}
			return false, errors.NewDBError(errors.WithError(err))
		}
		_, err = tx.Application.Update().Where(application.ID(appDB.ID)).SetDefaultProfile(defaultProfile).Save(ctx)
		if err != nil {
			return false, errors.NewDBError(errors.WithError(err))
		}
	}
	return true, nil
}

// Create any ignores resources for this application
func (g *Server) createIgnoredResources(ctx context.Context, tx *generated.Tx, app *catalogv3.Application, appDB *generated.Application) error {
	for _, ignoredResource := range app.IgnoredResources {
		_, err := tx.IgnoredResource.Create().
			SetApplicationFkID(appDB.ID).
			SetName(ignoredResource.Name).
			SetKind(ignoredResource.Kind).
			SetNamespace(ignoredResource.Namespace).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Retrieve registry by name and type for the specified project
func (g *Server) getRegistry(ctx context.Context, tx *generated.Tx, projectUUID string, registryName string, registryType string) (*generated.Registry, bool, error) {
	registry, err := tx.Registry.Query().
		Where(registry.ProjectUUID(projectUUID), registry.Name(registryName)).First(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, errors.NewDBError(errors.WithError(err))
	}
	if registry.Type != registryType {
		return nil, false, nil
	}
	return registry, true, nil
}

var applicationColumns = map[string]string{
	"publisher_name":     "",
	"name":               "name",
	"displayName":        "display_name",
	"description":        "description",
	"version":            "version",
	"chartName":          "chart_name",
	"chartVersion":       "chart_version",
	"helmRegistryName":   "helm_registry_name",
	"createTime":         "create_time",
	"updateTime":         "update_time",
	"defaultProfileName": "default_profile_name",
	"imageRegistryName":  "image_registry_name",
}

// ListApplications gets a list of all applications through gRPC
func (g *Server) ListApplications(ctx context.Context, req *catalogv3.ListApplicationsRequest) (*catalogv3.ListApplicationsResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	orderBys, err := parseOrderBy(req.OrderBy, errors.ApplicationType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}
	filters, err := parseFilter(req.Filter, errors.ApplicationType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	applications, _, totalElements, err := g.getApplications(ctx, tx, projectUUID, req.Kinds, orderBys, filters, req.PageSize, req.Offset)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "listed", "applications", projectUUID, "Total =>"+fmt.Sprintf("%d", totalElements))
	return &catalogv3.ListApplicationsResponse{Applications: applications, TotalElements: totalElements}, nil
}

func (g *Server) getApplications(ctx context.Context, tx *generated.Tx, projectUUID string, kinds []catalogv3.Kind,
	orderBys []*orderBy, filters []*filter, pageSize int32, offset int32) ([]*catalogv3.Application, []string, int32, error) {
	var err error
	var applicationsDB []*generated.Application
	var orderOptions []application.OrderOption
	applicationsQuery := tx.Application.Query()

	options, err := orderByOptions(orderBys, applicationColumns, errors.ApplicationType)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, pred := range options {
		orderOptions = append(orderOptions, pred)
	}
	applicationsQuery = applicationsQuery.Order(orderOptions...)

	filterPreds, err := filterPredicates(filters, applicationColumns, errors.ApplicationType)
	if err != nil {
		return nil, nil, 0, err
	}
	var applicationPreds []predicate.Application
	for _, pred := range filterPreds {
		applicationPreds = append(applicationPreds, pred)
	}
	applicationsQuery = applicationsQuery.Where(application.Or(applicationPreds...))

	kindFilter := kindPredicate(kinds)
	if kindFilter != nil {
		applicationsQuery = applicationsQuery.Where(kindFilter)
	}

	if projectUUID == "" {
		applicationsDB, err = applicationsQuery.All(ctx)
	} else {
		applicationsDB, err = applicationsQuery.Where(application.ProjectUUID(projectUUID)).All(ctx)
	}
	if err != nil {
		return nil, nil, 0, errors.NewDBError(errors.WithError(err))
	}

	applications, projectUUIDs, totalElements, err := g.applicationsExtract(ctx, applicationsDB, projectUUID, pageSize, offset)
	if err != nil {
		return nil, nil, 0, err
	}
	return applications, projectUUIDs, totalElements, nil
}

func (g *Server) applicationsExtract(ctx context.Context, appsDB []*generated.Application, publisherName string,
	pageSize int32, offset int32) ([]*catalogv3.Application, []string, int32, error) {
	var err error
	applications := make([]*catalogv3.Application, 0, len(appsDB))
	projectUUIDs := make([]string, 0, len(appsDB))
	startIndex, endIndex, totalElements, err := computePageRange(pageSize, offset, len(appsDB))
	if err != nil {
		return nil, nil, 0, err
	}

	if len(appsDB) == 0 {
		return []*catalogv3.Application{}, []string{}, 0, nil
	}

	for i := startIndex; i <= endIndex; i++ {
		appDB := appsDB[i]
		application, err := g.applicationExtract(ctx, appDB, publisherName)
		if err != nil {
			return nil, nil, 0, err
		}
		applications = append(applications, application)
		projectUUIDs = append(projectUUIDs, appDB.ProjectUUID)
	}
	return applications, projectUUIDs, totalElements, nil
}

func (g *Server) applicationExtract(ctx context.Context, appDB *generated.Application, _ string) (*catalogv3.Application, error) {
	helmRegistry, err := appDB.QueryRegistryFk().Only(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	profilesDB, err := appDB.QueryProfiles().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	profiles := make([]*catalogv3.Profile, 0, len(profilesDB))
	for _, profileDB := range profilesDB {
		requirements, err := g.extractDeploymentRequirements(ctx, profileDB)
		if err != nil {
			return nil, err
		}
		templates, err := g.extractParameterTemplates(ctx, profileDB)
		if err != nil {
			return nil, err
		}

		profiles = append(profiles, &catalogv3.Profile{
			Name:                  profileDB.Name,
			DisplayName:           profileDB.DisplayName,
			Description:           profileDB.Description,
			ChartValues:           profileDB.ChartValues,
			DeploymentRequirement: requirements,
			ParameterTemplates:    templates,
			CreateTime:            timestamppb.New(profileDB.CreateTime),
			UpdateTime:            timestamppb.New(profileDB.UpdateTime),
		})
	}

	defProfileDB, _ := appDB.QueryDefaultProfile().First(ctx)
	defaultProfileName := ""
	if defProfileDB != nil {
		defaultProfileName = defProfileDB.Name
	}

	ignoredResourcesDB, err := appDB.QueryIgnoredResources().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	ignoredResources := make([]*catalogv3.ResourceReference, 0, len(profilesDB))
	for _, irDB := range ignoredResourcesDB {
		ignoredResources = append(ignoredResources, &catalogv3.ResourceReference{Name: irDB.Name, Kind: irDB.Kind, Namespace: irDB.Namespace})
	}

	app := &catalogv3.Application{
		Name:               appDB.Name,
		DisplayName:        appDB.DisplayName,
		Description:        appDB.Description,
		Version:            appDB.Version,
		ChartName:          appDB.ChartName,
		ChartVersion:       appDB.ChartVersion,
		HelmRegistryName:   helmRegistry.Name,
		Profiles:           profiles,
		DefaultProfileName: defaultProfileName,
		IgnoredResources:   ignoredResources,
		Kind:               kindFromDB(appDB.Kind),
		CreateTime:         timestamppb.New(appDB.CreateTime),
		UpdateTime:         timestamppb.New(appDB.UpdateTime),
	}

	imageRegistry, err := appDB.QueryImageRegistryFk().Only(ctx)
	if err == nil {
		app.ImageRegistryName = imageRegistry.Name
	}

	return app, nil
}

// GetApplicationVersions gets all versions of a named Application through gRPC
func (g *Server) GetApplicationVersions(ctx context.Context, req *catalogv3.GetApplicationVersionsRequest) (*catalogv3.GetApplicationVersionsResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.ApplicationName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	applicationsDB, err := tx.Application.Query().
		Where(application.ProjectUUID(projectUUID), application.Name(req.ApplicationName)).All(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	}
	if len(applicationsDB) == 0 {
		g.rollbackTransaction(tx)
		return nil, errors.NewNotFound(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(req.ApplicationName))
	}

	applications, _, _, err := g.applicationsExtract(ctx, applicationsDB, projectUUID, 0, 0)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "listed", "application versions", projectUUID, req.ApplicationName, "total", fmt.Sprintf("%d", len(applications)))
	return &catalogv3.GetApplicationVersionsResponse{Application: applications}, nil
}

// GetApplication gets a single application through gRPC
func (g *Server) GetApplication(ctx context.Context, req *catalogv3.GetApplicationRequest) (*catalogv3.GetApplicationResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.ApplicationName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	applicationDB, err := tx.Application.Query().
		Where(
			application.ProjectUUID(projectUUID),
			application.Name(req.ApplicationName),
			application.Version(req.Version),
		).
		First(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		if generated.IsNotFound(err) {
			return nil, errors.NewNotFound(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithResourceName(req.ApplicationName),
				errors.WithResourceVersion(req.Version))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}

	application, err := g.applicationExtract(ctx, applicationDB, projectUUID)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "retrieved", "application", projectUUID, req.ApplicationName, req.Version)
	return &catalogv3.GetApplicationResponse{Application: application}, nil
}

// GetApplicationReferenceCount gets reference count for the specified application through gRPC
func (g *Server) GetApplicationReferenceCount(ctx context.Context, req *catalogv3.GetApplicationReferenceCountRequest) (*catalogv3.GetApplicationReferenceCountResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.ApplicationName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	if err := g.checkApplication(ctx, tx, req.ApplicationName, req.Version, projectUUID); err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	count, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.HasApplicationsWith(
				application.ProjectUUID(projectUUID),
				application.Name(req.ApplicationName),
				application.Version(req.Version),
			),
		).
		Count(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "retrieved", "application reference count", projectUUID, req.ApplicationName, req.Version)
	return &catalogv3.GetApplicationReferenceCountResponse{ReferenceCount: uint32(count)}, nil
}

type applicationChanges struct {
	kind             bool
	rootRecord       bool
	profiles         bool
	profile          bool
	ignoredResources bool
	helmRegistry     *generated.Registry
	imageRegistry    *generated.Registry
	newProfiles      []*catalogv3.Profile
}

func (c *applicationChanges) changed() bool {
	return c.rootRecord || c.profiles || c.profile
}

func (c *applicationChanges) changedKind() bool {
	return c.kind
}

// UpdateApplication updates an application through gRPC
func (g *Server) UpdateApplication(ctx context.Context, req *catalogv3.UpdateApplicationRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Application == nil || req.ApplicationName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	} else if err := req.Application.Validate(); err != nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage(err.Error()))
	} else if req.ApplicationName != req.Application.Name {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("name cannot be changed %s != %s",
				req.ApplicationName, req.Application.Name))
	} else if req.Version != req.Application.Version {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("version cannot be changed %s != %s",
				req.Version, req.Application.Version))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &ApplicationEvents{}
	if err = g.updateApplication(ctx, tx, projectUUID, req.Application, events); err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "updated", "application", projectUUID, req.GetApplicationName(), req.GetVersion())
	events.sendToAll(g.listeners)

	return &emptypb.Empty{}, nil
}

func (g *Server) updateApplication(ctx context.Context, tx *generated.Tx, projectUUID string, app *catalogv3.Application, events *ApplicationEvents) error {
	if len(app.Profiles) > 0 && app.DefaultProfileName == "" {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("default profile name must be specified"))
	}

	displayName, ok := validateDisplayName(app.Name, app.DisplayName)

	if !ok {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}
	// Get the application so that we can compute any changes
	appDB, ok, err := g.getApplication(ctx, tx, projectUUID, app.Name, app.Version)
	if err != nil {
		g.rollbackTransaction(tx)
		return errors.NewDBError(errors.WithError(err))
	} else if !ok {
		g.rollbackTransaction(tx)
		return errors.NewNotFound(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(app.Name),
			errors.WithResourceVersion(app.Version))
	}
	if app.Kind == catalogv3.Kind_KIND_UNSPECIFIED {
		app.Kind = kindFromDB(appDB.Kind) // keep the existing kind if not specified
	}
	app.DisplayName = displayName
	changes, err := g.computeApplicationChanges(ctx, tx, projectUUID, app, appDB)
	if err != nil {
		return err
	}

	// Make sure that the application doesn't belong to an already deployed deployment package
	// Changes to the kind field only are exempt.
	if changes.changedKind() && !changes.changed() {
		return g.updateApplicationKind(ctx, tx, projectUUID, app)
	} else if changes.changedKind() || changes.changed() {
		if err := g.checkApplicationNotInDeployedPackages(ctx, tx, projectUUID, app.Name, app.Version); err != nil {
			return err
		}
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkApplicationUniqueness(ctx, tx, "", app); err != nil {
		return err
	}

	stmt := tx.Application.Update().
		Where(
			application.ProjectUUID(projectUUID),
			application.Name(app.Name),
			application.Version(app.Version),
		).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(app.Description).
		SetRegistryFkID(changes.helmRegistry.ID).
		SetVersion(app.Version).
		SetChartName(app.ChartName).
		SetChartVersion(app.ChartVersion).
		SetKind(kindToDB(app.Kind))

	// If image registry has been changed, apply it as well.
	if len(app.ImageRegistryName) > 0 {
		stmt.SetImageRegistryFkID(changes.imageRegistry.ID)
	} else {
		stmt.ClearImageRegistryFk()
	}

	updateCount, err := stmt.Save(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	} else if updateCount == 0 {
		return errors.NewNotFound(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(app.Name),
			errors.WithResourceVersion(app.Version))
	}

	// Update the profiles, if necessary
	if changes.profiles {
		if err = g.updateProfiles(ctx, tx, projectUUID, app, appDB); err != nil {
			return err
		}
	} else {
		// Add any new profiles, if necessary
		if len(changes.newProfiles) > 0 {
			if err = g.createProfiles(ctx, tx, projectUUID, changes.newProfiles, appDB); err != nil {
				return err
			}
		}
	}

	// Update the default profile, if necessary
	if changes.profile || changes.profiles {
		if ok, err := g.updateDefaultProfile(ctx, tx, app.DefaultProfileName, appDB); err != nil {
			return err
		} else if !ok {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithMessage("could not update default profile"))
		}
	}

	// Update the ignored resources, if necessary
	if changes.ignoredResources {
		if err = g.updateIgnoredResources(ctx, tx, app, appDB); err != nil {
			return err
		}
	}
	events.append(UpdatedEvent, projectUUID, app)
	return nil
}

func (g *Server) updateApplicationKind(ctx context.Context, tx *generated.Tx, projectUUID string, app *catalogv3.Application) error {
	updateCount, err := tx.Application.Update().
		Where(
			application.ProjectUUID(projectUUID),
			application.Name(app.Name),
			application.Version(app.Version),
		).
		SetKind(kindToDB(app.Kind)).Save(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	} else if updateCount == 0 {
		return errors.NewNotFound(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(app.Name),
			errors.WithResourceVersion(app.Version))
	}
	return nil
}

func (g *Server) computeApplicationChanges(ctx context.Context, tx *generated.Tx, projectUUID string, app *catalogv3.Application, appDB *generated.Application) (*applicationChanges, error) {
	var err error
	changes := &applicationChanges{}
	registry, ok, err := g.getRegistry(ctx, tx, projectUUID, app.HelmRegistryName, helmType)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	} else if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(appDB.Name),
			errors.WithResourceVersion(appDB.Version),
			errors.WithMessage("helm registry %s not found", app.HelmRegistryName))
	}
	changes.helmRegistry = registry
	if len(app.ImageRegistryName) > 0 {
		registry, ok, err := g.getRegistry(ctx, tx, projectUUID, app.ImageRegistryName, imageType)
		if err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithResourceName(appDB.Name),
				errors.WithResourceVersion(appDB.Version),
				errors.WithMessage("image registry %s not found", app.ImageRegistryName))
		}
		changes.imageRegistry = registry
	}
	changes.kind = !isSameKind(app.Kind, appDB.Kind)
	if changes.rootRecord, err = g.applicationChanged(app, appDB, changes); err != nil {
		return nil, err
	}
	if changes.profiles, changes.newProfiles, err = g.applicationProfilesChanged(ctx, app, appDB); err != nil {
		return nil, err
	}
	if changes.profile, err = g.defaultApplicationProfileChanged(ctx, app, appDB); err != nil {
		return nil, err
	}
	if changes.ignoredResources, err = g.applicationIgnoredResourcesChanged(ctx, app, appDB); err != nil {
		return nil, err
	}

	return changes, nil
}

// Checks if the specified application belongs to any deployment_packages that are in the deployment state and returns
// and error if so
func (g *Server) checkApplicationNotInDeployedPackages(ctx context.Context, tx *generated.Tx, projectUUID string, appName string, version string) error {
	count, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.IsDeployed(true),
			deploymentpackage.HasApplicationsWith(
				application.ProjectUUID(projectUUID),
				application.Name(appName),
				application.Version(version),
			),
		).Count(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	if count > 0 {
		return errors.NewFailedPrecondition(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(appName),
			errors.WithResourceVersion(version),
			errors.WithMessage("cannot update application that is part of %d packages; please create a new version instead", count))
	}
	return nil
}

func (g *Server) applicationChanged(app *catalogv3.Application, appDB *generated.Application, changes *applicationChanges) (bool, error) {
	return app.DisplayName != appDB.DisplayName || app.Description != appDB.Description ||
		app.ChartName != appDB.ChartName || app.ChartVersion != appDB.ChartVersion ||
		app.HelmRegistryName != changes.helmRegistry.Name ||
		(changes.imageRegistry == nil && app.ImageRegistryName != "") ||
		(changes.imageRegistry != nil && app.ImageRegistryName != changes.imageRegistry.Name), nil
}

func (g *Server) applicationProfilesChanged(ctx context.Context, app *catalogv3.Application, appDB *generated.Application) (bool, []*catalogv3.Profile, error) {
	profiles, err := appDB.QueryProfiles().All(ctx)
	if err != nil {
		return false, nil, errors.NewDBError(errors.WithError(err))
	}

	// If number of existing profiles is greater than the number of new profiles, bail
	if len(profiles) > len(app.Profiles) {
		return true, nil, nil
	}

	// Otherwise, look for sameness within, but allow new profiles.
	existingProfiles := make(map[string]*generated.Profile, len(profiles))
	for _, pDB := range profiles {
		existingProfiles[pDB.Name] = pDB
	}

	newProfiles := make([]*catalogv3.Profile, 0)
	for _, p := range app.Profiles {
		if existingProfile, ok := existingProfiles[p.Name]; ok {
			// Profile is an existing one, make sure it's unchanged
			if !profilesAreSame(p, existingProfile) {
				return true, nil, nil // found a difference, so bail
			}
			if !requirementsAreSame(ctx, p, existingProfile) {
				return true, nil, nil
			}
			if templatesSame, err := g.parameterTemplatesAreSame(ctx, p, existingProfile); err != nil || !templatesSame {
				return true, nil, err
			}
			delete(existingProfiles, existingProfile.Name) // clear the key from the map of existing profiles
		} else {
			// Profile is a new one, add it to our list of new profiles
			newProfiles = append(newProfiles, p)
		}
	}

	// If there are any existing profiles remaining, this means deletion of an existing one was attempted
	if len(existingProfiles) > 0 {
		return true, nil, nil
	}

	return false, newProfiles, err
}

func requirementsAreSame(ctx context.Context, p *catalogv3.Profile, existingProfile *generated.Profile) bool {
	requirementsDB, err := existingProfile.QueryDeploymentRequirements().All(ctx)
	if err != nil {
		return false
	}

	if len(requirementsDB) != len(p.DeploymentRequirement) {
		return false
	}

	existingRequirements := make(map[string]*generated.DeploymentRequirement, len(requirementsDB))
	for _, requirementDB := range requirementsDB {
		name, err := requirementDB.QueryDeploymentPackageFk().First(ctx)
		if err != nil {
			return false
		}
		existingRequirements[fmt.Sprintf("%s:%s", name.Name, name.Version)] = requirementDB
	}

	for _, requirement := range p.DeploymentRequirement {
		requirementDB, ok := existingRequirements[fmt.Sprintf("%s:%s", requirement.Name, requirement.Version)]
		if !ok {
			return false
		}
		deploymentProfile, err := requirementDB.QueryDeploymentProfileFk().First(ctx)
		if err != nil {
			return false
		}
		if deploymentProfile.Name != requirement.DeploymentProfileName {
			return false
		}
	}
	return true
}

func profilesAreSame(profile *catalogv3.Profile, profileDB *generated.Profile) bool {
	return profile.Name == profileDB.Name && profile.DisplayName == profileDB.DisplayName && profile.Description == profileDB.Description &&
		profile.ChartValues == profileDB.ChartValues
}

func (g *Server) defaultApplicationProfileChanged(ctx context.Context, app *catalogv3.Application, appDB *generated.Application) (bool, error) {
	pDB, err := appDB.QueryDefaultProfile().Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return app.DefaultProfileName != "", nil
		}
		return false, errors.NewDBError(errors.WithError(err))
	}
	return app.DefaultProfileName != pDB.Name, nil
}

func (g *Server) applicationIgnoredResourcesChanged(ctx context.Context, app *catalogv3.Application, appDB *generated.Application) (bool, error) {
	if len(app.IgnoredResources) == 0 {
		return false, nil
	}

	ignoredResources, err := appDB.QueryIgnoredResources().All(ctx)
	if err != nil {
		return false, errors.NewDBError(errors.WithError(err))
	}

	// If number of existing and new ignored resources are not the same bail
	if len(ignoredResources) != len(app.IgnoredResources) {
		return true, nil
	}

	// Otherwise, look for sameness within
	existingIgnoredResources := make(map[string]*generated.IgnoredResource, len(ignoredResources))
	for _, ignoredResource := range ignoredResources {
		existingIgnoredResources[fmt.Sprintf("%s/%s/%s", ignoredResource.Name, ignoredResource.Namespace, ignoredResource.Kind)] = ignoredResource
	}

	for _, ignoredResource := range app.IgnoredResources {
		if _, ok := existingIgnoredResources[fmt.Sprintf("%s/%s/%s", ignoredResource.Name, ignoredResource.Namespace, ignoredResource.Kind)]; !ok {
			return true, nil // found a difference, so bail
		}
	}
	return false, nil
}

func (g *Server) parameterTemplatesAreSame(ctx context.Context, p *catalogv3.Profile, pDB *generated.Profile) (bool, error) {
	parameterTemplatesDB, err := pDB.QueryParameterTemplates().All(ctx)
	if err != nil {
		return false, errors.NewDBError(errors.WithError(err))
	}

	// If number of existing and new parameter templates are not the same bail
	if len(p.ParameterTemplates) != len(parameterTemplatesDB) {
		return false, nil
	}

	existingParameterTemplatesDB := make(map[string]*generated.ParameterTemplate, len(p.ParameterTemplates))
	for _, parameterTemplateDB := range parameterTemplatesDB {
		existingParameterTemplatesDB[parameterTemplateDB.Name] = parameterTemplateDB
	}

	for _, pt := range p.ParameterTemplates {
		ptDB, ok := existingParameterTemplatesDB[pt.Name]
		if !ok {
			return false, nil
		}
		if ptDB.Default != pt.Default ||
			ptDB.Type != pt.Type ||
			ptDB.DisplayName != pt.DisplayName ||
			len(ptDB.SuggestedValues) != len(pt.SuggestedValues) ||
			ptDB.Mandatory != pt.Mandatory ||
			ptDB.Secret != pt.Secret {
			return false, nil
		}
		// Check if the suggested values array match
		for i, svDB := range ptDB.SuggestedValues {
			sv := pt.SuggestedValues[i]
			if svDB != sv {
				return false, nil
			}
		}
	}
	return true, nil
}

// If profiles are given, use the specified list to completely replace the existing ones.
func (g *Server) updateProfiles(ctx context.Context, tx *generated.Tx, projectUUID string, app *catalogv3.Application, appDB *generated.Application) error {
	// Before we delete the exiting profiles, let's make sure none are presently being referred to by a deployment profile.
	givenProfiles := make(map[string]*catalogv3.Profile, 0)
	displayNames := make(map[string]*catalogv3.Profile, 0)
	for _, p := range app.Profiles {
		if _, ok := givenProfiles[p.Name]; ok {
			return errors.NewAlreadyExists(
				errors.WithResourceType(errors.ProfileType),
				errors.WithResourceName(p.Name))
		}
		if _, ok := displayNames[strings.ToLower(p.DisplayName)]; ok {
			return errors.NewAlreadyExists(
				errors.WithResourceType(errors.ProfileType),
				errors.WithResourceName(p.Name),
				errors.WithMessage("profile %s display name %s is not unique", p.Name, p.DisplayName))
		}
		givenProfiles[p.Name] = p
		displayNames[strings.ToLower(p.DisplayName)] = p
	}

	// Iterate over the existing profiles in the database and find those that are not in the new set, i.e. should be deleted
	profilesDB, err := appDB.QueryProfiles().All(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}

	// Scan over the existing profiles to make sure we're not attempting to delete any that have pending references
	// Otherwise either delete them or update them using what whas provided first.
	for _, profileDB := range profilesDB {
		if profile, ok := givenProfiles[profileDB.Name]; !ok {
			count, err := profileDB.QueryDeploymentProfiles().Count(ctx)
			if err != nil {
				return errors.NewDBError(errors.WithError(err))
			}

			// If a profile is to be deleted, make sure it is not being referred to by a deployment profile
			if count > 0 {
				return errors.NewFailedPrecondition(errors.WithMessage("profile %s cannot be deleted; it is in use by another deployment profile", profileDB.Name))
			}
			if err = tx.Profile.DeleteOneID(profileDB.ID).Exec(ctx); err != nil {
				return errors.NewDBError(errors.WithError(err))
			}
		} else {
			if err = g.updateProfile(ctx, tx, projectUUID, profile, profileDB); err != nil {
				return err
			}
		}

		// Expunge the profile from the new profiles as we do not need to create it later...
		delete(givenProfiles, profileDB.Name)
	}

	// Finally, convert the map of new profiles to a list and create them
	newProfiles := make([]*catalogv3.Profile, 0, len(givenProfiles))
	for _, p := range givenProfiles {
		newProfiles = append(newProfiles, p)
	}
	return g.createProfiles(ctx, tx, projectUUID, newProfiles, appDB)
}

// If ignored resources are given, use the specified list to completely replace the existing ones.
func (g *Server) updateIgnoredResources(ctx context.Context, tx *generated.Tx, app *catalogv3.Application, appDB *generated.Application) error {
	if _, err := tx.IgnoredResource.Delete().Where(ignoredresource.HasApplicationFkWith(application.ID(appDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return g.createIgnoredResources(ctx, tx, app, appDB)
}

// DeleteApplication deletes an application through gRPC
func (g *Server) DeleteApplication(ctx context.Context, req *catalogv3.DeleteApplicationRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.ApplicationName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	// Check to make sure no deployment_packages refer to this application first
	events := &ApplicationEvents{}
	count, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.HasApplicationsWith(
				application.ProjectUUID(projectUUID),
				application.Name(req.ApplicationName),
				application.Version(req.Version),
			),
		).Count(ctx)
	res, err := g.checkDeleteResult(ctx, nil, err, fmt.Sprintf("application %s:%s", req.ApplicationName, req.Version), projectUUID)
	if err != nil {
		g.rollbackTransaction(tx)
		return res, err
	}
	if count > 0 {
		g.rollbackTransaction(tx)
		return nil, errors.NewFailedPrecondition(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(req.ApplicationName),
			errors.WithResourceVersion(req.Version),
			errors.WithMessage("cannot delete application that is part of one or more %s", errors.DeploymentPackageType))
	}

	deleteCount, err := tx.Application.Delete().
		Where(
			application.ProjectUUID(projectUUID),
			application.Name(req.ApplicationName),
			application.Version(req.Version),
		).
		Exec(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	} else if deleteCount == 0 {
		g.rollbackTransaction(tx)
		return nil, errors.NewNotFound(
			errors.WithResourceType(errors.ApplicationType),
			errors.WithResourceName(req.ApplicationName),
			errors.WithResourceVersion(req.Version))
	}
	if _, err = g.checkDeleteResult(ctx, tx, err, fmt.Sprintf("application %s:%s", req.ApplicationName, req.Version), projectUUID); err != nil {
		return nil, err
	}
	events.append(DeletedEvent, projectUUID, &catalogv3.Application{Name: req.ApplicationName, Version: req.Version})
	events.sendToAll(g.listeners)
	logActivity(ctx, "deleted", "application", projectUUID, req.ApplicationName, req.Version)
	return &emptypb.Empty{}, nil
}

// WatchApplications watches inventory of applications for changes.
func (g *Server) WatchApplications(req *catalogv3.WatchApplicationsRequest, server catalogv3.CatalogService_WatchApplicationsServer) error {
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
			errors.WithResourceType(errors.ApplicationType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(server.Context(), req); err != nil {
		return err
	}

	ch := make(chan *catalogv3.WatchApplicationsResponse)

	// If replay requested
	if !req.NoReplay {
		// Get list of apps
		ctx := server.Context()
		tx, err := g.startTransaction(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}

		applications, projectUUIDs, _, err := g.getApplications(ctx, tx, projectUUID, req.Kinds, nil, nil, 0, 0)
		if err != nil {
			g.rollbackTransaction(tx)
			return err
		}

		events := &ApplicationEvents{}
		for i, app := range applications {
			events.append(ReplayedEvent, projectUUIDs[i], app)
		}

		// Send each replay event to the stream
		for _, e := range events.queue {
			if err = server.Send(e); err != nil {
				return err
			}
		}

		// Register the stream, so it can start receiving updates
		g.listeners.addApplicationListener(ch, req)

		err = g.commitTransaction(tx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	} else {
		// Register the stream, so it can start receiving updates
		g.listeners.addApplicationListener(ch, req)
	}
	defer g.listeners.deleteApplicationListener(ch)
	logActivity(server.Context(), "watching", "applications", projectUUID)
	return g.watchApplicationEvents(server, ch)
}

func (g *Server) watchApplicationEvents(server catalogv3.CatalogService_WatchApplicationsServer, ch chan *catalogv3.WatchApplicationsResponse) error {
	for e := range ch {
		if err := server.Send(e); err != nil {
			return err
		}
	}
	return nil
}
