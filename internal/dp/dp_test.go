// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dp

import (
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/helm"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func TestGenerateDeploymentPackage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-deployment-package")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	h := helm.HelmInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "test description",
		OCIRegistry: "oci://test",
		Username:    "testuser",
		Password:    "testpassword",
	}

	err = GenerateDeploymentPackage(h, "", tempDir, "", false)
	assert.NoError(t, err)

	dpFileName := fmt.Sprintf("%s/%s-deployment-package.yaml", tempDir, h.Name)
	appFileName := fmt.Sprintf("%s/%s-application.yaml", tempDir, h.Name)
	regFileName := fmt.Sprintf("%s/%s-registry.yaml", tempDir, h.Name)
	valFileName := fmt.Sprintf("%s/%s-values.yaml", tempDir, h.Name)

	assert.True(t, fileExists(dpFileName))
	assert.True(t, fileExists(appFileName))
	assert.True(t, fileExists(regFileName))
	assert.True(t, fileExists(valFileName))

	dpContent, err := os.ReadFile(dpFileName)
	assert.NoError(t, err)
	expectedDPContent := `---
specSchema: "DeploymentPackage"
schemaVersion: "0.1"
$schema: "https://schema.intel.com/catalog.orchestrator/0.1/schema"

name: test
version: 1.0.0
description: ""
applications:
- name: test
  version: 1.0.0
deploymentProfiles:
- name: default
  applicationProfiles:
  - application: test
    profile: default
`
	assert.Equal(t, expectedDPContent, string(dpContent))

	appContent, err := os.ReadFile(appFileName)
	assert.NoError(t, err)
	expectedAppContent := `---
specSchema: "Application"
schemaVersion: "0.1"
$schema: "https://schema.intel.com/catalog.orchestrator/0.1/schema"

name: test
version: 1.0.0
description: test description
helmRegistry: test-registry
chartName: test
chartVersion: 1.0.0
profiles:
- name: default
  valuesFileName: test-values.yaml
`
	assert.Equal(t, expectedAppContent, string(appContent))

	regContent, err := os.ReadFile(regFileName)
	assert.NoError(t, err)
	expectedRegContent := `---
specSchema: "Registry"
schemaVersion: "0.1"
$schema: "https://schema.intel.com/catalog.orchestrator/0.1/schema"

name: test-registry
description: OCI registry for test
type: HELM
rootUrl: oci://test
userName: ""
authToken: ""
`
	assert.Equal(t, expectedRegContent, string(regContent))

	valContent, err := os.ReadFile(valFileName)
	assert.NoError(t, err)
	expectedValContent := "# this file intentionally left blank\n"
	assert.Equal(t, expectedValContent, string(valContent))
}

func TestGenerateDeploymentPackageWithAuth(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-deployment-package")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	h := helm.HelmInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "test description",
		OCIRegistry: "oci://test",
		Username:    "testuser",
		Password:    "testpassword",
	}

	err = GenerateDeploymentPackage(h, "", tempDir, "", true)
	assert.NoError(t, err)

	regFileName := fmt.Sprintf("%s/%s-registry.yaml", tempDir, h.Name)

	assert.True(t, fileExists(regFileName))

	regContent, err := os.ReadFile(regFileName)
	assert.NoError(t, err)
	expectedRegContent := `---
specSchema: "Registry"
schemaVersion: "0.1"
$schema: "https://schema.intel.com/catalog.orchestrator/0.1/schema"

name: test-registry
description: OCI registry for test
type: HELM
rootUrl: oci://test
userName: testuser
authToken: testpassword
`
	assert.Equal(t, expectedRegContent, string(regContent))
}

func TestGenerateDeploymentPackageWithValues(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-deployment-package")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	h := helm.HelmInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "test description",
		OCIRegistry: "oci://test",
		Username:    "testuser",
		Password:    "testpassword",
	}

	inputValuesFileName := fmt.Sprintf("%s/values.yaml", os.TempDir())
	defer os.Remove(inputValuesFileName)
	sampleValues := `---
replicaCount: 2
image:
  repository: nginx
  tag: stable
  pullPolicy: IfNotPresent
service:
  type: ClusterIP
  port: 80
`
	err = os.WriteFile(inputValuesFileName, []byte(sampleValues), 0600)
	assert.NoError(t, err)

	err = GenerateDeploymentPackage(h, inputValuesFileName, tempDir, "", false)
	assert.NoError(t, err)

	valFileName := fmt.Sprintf("%s/%s-values.yaml", tempDir, h.Name)

	assert.True(t, fileExists(valFileName))

	valContent, err := os.ReadFile(valFileName)
	assert.NoError(t, err)
	assert.Equal(t, sampleValues, string(valContent))
}

func TestGenerateDeploymentPackageWithNamespace(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-deployment-package")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	h := helm.HelmInfo{
		Name:        "test",
		Version:     "1.0.0",
		Description: "test description",
		OCIRegistry: "oci://test",
		Username:    "testuser",
		Password:    "testpassword",
	}

	err = GenerateDeploymentPackage(h, "", tempDir, "testnamespace", false)
	assert.NoError(t, err)

	dpFileName := fmt.Sprintf("%s/%s-deployment-package.yaml", tempDir, h.Name)

	assert.True(t, fileExists(dpFileName))

	dpContent, err := os.ReadFile(dpFileName)
	assert.NoError(t, err)
	expectedDPContent := `---
specSchema: "DeploymentPackage"
schemaVersion: "0.1"
$schema: "https://schema.intel.com/catalog.orchestrator/0.1/schema"

name: test
version: 1.0.0
description: ""
applications:
- name: test
  version: 1.0.0
deploymentProfiles:
- name: default
  applicationProfiles:
  - application: test
    profile: default
defaultNamespaces:
  test: testnamespace
`
	assert.Equal(t, expectedDPContent, string(dpContent))
}
