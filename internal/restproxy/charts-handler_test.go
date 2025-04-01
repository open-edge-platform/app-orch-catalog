// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func (s *ProxyTestSuite) TestChartProxyAllChartsNoAuth() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"name":"chart1"},{"name":"chart2"},{"name":"chart3"}]`))
	}))
	defer server.Close()

	s.createTestRegistry(server.URL, "", "auth")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?publisher=pub&registry=reg"), 200,
		`["chart1","chart2","chart3"]`)
}

func ociURL(url string) string {
	return url + "/api/v2.0/projects/catalog-apps"
}

func (s *ProxyTestSuite) TestOCIChartProxyAllChartsNoAuth() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"name":"chart1"},{"name":"chart2"},{"name":"chart3"}]`))
	}))
	defer server.Close()

	fmt.Printf("serverURL=%s; inventoryURL=%s\n", server.URL, ociURL(server.URL))

	s.createTestRegistry(ociURL(server.URL), "", "auth")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?registry=reg"), 200,
		`["chart1","chart2","chart3"]`)
	s.cleanUpTestRegistry()
}

func (s *ProxyTestSuite) TestOCIChartProxyChartsVersionsNoAuth() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"extra_attrs":{"version":"v1.0"}},{"extra_attrs":{"version":"v2.0"}}]`))
	}))
	defer server.Close()

	s.createTestRegistry(ociURL(server.URL), "", "auth")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?registry=reg&chart=chart1"), 200,
		`["v1.0","v2.0"]`)
}

func (s *ProxyTestSuite) TestChartProxyAllChartsBasicAuth() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"name":"chart1"},{"name":"chart2"},{"name":"chart3"}]`))
	}))
	defer server.Close()

	s.createTestRegistry(server.URL, "user", "pwd")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?registry=reg"), 200,
		`["chart1","chart2","chart3"]`)
}

func (s *ProxyTestSuite) TestChartProxyAllChartsBearerAuth() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"name":"chart1"},{"name":"chart2"},{"name":"chart3"}]`))
	}))
	defer server.Close()

	s.createTestRegistry(server.URL, "", "token")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?registry=reg"), 200,
		`["chart1","chart2","chart3"]`)
}

func (s *ProxyTestSuite) TestChartProxyChartVersions() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"name":"chart1","version":"1.0"},{"name":"chart1","version":"2.0"},{"name":"chart1","version":"2.1"}]`))
	}))
	defer server.Close()

	s.createTestRegistry(server.URL+"/", "", "")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?registry=reg&chart=chart1"), 200,
		`["1.0","2.0","2.1"]`)
}

func (s *ProxyTestSuite) TestChartProxyNoInventoryURL() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`[{"name":"chart1","version":"1.0"},{"name":"chart1","version":"2.0"},{"name":"chart1","version":"2.1"}]`))
	}))
	defer server.Close()

	s.createTestRegistry("", "", "")
	s.checkRequestBody(s.newRequest("catalog.orchestrator.apis/charts?registry=reg&chart=chart1"), 204,
		``)
}

func (s *ProxyTestSuite) TestChartProxyBadPublisherOrRegistry() {
	s.checkRequest(s.newRequest("catalog.orchestrator.apis/charts?registry=bad"), 500)
}

func (s *ProxyTestSuite) TestChartProxyBadURL() {
	s.createTestRegistry("http://badurl", "", "")
	s.checkRequest(s.newRequest("catalog.orchestrator.apis/charts?registry=reg&chart=chart1"), 500)
}

func (s *ProxyTestSuite) TestChartProxyBadChartData() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = rw.Write([]byte(`["chart"]`))
	}))
	defer server.Close()

	s.createTestRegistry(server.URL, "", "")
	s.checkRequest(s.newRequest("catalog.orchestrator.apis/charts?registry=reg"), 500)
}

func (s *ProxyTestSuite) TestChartProxyRepoError() {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(500)
	}))
	defer server.Close()

	s.createTestRegistry(server.URL, "", "")
	s.checkRequest(s.newRequest("catalog.orchestrator.apis/charts?publisher=pub&registry=reg"), 500)
}

func (s *ProxyTestSuite) TestParseOCIChartNames() {
	body, err := os.ReadFile("testdata/repositories.json")
	s.NoError(err)

	names, err := parseChartNames(body)
	s.NoError(err)
	s.Len(names, 3)
	s.False(strings.Contains(names[0], "/"))
	s.False(strings.Contains(names[1], "/"))
	s.False(strings.Contains(names[2], "/"))
	fmt.Print(names)
}

func (s *ProxyTestSuite) TestParseOCIChartVersions() {
	body, err := os.ReadFile("testdata/artifacts.json")
	s.NoError(err)

	versions, err := parseChartVersions(body)
	s.NoError(err)
	s.Len(versions, 1)
}
