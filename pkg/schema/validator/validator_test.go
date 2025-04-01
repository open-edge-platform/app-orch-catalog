// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestValidateBytes(t *testing.T) {
	v, err := NewValidator()
	assert.NoError(t, err)

	yamlBytes, err := os.ReadFile("testdata/valid/registry-intel.yaml")
	assert.NoError(t, err)
	if v != nil {
		err = v.Validate(yamlBytes)
		assert.NoError(t, err)
	}
}

func TestValidateOKFiles(t *testing.T) {
	results, err := ValidateFiles("testdata/valid")
	assert.NoError(t, err)
	assert.Len(t, results, 4)
}

func TestValidateBadFiles(t *testing.T) {
	results, err := ValidateFiles("testdata/invalid")
	assert.Error(t, err)
	assert.Len(t, results, 6)
	count := 0
	for _, r := range results {
		if r.Err != nil && strings.Contains(r.Err.Error(), "does not validate with") {
			count++
		}
		fmt.Printf("%s: %v\n", r.Path, r.Err)
	}
	assert.Equal(t, 4, count)
}

func TestValidateBadMultipartFile(t *testing.T) {
	results, err := ValidateFiles("testdata/invalid/registry-intel.yaml")
	assert.Error(t, err)
	assert.Len(t, results, 1)
}
