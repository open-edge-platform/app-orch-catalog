// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	internaltesting "github.com/open-edge-platform/app-orch-catalog/internal/testing"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
	"os"
	"time"
)

func (s *NorthBoundTestSuite) getUpload(fileName string) *catalogv3.Upload {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		s.Failf("unable to load file %s: %+v", fileName, err)
		return nil
	}
	return &catalogv3.Upload{
		FileName: fileName,
		Artifact: bytes,
	}
}

func (s *NorthBoundTestSuite) uploadFile(ctx context.Context, fileName string, sessionID string, lastUpload bool) *catalogv3.UploadCatalogEntitiesResponse {
	resp, err := s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: sessionID, LastUpload: lastUpload, Upload: s.getUpload(fileName),
	})
	s.NoError(err)
	s.NotNil(resp)
	return resp
}

func (s *NorthBoundTestSuite) TestUploadBasics() {
	// Test create
	s.uploadThings()
	s.validateUpload()

	// Test no-op update
	s.uploadThings()
	s.validateUpload()
}

func (s *NorthBoundTestSuite) TestReUploadApplicationWithDependencies() {
	ctx := s.ProjectID(footen)
	resp, err := s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: false, Upload: s.getUpload("testdata/app-with-deps.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: true, Upload: s.getUpload("testdata/values.yaml"),
	})
	s.validateResponse(err, resp)

	gresp, err := s.client.GetApplication(ctx, &catalogv3.GetApplicationRequest{
		ApplicationName: "app-with-deps", Version: "1.0.1",
	})
	s.validateResponse(err, gresp)
	if s.Len(gresp.Application.Profiles, 1) {
		if s.Len(gresp.Application.Profiles[0].DeploymentRequirement, 1) {
			s.Equal("cp-1", gresp.Application.Profiles[0].DeploymentRequirement[0].DeploymentProfileName)
		}
	}

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: false, Upload: s.getUpload("testdata/app-with-deps-edited.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: true, Upload: s.getUpload("testdata/values.yaml"),
	})
	s.validateResponse(err, resp)
	gresp, err = s.client.GetApplication(ctx, &catalogv3.GetApplicationRequest{
		ApplicationName: "app-with-deps", Version: "1.0.1",
	})
	s.validateResponse(err, gresp)
	if s.Len(gresp.Application.Profiles, 1) {
		if s.Len(gresp.Application.Profiles[0].DeploymentRequirement, 1) {
			s.Equal("cp-2", gresp.Application.Profiles[0].DeploymentRequirement[0].DeploymentProfileName)
		}
	}
}

func (s *NorthBoundTestSuite) uploadThings() {
	ctx := s.ProjectID("intel")
	resp, err := s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: false, Upload: s.getUpload("testdata/registry-intel.yaml"),
	})
	s.validateResponse(err, resp)
	s.True(resp.SessionId != "")

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: false, Upload: s.getUpload("testdata/registry-new.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: false, Upload: s.getUpload("testdata/artifact.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: false, Upload: s.getUpload("testdata/application-librespeed.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: false, Upload: s.getUpload("testdata/application-librespeed-0.0.2.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: false, Upload: s.getUpload("testdata/deployment-package.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: false, Upload: s.getUpload("testdata/deployment-package-old.yaml"),
	})
	s.validateResponse(err, resp)

	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: resp.SessionId, LastUpload: true, Upload: s.getUpload("testdata/values.yaml"),
	})
	s.validateResponse(err, resp)
}

func (s *NorthBoundTestSuite) validateUpload() {
	ctx := s.ProjectID("intel")
	greg, err := s.client.GetRegistry(ctx, &catalogv3.GetRegistryRequest{RegistryName: "intel-harbor", ShowSensitiveInfo: true})
	s.validateResponse(err, greg)
	s.validateRegistry(greg.Registry, "intel-harbor", "intel-harbor", "The registry",
		"https://registry.intel.com/repo/", "", "",
		"-----BEGIN CERTIFICATE-----\nMIIF1DCCA7ygAwIBAgITEwDCtSGK4TAz...")

	greg, err = s.client.GetRegistry(ctx, &catalogv3.GetRegistryRequest{RegistryName: "intel-new", ShowSensitiveInfo: true})
	s.validateResponse(err, greg)
	s.validateRegistry(greg.Registry, "intel-new", "intel-new", "The registry",
		"https://registry.intel.com/repo/", "", "", "")

	gart, err := s.client.GetArtifact(ctx, &catalogv3.GetArtifactRequest{ArtifactName: "librespeed-thumbnail"})
	s.validateResponse(err, gart)
	s.validateArtifact(gart.Artifact, "librespeed-thumbnail", "librespeed-thumbnail", "Icon for librespeed application",
		"image/png", nil)

	gapp, err := s.client.GetApplication(ctx, &catalogv3.GetApplicationRequest{ApplicationName: "librespeed-vm", Version: "0.0.3"})
	s.validateResponse(err, gapp)
	s.validateApp(gapp.Application, "librespeed-vm", "0.0.3", "librespeed-vm",
		"This is a very lightweight Speedtest implemented in Javascript, using XMLHttpRequest and Web Workers.",
		1, "default", "librespeed-vm", "0.1.3", "intel-harbor")
	s.Len(gapp.Application.IgnoredResources, 1)

	gapp, err = s.client.GetApplication(ctx, &catalogv3.GetApplicationRequest{ApplicationName: "librespeed-vm", Version: "0.0.2"})
	s.validateResponse(err, gapp)
	s.validateApp(gapp.Application, "librespeed-vm", "0.0.2", "librespeed-vm",
		"This is a very lightweight Speedtest implemented in Javascript, using XMLHttpRequest and Web Workers.",
		1, "default", "librespeed-vm", "0.1.2", "intel-harbor")
	s.Len(gapp.Application.IgnoredResources, 0)
	templs := gapp.Application.Profiles[0].ParameterTemplates
	s.Len(templs, 3)
	s.validateParameterTemplate(templs[0], "t1", "string", "value3", []string{"value1", "value2", "value3"})
	s.validateParameterTemplate(templs[1], "t2", "number", "1", []string{"1", "2", "3", "4", "5"})
	s.validateParameterTemplate(templs[2], "t3", "string", "", []string{})
	s.True(templs[2].Secret)
	s.True(templs[2].Mandatory)

	gpkg, err := s.client.GetDeploymentPackage(ctx, &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "librespeed-app",
		Version:               "0.0.2",
	})
	s.validateResponse(err, gpkg)
	s.validateDeploymentPkg(gpkg.DeploymentPackage, "librespeed-app", "0.0.2", "librespeed-app",
		"This is a very lightweight Speedtest implemented in Javascript, using XMLHttpRequest and Web Workers.",
		"librespeed-thumbnail", "librespeed-thumbnail", 1, 0, 1, "default", 0, false)
	s.Len(gpkg.DeploymentPackage.Extensions, 1)
	s.Len(gpkg.DeploymentPackage.Namespaces, 2)

	gpkg, err = s.client.GetDeploymentPackage(ctx, &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "librespeed-app-old",
		Version:               "0.0.2",
	})
	s.validateResponse(err, gpkg)
	s.validateDeploymentPkg(gpkg.DeploymentPackage, "librespeed-app-old", "0.0.2", "librespeed-app-old",
		"This is a very lightweight Speedtest implemented in Javascript, using XMLHttpRequest and Web Workers.",
		"librespeed-thumbnail", "librespeed-thumbnail", 1, 0, 1, "default", 0, false)
}

func (s *NorthBoundTestSuite) TestUploadBadYaml() {
	var (
		err  error
		resp *catalogv3.UploadCatalogEntitiesResponse
	)
	files, err := os.ReadDir("testdata/badyaml")
	s.NoError(err)
	for _, file := range files {
		resp, err = s.client.UploadCatalogEntities(s.ctx, &catalogv3.UploadCatalogEntitiesRequest{
			SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/badyaml/" + file.Name()),
		})
		s.Error(err)
		s.Nil(resp)
	}
}

func (s *NorthBoundTestSuite) TestLPOD4317() {
	var (
		err  error
		resp *catalogv3.UploadCatalogEntitiesResponse
	)
	files, err := os.ReadDir("testdata/bloggen-sample")
	s.NoError(err)
	sessionID := ""
	for i, file := range files {
		last := i == len(files)-1
		resp, err = s.client.UploadCatalogEntities(s.ProjectID("intel"), &catalogv3.UploadCatalogEntitiesRequest{
			SessionId: sessionID, LastUpload: last, Upload: s.getUpload("testdata/bloggen-sample/" + file.Name()),
		})
		if err != nil {
			break
		}
		sessionID = resp.SessionId
	}
	s.Error(err)
	s.Contains(err.Error(), "document decode failed: yaml: unmarshal errors:\n  line 21: mapping key \"displayName\" already defined at line 20")
	s.Nil(resp)
}

func (s *NorthBoundTestSuite) TestUploadMalwareArtifact() {
	var (
		err  error
		resp *catalogv3.UploadCatalogEntitiesResponse
	)
	server := internaltesting.StartMalwareServer()
	malware.DefaultScanner = malware.NewScanner(":1123", time.Duration(5)*time.Second, false)

	defer func() {
		_ = server.Close()
		malware.DefaultScanner = nil
	}()

	ctx := s.ProjectID("intel")

	// Upload a text file with a test malware signature. Should fail
	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/malware/malware-artifact.yaml"),
	})
	s.Error(err)
	s.Contains(err.Error(), "artifact invalid: malware detected")
	s.Nil(resp)

	// Upload an ordinary text file. Should be OK
	resp, err = s.client.UploadCatalogEntities(ctx, &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/malware/ok-artifact.yaml"),
	})

	s.NoError(err)
	s.NotNil(resp)
	s.Nil(resp.ErrorMessages)
}

func (s *NorthBoundTestSuite) TestUploadMalware() {
	var (
		err  error
		resp *catalogv3.UploadCatalogEntitiesResponse
	)
	server := internaltesting.StartMalwareServer()
	malware.DefaultScanner = malware.NewScanner(":1123", time.Duration(5)*time.Second, false)

	defer func() {
		_ = server.Close()
		malware.DefaultScanner = nil
	}()

	time.Sleep(1 * time.Second) // FIXME: Pause briefly to allow the mock to startup

	// Upload a text file with a test malware signature. Should fail
	resp, err = s.client.UploadCatalogEntities(s.ProjectID("intel"), &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/malware/malware-values.yaml"),
	})
	s.Error(err)
	s.Contains(err.Error(), "invalid: malware detected")
	s.Nil(resp)

}

func (s *NorthBoundTestSuite) TestPermissiveMalwareFailure() {
	malware.DefaultScanner = malware.NewScanner(":1123", time.Duration(5)*time.Second, true)
	defer func() {
		malware.DefaultScanner = nil
	}()

	// Upload an ordinary text file. Should be OK
	resp, err := s.client.UploadCatalogEntities(s.ProjectID("intel"), &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/malware/ok-artifact.yaml"),
	})

	s.NoError(err)
	s.NotNil(resp)
	s.Nil(resp.ErrorMessages)
}

func (s *NorthBoundTestSuite) TestStrictMalwareFailure() {
	malware.DefaultScanner = malware.NewScanner(":1123", time.Duration(5)*time.Second, false)
	defer func() {
		malware.DefaultScanner = nil
	}()

	resp, err := s.client.UploadCatalogEntities(s.ProjectID("intel"), &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/malware/ok-artifact.yaml"),
	})

	s.Error(err)
	s.Nil(resp)
	s.Contains(err.Error(), "unavailable")
}

func (s *NorthBoundTestSuite) TestArtifactMalwareScanError() {
	u := &uploadSession{}
	d := upload.YamlSpec{}

	malware.DefaultScanner = malware.NewScanner(":1123", time.Duration(5)*time.Second, false)
	defer func() {
		malware.DefaultScanner = nil
	}()
	s.Error(u.loadArtifact(s.ctx, nil, d))
}

func (s *NorthBoundTestSuite) TestUploadBadBase64() {
	var (
		err  error
		resp *catalogv3.UploadCatalogEntitiesResponse
	)

	// Upload a text file with a test malware signature. Should fail
	resp, err = s.client.UploadCatalogEntities(s.ProjectID("intel"), &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/badyaml/bad-encoding-artifact.yaml"),
	})
	s.Error(err)
	s.Contains(err.Error(), "artifact invalid: error decoding base64")
	s.Nil(resp)

	// Upload an ordinary text file. Should be OK
	resp, err = s.client.UploadCatalogEntities(s.ProjectID("intel"), &catalogv3.UploadCatalogEntitiesRequest{
		SessionId: "", LastUpload: true, Upload: s.getUpload("testdata/malware/ok-artifact.yaml"),
	})

	s.NoError(err)
	s.NotNil(resp)
	s.Nil(resp.ErrorMessages)
}

func (s *NorthBoundTestSuite) TestUploadAppDeps() {
	var (
		err  error
		resp *catalogv3.UploadCatalogEntitiesResponse
	)
	ctx := s.ProjectID("intel")

	_ = s.uploadFile(ctx, "testdata/upload-apps/app-AB.yaml", "", true)
	resp = s.uploadFile(ctx, "testdata/upload-apps/values.yaml", "", false)
	_ = s.uploadFile(ctx, "testdata/upload-apps/app-C.yaml", resp.SessionId, true)

	// Check some values to be sure they are correct
	appB, err := s.client.GetDeploymentPackage(ctx, &catalogv3.GetDeploymentPackageRequest{DeploymentPackageName: "b", Version: "0.0.1"})
	s.validateResponse(err, appB)
	s.Len(appB.DeploymentPackage.ApplicationDependencies, 1)
	s.Equal("a", appB.DeploymentPackage.ApplicationReferences[0].Name)
	s.Len(appB.DeploymentPackage.ApplicationReferences, 2)
	expectedRefs := []string{"a", "b"}
	foundRefs := 0
	for i, expectedRef := range expectedRefs {
		if expectedRef == appB.DeploymentPackage.ApplicationReferences[i].Name {
			foundRefs++
			continue
		}
	}
	s.Equal(2, foundRefs)
	s.Len(appB.DeploymentPackage.DefaultNamespaces, 1)
	s.Equal("ns", appB.DeploymentPackage.DefaultNamespaces["b"])
}
