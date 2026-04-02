// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"context"
	"fmt"
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"google.golang.org/grpc/metadata"
	"net/http"
	"os"
)

func (s *TestSuite) createRESTRegistry(projectUUID string, name string, display string, description string, url string) *restapi.CatalogV3Registry {
	r, err := s.restClient.CatalogServiceCreateRegistryWithResponse(s.ProjectID(projectUUID), projectUUID, restapi.CatalogServiceCreateRegistryJSONRequestBody{
		Name: name, DisplayName: &display, Description: &description, RootUrl: url, Type: "HELM"}, addHeaders)
	if s.validateResponse(err, r) {
		if s.NotNil(r.JSON200) {
			s.validateRESTRegistry(r.JSON200.Registry, name, display, description)
			return &r.JSON200.Registry
		}
	}
	return nil
}

func (s *TestSuite) validateRESTRegistry(reg restapi.CatalogV3Registry, name string, display string, description string) {
	s.Equal(name, reg.Name)
	s.Equal(display, *reg.DisplayName)
	s.Equal(description, *reg.Description)
	s.Equal("HELM", reg.Type)
}

func (s *TestSuite) createRESTApplication(projectUUID string, reg string, name string, ver string, display string, description string,
	profiles []restapi.CatalogV3Profile, defaultProfile string) *restapi.CatalogV3Application {
	r, err := s.restClient.CatalogServiceCreateApplicationWithResponse(s.ProjectID(projectUUID), projectUUID, restapi.CatalogServiceCreateApplicationJSONRequestBody{
		Name: name, Version: ver, DisplayName: &display, Description: &description,
		ChartName: fmt.Sprintf("%s-chart", name), ChartVersion: ver, HelmRegistryName: reg,
		Profiles: &profiles, DefaultProfileName: &defaultProfile,
	}, addHeaders)
	if s.validateResponse(err, r) {
		if s.NotNil(r.JSON200) {
			s.validateRESTApplication(r.JSON200.Application, name, ver, display, description, len(profiles), defaultProfile)
			return &r.JSON200.Application
		}
	}
	return nil
}

func (s *TestSuite) validateRESTApplication(app restapi.CatalogV3Application, name string, ver string, display string, description string, profileCount int, defaultProfile string) {
	s.Equal(name, app.Name)
	s.Equal(ver, app.Version)
	s.Equal(display, *app.DisplayName)
	s.Equal(description, *app.Description)
	s.Equal(fmt.Sprintf("%s-chart", name), app.ChartName)
	s.Equal(ver, app.ChartVersion)
	s.Len(*app.Profiles, profileCount)
	s.Equal(defaultProfile, *app.DefaultProfileName)
}

func profileREST(name string, display string, description string, values string) restapi.CatalogV3Profile {
	return restapi.CatalogV3Profile{Name: name, DisplayName: &display, Description: &description, ChartValues: &values}
}

func (s *TestSuite) createRESTArtifact(projectUUID string, name string, display string, description string,
	mimeType string, path string) *restapi.CatalogV3Artifact {
	value, _ := os.ReadFile(path)
	r, err := s.restClient.CatalogServiceCreateArtifactWithResponse(s.ProjectID(projectUUID), projectUUID, restapi.CatalogServiceCreateArtifactJSONRequestBody{
		Name: name, DisplayName: &display, Description: &description, MimeType: mimeType, Artifact: value},
		addHeaders)
	if s.validateResponse(err, r) {
		if s.NotNil(r.JSON200) {
			s.validateRESTArtifact(r.JSON200.Artifact, name, display, description, mimeType, value)
			return &r.JSON200.Artifact
		}
	}
	return nil
}

func (s *TestSuite) validateRESTArtifact(app restapi.CatalogV3Artifact, name string, display string, description string, mimeType string, value []byte) {
	s.Equal(name, app.Name)
	s.Equal(display, *app.DisplayName)
	s.Equal(description, *app.Description)
	s.Equal(mimeType, app.MimeType)
	s.Equal(value, app.Artifact)
}

func (s *TestSuite) createRESTPackage(projectUUID string, name string, ver string, display string, description string, references []restapi.CatalogV3ApplicationReference,
	profiles []restapi.CatalogV3DeploymentProfile, defaultProfile string, extensions []restapi.CatalogV3APIExtension,
	artifacts []restapi.CatalogV3ArtifactReference) *restapi.CatalogV3DeploymentPackage {
	r, err := s.restClient.CatalogServiceCreateDeploymentPackageWithResponse(s.ProjectID(projectUUID), projectUUID, restapi.CatalogServiceCreateDeploymentPackageJSONRequestBody{
		Name: name, Version: ver, DisplayName: &display, Description: &description,
		ApplicationReferences: references, Profiles: &profiles, DefaultProfileName: &defaultProfile,
		Extensions: extensions, Artifacts: artifacts,
	}, addHeaders)
	if s.validateResponse(err, r) {
		if s.NotNil(r.JSON200) {
			s.validateRESTPackage(r.JSON200.DeploymentPackage, name, ver, display, description, len(references), len(profiles), defaultProfile, len(extensions), len(artifacts))
			return &r.JSON200.DeploymentPackage
		}
	}
	return nil
}

func (s *TestSuite) validateRESTPackage(pkg restapi.CatalogV3DeploymentPackage, name string, ver string, display string, description string, referenceCount int, profileCount int, defaultProfile string, extensionCount int, artifactCount int) {
	s.Equal(name, pkg.Name)
	s.Equal(ver, pkg.Version)
	s.Equal(display, *pkg.DisplayName)
	s.Equal(description, *pkg.Description)
	s.Len(pkg.ApplicationReferences, referenceCount)
	s.Len(*pkg.Profiles, profileCount)
	s.Equal(defaultProfile, *pkg.DefaultProfileName)
	s.Len(pkg.Extensions, extensionCount)
	s.Len(pkg.Artifacts, artifactCount)
}

func packageRESTProfile(name string, display string, description string, applicationProfiles map[string]string) restapi.CatalogV3DeploymentProfile {
	return restapi.CatalogV3DeploymentProfile{Name: name, DisplayName: &display, Description: &description, ApplicationProfiles: applicationProfiles}
}

func extensionREST(name string, version string, display string, description string, uiLabel string, uiService string, endpoints []restapi.CatalogV3Endpoint) restapi.CatalogV3APIExtension {
	ext := restapi.CatalogV3APIExtension{Name: name, Version: version, DisplayName: &display, Description: &description, Endpoints: &endpoints}
	if uiLabel != "" {
		ext.UiExtension = &restapi.CatalogV3UIExtension{AppName: name, ModuleName: name, Label: uiLabel, ServiceName: uiService, Description: description, FileName: "none"}
	}
	return ext
}

func findRESTExtension(name string, extensions []restapi.CatalogV3APIExtension) *restapi.CatalogV3APIExtension {
	for _, ext := range extensions {
		if ext.Name == name {
			return &ext
		}
	}
	return nil
}

func (s *TestSuite) validateRESTExtension(ext *restapi.CatalogV3APIExtension, name string, version string, display string, description string,
	uiLabel string, uiService string, endpointCount int) {
	s.Equal(name, ext.Name)
	s.Equal(version, ext.Version)
	s.Equal(display, *ext.DisplayName)
	s.Equal(description, *ext.Description)
	if uiLabel == "" {
		s.Nil(ext.UiExtension)
	} else {
		s.NotNil(ext.UiExtension)
		s.Equal(uiLabel, ext.UiExtension.Label)
		s.Equal(uiService, ext.UiExtension.ServiceName)
	}
	s.Len(*ext.Endpoints, endpointCount)
}

func endpointREST(service string, external string, internal string) restapi.CatalogV3Endpoint {
	return restapi.CatalogV3Endpoint{ServiceName: service, ExternalPath: external, InternalPath: internal}
}

func findRESTEndpoint(service string, endpoints []restapi.CatalogV3Endpoint) *restapi.CatalogV3Endpoint {
	for _, ep := range endpoints {
		if ep.ServiceName == service {
			return &ep
		}
	}
	return nil
}

func (s *TestSuite) validateRESTEndpoint(ep *restapi.CatalogV3Endpoint, service string, external string, internal string) {
	s.Equal(service, ep.ServiceName)
	s.Equal(external, ep.ExternalPath)
	s.Equal(internal, ep.InternalPath)
}

func artifactREST(name string, purpose string) restapi.CatalogV3ArtifactReference {
	return restapi.CatalogV3ArtifactReference{Name: name, Purpose: purpose}
}

func findRESTArtifactReference(name string, artifacts []restapi.CatalogV3ArtifactReference) *restapi.CatalogV3ArtifactReference {
	for _, ar := range artifacts {
		if ar.Name == name {
			return &ar
		}
	}
	return nil
}

func (s *TestSuite) validateRESTArtifactReference(ar *restapi.CatalogV3ArtifactReference, name string, purpose string) {
	s.Equal(name, ar.Name)
	s.Equal(purpose, ar.Purpose)
}

var authToken = ""

func addHeaders(ctx context.Context, req *http.Request) error {
	req.Header.Set("User-Agent", "tests")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		f := md.Get(ActiveProjectID)
		if len(f) > 0 {
			req.Header.Set(ActiveProjectID, f[0])
		}
	}
	return nil
}
