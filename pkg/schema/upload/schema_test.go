// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package upload

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var specs = YamlSpecs{
	{SpecSchema: RegistryType, Name: "reg1"},
	{SpecSchema: RegistryType, Name: "reg2"},
	{SpecSchema: ApplicationType, Name: "app1"},
	{SpecSchema: ApplicationType, Name: "app2"},
	{SpecSchema: ArtifactType, Name: "art1"},
	{SpecSchema: ArtifactType, Name: "art2"},
	{SpecSchema: DeploymentPackageType, Name: "pkg1"},
	{SpecSchema: DeploymentPackageType, Name: "pkg2"},
	{SpecSchema: DeploymentPackageLegacyType, Name: "lpkg1"},
	{SpecSchema: DeploymentPackageLegacyType, Name: "lpkg2"},
	{SpecSchema: "unknown", Name: "bar"},
	{SpecSchema: "unknown", Name: "foo"},
}

func TestNameOrdering(t *testing.T) {
	assert.True(t, specs.Less(0, 1))
	assert.True(t, specs.Less(2, 3))
	assert.True(t, specs.Less(4, 5))
	assert.True(t, specs.Less(6, 7))
	assert.True(t, specs.Less(8, 9))
	assert.True(t, specs.Less(10, 11))
}

func TestLenAndSwap(t *testing.T) {
	assert.Equal(t, specs.Len(), 12)
	specs.Swap(0, 1)
	assert.False(t, specs.Less(0, 1))
}

func TestTypeOrdering(t *testing.T) {
	assert.True(t, specs.Less(0, 2))
	assert.False(t, specs.Less(2, 0))
	assert.True(t, specs.Less(2, 6))
	assert.False(t, specs.Less(2, 1))
	assert.False(t, specs.Less(2, 4))
	assert.False(t, specs.Less(4, 2))
	assert.True(t, specs.Less(4, 6))
	assert.False(t, specs.Less(6, 2))

	assert.False(t, specs.Less(8, 6))
}

func TestGetRegistryType(t *testing.T) {
	assert.Equal(t, "new", YamlSpec{Type: "new", RegistryType: "old"}.GetRegistryType())
	assert.Equal(t, "old", YamlSpec{RegistryType: "old"}.GetRegistryType())
}

func TestGetRegistry(t *testing.T) {
	assert.Equal(t, "new", YamlSpec{HelmRegistry: "new", Registry: "old"}.GetHelmRegistry())
	assert.Equal(t, "old", YamlSpec{Registry: "old"}.GetHelmRegistry())
}

func TestGetArtifacts(t *testing.T) {
	assert.Len(t, YamlSpec{Artifacts: []ArtifactReference{{}, {}}, ArtifactReferences: []ArtifactReference{{}}}.GetArtifacts(), 2)
	assert.Len(t, YamlSpec{ArtifactReferences: []ArtifactReference{{}}}.GetArtifacts(), 1)
}
