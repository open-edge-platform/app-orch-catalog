// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
)

func (s *TestSuite) TestListExtensions() {
	// Form the request URL
	requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/v3/applications", s.CatalogRESTServerUrl)

	// Make the curl request using the access token and format the output with jq
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	s.T().Log(body)

	var result interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Print the formatted JSON output
	formattedJSON, err := json.MarshalIndent(result, "", "  ")
	assert.NoError(s.T(), err)
	fmt.Println(string(formattedJSON))
}
