// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	// Third-party imports
	"net/url"

	// Project-specific imports
	"github.com/open-edge-platform/app-orch-catalog/test/auth"

	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
)

const importEndpoint = "/catalog.orchestrator.apis/v3/import"

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

type ImportRequest struct {
	Url                       string `json:"url"`
	Username                  string `json:"username,omitempty"`
	AuthToken                 string `json:"auth_token,omitempty"`
	ChartValues               string `json:"chart_values,omitempty"`
	IncludeAuth               bool   `json:"include_auth,omitempty"`
	GenerateDefaultValues     bool   `json:"generate_default_values,omitempty"`
	GenerateDefaultParameters bool   `json:"generate_default_parameters,omitempty"`
	Namespace                 string `json:"namespace,omitempty"`
}

func (s *TestSuite) ImportHelmChart(importRequest *ImportRequest) (int, string) {
	params := url.Values{}
	params.Add("url", importRequest.Url)

	requestURL := fmt.Sprintf("%s%s?%s", s.CatalogRESTServerUrl, importEndpoint, params.Encode())

	req, err := http.NewRequest("POST", requestURL, nil)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	log.Printf("Importing Helm chart with request URL: %s", requestURL)

	res, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	s.Require().NoError(err)

	return res.StatusCode, string(body)
}

func (s *TestSuite) GetApplication(name string, version string) (*catalogv3.Application, error) {
	requestURL := fmt.Sprintf("%s%s/%s/versions/%s", s.CatalogRESTServerUrl, applicationsEndpoint, name, version)

	log.Printf("Retrieving application with request URL: %s", requestURL)

	req, err := http.NewRequest("GET", requestURL, nil)
	s.Require().NoError(err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get application: %s", res.Status)
	}

	var app catalogv3.Application
	err = json.NewDecoder(res.Body).Decode(&app)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (s *TestSuite) TestImportHelmChart() {
	importRequest := &ImportRequest{
		Url: "oci://ghcr.io/open-edge-platform/geti/helm/impt:2.9.0",
	}

	status, body := s.ImportHelmChart(importRequest)
	s.Equal(200, status, "Expected status code 200 for successful import")

	fmt.Println("Response body:", body)

	_ = body

	app, err := s.GetApplication("impt", "2.9.0")
	s.NoError(err, "Expected to retrieve application after import")

	_ = app
}
