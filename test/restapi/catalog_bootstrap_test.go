// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports
	"context"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"net/http"

	"github.com/stretchr/testify/assert"
)

func (s *TestSuite) TestListBootStrapExtensions() {
	ctx := context.TODO()
	applications, status, err := s.catalogClient.GetApplicationList(ctx)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for application list")

	assert.Equal(s.T(), len(types.GetApplications()), len(applications), "Mismatch in the number of applications")
	// Log application details for debugging purposes
	s.T().Log("Extensions:")
	for _, app := range applications {
		s.T().Logf("Name: %s, DisplayName: %s, Description: %s, Version: %s, Kind: %s, ChartName: %s, ChartVersion: %s, HelmRegistryName: %s",
			app.Name, *app.DisplayName, *app.Description, app.Version, *app.Kind, app.ChartName, app.ChartVersion, app.HelmRegistryName)
	}
}

func (s *TestSuite) TestListBootStrapDeploymentPackages() {
	ctx := context.TODO()

	orderBy := "name"
	var pageSize int32 = 10
	var offset int32 = 0

	deploymentPackages, status, err := s.catalogClient.ListDeploymentPackages(ctx, &restClient.CatalogServiceListDeploymentPackagesParams{
		OrderBy:  &orderBy,
		PageSize: &pageSize,
		Offset:   &offset,
		Kinds: &[]restClient.CatalogServiceListDeploymentPackagesParamsKinds{
			restClient.CatalogServiceListDeploymentPackagesParamsKindsKINDEXTENSION,
		},
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for deployment package list")

	assert.Equal(s.T(), len(types.GetDeploymentPackages()), len(deploymentPackages), "Mismatch in the number of deployment packages")

	// Log deployment package details for debugging purposes
	s.T().Log("Deployment Packages:")
	for _, pkg := range deploymentPackages {
		s.T().Logf("Name: %s, Description: %s, Version: %s, Kind: %s",
			pkg.Name, *pkg.Description, pkg.Version, *pkg.Kind)
	}
}

func (s *TestSuite) TestListBootStrapRegistries() {
	ctx := context.TODO()

	orderBy := "name"
	var pageSize int32 = 10
	var offset int32 = 0
	var showSensitiveInfo = true

	registries, status, err := s.catalogClient.ListRegistries(ctx, &restClient.CatalogServiceListRegistriesParams{
		OrderBy:           &orderBy,
		PageSize:          &pageSize,
		Offset:            &offset,
		ShowSensitiveInfo: &showSensitiveInfo,
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for registry list")

	// Assert that the size of the result.Registries matches the size of getRegistries
	assert.Equal(s.T(), len(s.catalogClient.GetRegistries()), len(registries), "Mismatch in the number of registries")
	// Log registry details for debugging purposes
	s.T().Log("Registries:")
	for _, registry := range registries {
		s.T().Logf("Name: %s, DisplayName: %s, Description: %s, RootURL: %s, Type: %s",
			registry.Name, *registry.DisplayName, *registry.Description, registry.RootUrl, registry.Type)
	}
}

func (s *TestSuite) TestVerifyBootstrappedRegistriesExist() {
	ctx := context.TODO()
	for _, registry := range s.catalogClient.GetRegistries() {
		result, status, err := s.catalogClient.GetRegistry(ctx, registry.Name)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for registry list")

		switch {
		case registry.Name != result.Name:
			assert.Equal(s.T(), registry.Name, result.Name, "Mismatch in 'Name' for registry: %s", registry.Name)
		case registry.DisplayName != result.DisplayName:
			assert.Equal(s.T(), registry.DisplayName, result.DisplayName, "Mismatch in 'DisplayName' for registry: %s", registry.Name)
		case registry.RootUrl != result.RootUrl:
			oldDockerURL := fmt.Sprintf("https://registry-oci.%s/", s.orchDomain)
			newDockerURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", s.orchDomain)
			// Docker Registry URL was changed recently. Avoid throwing errors in a development environment that's using the new URL
			// TODO: remove this special case when component-tests are moved forward
			if registry.RootUrl != oldDockerURL || result.RootUrl != newDockerURL {
				assert.Equal(s.T(), registry.RootUrl, result.RootUrl, "Mismatch in 'RootURL' for registry: %s", registry.Name)
			}
		case registry.Type != result.Type:
			assert.Equal(s.T(), registry.Type, result.Type, "Mismatch in 'Type' for registry: %s", registry.Name)
		}
		// assert.Equal(s.T(), registry.Description, result.Registry.Description)
	}
}

func (s *TestSuite) TestVerifyBootstrappedExtensionsExist() {
	ctx := context.TODO()
	for _, app := range types.GetApplications() {
		result, status, err := s.catalogClient.GetApplicationVersions(ctx, app.Name)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for application")
		assert.NotEmpty(s.T(), result, "Expected application to be found")
		gotApp := result[0]
		switch {
		case app.Name != gotApp.Name:
			assert.Equalf(s.T(), app.Name, gotApp.Name, "Mismatch in 'Name' for application: %s", app.Name)
		case app.DisplayName != gotApp.DisplayName:
			assert.Equalf(s.T(), app.DisplayName, gotApp.DisplayName, "Mismatch in 'DisplayName' for application: %s", app.Name)
		case app.ChartName != gotApp.ChartName:
			assert.Equalf(s.T(), app.ChartName, gotApp.ChartName, "Mismatch in 'ChartName' for application: %s", app.Name)
		case app.Kind != gotApp.Kind:
			assert.Equalf(s.T(), app.Kind, gotApp.Kind, "Mismatch in 'Kind' for application: %s", app.Name)
		case app.HelmRegistryName != gotApp.HelmRegistryName:
			assert.Equalf(s.T(), app.HelmRegistryName, gotApp.HelmRegistryName, "Mismatch in 'HelmRegistryName' for application: %s", app.Name)
		}
		//assert.Equal(s.T(), app.Description, result.Application.Description)
	}

}

func (s *TestSuite) TestVerifyBootstrappedDeploymentPackagesExist() {
	ctx := context.TODO()
	for _, pkg := range types.GetDeploymentPackages() {
		result, status, err := s.catalogClient.GetDeploymentPackageVersions(ctx, pkg.Name)
		assert.NoError(s.T(), err, "Expected to get deployment package without error")
		assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for deployment package")
		assert.NotEmpty(s.T(), result, "Expected deployment package to be found")
		gotPkg := result[0]
		switch {
		case pkg.Name != gotPkg.Name:
			assert.Equalf(s.T(), pkg.Name, gotPkg.Name, "Mismatch in 'Name' for deployment package: %s", pkg.Name)
		case pkg.Kind != gotPkg.Kind:
			assert.Equalf(s.T(), pkg.Kind, gotPkg.Kind, "Mismatch in 'Kind' for deployment package: %s", pkg.Name)
		}

	}
}
