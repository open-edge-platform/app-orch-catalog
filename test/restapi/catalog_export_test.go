// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"archive/tar"
	"compress/gzip"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"io"
	"log"
	"net/http"
	// Third-party imports
)

func (s *TestSuite) TestExportDeploymentPackage() {
	s.T().Skip()
	// Before we can test export, first import the wordpress package
	_, err := s.catalogClient.UploadTarball(types.WordpressTarballPathName)
	s.Require().NoError(err, "Expected to upload tarball before exporting")

	res, err := s.catalogClient.ExportDeploymentPackage("test-wordpress", "0.1.1")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, res.StatusCode, "Expected HTTP status code 200 OK for export")
	defer res.Body.Close()

	files := make(map[string][]byte)

	// decompress the tarball
	gzReader, err := gzip.NewReader(res.Body)
	s.Require().NoError(err, "Expected to create gzip reader")
	defer gzReader.Close()

	// read the files from the tarbvall
	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		s.Require().NoError(err, "Error reading tar entry")

		if header.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tarReader)
			s.Require().NoError(err, "Error reading file content")
			files[header.Name] = content
		}
	}

	for fileName, content := range files {
		log.Printf("Found file in tarball: %s, size: %d bytes", fileName, len(content))
	}

	_, ok := files["test-wordpress-deployment-package.yaml"]
	s.True(ok, "Expected to find 'wordpress-deployment-package.yaml' in the tarball")

	_, ok = files["test-wordpress-application.yaml"]
	s.True(ok, "Expected to find 'wordpress-application.yaml' in the tarball")

	_, ok = files["test-bitnami-registry.yaml"]
	s.True(ok, "Expected to find 'bitnami-registry.yaml' in the tarball")

	_, ok = files["values-test-wordpress-0.1.1-default.yaml"]
	s.True(ok, "Expected to find 'values-wordpress-0.1.1-default.yaml' in the tarball")

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after export")
	s.NoError(s.catalogClient.DeleteApplication(types.WordpressName, types.WordpressVersion, true), "Expected to delete application after export")
	s.NoError(s.catalogClient.DeleteRegistry(types.WordpressRegistryName, true), "Expected to delete registry after export")
}

func (s *TestSuite) TestExportDeploymentPackageNoExist() {
	s.T().Skip()
	res, err := s.catalogClient.ExportDeploymentPackage("not-a-real-package", "0.1.1")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotFound, res.StatusCode, "Expected HTTP status code 404 for export")
	defer res.Body.Close()
}
