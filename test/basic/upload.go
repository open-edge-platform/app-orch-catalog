// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"fmt"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"os"
)

const librespeedDescription = "This is a very lightweight Speedtest implemented in Javascript, using XMLHttpRequest and Web Workers."

func (s *TestSuite) checkIntelRegistry() {
	resp, err := s.client.GetRegistry(s.AddHeaders("intel"),
		&catalogv3.GetRegistryRequest{
			RegistryName: "intel-harbor",
		})
	s.validateResponse(err, resp)
	s.validateRegistry(resp.Registry, "intel-harbor", "intel-harbor", "The Intel amr caas registry")
}

func (s *TestSuite) checkDeploymentPackage() {
	pkg, err := s.client.GetDeploymentPackage(s.AddHeaders("intel"),
		&catalogv3.GetDeploymentPackageRequest{
			DeploymentPackageName: "librespeed-app",
			Version:               "0.0.2",
		})
	s.validateResponse(err, pkg)
	s.validatePackage(pkg.DeploymentPackage, "librespeed-app", "0.0.2", "librespeed-app", librespeedDescription,
		1, 1, "default", 1, 0)
	s.validateExtension(pkg.DeploymentPackage.Extensions[0], "ext1", "1.2.3",
		"extension 1", "This is Extension #1", "label1", "service1", 2)
	s.validateEndpoint(pkg.DeploymentPackage.Extensions[0].Endpoints[0], "service1", "EPath1", "IPath1")
	s.validateEndpoint(pkg.DeploymentPackage.Extensions[0].Endpoints[1], "service2", "EPath2", "IPath2")
}

func (s *TestSuite) checkApplication() {
	pkg, err := s.client.GetApplication(s.AddHeaders("intel"),
		&catalogv3.GetApplicationRequest{
			ApplicationName: "",
			Version:         "0.0.2",
		})
	s.validateResponse(err, pkg)
	s.validateApplication(pkg.Application, "librespeed-vm", "0.0.2", "librespeed-vm", librespeedDescription,
		1, "default")
	s.Equal("default", pkg.Application.Profiles[0].Name)
	s.Equal("Default", pkg.Application.Profiles[0].DisplayName)
	s.Contains(pkg.Application.Profiles[0].ChartValues, "format: iso8601")
}

func (s *TestSuite) checkThumbnailArtifact() {
	art, err := s.client.GetArtifact(s.AddHeaders("intel"),
		&catalogv3.GetArtifactRequest{
			ArtifactName: "librespeed-thumbnail",
		})
	s.validateResponse(err, art)
	s.validateArtifact(art.Artifact, "librespeed-thumbnail", "librespeed-thumbnail",
		"Icon for librespeed application", "text/plain", []byte("text"))
}

func (s *TestSuite) uploadFile(file string, uploadNumber uint32, lastUpload bool, sessionID string) (string, uint32, error) {
	fmt.Fprintf(os.Stderr, "uploadFile(%s,%d,%v,%s)\n", file, uploadNumber, lastUpload, sessionID)
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return "", 0, err
	}
	r, err := s.client.UploadCatalogEntities(s.AddHeaders("intel"), &catalogv3.UploadCatalogEntitiesRequest{
		SessionId:    sessionID,
		UploadNumber: uploadNumber,
		LastUpload:   lastUpload,
		Upload: &catalogv3.Upload{
			FileName: file,
			Artifact: fileBytes,
		},
	})
	if s.validateResponse(err, r) {
		return r.SessionId, r.UploadNumber, err
	}
	return sessionID, 0, err
}

// TestUpload tests uploading YAML specs via the catalog gRPC API
func (s *TestSuite) TestUpload() {
	s.WipeOutData()

	var err error
	sessionID := ""
	count := 0
	uploadNumber := uint32(1)

	// map is used to guarantee random order in loading the files
	uploads := map[string]bool{
		"/test/basic/testdata/registry-intel.yaml":               false,
		"/test/basic/testdata/artifact.yaml":                     false,
		"/test/basic/testdata/application-librespeed-0.0.2.yaml": false,
		"/test/basic/testdata/deployment-package.yaml":           false,
		"/test/basic/testdata/values.yaml":                       false,
	}

	for file := range uploads {
		sessionID, uploadNumber, err = s.uploadFile(file, uploadNumber, count == len(uploads)-1, sessionID)
		if !s.NoError(err) {
			fmt.Fprintf(os.Stderr, "upload failed: %+v\n", err)
			return
		}
		s.NotEqual("", sessionID)
		count++
	}

	s.checkIntelRegistry()
	s.checkThumbnailArtifact()
	s.checkDeploymentPackage()
}
