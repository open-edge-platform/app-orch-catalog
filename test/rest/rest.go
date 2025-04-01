// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"net/http"
	"net/url"
	"time"
)

// listDeploymentPackages allows the verb to be overridden, for tests related to http verb restriction
func (s *TestSuite) listDeploymentPackages(server string, verb string) (*http.Response, error) {
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
	auth.AddRestAuthHeader(req, s.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, err
}

// listPublishers does a GET operation to fetch the list of publishers
func (s *TestSuite) listPublishers(server string) (*http.Response, error) {
	var err error
	req, err := http.NewRequest(http.MethodGet, server+"catalog.orchestrator.apis/v1/publishers", nil)
	if err != nil {
		return nil, err
	}
	auth.AddRestAuthHeader(req, s.token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, err
}

// Rest-Proxy will return error responses to all requests for a few seconds after starting.
// TODO: Investigate Ready-Check
func (s *TestSuite) waitForProxyUsable() error {
	// Wait up to 30 seconds, in 5-second intervals
	for i := 0; i < 6; i++ {
		res, _ := s.listPublishers(s.serverUrl)
		if res != nil && res.StatusCode == 200 {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for rest-proxy to be usable")
}

// TestRest tests basics of exercising the REST API of the catalog service.
func (s *TestSuite) TestRest() {
	s.NoError(s.waitForProxyUsable())

	res, err := s.listDeploymentPackages(s.serverUrl, http.MethodGet)
	s.NoError(err)
	s.Equal("200 OK", res.Status)

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodPost)
	s.NoError(err)
	s.Equal("400 Bad Request", res.Status) /* legitimately a bad request, as we send empty body */

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodTrace)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodPatch)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodDelete)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodPut)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodHead)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)

	res, err = s.listDeploymentPackages(s.serverUrl, http.MethodConnect)
	s.NoError(err)
	s.Equal("405 Method Not Allowed", res.Status)
}
