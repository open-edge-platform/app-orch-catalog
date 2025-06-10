// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	// Third-party imports

	// Project-specific imports
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
)

const exportEndpoint = "/catalog.orchestrator.apis/v3/export"

func (s *TestSuite) ExportDeploymentPackage(name string, version string) (int, io.Reader) {
	params := url.Values{}
	params.Add("deployment_package_name", name)
	params.Add("version", version)

	requestURL := fmt.Sprintf("%s%s?%s", s.CatalogRESTServerUrl, exportEndpoint, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
	s.Require().NoError(err, "Expected to create HTTP request for exporting Deployment Package")
	auth.AddRestAuthHeader(req, s.token, s.projectID)

	log.Printf("Exporting Helm chart with request URL: %s", requestURL)

	res, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer res.Body.Close()

	//	body, err := io.ReadAll(res.Body)
	//	s.Require().NoError(err)

	return res.StatusCode, res.Body
}

func (s *TestSuite) TestExportDeploymentPackage() {
	_, err := s.UploadTarball(wordpressTarballPathName)
	s.Require().NoError(err, "Expected to upload tarball before exporting")

	statusCode, body := s.ExportDeploymentPackage("wordpress", "1.0.0")
	s.Require().Equal(http.StatusOK, statusCode, "Expected HTTP status code 200 OK for export")

	files := make(map[string][]byte)

	// decompress the tarball
	gzReader, err := gzip.NewReader(body)
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

	_, ok := files["wordpress-deployment-package.yaml"]
	s.True(ok, "Expected to find 'wordpress-deployment-package.yaml' in the tarball")

	_, ok = files["wordpress-application.yaml"]
	s.True(ok, "Expected to find 'wordpress-application.yaml' in the tarball")

	_, ok = files["bitnami-registry.yaml"]
	s.True(ok, "Expected to find 'bitnami-registry.yaml' in the tarball")

	_, ok = files["values-wordpress-0.1.1-default.yaml"]
	s.True(ok, "Expected to find 'values-wordpress-0.1.1-default.yaml' in the tarball")
}
