// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	// Third-party imports

	"github.com/open-edge-platform/app-orch-catalog/internal/yamlreader"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
)

// We use the Geti helm chart for testing purposes as it is located at ghcr.io and is a know valid OCI Helm chart.

const (
	helmToDpTool  = "../../build/_output/helm-to-dp"
	getiHelmChart = "oci://ghcr.io/open-edge-platform/geti/helm/impt:2.9.0"
	badHelmChart  = "oci://ghcr.invalid/open-edge-platform/geti/helm/impt:2.9.0-bad-chart"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func (s *TestSuite) runHelmToDp(args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Executing command: %s %v", helmToDpTool, args)

	c := exec.CommandContext(ctx, helmToDpTool, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	outStr := stdout.String()
	errStr := stderr.String()
	return outStr, errStr, err
}

// LoadDeploymentPackage reads a deployment package from tempdir.
// It returns one DeploymentPackage, one Application, and one Registry.
func (s *TestSuite) LoadDeploymentPackage(tempdir string) (*catalogv3.DeploymentPackage, *catalogv3.Application, *catalogv3.Registry) {
	var dp *catalogv3.DeploymentPackage
	var app *catalogv3.Application
	var reg *catalogv3.Registry

	r := &yamlreader.YamlReader{}

	files, err := r.ReadYamlFilesFromDir(tempdir)
	s.Require().NoError(err, "Expected to read YAML files from directory")

	orderedSpecs, err := r.LoadYamlSpecs(files)
	s.Require().NoError(err, "Expected to load YAML specs from FileSet")

	for _, spec := range orderedSpecs {
		switch spec.SpecSchema {
		case upload.DeploymentPackageType:
			s.Nil(dp, "Expected Deployment Package to be nil before reading")
			dp, err = r.ReadDeploymentPackage(spec)
			s.NoError(err, "Expected to read Deployment Package from spec")
		case upload.DeploymentPackageLegacyType:
			s.Nil(dp, "Expected Deployment Package to be nil before reading")
			dp, err = r.ReadDeploymentPackage(spec)
			s.NoError(err, "Expected to read Deployment Package from legacy spec")
		case upload.ApplicationType:
			s.Nil(app, "Expected Application to be nil before reading")
			app, err = r.ReadApplication(spec, files) // application uses the FileSet to lookup profiles
			s.NoError(err, "Expected to read Application from spec")
		case upload.RegistryType:
			s.Nil(reg, "Expected Registry to be nil before reading")
			reg, err = r.ReadRegistry(spec)
			s.NoError(err, "Expected to read Registry from spec")
		default:
			s.Failf("Unhandled spec type", "Unhandled type %q in spec %q", spec.SpecSchema, spec.FileName)
		}
	}

	return dp, app, reg
}

func (s *TestSuite) TestHelmToDpGoodURL() {
	tempDir, err := os.MkdirTemp("", "catalog_test_*")
	s.Require().NoError(err)
	defer os.RemoveAll(tempDir)

	_, stderr, err := s.runHelmToDp(getiHelmChart, "-o", tempDir)
	s.NoError(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stderr, "Expected no error output when running catalog-schema on a good package")

	// Now load a deployment package from the temp directory and verify it is correct.

	dp, app, reg := s.LoadDeploymentPackage(tempDir)
	s.Require().NotNil(dp, "Expected Deployment Package to be loaded")
	s.Require().NotNil(app, "Expected Application to be loaded")
	s.Require().NotNil(reg, "Expected Registry to be loaded")

	s.Equal("impt", dp.Name, "Expected Deployment Package name to be 'impt'")
	s.Equal("2.9.0", dp.Version, "Expected Deployment Package version to be '2.9.0'")

	s.Equal("impt", app.Name, "Expected Application name to be 'impt'")
	s.Equal("2.9.0", app.Version, "Expected Application version to be '2.9.0'")
	s.Equal("impt-registry", app.HelmRegistryName, "Expected Application registry to be 'impt-registry'")
	s.Equal("impt", app.ChartName, "Expected Application chart name to be 'impt'")
	s.Equal("2.9.0", app.ChartVersion, "Expected Application chart version to be '2.9.0'")

	s.Equal("impt-registry", reg.Name, "Expected Registry name to be 'impt-registry'")
	s.Equal("oci://ghcr.io/open-edge-platform/geti/helm", reg.RootUrl, "Expected Registry URL to match the input")
}

func (s *TestSuite) TestHelmToDpBadURL() {
	tempDir, err := os.MkdirTemp("", "catalog_test_*")
	s.Require().NoError(err)
	defer os.RemoveAll(tempDir)

	_, stderr, err := s.runHelmToDp(badHelmChart, "-o", tempDir)
	s.Error(err, "Expected error when running catalog-schema on a bad URL")

	if !(strings.Contains(stderr, "failed to resolve") || strings.Contains(stderr, "failed to verify certificate")) {
		s.T().Logf("Unexpected error message: %s", stderr)
		s.Fail("Expected error message to contain 'failed to resolve' or 'failed to verify certificate'")
	}
}
