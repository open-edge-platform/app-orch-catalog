// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
)

// TestRESTBasics tests basics of exercising the REST API of the catalog service.
func (s *TestSuite) TestRESTBasics() {
	// Create some registries
	s.createRESTRegistry("footen", "fooreg", "Foo Registry", "Registry for foos", "https://reg.foo.com/")
	s.createRESTRegistry("footen", "fooregalt", "Foo Alt Registry", "Alternate registry for foos", "https://regalt.foo.com/")

	s.createRESTRegistry("barten", "barreg", "Bar Registry", "Registry for bars", "https://reg.bar.com/")

	// Create a few applications with embedded profiles
	s.createRESTApplication("footen", "fooreg", "ap1", "1.1",
		"App One", "First application", []restapi.Profile{
			profileREST("p1", "Profile One", "First profile", "some yaml goes here"),
			profileREST("p2", "Profile Two", "Second profile", "some other yaml goes here"),
		}, "p2")

	s.createRESTApplication("footen", "fooregalt", "ap2", "1.2",
		"App Two", "Second application", []restapi.Profile{
			profileREST("p1", "Profile One", "First profile", "some odd yaml goes here"),
		}, "p1")

	s.createRESTApplication("barten", "barreg", "ap3", "1.3",
		"App Three", "Third application", []restapi.Profile{
			profileREST("p1", "Profile One", "First profile", "some odd yaml goes here"),
			profileREST("p2", "Profile Two", "Second profile", "some other yaml goes here"),
			profileREST("p3", "Profile Three", "Third profile", "some weird yaml here"),
		}, "p2")

	s.createRESTArtifact("footen", "icon", "Foo Icon", "Icon for foos",
		"image/png", "test/basic/1x1.png")
	s.createRESTArtifact("footen", "thumb", "Foo Thumbnail", "Thumbnail for foos",
		"image/png", "test/basic/1x1.png")

	// Create a new package
	s.createRESTPackage("footen", "cap1", "1.1", "Package One", "First Package",
		[]restapi.ApplicationReference{
			{Name: "ap1", Version: "1.1"},
			{Name: "ap2", Version: "1.2"},
		},
		[]restapi.DeploymentProfile{
			packageRESTProfile("p1", "Profile One", "First profile", map[string]string{"ap1": "p2", "ap2": "p1"}),
			packageRESTProfile("p2", "Profile Two", "Second profile", map[string]string{"ap1": "p1", "ap2": "p1"}),
		},
		"p1",
		[]restapi.APIExtension{
			extensionREST("ext1", "v1.1", "Extension 1", "First Extension", "", "",
				[]restapi.Endpoint{
					endpointREST("svc1", "external/path", "internal/path"),
					endpointREST("svc2", "another/external/path", "another/internal/path"),
				}),
			extensionREST("ext2", "v1.2", "Extension 2", "Second Extension", "Service Two", "svc2",
				[]restapi.Endpoint{
					endpointREST("svc1", "some/external/path", "some/internal/path"),
					endpointREST("svc2", "other/external/path", "other/internal/path"),
				}),
		},
		[]restapi.ArtifactReference{
			artifactREST("icon", "icon"),
			artifactREST("thumb", "thumbnail"),
		},
	)

	s.TestValidateRESTBasics()
}

func (s *TestSuite) TestUpdateApplicationWithDeploymentRequirements() {
	s.createRESTRegistry("barten", "barreg", "Bar Registry", "Registry for bars", "https://reg.bar.com/")

	s.createRESTApplication("barten", "barreg", "app", "1.1",
		"App", "Application", []restapi.Profile{
			profileREST("p1", "Profile", "A Profile", "some odd yaml goes here"),
		}, "p1")

	s.createRESTPackage("barten", "pkg", "1.0", "Package", "Package",
		[]restapi.ApplicationReference{
			{Name: "app", Version: "1.1"},
		}, nil, "", nil, nil)

	// Create an app with deployment requirement it its profile
	requirements := []restapi.DeploymentRequirement{
		{Name: "pkg", Version: "1.0"},
	}

	profiles := []restapi.Profile{
		profileREST("p1", "Profile One", "First profile", "some odd yaml goes here"),
	}
	profiles[0].DeploymentRequirement = &requirements

	cresp, err := s.restClient.CatalogServiceCreateApplicationWithResponse(s.ProjectID("barten"), restapi.CatalogServiceCreateApplicationJSONRequestBody{
		Name:               "topapp",
		Version:            "1.0",
		ChartName:          "topapp-chart",
		ChartVersion:       "1.0",
		DefaultProfileName: &profiles[0].Name,
		HelmRegistryName:   "barreg",
		Profiles:           &profiles,
	}, addHeaders)
	s.validateResponse(err, cresp.JSON200)

	// Retrieve the app and validate that the deployment requirements are there
	resp, err := s.restClient.CatalogServiceGetApplicationWithResponse(s.ProjectID("barten"), "topapp", "1.0", addHeaders)
	s.validateResponse(err, resp.JSON200)
	if resp.JSON200.Application.Profiles != nil && s.Len(*resp.JSON200.Application.Profiles, 1) {
		s.Len(*(*resp.JSON200.Application.Profiles)[0].DeploymentRequirement, 1)
	}

	// Tweak the one profile
	newValues := "different chart values"
	profiles[0].ChartValues = &newValues

	// Update the app
	_, err = s.restClient.CatalogServiceUpdateApplicationWithResponse(s.ProjectID("barten"), "topapp", "1.0",
		restapi.CatalogServiceUpdateApplicationJSONRequestBody{
			Name:               "topapp",
			Version:            "1.0",
			ChartName:          "topapp-chart",
			ChartVersion:       "1.0c",
			DefaultProfileName: &profiles[0].Name,
			HelmRegistryName:   "barreg",
			Profiles:           &profiles,
		}, addHeaders)
	s.NoError(err)

	// Retrieve the app and validate that the deployment requirements are there
	resp, err = s.restClient.CatalogServiceGetApplicationWithResponse(s.ProjectID("barteb"), "topapp", "1.0", addHeaders)
	s.validateResponse(err, resp.JSON200)
	if resp.JSON200.Application.Profiles != nil && s.Len(*resp.JSON200.Application.Profiles, 1) {
		s.Len(*(*resp.JSON200.Application.Profiles)[0].DeploymentRequirement, 1)
		s.Equal(newValues, *(*resp.JSON200.Application.Profiles)[0].ChartValues)
	}
}

// TestValidateRESTBasics only validates that the entities created by the TestRESTBasics are intact.
func (s *TestSuite) TestValidateRESTBasics() {
	// Get all registries for foo project
	showSensitiveInfo := true
	lr, err := s.restClient.CatalogServiceListRegistriesWithResponse(s.ProjectID("footen"), &restapi.CatalogServiceListRegistriesParams{
		ShowSensitiveInfo: &showSensitiveInfo,
	}, addHeaders)
	s.validateResponse(err, lr)
	s.Len(lr.JSON200.Registries, 2)

	// Get all registries for bar project
	showSensitiveInfo = true
	lr, err = s.restClient.CatalogServiceListRegistriesWithResponse(s.ProjectID("barten"), &restapi.CatalogServiceListRegistriesParams{
		ShowSensitiveInfo: &showSensitiveInfo,
	}, addHeaders)
	s.validateResponse(err, lr)
	s.Len(lr.JSON200.Registries, 1)

	// Get all applications for foo project
	lar, err := s.restClient.CatalogServiceListApplicationsWithResponse(s.ProjectID("footen"), &restapi.CatalogServiceListApplicationsParams{}, addHeaders)
	s.validateResponse(err, lar)
	s.Len(lar.JSON200.Applications, 2)

	// Get all applications for bar project
	lar, err = s.restClient.CatalogServiceListApplicationsWithResponse(s.ProjectID("barten"), &restapi.CatalogServiceListApplicationsParams{}, addHeaders)
	s.validateResponse(err, lr)
	s.Len(lar.JSON200.Applications, 1)

	// Get all artifacts for foo project
	lir, err := s.restClient.CatalogServiceListArtifactsWithResponse(s.ProjectID("footen"), &restapi.CatalogServiceListArtifactsParams{}, addHeaders)
	s.validateResponse(err, lir)
	s.Len(lir.JSON200.Artifacts, 2)

	resp, err := s.restClient.CatalogServiceGetDeploymentPackageWithResponse(s.ProjectID("footen"), "cap1", "1.1", addHeaders)
	s.validateResponse(err, resp)
	s.validateRESTPackage(resp.JSON200.DeploymentPackage, "cap1", "1.1", "Package One", "First Package",
		2, 2, "p1", 2, 2)

	// Validate the extensions and their endpoints
	ext := findRESTExtension("ext1", resp.JSON200.DeploymentPackage.Extensions)
	s.validateRESTExtension(ext, "ext1", "v1.1", "Extension 1", "First Extension", "", "", 2)
	ep := findRESTEndpoint("svc1", *ext.Endpoints)
	s.validateRESTEndpoint(ep, "svc1", "external/path", "internal/path")
	ep = findRESTEndpoint("svc2", *ext.Endpoints)
	s.validateRESTEndpoint(ep, "svc2", "another/external/path", "another/internal/path")

	ext = findRESTExtension("ext2", resp.JSON200.DeploymentPackage.Extensions)
	s.validateRESTExtension(ext, "ext2", "v1.2", "Extension 2", "Second Extension", "Service Two", "svc2", 2)
	ep = findRESTEndpoint("svc1", *ext.Endpoints)
	s.validateRESTEndpoint(ep, "svc1", "some/external/path", "some/internal/path")
	ep = findRESTEndpoint("svc2", *ext.Endpoints)
	s.validateRESTEndpoint(ep, "svc2", "other/external/path", "other/internal/path")

	// Validate the artifact references
	ar := findRESTArtifactReference("icon", resp.JSON200.DeploymentPackage.Artifacts)
	s.validateRESTArtifactReference(ar, "icon", "icon")
	ar = findRESTArtifactReference("thumb", resp.JSON200.DeploymentPackage.Artifacts)
	s.validateRESTArtifactReference(ar, "thumb", "thumbnail")

	// Get all packages for bar project
	lbr, err := s.restClient.CatalogServiceListDeploymentPackagesWithResponse(s.ProjectID("footen"), &restapi.CatalogServiceListDeploymentPackagesParams{}, addHeaders)
	s.validateResponse(err, lbr)
	s.Len(lbr.JSON200.DeploymentPackages, 1)

	// Get all packages for foo project
	lbr, err = s.restClient.CatalogServiceListDeploymentPackagesWithResponse(s.ProjectID("barten"), &restapi.CatalogServiceListDeploymentPackagesParams{}, addHeaders)
	s.validateResponse(err, lbr)
	s.Len(lbr.JSON200.DeploymentPackages, 0)
}
