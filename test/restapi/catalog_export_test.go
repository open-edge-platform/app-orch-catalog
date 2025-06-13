// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"io"
	"log"
	"net/http"
)

func (s *TestSuite) TestExportDeploymentPackage() {
	ctx := context.TODO()
	// Before we can test export, first import the wordpress package
	_, status, err := s.catalogClient.UploadTarball(ctx, types.WordpressTarballPathName)
	s.Require().NoError(err, "Expected to upload tarball before exporting")
	s.Require().Equal(http.StatusOK, status, "Expected HTTP status code 200 for upload")

	resp, status, err := s.catalogClient.ExportDeploymentPackage(ctx, "test-wordpress", "0.1.1")
	s.Require().NoError(err, "Expected to export deployment package without error")
	s.Require().Equal(http.StatusOK, status, "Expected HTTP status code 200 for export")

	files := make(map[string][]byte)

	// decompress the tarball
	gzReader, err := gzip.NewReader(bytes.NewReader(resp.Body))
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
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after export")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after export")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after export")
}

func (s *TestSuite) TestExportDeploymentPackageNoExist() {
	ctx := context.TODO()
	_, status, err := s.catalogClient.ExportDeploymentPackage(ctx, "not-a-real-package", "0.1.1")
	s.Require().Error(err)
	s.Require().Equal(http.StatusNotFound, status, "Expected HTTP status code 404 for export")
}
