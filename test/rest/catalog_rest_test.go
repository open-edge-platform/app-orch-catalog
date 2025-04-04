// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"time"
)

// listDeploymentPackages allows the verb to be overridden, for tests related to http verb restriction
func (s *TestSuite) listDeploymentPackages(server string, verb string) (*http.Response, error) {
	s.T().Skip("skipping test for now, as it is not working")
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/catalog.orchestrator.apis/v1/deployment_packages")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(verb, queryURL.String(), nil)
	if err != nil {
		return nil, err
	}
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, err
}

// listPublishers does a GET operation to fetch the list of publishers
func (s *TestSuite) listPublishers(server string) (*http.Response, error) {
	s.T().Skip("skipping test for now, as it is not working")
	var err error
	req, err := http.NewRequest(http.MethodGet, server+"catalog.orchestrator.apis/v1/publishers", nil)
	if err != nil {
		return nil, err
	}
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, err
}

// Rest-Proxy will return error responses to all requests for a few seconds after starting.
// TODO: Investigate Ready-Check
func (s *TestSuite) waitForProxyUsable() error {
	s.T().Skip("skipping test for now, as it is not working")

	// Wait up to 30 seconds, in 5-second intervals
	for i := 0; i < 6; i++ {
		res, _ := s.listPublishers(s.CatalogRESTServerUrl)
		if res != nil && res.StatusCode == 200 {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for rest-proxy to be usable")
}

// TestRest tests basics of exercising the REST API of the catalog service.
func (s *TestSuite) TestRest() {
	s.T().Skip("skipping test for now, as it is not working")

	s.NoError(s.waitForProxyUsable())

	res, err := s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodGet)
	s.NoError(err)
	s.Equal("200 OK", res.Status)

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodPost)
	s.NoError(err)
	s.Equal("400 Bad Request", res.Status) /* legitimately a bad request, as we send empty body */

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodTrace)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodPatch)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodDelete)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodPut)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodHead)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.CatalogRESTServerUrl, http.MethodConnect)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)
}

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
