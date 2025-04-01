// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package basic is a suite of basic functionality tests for the catalog service
package basic

import (
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/stretchr/testify/suite"
	"path/filepath"
)

// TestSuite is the basic test suite
type TestSuite struct {
	suite.Suite
	serverUrl      string
	token          string
	KeycloakServer string
}

func (s *TestSuite) findCharts(chartName string) string {
	matches, err := filepath.Glob("build/_output/" + chartName + "-*.tgz")
	s.NoError(err)
	return matches[0]
}

// SetupSuite sets-up the integration tests for the application catalog basic test suite
func (s *TestSuite) SetupSuite() {
}

// SetupTest sets up for each integration test
func (s *TestSuite) SetupTest() {
	s.serverUrl = "http://catalog-service-rest-proxy:8081/"
	s.token = auth.SetUpAccessToken(s.T(), s.KeycloakServer)
}

// TearDownTest tears down remnants of each integration test
func (s *TestSuite) TearDownTest() {
}
