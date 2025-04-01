// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentprofile"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/profile"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"strings"
)

func (g *Server) injectDeploymentProfile(ctx context.Context, tx *generated.Tx, profile *catalogv3.DeploymentProfile, pkgDB *generated.DeploymentPackage) (*generated.DeploymentProfile, error) {
	displayName, ok := validateDisplayName(profile.Name, profile.DisplayName)
	if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentProfileType),
			errors.WithResourceName(profile.Name),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	if err := g.checkDeploymentProfileUniqueness(ctx, profile, pkgDB); err != nil {
		return nil, err
	}

	profiles, err := g.validateAllProfiles(ctx, pkgDB, profile)
	if err != nil {
		g.rollbackTransaction(tx)
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentProfileType),
			errors.WithResourceName(profile.Name),
			errors.WithMessage("application profiles are incongruent"))
	}

	created, err := tx.DeploymentProfile.Create().
		SetDeploymentPackageFkID(pkgDB.ID).
		SetName(profile.GetName()).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(profile.GetDescription()).
		AddProfiles(profiles...).
		Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, errors.NewInvalidArgument(
				errors.WithResourceType(errors.DeploymentProfileType),
				errors.WithResourceName(profile.Name),
				errors.WithMessage("deployment profile already exists"))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}
	return created, nil
}

func (g *Server) updateDeploymentProfile(ctx context.Context, tx *generated.Tx, deploymentProfile *catalogv3.DeploymentProfile, deploymentProfileDB *generated.DeploymentProfile, pkgDB *generated.DeploymentPackage) error {
	displayName, ok := validateDisplayName(deploymentProfile.Name, deploymentProfile.DisplayName)
	if !ok {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentProfileType),
			errors.WithResourceName(deploymentProfile.Name),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	profiles, err := g.validateAllProfiles(ctx, pkgDB, deploymentProfile)
	if err != nil {
		g.rollbackTransaction(tx)
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.DeploymentProfileType),
			errors.WithResourceName(deploymentProfile.Name),
			errors.WithMessage("application profiles are incongruent"))
	}

	_, err = tx.DeploymentProfile.Update().Where(deploymentprofile.ID(deploymentProfileDB.ID)).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(deploymentProfile.Description).
		AddProfiles(profiles...).
		Save(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}

	return nil
}

func (g *Server) getDeploymentPackage(ctx context.Context, tx *generated.Tx, projectUUID string, pkgName string, pkgVersion string) (*generated.DeploymentPackage, bool, error) {
	deploymentPkg, err := tx.DeploymentPackage.Query().
		Where(
			deploymentpackage.ProjectUUID(projectUUID),
			deploymentpackage.Name(pkgName),
			deploymentpackage.Version(pkgVersion),
		).Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, errors.NewDBError(errors.WithError(err))
	}
	return deploymentPkg, true, nil
}

// Returns an error if the profile display name is not unique
func (g *Server) checkDeploymentProfileUniqueness(ctx context.Context, p *catalogv3.DeploymentProfile, pkgDB *generated.DeploymentPackage) error {
	if len(p.DisplayName) > 0 {
		existing, err := pkgDB.QueryDeploymentProfiles().
			Where(
				deploymentprofile.DisplayNameLc(strings.ToLower(p.DisplayName)),
				deploymentprofile.Not(deploymentprofile.Name(p.Name)),
			).
			Count(ctx)
		if err = checkUniqueness(existing, err, "deployment profile", p.Name, p.DisplayName, errors.DeploymentProfileType); err != nil {
			return err
		}
	}
	return nil
}

// Validates that the list of profiles named in the given deployment profile indeed corresponds to the
// valid profiles of applications named in the deployment package to which the given deployment profile belongs.
// Returns the list of those application profiles if all named profiles are valid; otherwise, returns an error.
func (g *Server) validateAllProfiles(ctx context.Context, pkg *generated.DeploymentPackage, cp *catalogv3.DeploymentProfile) ([]*generated.Profile, error) {
	profiles := make([]*generated.Profile, 0, len(cp.ApplicationProfiles))
	for appReference, profileName := range cp.ApplicationProfiles {
		appProfile, err := g.validateProfile(ctx, pkg, appReference, profileName)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, appProfile)
	}
	return profiles, nil
}

// Validates that the given application - part of a deployment package - has the specified profile.
// Returns the actual profile if valid; otherwise, returns an error.
func (g *Server) validateProfile(ctx context.Context, pkg *generated.DeploymentPackage, appReference string, profileName string) (*generated.Profile, error) {
	appDB, err := g.findApplicationInPackage(ctx, pkg, appReference)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	return appDB.QueryProfiles().Where(profile.Name(profileName)).Only(ctx)
}

func (g *Server) findApplicationInPackage(ctx context.Context, pkg *generated.DeploymentPackage, appReference string) (*generated.Application, error) {
	// Split the app reference into fields by ":"
	fields := strings.Split(appReference, ":")
	appName := fields[0]
	appsDB, err := pkg.QueryApplications().Where(application.Name(appName)).All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	// If we have just the app name and the query by name yielded at least one item, just return it
	if len(fields) == 1 && len(appsDB) > 0 {
		return appsDB[0], nil
	}

	// Otherwise, start filtering out apps by their version and publisher specified in the reference
	appVersion := ""
	if len(fields) > 1 {
		appVersion = fields[1]
	}
	for _, appDB := range appsDB {
		// If version was specified, it must match the app version.
		if appVersion == "" || appDB.Version == appVersion {
			return appDB, nil
		}
	}

	// If we got through all the apps, it must mean the reference is incongruous
	return nil, errors.NewNotFound(errors.WithMessage("application %s:%s not found", appName, appVersion))
}

// Composes map of application name to profile name mappings for the given deployment profile.
func composeAppProfileMap(ctx context.Context, cp *generated.DeploymentProfile, useFQNames bool) (map[string]string, error) {
	// Reconstruct the application->profiles map
	profiles := make(map[string]string, 0)
	cpProfiles, err := cp.QueryProfiles().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	for _, profile := range cpProfiles {
		app, err1 := profile.QueryApplicationFk().Only(ctx)
		if err1 != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		ref := app.Name
		if useFQNames {
			ref = fmt.Sprintf("%s:%s", app.Name, app.Version)
		}
		profiles[ref] = profile.Name
	}
	return profiles, nil
}

func implicitDefaultDeploymentProfile(ctx context.Context, pkgDB *generated.DeploymentPackage) (*catalogv3.DeploymentProfile, error) {
	profiles := make(map[string]string, 0)

	// Query all applications that have default profiles
	appsDB, err := pkgDB.QueryApplications().Where(application.HasDefaultProfile()).All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}

	// If there are no such apps, return nil to signify no implicit default profile
	if len(appsDB) == 0 {
		return nil, nil
	}

	// Otherwise, iterate over the apps and register their default profiles as part of this synthetic default deployment profile
	for _, appDB := range appsDB {
		profileDB, err := appDB.QueryDefaultProfile().First(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
		profiles[appDB.Name] = profileDB.Name
	}
	return &catalogv3.DeploymentProfile{
		Name:                "implicit-default",
		DisplayName:         "Implicit Default",
		Description:         "Implicit default deployment profile",
		ApplicationProfiles: profiles,
	}, nil
}
