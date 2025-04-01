// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"fmt"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	wiper2 "github.com/open-edge-platform/app-orch-catalog/pkg/wiper"
)

func (s *TestSuite) WipeOutData() {
	wiper := wiper2.NewGRPCWiper(s.client)
	errors := wiper.Wipe(s.AddHeaders("footen"), "")
	if len(errors) > 0 {
		fmt.Printf("footen wipe errors: [%+v]\n", errors)
	}
	errors = wiper.Wipe(s.AddHeaders("barten"), "")
	if len(errors) > 0 {
		fmt.Printf("barten wipe errors: [%+v]\n", errors)
	}
}

// TestBasics tests basics of exercising the gRPC API of the catalog service.
func (s *TestSuite) TestBasics() {
	// Create some registries
	s.createRegistry("footen", "fooreg", "Foo Registry", "Registry for foos", "https://reg.foo.com/")
	s.createRegistry("footen", "fooregalt", "Foo Alt Registry", "Alternate registry for foos", "https://regalt.foo.com/")

	s.createRegistry("barten", "barreg", "Bar Registry", "Registry for bars", "https://reg.bar.com/")

	// Create a few applications with embedded profiles
	s.createApplication("footen", "fooreg", "ap1", "1.1",
		"App One", "First application", []*catalogv3.Profile{
			profile("p1", "Profile One", "First profile", "some yaml goes here"),
			profile("p2", "Profile Two", "Second profile", "some other yaml goes here"),
		}, "p2")

	s.createApplication("footen", "fooregalt", "ap2", "1.2",
		"App Two", "Second application", []*catalogv3.Profile{
			profile("p1", "Profile One", "First profile", "some odd yaml goes here"),
		}, "p1")

	s.createApplication("barten", "barreg", "ap3", "1.3",
		"App Three", "Third application", []*catalogv3.Profile{
			profile("p1", "Profile One", "First profile", "some odd yaml goes here"),
			profile("p2", "Profile Two", "Second profile", "some other yaml goes here"),
			profile("p3", "Profile Three", "Third profile", "some weird yaml here"),
		}, "p2")

	s.createArtifact("footen", "icon", "Foo Icon", "Icon for foos",
		"image/png", "test/basic/1x1.png")
	s.createArtifact("footen", "thumb", "Foo Thumbnail", "Thumbnail for foos",
		"image/png", "test/basic/1x1.png")

	// Create a new package
	s.createPackage("footen", "cap1", "1.1", "Package One", "First Package",
		[]*catalogv3.ApplicationReference{
			{Name: "ap1", Version: "1.1"},
			{Name: "ap2", Version: "1.2"},
		},
		[]*catalogv3.DeploymentProfile{
			packageProfile("p1", "Profile One", "First profile", map[string]string{"ap1": "p2", "ap2": "p1"}),
			packageProfile("p2", "Profile Two", "Second profile", map[string]string{"ap1": "p1", "ap2": "p1"}),
		},
		"p1",
		[]*catalogv3.APIExtension{
			extension("ext1", "v1.1", "Extension 1", "First Extension", "", "",
				[]*catalogv3.Endpoint{
					endpoint("svc1", "external/path", "internal/path"),
					endpoint("svc2", "another/external/path", "another/internal/path"),
				}),
			extension("ext2", "v1.2", "Extension 2", "Second Extension", "Service Two", "svc2",
				[]*catalogv3.Endpoint{
					endpoint("svc1", "some/external/path", "some/internal/path"),
					endpoint("svc2", "other/external/path", "other/internal/path"),
				}),
		},
		[]*catalogv3.ArtifactReference{
			artifact("icon", "icon"),
			artifact("thumb", "thumbnail"),
		},
	)

	s.TestValidateBasics()
}

// TestValidateBasics only validates that the entities created by the TestBasics are intact.
func (s *TestSuite) TestValidateBasics() {
	// Get all registries for foo project
	lr, err := s.client.ListRegistries(s.AddHeaders("footen"), &catalogv3.ListRegistriesRequest{})
	s.NoError(err)
	s.Len(lr.Registries, 2)

	// Get all registries for bar project
	lr, err = s.client.ListRegistries(s.AddHeaders("barten"), &catalogv3.ListRegistriesRequest{})
	s.NoError(err)
	s.Len(lr.Registries, 1)

	// Get all applications for foo project
	lar, err := s.client.ListApplications(s.AddHeaders("footen"), &catalogv3.ListApplicationsRequest{})
	s.NoError(err)
	s.Len(lar.Applications, 2)

	// Get all applications for bar project
	lar, err = s.client.ListApplications(s.AddHeaders("barten"), &catalogv3.ListApplicationsRequest{})
	s.NoError(err)
	s.Len(lar.Applications, 1)

	// Get all artifacts for foo project
	lir, err := s.client.ListArtifacts(s.AddHeaders("footen"), &catalogv3.ListArtifactsRequest{})
	s.NoError(err)
	s.Len(lir.Artifacts, 2)

	resp, err := s.client.GetDeploymentPackage(s.AddHeaders("footen"), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "cap1", Version: "1.1",
	})
	s.NoError(err)
	s.validatePackage(resp.DeploymentPackage, "cap1", "1.1", "Package One", "First Package",
		2, 2, "p1", 2, 2)

	// Validate the extensions and their endpoints
	ext := findExtension("ext1", resp.DeploymentPackage.Extensions)
	s.validateExtension(ext, "ext1", "v1.1", "Extension 1", "First Extension", "", "", 2)
	ep := findEndpoint("svc1", ext.Endpoints)
	s.validateEndpoint(ep, "svc1", "external/path", "internal/path")
	ep = findEndpoint("svc2", ext.Endpoints)
	s.validateEndpoint(ep, "svc2", "another/external/path", "another/internal/path")

	ext = findExtension("ext2", resp.DeploymentPackage.Extensions)
	s.validateExtension(ext, "ext2", "v1.2", "Extension 2", "Second Extension", "Service Two", "svc2", 2)
	ep = findEndpoint("svc1", ext.Endpoints)
	s.validateEndpoint(ep, "svc1", "some/external/path", "some/internal/path")
	ep = findEndpoint("svc2", ext.Endpoints)
	s.validateEndpoint(ep, "svc2", "other/external/path", "other/internal/path")

	// Validate the artifact references
	ar := findArtifactReference("icon", resp.DeploymentPackage.Artifacts)
	s.validateArtifactReference(ar, "icon", "icon")
	ar = findArtifactReference("thumb", resp.DeploymentPackage.Artifacts)
	s.validateArtifactReference(ar, "thumb", "thumbnail")

	// Validate the artifact references
	ar = findArtifactReference("icon", resp.DeploymentPackage.Artifacts)
	s.validateArtifactReference(ar, "icon", "icon")
	ar = findArtifactReference("thumb", resp.DeploymentPackage.Artifacts)
	s.validateArtifactReference(ar, "thumb", "thumbnail")

	// Get all deployment packages for project footen
	lbr, err := s.client.ListDeploymentPackages(s.AddHeaders("footen"), &catalogv3.ListDeploymentPackagesRequest{})
	s.NoError(err)
	s.Len(lbr.DeploymentPackages, 1)

	// Get all deployment packages for project barten
	lbr, err = s.client.ListDeploymentPackages(s.AddHeaders("barten"), &catalogv3.ListDeploymentPackagesRequest{})
	s.NoError(err)
	s.Len(lbr.DeploymentPackages, 0)
}
