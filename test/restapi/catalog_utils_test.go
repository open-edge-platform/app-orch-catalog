// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient/utilities"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"github.com/stretchr/testify/assert"

	"io"
	"net/http"
)

// processResponse validates response status and returns body
func (s *TestSuite) processResponse(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	return body, err
}

// unmarshalJSON unmarshals response body into provided result struct
func (s *TestSuite) unmarshalJSON(body []byte, result interface{}) error {
	err := json.Unmarshal(body, result)
	assert.NoError(s.T(), err)
	return err
}

func (s *TestSuite) TestUploadTarball() {
	s.T().Skip()
	ctx := context.TODO()
	_, status, err := s.catalogClient.UploadTarball(ctx, types.WordpressTarballPathName)
	assert.NoError(s.T(), err, "Expected to upload tarball without error")
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for upload")

	// Make sure the wordpress DP was created

	dp, status, err := s.catalogClient.GetDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion)
	assert.NoError(s.T(), err, "Expected to retrieve deployment package after upload")
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for deployment package retrieval")
	assert.NotNil(s.T(), dp, "Expected deployment package to be non-nil after upload")

	assert.Equal(s.T(), types.WordpressName, dp.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), types.WordpressVersion, dp.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after upload")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after upload")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after upload")
}

func (s *TestSuite) TestUploadSeparateFiles() {
	s.T().Skip()
	ctx := context.TODO()

	pathNames := []string{"../testdata/wordpress/app-wordpress-0.1.1.yaml",
		"../testdata/wordpress/dp-wordpress-0.1.1.yaml",
		"../testdata/wordpress/registry-bitnami.yaml",
		"../testdata/wordpress/values-wordpress-0.1.1.yaml",
	}

	for _, pathName := range pathNames {
		resp, status, err := s.catalogClient.UploadTarball(ctx, pathName)
		assert.NoError(s.T(), err, fmt.Sprintf("Expected to upload file %s without error", pathName))
		assert.Equal(s.T(), http.StatusOK, status, fmt.Sprintf("Expected HTTP status code 200 OK for upload of %s", pathName))
		assert.NotNil(s.T(), resp, fmt.Sprintf("Expected response to be non-nil for upload of %s", pathName))

	}

	dp, status, err := s.catalogClient.GetDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion)
	assert.NoError(s.T(), err, "Expected to retrieve deployment package after upload")
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for deployment package retrieval")
	assert.NotNil(s.T(), dp, "Expected deployment package to be non-nil after upload")

	assert.Equal(s.T(), "test-wordpress", dp.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), "0.1.1", dp.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after upload")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after upload")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after upload")
}

func (s *TestSuite) TestGetCharts() {
	ctx := context.TODO()

	resp, status, err := s.catalogClient.GetCharts(ctx, &utilities.CatalogServiceGetRegistryChartsParams{
		Registry: types.GetPointerString("harbor-helm-oci"),
	})
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, status, "Expected HTTP status code 200 OK for getting charts")

	// On a fresh orchestrator there should be no charts in the registry
	assert.Equal(s.T(), "null", string(resp.Body), "Expected the response body to be empty")
}
