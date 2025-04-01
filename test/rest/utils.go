// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func GetAccessToken(t *testing.T, username, password, serviceDomainWithPort string) string {
	data := url.Values{}
	data.Set("client_id", "system-client")
	data.Set("username", username)
	data.Set("password", password)
	data.Set("grant_type", "password")

	req, err := http.NewRequest("POST", fmt.Sprintf("https://keycloak.%s/realms/master/protocol/openid-connect/token", serviceDomainWithPort), bytes.NewBufferString(data.Encode()))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	var tokenData map[string]interface{}
	err = json.Unmarshal(body, &tokenData)
	assert.NoError(t, err)

	accessToken, ok := tokenData["access_token"].(string)
	assert.True(t, ok)

	return accessToken
}
