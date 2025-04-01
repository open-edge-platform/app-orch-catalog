// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
)

func (s *NorthBoundTestSuite) TestNoSyntheticDefaultDeploymentProfile() {
	s.createApp(footen, fooreg, "app1", "v0.1.0", 0)
	s.createApp(footen, fooreg, "app2", "v0.2.0", 0)
	s.createDeploymentPkg(footen, "ca-simple", "v0.1.0", "app1:v0.1.0", "app2:v0.2.0")

	ca, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-simple", Version: "v0.1.0",
	})
	s.NoError(err)
	s.NotNil(ca)
	s.validateDeploymentPkg(ca.DeploymentPackage, "ca-simple", "v0.1.0", "Deployment Package ca-simple",
		"This is deployment package ca-simple", "", "", 2, 0, 0, "", 0, false)
}

func (s *NorthBoundTestSuite) TestSyntheticDefaultDeploymentProfile() {
	s.createApp(footen, fooreg, "app1", "v0.1.0", 0)
	s.createDeploymentPkg(footen, "ca-simple", "v0.1.0", "foo:v0.1.0", "app1:v0.1.0")

	ca, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-simple", Version: "v0.1.0",
	})
	s.NoError(err)
	s.NotNil(ca)
	s.validateDeploymentPkg(ca.DeploymentPackage, "ca-simple", "v0.1.0", "Deployment Package ca-simple",
		"This is deployment package ca-simple", "", "", 2, 0, 1, "implicit-default", 0, false)

	profile := ca.DeploymentPackage.Profiles[0]
	s.Equal(profile.Name, "implicit-default")
	s.Len(profile.ApplicationProfiles, 1)
	s.Equal(profile.ApplicationProfiles["foo"], "p1")
}
