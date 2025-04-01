// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"encoding/json"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"time"
)

func (s *ProxyTestSuite) subscribeCmd(cmd string, resp string, kind string) {
	err := s.ws.WriteJSON(Message{Op: cmd, Kind: kind, Project: "project"})
	s.NoError(err)

	// Read confirmation of subscription
	msg := &Message{}
	err = s.ws.ReadJSON(msg)
	s.NoError(err)
	s.Equal(resp, msg.Op)
	s.Equal(kind, msg.Kind)
}

func (s *ProxyTestSuite) subscribe(kind string) {
	s.subscribeCmd(subscribeOp, subscribedOp, kind)
}

func (s *ProxyTestSuite) unsubscribe(kind string) {
	s.subscribeCmd(unsubscribeOp, unsubscribedOp, kind)
}

func (s *ProxyTestSuite) readEvent(kind string, eventType string, name string) {
	msg := &Message{}
	err := s.ws.ReadJSON(msg)
	s.NoError(err)
	s.Equal(eventType, msg.Op)
	s.Equal(kind, msg.Kind)
	dict := make(map[string]interface{})
	s.NoError(json.Unmarshal(msg.Payload, &dict))
	s.Equal(name, dict["name"])
}

func (s *ProxyTestSuite) createTestRegistry(inventoryURL string, username string, auth string) {
	s.cleanUpTestRegistry()
	_, err := s.client.CreateRegistry(s.ctx, &catalogv3.CreateRegistryRequest{Registry: &catalogv3.Registry{
		Name: "reg", RootUrl: "http://foobar.com", Type: "HELM", InventoryUrl: inventoryURL, Username: username, AuthToken: auth,
	}})
	s.NoError(err)
}

func (s *ProxyTestSuite) cleanUpTestRegistry() {
	_, _ = s.client.DeleteRegistry(s.ctx, &catalogv3.DeleteRegistryRequest{RegistryName: "reg"})
}

func (s *ProxyTestSuite) createThings() {
	s.createTestRegistry("http://silly/", "", "")

	_, err := s.client.CreateArtifact(s.ctx, &catalogv3.CreateArtifactRequest{Artifact: &catalogv3.Artifact{
		Name:     "art",
		MimeType: "text/plain",
		Artifact: []byte("nothing here"),
	}})
	s.NoError(err)

	_, err = s.client.CreateApplication(s.ctx, &catalogv3.CreateApplicationRequest{Application: &catalogv3.Application{
		Name:             "app",
		Version:          "1.0",
		HelmRegistryName: "reg",
		ChartName:        "chart",
		ChartVersion:     "1.0",
	}})
	s.NoError(err)

	_, err = s.client.CreateDeploymentPackage(s.ctx, &catalogv3.CreateDeploymentPackageRequest{DeploymentPackage: &catalogv3.DeploymentPackage{
		Name:                  "pkg",
		Version:               "1.0",
		ApplicationReferences: []*catalogv3.ApplicationReference{{Name: "app", Version: "1.0"}},
	}})
	s.NoError(err)
}

func (s *ProxyTestSuite) deleteThings() {
	_, err := s.client.DeleteDeploymentPackage(s.ctx, &catalogv3.DeleteDeploymentPackageRequest{DeploymentPackageName: "pkg", Version: "1.0"})
	s.NoError(err)
	_, err = s.client.DeleteApplication(s.ctx, &catalogv3.DeleteApplicationRequest{ApplicationName: "app", Version: "1.0"})
	s.NoError(err)
	_, err = s.client.DeleteArtifact(s.ctx, &catalogv3.DeleteArtifactRequest{ArtifactName: "art"})
	s.NoError(err)
	_, err = s.client.DeleteRegistry(s.ctx, &catalogv3.DeleteRegistryRequest{RegistryName: "reg"})
	s.NoError(err)
}

func (s *ProxyTestSuite) testKindBasics(kind string, name string) {
	s.setupWebSocket()
	s.subscribe(kind)

	// Create a thing and read an event for it
	s.createThings()
	s.readEvent(kind, "created", name)

	// Delete the thing and read an event for it
	s.deleteThings()
	s.readEvent(kind, "deleted", name)

	s.unsubscribe(kind)
	s.closeWebSocket()
}

func (s *ProxyTestSuite) TestRegistryBasics() {
	s.testKindBasics(registryKind, "reg")
}

func (s *ProxyTestSuite) TestArtifactBasics() {
	s.testKindBasics(artifactKind, "art")
}

func (s *ProxyTestSuite) TestApplicationBasics() {
	s.testKindBasics(applicationKind, "app")
}

func (s *ProxyTestSuite) TestDeploymentPackageBasics() {
	s.testKindBasics(deploymentPackageKind, "pkg")
}

func (s *ProxyTestSuite) TestApplicationAndPackageBasics() {
	s.T().Skip()
	pingPeriod = 100 * time.Millisecond
	maxPongWait = 200 * time.Millisecond

	s.setupWebSocket()
	s.subscribe(applicationKind)
	s.subscribe(deploymentPackageKind)

	s.createThings()
	s.readEvent(applicationKind, "created", "app")
	s.readEvent(deploymentPackageKind, "created", "pkg")

	s.deleteThings()
	s.readEvent(deploymentPackageKind, "deleted", "pkg")
	s.readEvent(applicationKind, "deleted", "app")

	s.unsubscribe(applicationKind)
	s.unsubscribe(deploymentPackageKind)

	time.Sleep(5 * pingPeriod)
	s.closeWebSocket()
}
