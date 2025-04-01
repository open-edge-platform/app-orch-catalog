// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"fmt"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"math/rand"
	"os"
)

// countRange represents range of entity counts
type countRange struct {
	min int
	max int
}

func (r *countRange) getCount() int {
	if r.max == r.min {
		return r.min
	}
	return r.min + rand.Intn(r.max-r.min)
}

func oneOf(count int) int {
	if count == 0 {
		return 1
	}
	return 1 + rand.Intn(count)
}

func oneItem[T any](list []*T) *T {
	if len(list) == 0 {
		return nil
	}
	return list[rand.Intn(len(list))]
}

type projectSpec struct {
	registries     countRange
	artifacts      countRange
	apps           countRange
	appSpec        *appSpec
	packages       countRange
	appPackageSpec *appPackageSpec
}

// appSpec characterizes a randomly generated application record
type appSpec struct {
	versions countRange
	profiles countRange
}

// appPackageSpec characterizes a randomly generated application package record
type appPackageSpec struct {
	versions   countRange
	profiles   countRange
	apps       countRange
	deps       countRange
	artifacts  countRange
	extensions countRange
	endpoints  countRange
	namespaces countRange
}

func (s *TestSuite) validateResponse(err error, r interface{}) bool {
	return s.NoError(err) && s.NotNil(r)
}

func (s *TestSuite) createRegistry(projectUUID string, name string, display string, description string, url string) *catalogv3.Registry {
	r, err := s.client.CreateRegistry(s.AddHeaders(projectUUID), &catalogv3.CreateRegistryRequest{
		Registry: &catalogv3.Registry{Name: name, DisplayName: display, Description: description, RootUrl: url, Type: "HELM"}})
	if s.validateResponse(err, r) {
		s.validateRegistry(r.Registry, name, display, description)
		return r.Registry
	}
	return nil
}

func (s *TestSuite) validateRegistry(reg *catalogv3.Registry, name string, display string, description string) {
	s.Equal(name, reg.Name)
	s.Equal(display, reg.DisplayName)
	s.Equal(description, reg.Description)
	s.Equal("HELM", reg.Type)
}

func (s *TestSuite) createApplication(projectUUID string, reg string, name string, ver string, display string, description string,
	profiles []*catalogv3.Profile, defaultProfile string) *catalogv3.Application {
	r, err := s.client.CreateApplication(s.AddHeaders(projectUUID), &catalogv3.CreateApplicationRequest{
		Application: &catalogv3.Application{
			Name: name, Version: ver, DisplayName: display, Description: description,
			ChartName: fmt.Sprintf("%s-chart", name), ChartVersion: ver, HelmRegistryName: reg,
			Profiles: profiles, DefaultProfileName: defaultProfile,
		},
	})
	if r == nil {
		fmt.Printf("%v\n", err)
	}
	if s.validateResponse(err, r) {
		s.validateApplication(r.Application, name, ver, display, description, len(profiles), defaultProfile)
		return r.Application
	}
	return nil
}

func (s *TestSuite) validateApplication(app *catalogv3.Application, name string, ver string, display string,
	description string, profileCount int, defaultProfile string) {
	s.Equal(name, app.Name)
	s.Equal(ver, app.Version)
	s.Equal(display, app.DisplayName)
	s.Equal(description, app.Description)
	s.Equal(fmt.Sprintf("%s-chart", name), app.ChartName)
	s.Equal(ver, app.ChartVersion)
	s.Len(app.Profiles, profileCount)
	s.Equal(defaultProfile, app.DefaultProfileName)
}

func profile(name string, display string, description string, values string) *catalogv3.Profile {
	return &catalogv3.Profile{Name: name, DisplayName: display, Description: description, ChartValues: values}
}

func (s *TestSuite) createArtifact(projectUUID string, name string, display string, description string,
	mimeType string, path string) *catalogv3.Artifact {
	value, _ := os.ReadFile(path)
	r, err := s.client.CreateArtifact(s.AddHeaders(projectUUID), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: name, DisplayName: display, Description: description,
			MimeType: mimeType, Artifact: value},
	})
	if s.validateResponse(err, r) {
		s.validateArtifact(r.Artifact, name, display, description, mimeType, value)
		return r.Artifact
	}
	return nil
}

func (s *TestSuite) validateArtifact(app *catalogv3.Artifact, name string, display string, description string, mimeType string, value []byte) {
	s.Equal(name, app.Name)
	s.Equal(display, app.DisplayName)
	s.Equal(description, app.Description)
	s.Equal(mimeType, app.MimeType)
	s.Equal(value, app.Artifact)
}

func (s *TestSuite) createPackage(projectUUID string, name string, ver string, display string, description string, references []*catalogv3.ApplicationReference,
	profiles []*catalogv3.DeploymentProfile, defaultProfile string, extensions []*catalogv3.APIExtension,
	artifacts []*catalogv3.ArtifactReference) *catalogv3.DeploymentPackage {
	r, err := s.client.CreateDeploymentPackage(s.AddHeaders(projectUUID), &catalogv3.CreateDeploymentPackageRequest{
		DeploymentPackage: &catalogv3.DeploymentPackage{
			Name: name, Version: ver, DisplayName: display, Description: description,
			ApplicationReferences: references, Profiles: profiles, DefaultProfileName: defaultProfile,
			Extensions: extensions, Artifacts: artifacts,
		},
	})
	if s.validateResponse(err, r) {
		s.validatePackage(r.DeploymentPackage, name, ver, display, description, len(references), len(profiles), defaultProfile, len(extensions), len(artifacts))
		return r.DeploymentPackage
	}
	return nil
}

func (s *TestSuite) validatePackage(app *catalogv3.DeploymentPackage, name string, ver string, display string,
	description string, referenceCount int, profileCount int, defaultProfile string, extensionCount int, artifactCount int) {
	s.Equal(name, app.Name)
	s.Equal(ver, app.Version)
	s.Equal(display, app.DisplayName)
	s.Equal(description, app.Description)
	s.Len(app.ApplicationReferences, referenceCount)
	s.Len(app.Profiles, profileCount)
	s.Equal(defaultProfile, app.DefaultProfileName)
	s.Len(app.Extensions, extensionCount)
	s.Len(app.Artifacts, artifactCount)
}

func packageProfile(name string, display string, description string, applicationProfiles map[string]string) *catalogv3.DeploymentProfile {
	return &catalogv3.DeploymentProfile{Name: name, DisplayName: display, Description: description, ApplicationProfiles: applicationProfiles}
}

func extension(name string, version string, display string, description string, uiLabel string, uiService string, endpoints []*catalogv3.Endpoint) *catalogv3.APIExtension {
	ext := &catalogv3.APIExtension{Name: name, Version: version, DisplayName: display, Description: description, Endpoints: endpoints}
	if uiLabel != "" {
		ext.UiExtension = &catalogv3.UIExtension{AppName: name, ModuleName: name, Label: uiLabel, ServiceName: uiService, Description: description, FileName: "none"}
	}
	return ext
}

func findExtension(name string, extensions []*catalogv3.APIExtension) *catalogv3.APIExtension {
	for _, ext := range extensions {
		if ext.Name == name {
			return ext
		}
	}
	return nil
}

func (s *TestSuite) validateExtension(ext *catalogv3.APIExtension, name string, version string, display string, description string,
	uiLabel string, uiService string, endpointCount int) {
	s.Equal(name, ext.Name)
	s.Equal(version, ext.Version)
	s.Equal(display, ext.DisplayName)
	s.Equal(description, ext.Description)
	s.Equal(name, ext.Name)
	if uiLabel == "" {
		s.Nil(ext.UiExtension)
	} else {
		s.NotNil(ext.UiExtension)
		s.Equal(uiLabel, ext.UiExtension.Label)
		s.Equal(uiService, ext.UiExtension.ServiceName)
	}
	s.Len(ext.Endpoints, endpointCount)
}

func endpoint(service string, external string, internal string) *catalogv3.Endpoint {
	return &catalogv3.Endpoint{ServiceName: service, ExternalPath: external, InternalPath: internal}
}

func findEndpoint(service string, endpoints []*catalogv3.Endpoint) *catalogv3.Endpoint {
	for _, ep := range endpoints {
		if ep.ServiceName == service {
			return ep
		}
	}
	return nil
}

func (s *TestSuite) validateEndpoint(ep *catalogv3.Endpoint, service string, external string, internal string) {
	s.Equal(service, ep.ServiceName)
	s.Equal(external, ep.ExternalPath)
	s.Equal(internal, ep.InternalPath)
}

func artifact(name string, purpose string) *catalogv3.ArtifactReference {
	return &catalogv3.ArtifactReference{Name: name, Purpose: purpose}
}

func findArtifactReference(name string, artifacts []*catalogv3.ArtifactReference) *catalogv3.ArtifactReference {
	for _, ar := range artifacts {
		if ar.Name == name {
			return ar
		}
	}
	return nil
}

func (s *TestSuite) validateArtifactReference(ar *catalogv3.ArtifactReference, name string, purpose string) {
	s.Equal(name, ar.Name)
	s.Equal(purpose, ar.Purpose)
}
