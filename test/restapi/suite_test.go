// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package rest is a suite of REST API functionality tests for the catalog service
package restapi

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/stretchr/testify/suite"
	"time"
)

const (
	RestAddress            = "app-orch-catalog-rest-proxy:8081/"
	RestAddressPortForward = "127.0.0.1"
	KeycloakServer         = "keycloak.kind.internal"

	PortForwardServiceNamespace = "orch-app"
	PortForwardService          = "svc/app-orch-catalog-rest-proxy"
	PortForwardLocalPort        = "8081"
	PortForwardAddress          = "0.0.0.0"
	PortForwardRemotePort       = "8081"
)

// TestSuite is the basic test suite
type TestSuite struct {
	suite.Suite
	KeycloakServer       string
	CatalogRESTServerUrl string
	token                string
	projectID            string
	cmd                  *exec.Cmd
}

// SetupSuite sets-up the integration tests for the Catalog basic test suite
func (s *TestSuite) SetupSuite() {
	s.KeycloakServer = KeycloakServer
	s.CatalogRESTServerUrl = RestAddress
}

// SetupTest sets up for each integration test
func (s *TestSuite) SetupTest() {
	var err error
	s.token = auth.SetUpAccessToken(s.T(), s.KeycloakServer)
	s.CatalogRESTServerUrl = fmt.Sprintf("http://%s:%s", RestAddressPortForward, PortForwardRemotePort)
	s.projectID = "sample-project"
	s.NoError(err)
	s.cmd, err = portForwardToCatalog()
	s.NoError(err)
}

func killportForwardToCatalog(cmd *exec.Cmd) error {
	fmt.Println("kill process that port-forwards network to app-orch-catalog")
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Kill()
	}
	return nil
}

func portForwardToCatalog() (*exec.Cmd, error) {
	fmt.Println("port-forward to app-deployment-manager")

	cmd := exec.Command("kubectl", "port-forward", "-n", PortForwardServiceNamespace, PortForwardService, fmt.Sprintf("%s:%s", PortForwardLocalPort, PortForwardRemotePort), "--address", PortForwardAddress)
	err := cmd.Start()
	time.Sleep(5 * time.Second) // Give some time for port-forwarding to establish

	return cmd, err
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// TearDownTest tears down remnants of each integration test
func (s *TestSuite) TearDownTest() {
	err := killportForwardToCatalog(s.cmd)
	s.NoError(err)
}
