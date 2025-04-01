// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentpackage"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentprofile"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/deploymentrequirement"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/parametertemplate"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/application"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/profile"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
)

func validateParameterTemplates(profile *catalogv3.Profile) error {
	ptNames := map[string]*catalogv3.ParameterTemplate{}
	for _, pt := range profile.ParameterTemplates {
		_, dup := ptNames[pt.Name]
		if dup {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithResourceName(profile.Name),
				errors.WithMessage("duplicate parameter template %s", pt.Name))
		}
		_, ok := validateDisplayName(pt.Name, pt.DisplayName)
		if !ok {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ProfileType),
				errors.WithResourceName(profile.Name),
				errors.WithMessage("display name cannot contain leading or trailing spaces"))
		}
		if (pt.Mandatory || pt.Secret) && pt.Default != "" {
			return errors.NewInvalidArgument(
				errors.WithResourceType(errors.ApplicationType),
				errors.WithResourceName(profile.Name),
				errors.WithMessage("mandatory or secret parameter template %s should have no default value", pt.Name))
		}
		ptNames[pt.Name] = pt
	}
	return nil
}

func (g *Server) injectProfile(ctx context.Context, tx *generated.Tx, projectUUID string, profile *catalogv3.Profile, app *generated.Application) (*generated.Profile, error) {
	displayName, ok := validateDisplayName(profile.Name, profile.DisplayName)
	if !ok {
		return nil, errors.NewInvalidArgument(
			errors.WithResourceType(errors.ProfileType),
			errors.WithResourceName(profile.Name),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	if err := g.checkProfileUniqueness(ctx, profile, app); err != nil {
		return nil, err
	}

	created, err := tx.Profile.Create().
		SetApplicationFkID(app.ID).
		SetName(profile.GetName()).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(profile.GetDescription()).
		SetChartValues(profile.ChartValues).
		Save(ctx)
	if err != nil {
		if generated.IsConstraintError(err) {
			return nil, errors.NewAlreadyExists(
				errors.WithResourceType(errors.ProfileType),
				errors.WithResourceName(profile.Name))
		}
		return nil, errors.NewDBError(errors.WithError(err))
	}

	for _, dr := range profile.DeploymentRequirement {
		if err = g.injectDeploymentRequirement(ctx, tx, projectUUID, created, dr); err != nil {
			return nil, err
		}
	}

	err = validateParameterTemplates(profile)
	if err != nil {
		return nil, err
	}

	for _, pt := range profile.ParameterTemplates {
		_, err = tx.ParameterTemplate.Create().
			SetName(pt.Name).
			SetDisplayName(pt.DisplayName).
			SetDefault(pt.Default).
			SetType(pt.Type).
			SetValidator(pt.Validator).
			SetSuggestedValues(pt.SuggestedValues).
			SetMandatory(pt.Mandatory).
			SetSecret(pt.Secret).
			SetProfileFkID(created.ID).
			Save(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}
	}
	return created, nil
}

func (g *Server) injectDeploymentRequirement(ctx context.Context, tx *generated.Tx, projectUUID string, profileDB *generated.Profile, requirement *catalogv3.DeploymentRequirement) error {
	log.Infof("Injecting DR: %+v", requirement)
	pkgID, err := tx.DeploymentPackage.Query().Where(
		deploymentpackage.ProjectUUID(projectUUID),
		deploymentpackage.Name(requirement.Name),
		deploymentpackage.Version(requirement.Version)).FirstID(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return errors.NewNotFound(
				errors.WithResourceType(errors.DeploymentPackageType),
				errors.WithMessage("deployment package %s not found", requirement.Name))
		}
		return errors.NewDBError(errors.WithError(err))
	}
	drCreateStmt := tx.DeploymentRequirement.Create().
		SetProfileFk(profileDB).
		SetDeploymentPackageFkID(pkgID)
	if requirement.DeploymentProfileName != "" {
		dpID, err := tx.DeploymentProfile.Query().Where(
			deploymentprofile.HasDeploymentPackageFkWith(deploymentpackage.ID(pkgID)),
			deploymentprofile.Name(requirement.DeploymentProfileName)).FirstID(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
		drCreateStmt.SetDeploymentProfileFkID(dpID)
	}
	if err = drCreateStmt.Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}
	return nil
}

func (g *Server) getApplication(ctx context.Context, tx *generated.Tx, projectUUID string, applicationName string, applicationVersion string) (*generated.Application, bool, error) {
	app, err := tx.Application.Query().
		Where(
			application.ProjectUUID(projectUUID),
			application.Name(applicationName),
			application.Version(applicationVersion),
		).
		First(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, errors.NewDBError(errors.WithError(err))
	}
	return app, true, nil
}

// Returns an error if the profile display name is not unique
func (g *Server) checkProfileUniqueness(ctx context.Context, p *catalogv3.Profile, app *generated.Application) error {
	if len(p.DisplayName) > 0 {
		existing, err := app.QueryProfiles().
			Where(profile.DisplayNameLc(strings.ToLower(p.DisplayName)), profile.Not(profile.Name(p.Name))).
			Count(ctx)
		if err = checkUniqueness(existing, err, "profile", p.Name, p.DisplayName, errors.ProfileType); err != nil {
			return err
		}
	}
	return nil
}

func (g *Server) extractDeploymentRequirements(ctx context.Context, profileDB *generated.Profile) ([]*catalogv3.DeploymentRequirement, error) {
	requirementsDB, err := profileDB.QueryDeploymentRequirements().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	requirements := make([]*catalogv3.DeploymentRequirement, 0, len(requirementsDB))
	for _, drDB := range requirementsDB {
		dpkgDB, err := drDB.QueryDeploymentPackageFk().Only(ctx)
		if err != nil {
			return nil, errors.NewDBError(errors.WithError(err))
		}

		dprofName := ""
		dprofDB, err := drDB.QueryDeploymentProfileFk().Only(ctx)
		if err == nil {
			dprofName = dprofDB.Name
		}

		requirement := &catalogv3.DeploymentRequirement{
			Name:                  dpkgDB.Name,
			Version:               dpkgDB.Version,
			DeploymentProfileName: dprofName,
		}

		dpProfileDB, err := drDB.QueryDeploymentProfileFk().First(ctx)
		if err == nil {
			requirement.DeploymentProfileName = dpProfileDB.Name
		}

		requirements = append(requirements, requirement)
	}
	return requirements, nil
}

func (g *Server) extractParameterTemplates(ctx context.Context, profileDB *generated.Profile) ([]*catalogv3.ParameterTemplate, error) {
	parameterTemplatesDB, err := profileDB.QueryParameterTemplates().All(ctx)
	if err != nil {
		return nil, errors.NewDBError(errors.WithError(err))
	}
	parameterTemplates := make([]*catalogv3.ParameterTemplate, 0, len(parameterTemplatesDB))
	for _, ptDB := range parameterTemplatesDB {
		parameterTemplates = append(parameterTemplates, &catalogv3.ParameterTemplate{
			Name:            ptDB.Name,
			DisplayName:     ptDB.DisplayName,
			Default:         ptDB.Default,
			Type:            ptDB.Type,
			Validator:       ptDB.Validator,
			SuggestedValues: ptDB.SuggestedValues,
			Mandatory:       ptDB.Mandatory,
			Secret:          ptDB.Secret,
		})
	}
	return parameterTemplates, nil
}

func (g *Server) updateProfile(ctx context.Context, tx *generated.Tx, projectUUID string, p *catalogv3.Profile, pDB *generated.Profile) error {
	displayName, ok := validateDisplayName(p.Name, p.DisplayName)
	if !ok {
		return errors.NewInvalidArgument(
			errors.WithResourceType(errors.ProfileType),
			errors.WithResourceName(p.Name),
			errors.WithMessage("display name cannot contain leading or trailing spaces"))
	}

	_, err := tx.Profile.Update().Where(profile.ID(pDB.ID)).
		SetDisplayName(displayName).
		SetDisplayNameLc(strings.ToLower(displayName)).
		SetDescription(p.Description).
		SetChartValues(p.ChartValues).
		Save(ctx)
	if err != nil {
		return errors.NewDBError(errors.WithError(err))
	}

	// Delete and re-create all deployment requirements
	if _, err := tx.DeploymentRequirement.Delete().Where(deploymentrequirement.HasProfileFkWith(profile.ID(pDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}

	for _, dr := range p.DeploymentRequirement {
		if err = g.injectDeploymentRequirement(ctx, tx, projectUUID, pDB, dr); err != nil {
			return err
		}
	}

	// Delete and re-create all parameter templates...
	if _, err := tx.ParameterTemplate.Delete().Where(parametertemplate.HasProfileFkWith(profile.ID(pDB.ID))).Exec(ctx); err != nil {
		return errors.NewDBError(errors.WithError(err))
	}

	err = validateParameterTemplates(p)
	if err != nil {
		return err
	}
	for _, pt := range p.ParameterTemplates {
		_, err := tx.ParameterTemplate.Create().
			SetName(pt.Name).
			SetDisplayName(pt.DisplayName).
			SetDefault(pt.Default).
			SetType(pt.Type).
			SetValidator(pt.Validator).
			SetSuggestedValues(pt.SuggestedValues).
			SetMandatory(pt.Mandatory).
			SetSecret(pt.Secret).
			SetProfileFkID(pDB.ID).
			Save(ctx)
		if err != nil {
			return errors.NewDBError(errors.WithError(err))
		}
	}
	return nil
}
