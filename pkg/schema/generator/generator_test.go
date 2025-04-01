// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func prefix(s string) string {
	return fmt.Sprintf("../../../%s", s)
}

func TestSchemaGenerator(t *testing.T) {
	assert.NoError(t, generateSchema(prefix(schemaBaseFile), prefix(openapiSpecFile), "/tmp/catalog-schema"))
}
