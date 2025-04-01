// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/namespace"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/predicate"
	"reflect"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/applicationdependency"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/applicationnamespace"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/artifact"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/artifactreference"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentprofile"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/extension"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateDeploymentPackage creates an CreateDeploymentPackage from gRPC request
func (g *Server) CreateDeploymentPackage(ctx context.Context, req *catalogv3.CreateDeploymentPackageRequest) (*catalogv3.CreateDeploymentPackageResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.DeploymentPackage == nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("incomplete request"))
	} else if err := req.DeploymentPackage.Validate(); err != nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage(err.Error()))
	} else if err := validateDeploymentProfiles(req.DeploymentPackage); err != nil {
		return nil, err
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &DeploymentPackageEvents{}
	created, err := g.createDeploymentPackage(ctx, tx, projectUUID, req.DeploymentPackage, events)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "created", "deployment-package", projectUUID, req.DeploymentPackage.Name, req.DeploymentPackage.Version)
	events.sendToAll(g.listeners)

	pkg := req.DeploymentPackage
	return &catalogv3.CreateDeploymentPackageResponse{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name:                       created.Name,
			DisplayName:                created.DisplayName,
			Description:                created.Description,
			Version:                    created.Version,
			IsVisible:                  created.IsVisible,
			IsDeployed:                 created.IsDeployed,
			ApplicationReferences:      pkg.ApplicationReferences,
			ApplicationDependencies:    pkg.ApplicationDependencies,
			DefaultNamespaces:          pkg.DefaultNamespaces,
			Namespaces:                 pkg.Namespaces,
			Profiles:                   pkg.Profiles,
			DefaultProfileName:         pkg.DefaultProfileName,
			Extensions:                 pkg.Extensions,
			Artifacts:                  pkg.Artifacts,
			ForbidsMultipleDeployments: pkg.ForbidsMultipleDeployments,
			Kind:                       kindFromDB(created.Kind),
			CreateTime:                 timestamppb.New(created.CreateTime),
		},
	}, nil
}

func validateDeploymentProfiles(pkg *catalogv3.DeploymentPackage) error {
	// If there are no deployment profiles or if there is exactly one app in the package, bail with no issue
	if len(pkg.Profiles) == 0 || len(pkg.ApplicationReferences) == 1 {
		return nil
	}

	// See if all deployment packages use fully qualified app names
	fqan := true
	for _, dp := range pkg.Profiles {
		if fqan {
			for an := range dp.ApplicationProfiles {
				if !strings.Contains(an, ":") {
					fqan = false
					break
				}
			}
		}
	}

	// If all deployment packages use fully qualified app names, bail with no issue - for now
	if fqan {
		return nil
	}

	// Otherwise, check if there are any duplicate app names
	if hasDuplicateAppNames(pkg.ApplicationReferences) {
		return errors.NewInvalidArgument(errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("package %s contains duplicate application names, "+
				"but does not use fully qualified profile references", pkg.Name))
	}
	return nil
}

func hasDuplicateAppNames(refs []*catalogv3.ApplicationReference) bool {
	appNames := make(map[string]string, len(refs))
	for _, ar := range refs {
		appNames[ar.Name] = ar.Name
	}
	return len(appNames) != len(refs)
}

func (g *Server) createDeploymentPackage(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage, events *DeploymentPackageEvents) (*generated.DeploymentPackage, error) {
	if len(pkg.Profiles) > 0 && pkg.DefaultProfileName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("default profile name must be specified"))
	}

	displayName, ok := validateDisplayName(pkg.Name, pkg.DisplayName)
	if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkDeploymentPackageUniqueness(ctx, tx, projectUUID, pkg); err != nil {
		return nil, err
	}

	stmt := tx.DeploymentPackage.Create().
		SetProjectUUID(projectUUID).
		SetName(pkg.Name).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(pkg.Description).
		SetVersion(pkg.Version).
		SetIsVisible(pkg.IsVisible).
		SetIsDeployed(pkg.IsDeployed).
		SetAllowsMultipleDeployments(!pkg.ForbidsMultipleDeployments).
		SetKind(kindToDB(pkg.Kind))

	created, err := stmt.Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("deployment package %s already exists", pkg.Name))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}

	// Create any application references
	if err = g.createApplicationReferences(ctx, tx, projectUUID, pkg, created); err != nil {
		return nil, err
	}

	// Create any application dependencies
	if err = g.createApplicationDependencies(ctx, tx, pkg, created); err != nil {
		return nil, err
	}

	// Create any default application namespaces
	if err = g.createApplicationNamespaces(ctx, tx, pkg, created); err != nil {
		return nil, err
	}

	// Create any prescribed namespaces
	if err = g.createNamespaces(ctx, tx, pkg, created); err != nil {
		return nil, err
	}

	// Create any deployment profiles and record a default profile
	if err = g.createDeploymentProfiles(ctx, tx, pkg.Profiles, pkg, created); err != nil {
		return nil, err
	}

	// Update the default profile
	if err = g.updateDefaultDeploymentProfile(ctx, tx, projectUUID, pkg.DefaultProfileName, pkg.Name, pkg.Version); err != nil {
		return nil, err
	}

	// Create any extensions
	if err = g.createExtensions(ctx, tx, pkg, created); err != nil {
		return nil, err
	}

	// Create any artifact references
	if err = g.createArtifactReferences(ctx, tx, projectUUID, pkg, created); err != nil {
		return nil, err
	}

	events.append(CreatedEvent, projectUUID, pkg)
	return created, nil
}

// Returns an error if the deployment package display name is not unique
func (g *Server) checkDeploymentPackageUniqueness(ctx context.Context, tx *generated.Tx, projectUUID string, a *catalogv3.DeploymentPackage) error {
	if len(a.DisplayName) > 0 {
		existing, err := tx.DeploymentPackage.Query().
			Where(
				deploymentpackage.ProjectUUID(projectUUID),
				deploymentpackage.DisplayNameLc(strings.ToLower(a.DisplayName)),
				deploymentpackage.Not(deploymentpackage.Name(a.Name))).
			Count(ctx)
		if err = checkUniqueness(existing, err, "deployment package", a.Name, a.DisplayName, errors.DeploymentPackageType); err != nil {
			return err
		}
	}
	return nil
}

// Returns an error if the deployment package is deployed
func (g *Server) checkDeploymentPackageNotDeployed(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage) error {
	first, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.And(deploymentpackage.Name(pkg.Name)),
			deploymentpackage.And(deploymentpackage.Version(pkg.Version))).First(ctx)

	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}
		return errors.NewDBError(errors.WithError(err))
	}

	if first.IsDeployed && pkg.IsDeployed {
		return errors.NewFailedPrecondition(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceName(pkg.Name),
			errors.WithResourceVersion(pkg.Version),
			errors.WithMessage("cannot modify deployed package"))
	}
	return nil
}

func (g *Server) createApplicationReferences(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	appIDs := make([]uint64, 0)
	for _, appRef := range pkg.ApplicationReferences {
		appID, err := tx.Application.Query().
			Where(
				application.ProjectUUID(projectUUID),
				application.Name(appRef.Name),
				application.Version(appRef.Version),
			).FirstID(ctx)
		appIDs = append(appIDs, appID)
		if err != nil {
			if generated.IsNotFound(err) {
				return errors.NewInvalidArgument(
					errors.WithResourceType(errors.ApplicationReferenceType),
					errors.WithMessage("application reference not found"),
					errors.WithResourceName(appRef.Name),
					errors.WithResourceVersion(appRef.Version))
			}
			return errors.NewDBError(errors.WithError(err))
		}
	}
	err := pkgDB.Update().ClearApplications().AddApplicationIDs(appIDs...).Exec(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return nil
}

// Create any deployment profiles and record a default profile
func (g *Server) createDeploymentProfiles(ctx context.Context, tx *generated.Tx, profiles []*catalogv3.DeploymentProfile, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	var defaultProfile *generated.DeploymentProfile
	for _, profile := range profiles {
		createdProfile, err := g.injectDeploymentProfile(ctx, tx, profile, pkgDB)
		if err != nil {
			return err
		}
		if pkg.DefaultProfileName != "" && pkg.DefaultProfileName == profile.Name {
			defaultProfile = createdProfile
		}
	}

	// Retroactively update the deployment package with its default profile, if there was one
	if defaultProfile != nil {
		_, err := tx.DeploymentPackage.Update().Where(deploymentpackage.ID(pkgDB.ID)).SetDefaultProfile(defaultProfile).Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	}
	return nil
}

// Create any application dependencies
func (g *Server) createApplicationDependencies(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	ars := make(map[string]*generated.Application, 0)
	appsDB, err := pkgDB.QueryApplications().All(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	for _, ar := range appsDB {
		ars[ar.Name] = ar
	}

	for _, appDep := range pkg.ApplicationDependencies {
		if appDep.Name == appDep.Requires {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("application %s cannot depend on itself", appDep.Name))
		}
		source, ok := ars[appDep.Name]
		if !ok {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("dependency for application %s does not exist", appDep.Name))
		}
		target, ok := ars[appDep.Requires]
		if !ok {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("dependency requirement %s does not exist", appDep.Requires))
		}
		_, err := tx.ApplicationDependency.Create().
			SetDeploymentPackageFkID(pkgDB.ID).
			SetSourceFk(source).
			SetTargetFk(target).
			Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	}
	return nil
}

// Create any application namespaces
func (g *Server) createApplicationNamespaces(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	namespaces := make(map[string]*generated.Application, 0)
	appsDB, err := pkgDB.QueryApplications().All(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	for _, aDB := range appsDB {
		namespaces[aDB.Name] = aDB
	}

	for appName, appNamespace := range pkg.DefaultNamespaces {
		aDB, ok := namespaces[appName]
		if !ok {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("application %s does not exist", appName))
		}
		_, err := tx.ApplicationNamespace.Create().
			SetDeploymentPackageFkID(pkgDB.ID).
			SetSourceFk(aDB).
			SetNamespace(appNamespace).
			Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	}
	return nil
}

const (
	namespaceLabelType      = "label"
	namespaceAnnotationType = "annotation"
)

// Create any prescribed namespaces
func (g *Server) createNamespaces(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	for _, namespace := range pkg.Namespaces {
		nsDB, err := tx.Namespace.Create().
			SetDeploymentPackageFkID(pkgDB.ID).
			SetName(namespace.Name).
			Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}

		for key, value := range namespace.Labels {
			_, err := tx.NamespaceAdornment.Create().SetNamespaceFkID(nsDB.ID).SetType(namespaceLabelType).
				SetKey(key).SetValue(value).Save(ctx)
			if err != nil {
				return errors.NewDBError(errors.WithError(err))
			}
		}

		for key, value := range namespace.Annotations {
			_, err := tx.NamespaceAdornment.Create().SetNamespaceFkID(nsDB.ID).SetType(namespaceAnnotationType).
				SetKey(key).SetValue(value).Save(ctx)
			if err != nil {
				return errors.NewDBError(errors.WithError(err))
			}
		}
	}
	return nil
}

func (g *Server) createExtensions(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	for _, ext := range pkg.Extensions {
		stmt := tx.Extension.Create().
			SetDeploymentPackageFkID(pkgDB.ID).
			SetName(ext.Name).
			SetVersion(ext.Version).
			SetDisplayName(ext.DisplayName).
			SetDescription(ext.Description)
		if ext.UiExtension != nil {
			stmt.SetUILabel(ext.UiExtension.Label).
				SetUIServiceName(ext.UiExtension.ServiceName).
				SetUIDescription(ext.UiExtension.Description).
				SetUIFileName(ext.UiExtension.FileName).
				SetUIAppName(ext.UiExtension.AppName).
				SetUIModuleName(ext.UiExtension.ModuleName)
		}

		extDB, err := stmt.Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}

		for _, ep := range ext.Endpoints {
			_, err = tx.Endpoint.Create().
				SetExtensionFkID(extDB.ID).
				SetServiceName(ep.ServiceName).
				SetExternalPath(ep.ExternalPath).
				SetInternalPath(ep.InternalPath).
				SetScheme(ep.Scheme).
				SetAuthType(ep.AuthType).
				SetAppName(ep.AppName).
				Save(ctx)
			if err != nil {
				return errors.NewDBError(errors.WithError(err))
			}
		}
	}
	return nil
}

func (g *Server) createArtifactReferences(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	for _, art := range pkg.Artifacts {
		artifactDB, err := tx.Artifact.Query().
			Where(
				artifact.ProjectUUID(projectUUID),
				artifact.Name(art.Name),
			).Only(ctx)
		if err != nil {
			if generated.IsNotFound(err) {
				return errors.NewInvalidArgument(
					errors.WithResourceType(errors.DeploymentPackageType),
					errors.WithResourceName(pkg.Name),
					errors.WithResourceVersion(pkg.Version),
					errors.WithMessage("artifact %s not found", art.Name))
			}
			return errors.NewDBError(errors.WithError(err))
		}
		_, err = tx.ArtifactReference.Create().
			SetDeploymentPackageFkID(pkgDB.ID).
			SetPurpose(art.Purpose).
			SetArtifact(artifactDB).
			Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	}
	return nil
}

// ListDeploymentPackages gets a list of all deployment-packages through gRPC
func (g *Server) ListDeploymentPackages(ctx context.Context, req *catalogv3.ListDeploymentPackagesRequest) (*catalogv3.ListDeploymentPackagesResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	orderBys, err := parseOrderBy(req.OrderBy, errors.DeploymentPackageType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	filters, err := parseFilter(req.Filter, errors.DeploymentPackageType)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	packages, _, totalElements, err := g.getDeploymentPackages(ctx, tx, projectUUID, req.Kinds, orderBys, filters, req.PageSize, req.Offset)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "listed", "deployment-packages", projectUUID, "Total => "+fmt.Sprintf("%d", totalElements))
	return &catalogv3.ListDeploymentPackagesResponse{DeploymentPackages: packages, TotalElements: totalElements}, nil
}

var dpColumns = map[string]string{
	"publisherName": "",
	"name":          "name",
	"displayName":   "display_name",
	"description":   "description",
	"version":       "version",
	"createTime":    "create_time",
	"updateTime":    "update_time",
}

func (g *Server) getDeploymentPackages(ctx context.Context, tx *generated.Tx, projectUUID string, kinds []catalogv3.Kind,
	orderBys []*orderBy, filters []*filter, pageSize int32, offset int32) ([]*catalogv3.DeploymentPackage, []string, int32, error) {
	var err error
	var pkgsDB []*generated.DeploymentPackage
	var orderOptions []deploymentpackage.OrderOption
	dpQuery := tx.DeploymentPackage.Query()

	options, err := orderByOptions(orderBys, dpColumns, errors.DeploymentPackageType)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, pred := range options {
		orderOptions = append(orderOptions, pred)
	}
	dpQuery = dpQuery.Order(orderOptions...)

	filterPreds, err := filterPredicates(filters, dpColumns, errors.DeploymentPackageType)
	if err != nil {
		return nil, nil, 0, err
	}
	var dpPreds []predicate.DeploymentPackage
	for _, pred := range filterPreds {
		dpPreds = append(dpPreds, pred)
	}
	dpQuery = dpQuery.Where(deploymentpackage.Or(dpPreds...))

	kindFilter := kindPredicate(kinds)
	if kindFilter != nil {
		dpQuery = dpQuery.Where(kindFilter)
	}

	if projectUUID == "" {
		pkgsDB, err = dpQuery.All(ctx)
	} else {
		pkgsDB, err = dpQuery.Where(deploymentpackage.ProjectUUID(projectUUID)).All(ctx)
	}
	if err != nil {
		return nil, nil, 0, errors.NewDBError(errors.WithError(err))
	}

	totalElements := int32(len(pkgsDB))
	startIndex, endIndex, _, err := computePageRange(pageSize, offset, len(pkgsDB))
	if err != nil {
		return nil, nil, 0, err
	}
	if len(pkgsDB) == 0 {
		return []*catalogv3.DeploymentPackage{}, []string{}, 0, nil
	}

	packages := make([]*catalogv3.DeploymentPackage, 0, len(pkgsDB))
	projectUUIDs := make([]string, 0, len(pkgsDB))

	for i := startIndex; i <= endIndex; i++ {
		pkgDB := pkgsDB[i]
		pkg, err := extractDeploymentPackage(ctx, pkgDB)
		if err != nil {
			return nil, nil, 0, err
		}
		packages = append(packages, pkg)
		projectUUIDs = append(projectUUIDs, pkgDB.ProjectUUID)
	}
	return packages, projectUUIDs, totalElements, nil
}

func extractDeploymentPackage(ctx context.Context, pkgDB *generated.DeploymentPackage) (*catalogv3.DeploymentPackage, error) {
	applications, err := extractApplicationReferences(ctx, pkgDB)
	if err != nil {
		return nil, err
	}

	duplicateAppNames := hasDuplicateAppNames(applications)

	profiles, err := extractDeploymentProfiles(ctx, pkgDB, duplicateAppNames, true)
	if err != nil {
		return nil, err
	}

	applicationDependencies, err := extractApplicationDependencies(ctx, pkgDB)
	if err != nil {
		return nil, err
	}

	applicationNamespaces, err := extractApplicationNamespaces(ctx, pkgDB)
	if err != nil {
		return nil, err
	}

	namespaces, err := extractNamespaces(ctx, pkgDB)
	if err != nil {
		return nil, err
	}

	extensions, err := extractExtensions(ctx, pkgDB)
	if err != nil {
		return nil, err
	}

	artifacts, err := extractArtifactReferences(ctx, pkgDB)
	if err != nil {
		return nil, err
	}

	dp := &catalogv3.DeploymentPackage{
		Name:                       pkgDB.Name,
		DisplayName:                pkgDB.DisplayName,
		Description:                pkgDB.Description,
		Version:                    pkgDB.Version,
		IsVisible:                  pkgDB.IsVisible,
		IsDeployed:                 pkgDB.IsDeployed,
		ApplicationReferences:      applications,
		ApplicationDependencies:    applicationDependencies,
		DefaultNamespaces:          applicationNamespaces,
		Namespaces:                 namespaces,
		Profiles:                   profiles,
		Extensions:                 extensions,
		Artifacts:                  artifacts,
		ForbidsMultipleDeployments: !pkgDB.AllowsMultipleDeployments,
		Kind:                       kindFromDB(pkgDB.Kind),
		CreateTime:                 timestamppb.New(pkgDB.CreateTime),
		UpdateTime:                 timestamppb.New(pkgDB.UpdateTime),
	}

	// Fetch default profile for this deployment package
	profileDB, err := pkgDB.QueryDefaultProfile().Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			// If we did not find a default profile and there is only one profile, it should be the implicit default
			if len(profiles) == 1 {
				dp.DefaultProfileName = profiles[0].Name
			}
		} else {
			return nil, errors.NewDBError(errors.WithError(err))
		}
	} else {
		dp.DefaultProfileName = profileDB.Name
	}

	return dp, nil
}

// Fetch application references for this deployment package
func extractApplicationReferences(ctx context.Context, pkgDB *generated.DeploymentPackage) ([]*catalogv3.ApplicationReference, error) {
	applicationsDB, err := pkgDB.QueryApplications().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	applications := make([]*catalogv3.ApplicationReference, 0, len(applicationsDB))
	for _, appApplicationDB := range applicationsDB {
		applications = append(applications, &catalogv3.ApplicationReference{
			Name:    appApplicationDB.Name,
			Version: appApplicationDB.Version,
		})
	}
	return applications, err
}

// Fetch deployment profiles for this deployment package
func extractDeploymentProfiles(ctx context.Context, pkgDB *generated.DeploymentPackage, useFQNames bool, includeSyntheticDefault bool) ([]*catalogv3.DeploymentProfile, error) {
	profilesDB, err := pkgDB.QueryDeploymentProfiles().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	profiles := make([]*catalogv3.DeploymentProfile, 0, len(profilesDB))
	if len(profilesDB) > 0 {
		for _, profileDB := range profilesDB {
			// Produce application -> profile name map
			applicationProfiles, err := composeAppProfileMap(ctx, profileDB, useFQNames)
			if err != nil {
				return nil, err
			}

			profiles = append(profiles, &catalogv3.DeploymentProfile{
				Name:                profileDB.Name,
				DisplayName:         profileDB.DisplayName,
				Description:         profileDB.Description,
				ApplicationProfiles: applicationProfiles,
				CreateTime:          timestamppb.New(profileDB.CreateTime),
				UpdateTime:          timestamppb.New(profileDB.UpdateTime),
			})
		}
	} else if includeSyntheticDefault {
		defaultProfile, err := implicitDefaultDeploymentProfile(ctx, pkgDB)
		if err != nil {
			return nil, err
		}

		// If there is an implicit default profile, add it to the profiles
		if defaultProfile != nil {
			profiles = append(profiles, defaultProfile)
		}
	}
	return profiles, nil
}

// Fetch app dependencies for the given deployment package
func extractApplicationDependencies(ctx context.Context, pkgDB *generated.DeploymentPackage) ([]*catalogv3.ApplicationDependency, error) {
	dependenciesDB, err := pkgDB.QueryApplicationDependencies().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	dependencies := make([]*catalogv3.ApplicationDependency, 0, len(dependenciesDB))
	for _, dependencyDB := range dependenciesDB {
		source, err := dependencyDB.QuerySourceFk().Only(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		target, err := dependencyDB.QueryTargetFk().Only(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		dependencies = append(dependencies, &catalogv3.ApplicationDependency{Name: source.Name, Requires: target.Name})
	}
	return dependencies, nil
}

// Fetch app default namespaces for the given deployment package
func extractApplicationNamespaces(ctx context.Context, pkgDB *generated.DeploymentPackage) (map[string]string, error) {
	namespacesDB, err := pkgDB.QueryApplicationNamespaces().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	namespaces := make(map[string]string, len(namespacesDB))
	for _, namespaceDB := range namespacesDB {
		sourceDB, err := namespaceDB.QuerySourceFk().Only(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		namespaces[sourceDB.Name] = namespaceDB.Namespace
	}
	return namespaces, nil
}

// Fetch prescribed namespaces for the given deployment package
func extractNamespaces(ctx context.Context, pkgDB *generated.DeploymentPackage) ([]*catalogv3.Namespace, error) {
	namespacesDB, err := pkgDB.QueryNamespaces().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	namespaces := make([]*catalogv3.Namespace, 0, len(namespacesDB))
	for _, namespaceDB := range namespacesDB {
		namespace := &catalogv3.Namespace{Name: namespaceDB.Name, Labels: map[string]string{}, Annotations: map[string]string{}}
		adornmentsDB, err := namespaceDB.QueryAdornments().All(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		for _, adornmentDB := range adornmentsDB {
			if adornmentDB.Type == namespaceLabelType {
				namespace.Labels[adornmentDB.Key] = adornmentDB.Value
			}
			if adornmentDB.Type == namespaceAnnotationType {
				namespace.Annotations[adornmentDB.Key] = adornmentDB.Value
			}
		}
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

// Extract extensions for this specified deployment package
func extractExtensions(ctx context.Context, pkgDB *generated.DeploymentPackage) ([]*catalogv3.APIExtension, error) {
	extensionsDB, err := pkgDB.QueryExtensions().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	extensions := make([]*catalogv3.APIExtension, 0, len(extensionsDB))
	for _, extDB := range extensionsDB {
		ext := &catalogv3.APIExtension{
			Name:        extDB.Name,
			Version:     extDB.Version,
			DisplayName: extDB.DisplayName,
			Description: extDB.Description,
		}
		if extDB.UILabel != "" {
			ext.UiExtension = &catalogv3.UIExtension{
				Label:       extDB.UILabel,
				ServiceName: extDB.UIServiceName,
				Description: extDB.UIDescription,
				FileName:    extDB.UIFileName,
				AppName:     extDB.UIAppName,
				ModuleName:  extDB.UIModuleName,
			}
		}

		epsDB, err := extDB.QueryEndpoints().All(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}

		eps := make([]*catalogv3.Endpoint, 0, len(epsDB))
		for _, epDB := range epsDB {
			eps = append(eps, &catalogv3.Endpoint{
				ServiceName:  epDB.ServiceName,
				ExternalPath: epDB.ExternalPath,
				InternalPath: epDB.InternalPath,
				Scheme:       epDB.Scheme,
				AuthType:     epDB.AuthType,
				AppName:      epDB.AppName,
			})
		}
		ext.Endpoints = eps

		extensions = append(extensions, ext)
	}
	return extensions, nil
}

// Extract artifact references for this specified deployment package
func extractArtifactReferences(ctx context.Context, pkgDB *generated.DeploymentPackage) ([]*catalogv3.ArtifactReference, error) {
	artifactsDB, err := pkgDB.QueryArtifacts().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	artifacts := make([]*catalogv3.ArtifactReference, 0, len(artifactsDB))
	for _, refDB := range artifactsDB {
		artifactDB, err := refDB.QueryArtifact().Only(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		artifacts = append(artifacts, &catalogv3.ArtifactReference{
			Name:    artifactDB.Name,
			Purpose: refDB.Purpose,
		})
	}
	return artifacts, nil
}

// GetDeploymentPackageVersions gets all versions of a named DeploymentPackage through gRPC
func (g *Server) GetDeploymentPackageVersions(ctx context.Context, req *catalogv3.GetDeploymentPackageVersionsRequest) (*catalogv3.GetDeploymentPackageVersionsResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.DeploymentPackageName == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	pkgsDB, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.Name(req.DeploymentPackageName),
		).
		All(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	}
	if len(pkgsDB) == 0 {
		g.rollbackTransaction(tx)
		return nil, errors.NewNotFound(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceName(req.DeploymentPackageName))
	}

	pkgs := make([]*catalogv3.DeploymentPackage, 0, len(pkgsDB))
	for _, pkgDB := range pkgsDB {
		dp, err := extractDeploymentPackage(ctx, pkgDB)
		if err != nil {
			g.rollbackTransaction(tx)
			return nil, err
		}
		pkgs = append(pkgs, dp)
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "got all version of", "deployment-package", projectUUID, "name => "+req.DeploymentPackageName,
		"total versions => "+fmt.Sprintf("%d", len(pkgs)))
	return &catalogv3.GetDeploymentPackageVersionsResponse{DeploymentPackages: pkgs}, nil
}

// GetDeploymentPackage gets a single application through gRPC
func (g *Server) GetDeploymentPackage(ctx context.Context, req *catalogv3.GetDeploymentPackageRequest) (*catalogv3.GetDeploymentPackageResponse, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.DeploymentPackageName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	pkgDB, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.Name(req.DeploymentPackageName),
			deploymentpackage.Version(req.Version),
		).
		Only(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		if generated.IsNotFound(err) {
			return nil, errors.NewNotFound(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(req.DeploymentPackageName),
				errors.WithResourceVersion(req.Version))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}

	ca, err := extractDeploymentPackage(ctx, pkgDB)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	logActivity(ctx, "got", "deployment-package", projectUUID, req.DeploymentPackageName, req.Version)
	return &catalogv3.GetDeploymentPackageResponse{DeploymentPackage: ca}, nil
}

type packageChanges struct {
	kindOrState       bool
	rootRecord        bool
	applications      bool
	profiles          bool
	profile           bool
	dependencies      bool
	defaultNamespaces bool
	namespaces        bool
	extensions        bool
	artifacts         bool
	newProfiles       []*catalogv3.DeploymentProfile
}

func (c *packageChanges) changed() bool {
	return c.rootRecord || c.applications || c.profiles || c.profile || c.dependencies || c.defaultNamespaces || c.namespaces || c.extensions || c.artifacts
}

func (c *packageChanges) changedKindOrDeployedState() bool {
	return c.kindOrState
}

// UpdateDeploymentPackage updates an application through gRPC
func (g *Server) UpdateDeploymentPackage(ctx context.Context, req *catalogv3.UpdateDeploymentPackageRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.DeploymentPackage == nil || req.DeploymentPackageName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("incomplete request"))
	} else if err := req.DeploymentPackage.Validate(); err != nil {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage(err.Error()))
	} else if req.DeploymentPackageName != req.DeploymentPackage.Name {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("name cannot be changed %s != %s", req.DeploymentPackageName, req.DeploymentPackage.Name))
	} else if req.Version != req.DeploymentPackage.Version {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("version cannot be changed %s != %s", req.Version, req.DeploymentPackage.Version))
	} else if err := validateDeploymentProfiles(req.DeploymentPackage); err != nil {
		return nil, err
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	events := &DeploymentPackageEvents{}
	if err = g.updateDeploymentPackage(ctx, tx, projectUUID, req.DeploymentPackage, events); err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	err = g.commitTransaction(tx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	logActivity(ctx, "updated", "deployment-package", projectUUID, req.DeploymentPackageName, req.Version)
	events.sendToAll(g.listeners)

	return &emptypb.Empty{}, nil
}

func (g *Server) updateDeploymentPackage(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage, events *DeploymentPackageEvents) error {
	if len(pkg.Profiles) > 0 && pkg.DefaultProfileName == "" {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("default profile name must be specified"))
	}

	displayName, ok := validateDisplayName(pkg.Name, pkg.DisplayName)
	if !ok {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceName(pkg.Name),
			errors.WithResourceVersion(pkg.Version),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	pkgDB, ok, err := g.getDeploymentPackage(ctx, tx, projectUUID, pkg.Name, pkg.Version)
	if err != nil {
		return err
	} else if !ok {
		return errors.NewNotFound(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceVersion(pkg.Version),
			errors.WithResourceName(pkg.Name))
	}

	changes, err := g.computePackageChanges(ctx, pkg, pkgDB)
	if err != nil {
		return err
	}

	// If there are any changes (other than changing the isDeployed bit)...
	// Changes to the kind field only are exempt.
	if changes.changedKindOrDeployedState() && !changes.changed() {
		return g.updatePackageKindOrDeployedState(ctx, tx, projectUUID, pkg)
	} else if changes.changed() {
		// Make sure that CA is not already deployed
		if err := g.checkDeploymentPackageNotDeployed(ctx, tx, projectUUID, pkg); err != nil {
			return err
		}
	}

	// Make sure that the display name, if specified is unique
	if err := g.checkDeploymentPackageUniqueness(ctx, tx, "", pkg); err != nil {
		return err
	}

	updateCount, err := tx.DeploymentPackage.Update().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.Name(pkg.Name),
			deploymentpackage.Version(pkg.Version),
		).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(pkg.Description).
		SetVersion(pkg.Version).
		SetIsVisible(pkg.IsVisible).
		SetIsDeployed(pkg.IsDeployed).
		SetAllowsMultipleDeployments(!pkg.ForbidsMultipleDeployments).
		SetKind(kindToDB(pkg.Kind)).
		Save(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	} else if updateCount == 0 {
		return errors.NewNotFound(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceName(pkg.Name),
			errors.WithResourceVersion(pkg.Version))
	}

	// Update the application references, if necessary
	if changes.applications {
		if err = g.createApplicationReferences(ctx, tx, projectUUID, pkg, pkgDB); err != nil {
			return err
		}
	}

	// Update the app dependencies, if necessary
	if changes.dependencies || changes.applications {
		if err = g.updateApplicationDependencies(ctx, tx, pkg, pkgDB); err != nil {
			return err
		}
	}

	// Update the app namespaces, if necessary
	if changes.defaultNamespaces || changes.applications {
		if err = g.updateApplicationNamespaces(ctx, tx, pkg, pkgDB); err != nil {
			return err
		}
	}

	// Update the app namespaces, if necessary
	if changes.namespaces {
		if err = g.updateNamespaces(ctx, tx, pkg, pkgDB); err != nil {
			return err
		}
	}

	// Update the deployment profiles, if necessary
	if changes.profiles || changes.applications {
		if err = g.updateDeploymentProfiles(ctx, tx, pkg, pkgDB); err != nil {
			return err
		}
	} else {
		if len(changes.newProfiles) > 0 {
			if err = g.createDeploymentProfiles(ctx, tx, changes.newProfiles, pkg, pkgDB); err != nil {
				return err
			}
		}
	}

	// Update the default profile
	if changes.profile || changes.profiles {
		if err = g.updateDefaultDeploymentProfile(ctx, tx, projectUUID, pkg.DefaultProfileName, pkg.Name, pkg.Version); err != nil {
			return err
		}
	}

	// Update the extensions, if necessary
	if changes.extensions {
		if err = g.updateExtensions(ctx, tx, pkg, pkgDB); err != nil {
			return err
		}
	}

	// Update the artifacts, if necessary
	if changes.artifacts {
		if err = g.updateArtifactReferences(ctx, tx, projectUUID, pkg, pkgDB); err != nil {
			return err
		}
	}

	events.append(UpdatedEvent, projectUUID, pkg)
	return nil
}

func (g *Server) updatePackageKindOrDeployedState(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage) error {
	updateCount, err := tx.DeploymentPackage.Update().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.Name(pkg.Name),
			deploymentpackage.Version(pkg.Version),
		).
		SetIsDeployed(pkg.IsDeployed).
		SetKind(kindToDB(pkg.Kind)).
		Save(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	} else if updateCount == 0 {
		return errors.NewNotFound(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceName(pkg.Name),
			errors.WithResourceVersion(pkg.Version))
	}
	return nil
}

func (g *Server) computePackageChanges(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (*packageChanges, error) {
	var err error
	changes := &packageChanges{}
	changes.kindOrState = !isSameKind(pkg.Kind, pkgDB.Kind) || pkg.IsDeployed != pkgDB.IsDeployed
	changes.rootRecord = g.deploymentPackageChanged(pkg, pkgDB)

	if changes.applications, err = g.applicationReferencesChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.profiles, changes.newProfiles, err = g.deploymentProfilesChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.profile, err = g.defaultDeploymentProfileChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.dependencies, err = g.applicationDependenciesChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.defaultNamespaces, err = g.applicationNamespacesChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.namespaces, err = g.namespacesChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.extensions, err = g.extensionsChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	if changes.artifacts, err = g.artifactReferencesChanged(ctx, pkg, pkgDB); err != nil {
		return nil, err
	}
	return changes, nil
}

// Determines if deployment package root record fields have changed - isDeployed is exempt
func (g *Server) deploymentPackageChanged(pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) bool {
	return pkg.DisplayName != pkgDB.DisplayName || pkg.Description != pkgDB.Description || pkg.IsVisible != pkgDB.IsVisible ||
		pkg.ForbidsMultipleDeployments == pkgDB.AllowsMultipleDeployments
}

// Determines if application references have changed
func (g *Server) applicationReferencesChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	appRefs, err := extractApplicationReferences(ctx, pkgDB)
	if err != nil {
		return false, err
	}

	// If number of existing and new references are not the same, bail
	if len(appRefs) != len(pkg.ApplicationReferences) {
		return true, nil
	}

	// Otherwise, look for sameness within
	existingRefs := make(map[string]*catalogv3.ApplicationReference, len(appRefs))
	for _, ref := range appRefs {
		existingRefs[fmt.Sprintf("%s@%s", ref.Name, ref.Version)] = ref
	}

	for _, ref := range pkg.ApplicationReferences {
		if existingRef, ok := existingRefs[fmt.Sprintf("%s@%s", ref.Name, ref.Version)]; !ok || !appReferencesAreSame(ref, existingRef) {
			return true, nil // found a difference, so bail
		}
	}
	return false, nil
}

func appReferencesAreSame(a, b *catalogv3.ApplicationReference) bool {
	return a.Name == b.Name && a.Version == b.Version
}

// Determines if deployment profiles have changed
func (g *Server) deploymentProfilesChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, []*catalogv3.DeploymentProfile, error) {
	profiles, err := extractDeploymentProfiles(ctx, pkgDB, false, false)
	if err != nil {
		return false, nil, err
	}

	// If number of existing and new profiles are not the same bail
	if len(profiles) > len(pkg.Profiles) {
		return true, nil, nil
	}

	// Otherwise, look for sameness within
	existingProfiles := make(map[string]*catalogv3.DeploymentProfile, len(profiles))
	for _, profile := range profiles {
		existingProfiles[profile.Name] = profile
	}

	newProfiles := make([]*catalogv3.DeploymentProfile, 0)
	for _, p := range pkg.Profiles {
		if existingProfile, ok := existingProfiles[p.Name]; ok {
			if !deploymentProfilesAreSame(p, existingProfile) {
				return true, nil, nil // found a difference, so bail
			}
			delete(existingProfiles, existingProfile.Name)
		} else {
			newProfiles = append(newProfiles, p)
		}
	}

	// If there are any existing profiles remaining, this means deletion of an existing one was attempted
	if len(existingProfiles) > 0 {
		return true, nil, nil
	}

	return false, newProfiles, nil
}

func (g *Server) updateDeploymentProfiles(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	givenProfiles := make(map[string]*catalogv3.DeploymentProfile, 0)
	displayNames := make(map[string]*catalogv3.DeploymentProfile, 0)
	for _, p := range pkg.Profiles {
		if _, ok := givenProfiles[p.Name]; ok {
			return errors.NewAlreadyExists(
				errors.WithResourceType(errors.DeploymentProfileType),
				errors.WithResourceName(p.Name))
		}
		if _, ok := displayNames[strings.ToLower(p.DisplayName)]; ok {
			return errors.NewAlreadyExists(
				errors.WithResourceType(errors.DeploymentProfileType),
				errors.WithResourceName(p.Name),
				errors.WithMessage("deployment profile %s display name %s is not unique", p.Name, p.DisplayName))
		}
		givenProfiles[p.Name] = p
		displayNames[strings.ToLower(p.DisplayName)] = p
	}

	// Iterate over the existing profiles in the database and find those that are not in the new set, i.e. should be deleted
	profilesDB, err := pkgDB.QueryDeploymentProfiles().All(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}

	// Scan over the existing profiles to make sure we're not attempting to delete any that have pending references
	// Otherwise either delete them or update them using what whas provided first.
	for _, profileDB := range profilesDB {
		if profile, ok := givenProfiles[profileDB.Name]; !ok {
			if err = tx.DeploymentProfile.DeleteOneID(profileDB.ID).Exec(ctx); err != nil {
				return errors.NewDBError(errors.WithError(err))
			}
		} else {
			if err = g.updateDeploymentProfile(ctx, tx, profile, profileDB, pkgDB); err != nil {
				return err
			}
		}

		// Expunge the profile from the new profiles as we do not need to create it later...
		delete(givenProfiles, profileDB.Name)
	}

	// Finally, convert the map of new profiles to a list and create them
	newProfiles := make([]*catalogv3.DeploymentProfile, 0, len(givenProfiles))
	for _, p := range givenProfiles {
		newProfiles = append(newProfiles, p)
	}
	return g.createDeploymentProfiles(ctx, tx, newProfiles, pkg, pkgDB)
}

func deploymentProfilesAreSame(a, b *catalogv3.DeploymentProfile) bool {
	return a.Name == b.Name && a.DisplayName == b.DisplayName && a.Description == b.Description &&
		reflect.DeepEqual(a.ApplicationProfiles, b.ApplicationProfiles)
}

// Determines if application dependencies have changed
func (g *Server) applicationDependenciesChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	appDeps, err := extractApplicationDependencies(ctx, pkgDB)
	if err != nil {
		return false, err
	}

	// If number of existing and new dependencies are not the same, bail.
	if len(appDeps) != len(pkg.ApplicationDependencies) {
		return true, nil
	}

	// Otherwise, look for sameness within
	existingDeps := make(map[string]*catalogv3.ApplicationDependency, len(appDeps))
	for _, dep := range appDeps {
		existingDeps[fmt.Sprintf("%s/%s", dep.Name, dep.Requires)] = dep
	}

	apps := make(map[string]*catalogv3.ApplicationReference, 0)
	for _, ar := range pkg.ApplicationReferences {
		apps[ar.Name] = ar
	}

	for _, dep := range pkg.ApplicationDependencies {
		if _, ok := existingDeps[fmt.Sprintf("%s/%s", dep.Name, dep.Requires)]; !ok {
			return true, nil // found a difference, so bail
		}

		// Make sure that the existing dependency makes sense in the context of the specified pkg references
		if _, ok := apps[dep.Name]; !ok {
			return false, errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("dependency source %s not found", dep.Name))
		}
		if _, ok := apps[dep.Requires]; !ok {
			return false, errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("dependency target %s not found", dep.Requires))
		}
	}
	return false, nil
}

func (g *Server) updateApplicationDependencies(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	if _, err := tx.ApplicationDependency.Delete().
		Where(applicationdependency.HasDeploymentPackageFkWith(deploymentpackage.ID(pkgDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return g.createApplicationDependencies(ctx, tx, pkg, pkgDB)
}

// Determines if application namespaces have changed
func (g *Server) applicationNamespacesChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	existingNamespaces, err := extractApplicationNamespaces(ctx, pkgDB)
	if err != nil {
		return false, err
	}

	appRefs, err := extractApplicationReferences(ctx, pkgDB)
	if err != nil {
		return false, err
	}
	appRefMap := make(map[string]string)
	for _, appRef := range appRefs {
		appRefMap[appRef.Name] = appRef.Version
	}

	// If number of existing and new namespaces are not the same, bail
	if len(existingNamespaces) != len(pkg.DefaultNamespaces) {
		return true, nil
	}

	// Otherwise, look for sameness within
	for appName, appNamespace := range pkg.DefaultNamespaces {
		if namespace, ok := existingNamespaces[appName]; !ok || appNamespace != namespace {
			return false, errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithResourceName(pkg.Name),
				errors.WithResourceVersion(pkg.Version),
				errors.WithMessage("application %s does not exist", appName))
		}
		if _, appRefOk := appRefMap[appName]; !appRefOk {
			return true, nil // app ref does not exist, so bail
		}
	}
	return false, nil
}

func (g *Server) updateApplicationNamespaces(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	if _, err := tx.ApplicationNamespace.Delete().
		Where(applicationnamespace.HasDeploymentPackageFkWith(deploymentpackage.ID(pkgDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return g.createApplicationNamespaces(ctx, tx, pkg, pkgDB)
}

// Determines if prescribed namespaces have changed
func (g *Server) namespacesChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	namespaces, err := extractNamespaces(ctx, pkgDB)
	if err != nil {
		return false, err
	}

	// If number of existing and new namespaces are not the same, bail
	if len(namespaces) != len(pkg.Namespaces) {
		return true, nil
	}

	// Produce a map of existing namespaces to use for reconciliation
	existingNamespaces := map[string]*catalogv3.Namespace{}
	for _, namespace := range namespaces {
		existingNamespaces[namespace.Name] = namespace
	}

	// Otherwise, look for sameness within
	for _, namespace := range pkg.Namespaces {
		ns, ok := existingNamespaces[namespace.Name]
		if !ok || !reflect.DeepEqual(ns.Labels, namespace.Labels) || !reflect.DeepEqual(ns.Annotations, namespace.Annotations) {
			return true, nil
		}
	}
	return false, nil
}

func (g *Server) updateNamespaces(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	if _, err := tx.Namespace.Delete().
		Where(namespace.HasDeploymentPackageFkWith(deploymentpackage.ID(pkgDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return g.createNamespaces(ctx, tx, pkg, pkgDB)
}

// Determines if default deployment profile has changed
func (g *Server) defaultDeploymentProfileChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	dpDB, err := pkgDB.QueryDefaultProfile().Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) { // If there were no profiles it's OK to have blank default profile
			return pkg.DefaultProfileName != "", nil
		}
		return false, errors.NewDBError(errors.WithError(err))
	}
	return pkg.DefaultProfileName != dpDB.Name, nil
}

func (g *Server) updateDefaultDeploymentProfile(ctx context.Context, tx *generated.Tx, projectUUID string, name string, pkgName string, version string) error {
	if name != "" {
		// Find the named profile in the database.
		dpDB, err := tx.DeploymentProfile.Query().
			Where(
				deploymentprofile.Name(name),
				deploymentprofile.HasDeploymentPackageFkWith(
					deploymentpackage.ProjectUUID(projectUUID),
					deploymentpackage.Name(pkgName),
					deploymentpackage.Version(version),
				),
			).First(ctx)
		if err != nil {
			if generated.IsNotFound(err) {
				return errors.NewInvalidArgument(
					errors.WithResourceType(errors.DeploymentPackageType),
					errors.WithResourceName(pkgName),
					errors.WithResourceVersion(version),
					errors.WithMessage("%s %s not found", errors.DeploymentProfileType, name))
			}
			return errors.NewDBError(errors.WithError(err))
		}
		_, err = tx.DeploymentPackage.Update().
			Where(
				deploymentpackage.ProjectUUID(projectUUID),
				deploymentpackage.Name(pkgName),
				deploymentpackage.Version(version),
			).
			SetDefaultProfileID(dpDB.ID).
			Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	}
	return nil
}

// Determines if extensions have changed
func (g *Server) extensionsChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	extensions, err := extractExtensions(ctx, pkgDB)
	if err != nil {
		return false, err
	}

	// If number of existing and new extensions are not the same, bail
	if len(extensions) != len(pkg.Extensions) {
		return true, nil
	}

	// Otherwise, look for sameness within
	existingExtensions := make(map[string]*catalogv3.APIExtension, len(extensions))
	for _, ext := range extensions {
		existingExtensions[ext.Name] = ext
	}

	for _, ext := range pkg.Extensions {
		if existingExt, ok := existingExtensions[ext.Name]; !ok || !extensionsAreSame(ext, existingExt) {
			return true, nil // found a difference, so bail
		}
	}
	return false, nil
}

func (g *Server) updateExtensions(ctx context.Context, tx *generated.Tx, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	if _, err := tx.Extension.Delete().
		Where(extension.HasDeploymentPackageFkWith(deploymentpackage.ID(pkgDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return g.createExtensions(ctx, tx, pkg, pkgDB)
}

func extensionsAreSame(a, b *catalogv3.APIExtension) bool {
	return a.Name == b.Name && a.Version == b.Version && a.DisplayName == b.DisplayName && a.Description == b.Description &&
		reflect.DeepEqual(a.Endpoints, b.Endpoints) && reflect.DeepEqual(a.UiExtension, b.UiExtension)
}

// Determines if artifact references have changed
func (g *Server) artifactReferencesChanged(ctx context.Context, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) (bool, error) {
	artifacts, err := extractArtifactReferences(ctx, pkgDB)
	if err != nil {
		return false, err
	}

	// If number of existing and new artifacts are not the same, bail.
	if len(artifacts) != len(pkg.Artifacts) {
		return true, nil
	}
	// Otherwise, look for sameness within
	existingArtifacts := make(map[string]*catalogv3.ArtifactReference, len(artifacts))
	for _, ref := range artifacts {
		existingArtifacts[ref.Name] = ref
	}

	for _, ref := range pkg.Artifacts {
		if existingRef, ok := existingArtifacts[ref.Name]; !ok || !artifactsAreSame(ref, existingRef) {
			return true, nil
		}
	}
	return false, nil
}

func (g *Server) updateArtifactReferences(ctx context.Context, tx *generated.Tx, projectUUID string, pkg *catalogv3.DeploymentPackage, pkgDB *generated.DeploymentPackage) error {
	if _, err := tx.ArtifactReference.Delete().
		Where(artifactreference.HasDeploymentPackageFkWith(deploymentpackage.ID(pkgDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return g.createArtifactReferences(ctx, tx, projectUUID, pkg, pkgDB)
}

func artifactsAreSame(a, b *catalogv3.ArtifactReference) bool {
	return a.Name == b.Name && a.Purpose == b.Purpose
}

// DeleteDeploymentPackage deletes an application through gRPC
func (g *Server) DeleteDeploymentPackage(ctx context.Context, req *catalogv3.DeleteDeploymentPackageRequest) (*emptypb.Empty, error) {
	projectUUID, err := GetActiveProjectID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil || req.DeploymentPackageName == "" || req.Version == "" {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(ctx, req); err != nil {
		return nil, err
	}

	tx, err := g.startTransaction(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	// Make sure that CA is not already deployed
	if err := g.checkDeploymentPackageNotDeployed(ctx, tx, projectUUID,
		&catalogv3.DeploymentPackage{
			Name:       req.DeploymentPackageName,
			Version:    req.Version,
			IsDeployed: true,
		}); err != nil {
		g.rollbackTransaction(tx)
		return nil, err
	}

	events := &DeploymentPackageEvents{}
	deleteCount, err := tx.DeploymentPackage.Delete().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.Name(req.DeploymentPackageName),
			deploymentpackage.Version(req.Version),
		).
		Exec(ctx)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewDBError(errors.WithError(err))
	} else if deleteCount == 0 {
		g.rollbackTransaction(tx)
		return nil, errors.NewNotFound(
			errors.WithResourceType(errors.DeploymentPackageType),
			errors.WithResourceName(req.DeploymentPackageName),
			errors.WithResourceVersion(req.Version))
	}
	if _, err = g.checkDeleteResult(ctx, tx, err, fmt.Sprintf("deployment package %s:%s", req.DeploymentPackageName, req.Version), projectUUID); err != nil {
		return nil, err
	}
	events.append(DeletedEvent, projectUUID, &catalogv3.DeploymentPackage{Name: req.DeploymentPackageName, Version: req.Version})
	events.sendToAll(g.listeners)

	return &emptypb.Empty{}, nil
}

// WatchDeploymentPackages watches inventory of deployment packages for changes.
func (g *Server) WatchDeploymentPackages(req *catalogv3.WatchDeploymentPackagesRequest, server catalogv3.CatalogService_WatchDeploymentPackagesServer) error {
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
			errors.WithResourceType(errors.DeploymentProfileType),
			errors.WithMessage("incomplete request"))
	}

	if err := g.authCheckAllowed(server.Context(), req); err != nil {
		return err
	}

	ch := make(chan *catalogv3.WatchDeploymentPackagesResponse)

	// If replay requested
	if !req.NoReplay {
		// Get list of packages
		ctx := server.Context()
		tx, err := g.startTransaction(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}

		deploymentPackages, projectUUIDs, _, err := g.getDeploymentPackages(ctx, tx, projectUUID, req.Kinds, nil, nil, 0, 0)
		if err != nil {
			g.rollbackTransaction(tx)
			return err
		}

		events := &DeploymentPackageEvents{}
		for i, pkg := range deploymentPackages {
			events.append(ReplayedEvent, projectUUIDs[i], pkg)
		}

		// Send each replay event to the stream
		for _, e := range events.queue {
			if err = server.Send(e); err != nil {
				g.rollbackTransaction(tx)
				return err
			}
		}

		// Register the stream, so it can start receiving updates
		g.listeners.addDeploymentPackageListener(ch, req)

		err = g.commitTransaction(tx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	} else {
		// Register the stream, so it can start receiving updates
		g.listeners.addDeploymentPackageListener(ch, req)
	}
	defer g.listeners.deleteDeploymentPackageListener(ch)
	logActivity(server.Context(), "watching", "deployment-packages", projectUUID)
	return g.watchDeploymentPackageEvents(server, ch)
}

func (g *Server) watchDeploymentPackageEvents(server catalogv3.CatalogService_WatchDeploymentPackagesServer, ch chan *catalogv3.WatchDeploymentPackagesResponse) error {
	for e := range ch {
		if err := server.Send(e); err != nil {
			return err
		}
	}
	return nil
}
