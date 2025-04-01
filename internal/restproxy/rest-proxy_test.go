// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"context"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
)

func (s *ProxyTestSuite) newRequest(path string) *http.Request {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("http://localhost:6942/%s", path), http.NoBody)
	s.NoError(err)
	req.Header.Set(ActiveProjectID, "project")
	return req
}

func (s *ProxyTestSuite) checkRequest(req *http.Request, status int) {
	resp, err := s.httpClient.Do(req)
	s.NoError(err)
	if s.NotNil(resp) {
		s.Equal(status, resp.StatusCode)
	}
}

func (s *ProxyTestSuite) checkRequestBody(req *http.Request, status int, body string) {
	resp, err := s.httpClient.Do(req)
	s.NoError(err)
	if s.NotNil(resp) {
		s.Equal(status, resp.StatusCode)
		resBody, err := io.ReadAll(resp.Body)
		s.NoError(err)
		s.Equal(body, string(resBody))
	}
}

func (s *ProxyTestSuite) TestEndpointHealthz() {
	s.checkRequest(s.newRequest("healthz"), 200)
}

func (s *ProxyTestSuite) TestEndpointTest() {
	s.checkRequest(s.newRequest("test"), 200)
}

func (s *ProxyTestSuite) TestOIDCExternalTest() {
	s.checkRequest(s.newRequest("openidc-issuer"), 200)
}
