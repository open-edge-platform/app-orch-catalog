// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
)

func (s *TestSuite) TestListExtensions() {
	// Form the request URL
	requestURL := fmt.Sprintf("%s/v3/projects/%s/catalog/applications?orderBy=name+asc&pageSize=10&offset=0&kinds=KIND_EXTENSION", s.CatalogRESTServerUrl, s.projectID)

	// Make the curl request using the access token and format the output with jq
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+s.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	var result interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Print the formatted JSON output
	formattedJSON, err := json.MarshalIndent(result, "", "  ")
	assert.NoError(s.T(), err)
	fmt.Println(string(formattedJSON))
}
