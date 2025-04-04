// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package auth contains utilities for keycloak authentication
package auth

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const (
	SampleProject = "sample-project"
)

func SetUpAccessToken(t *testing.T, server string) string {
	c := &http.Client{
		Transport: &http.Transport{},
	}
	data := url.Values{}
	data.Set("client_id", "system-client")
	data.Set("username", fmt.Sprintf("%s-edge-mgr", SampleProject))
	data.Set("password", "ChangeMeOn1stLogin!")
	data.Set("grant_type", "password")
	url := "https://" + server + "/realms/master/protocol/openid-connect/token"
	req, err := http.NewRequest(http.MethodPost,
		url,
		strings.NewReader(data.Encode()))
	assert.NoError(t, err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	assert.NoError(t, err)
	if resp == nil {
		fmt.Fprintf(os.Stderr, "No response from keycloak: %s\n", url)
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, resp.StatusCode, http.StatusOK)
	rawTokenData, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	tokenData := map[string]interface{}{}
	err = json.Unmarshal(rawTokenData, &tokenData)
	assert.NoError(t, err)

	accessToken := tokenData["access_token"].(string)
	assert.NotContains(t, accessToken, `named cookie not present`)
	return accessToken
}

func AddRestAuthHeader(req *http.Request, token string, projectID string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Activeprojectid", fmt.Sprintf("%s", projectID))
}
