// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

const (
	username              = "sample-project-edge-mgr"
	password              = "ChangeMeOn1stLogin!"
	serviceDomainWithPort = "kind.internal"
)

func TestCatalogAPI(t *testing.T) {
	// Get the access token using the utility function
	accessToken := GetAccessToken(t, username, password, serviceDomainWithPort)

	// Make the curl request using the access token and format the output with jq
	req, err := http.NewRequest("GET", "https://api.kind.internal/v3/projects/sample-project/catalog/applications?orderBy=name+asc&pageSize=10&offset=0&kinds=KIND_EXTENSION", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var result interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Print the formatted JSON output
	formattedJSON, err := json.MarshalIndent(result, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(formattedJSON))
}
