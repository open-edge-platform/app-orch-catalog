// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package rest is a suite of REST API functionality tests for the catalog service
package restapi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"time"

	"github.com/open-edge-platform/app-orch-catalog/test/auth"
	"github.com/stretchr/testify/suite"
)

const (
	RestAddress            = "app-orch-catalog-rest-proxy:8081/"
	RestAddressPortForward = "127.0.0.1"

	PortForwardServiceNamespace = "orch-app"
	PortForwardService          = "svc/app-orch-catalog-rest-proxy"
	PortForwardLocalPort        = "8081"
	PortForwardAddress          = "0.0.0.0"
	PortForwardRemotePort       = "8081"
)

// TestSuite is the basic test suite
type TestSuite struct {
	suite.Suite
	orchDomain		     string
	KeycloakServer       string
	CatalogRESTServerUrl string
	token                string
	projectID            string
	cmd                  *exec.Cmd
}

// SetupSuite sets-up the integration tests for the Catalog basic test suite
func (s *TestSuite) SetupSuite() {
	s.CatalogRESTServerUrl = RestAddress

	// To use the component-tests with a domain other than kind.internal, ensure
	// the ORCH_DOMAIN environment variable is set.
	s.orchDomain = os.Getenv("ORCH_DOMAIN")
	if s.orchDomain == "" {
		s.orchDomain = "kind.internal"
	}
	s.KeycloakServer = fmt.Sprintf("keycloak.%s", s.orchDomain)

	var err error
	s.token = auth.SetUpAccessToken(s.T(), s.KeycloakServer)
	s.CatalogRESTServerUrl = fmt.Sprintf("http://%s:%s", RestAddressPortForward, PortForwardRemotePort)
	s.projectID, err = auth.GetProjectId(context.TODO())
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
	fmt.Println("port-forward to app-orch-catalog")

	cmd := exec.Command("kubectl", "port-forward", "-n", PortForwardServiceNamespace, PortForwardService,
		fmt.Sprintf("%s:%s", PortForwardLocalPort, PortForwardRemotePort),
		"--address", PortForwardAddress)
	err := cmd.Start()
	time.Sleep(5 * time.Second) // Give some time for port-forwarding to establish

	return cmd, err
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TearDownSuite() {
	err := killportForwardToCatalog(s.cmd)
	s.NoError(err)
}
