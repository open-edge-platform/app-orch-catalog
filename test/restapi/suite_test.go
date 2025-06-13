// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package rest is a suite of REST API functionality tests for the catalog service
package restapi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/open-edge-platform/app-orch-catalog/test/utils/auth"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/methods"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/portforward"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"

	"github.com/stretchr/testify/suite"
)

// TestSuite is the basic test suite
type TestSuite struct {
	suite.Suite
	orchDomain     string
	KeycloakServer string
	catalogClient  *methods.CatalogClient
	cmd            *exec.Cmd
	token          string
}

// SetupSuite sets-up the integration tests for the Catalog basic test suite
func (s *TestSuite) SetupSuite() {

	// To use the component-tests with a domain other than kind.internal, ensure
	// the ORCH_DOMAIN and AUTO_CERT environment variables are set.
	autoCert, err := strconv.ParseBool(os.Getenv("AUTO_CERT"))
	s.orchDomain = os.Getenv("ORCH_DOMAIN")
	if err != nil || !autoCert || s.orchDomain == "" {
		s.orchDomain = "kind.internal"
	}
	s.T().Log("Orchestration domain set to:", s.orchDomain)
	s.KeycloakServer = fmt.Sprintf("keycloak.%s", s.orchDomain)
	catalogRESTServerUrl := fmt.Sprintf("http://%s:%s", types.RestAddressPortForward, types.PortForwardRemotePort)
	s.token = auth.SetUpAccessToken(s.T(), s.KeycloakServer)
	projectID, err := auth.GetProjectId(context.TODO(), types.SampleProject, types.SampleOrg)

	s.catalogClient = methods.NewCatalogClient(catalogRESTServerUrl, s.token, projectID, s.orchDomain)

	s.NoError(err)
	s.cmd, err = portforward.PortForwardToCatalog()
	s.NoError(err)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TearDownSuite() {
	err := portforward.KillportForwardToCatalog(s.cmd)
	s.NoError(err)
}
