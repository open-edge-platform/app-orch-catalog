package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
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
	ctx := context.TODO()
	res, err := s.catalogClient.UploadTarball(ctx, types.WordpressTarballPathName)
	assert.NoError(s.T(), err, "Expected to upload tarball without error")
	assert.Equal(s.T(), http.StatusOK, res.StatusCode, "Expected HTTP status code 200 OK for upload")

	defer res.Body.Close()
	if res.Status != "200 OK" {
		// print response message if something has gone wrong, for debugging
		bodyBytes, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)
		s.T().Logf("Response Body: %s", string(bodyBytes))
	}

	// Make sure the wordpress DP was created
	endpoint := fmt.Sprintf("%s/test-wordpress/versions/0.1.1", types.DeploymentPackagesEndpoint)
	queryParams := map[string]string{
		"orderBy":  "name",
		"pageSize": "10",
		"offset":   "0",
	}

	res, err = s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, queryParams)
	assert.NoError(s.T(), err)

	body, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	var result struct {
		DeploymentPackage restClient.DeploymentPackage `json:"deploymentPackage"`
	}
	s.unmarshalJSON(body, &result)

	assert.Equal(s.T(), types.WordpressName, result.DeploymentPackage.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), types.WordpressVersion, result.DeploymentPackage.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after upload")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after upload")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after upload")
}

func (s *TestSuite) TestUploadSeparateFiles() {
	ctx := context.TODO()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	pathNames := []string{"../testdata/wordpress/app-wordpress-0.1.1.yaml",
		"../testdata/wordpress/dp-wordpress-0.1.1.yaml",
		"../testdata/wordpress/registry-bitnami.yaml",
		"../testdata/wordpress/values-wordpress-0.1.1.yaml",
	}

	for _, pathName := range pathNames {
		file, err := os.Open(pathName)
		assert.NoError(s.T(), err)
		defer file.Close()

		fileName := pathName[strings.LastIndex(pathName, "/")+1:]

		part, _ := writer.CreateFormFile("files", fileName)
		_, err = io.Copy(part, file)
		assert.NoError(s.T(), err)
	}

	writer.Close()

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	res, err := s.catalogClient.MakeAuthenticatedRequest("POST", types.UploadEndpoint, body, nil, headers)
	assert.NoError(s.T(), err)

	defer res.Body.Close()
	assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for upload")
	if res.Status != "200 OK" {
		// print response message if something has gone wrong, for debugging
		bodyBytes, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)
		s.T().Logf("Response Body: %s", string(bodyBytes))
	}

	// Make sure the wordpress DP was created
	endpoint := fmt.Sprintf("%s/test-wordpress/versions/0.1.1", types.DeploymentPackagesEndpoint)
	queryParams := map[string]string{
		"orderBy":  "name",
		"pageSize": "10",
		"offset":   "0",
	}

	res, err = s.catalogClient.MakeAuthenticatedRequest("GET", endpoint, nil, queryParams)
	assert.NoError(s.T(), err)

	resBody, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	var result struct {
		DeploymentPackage restClient.DeploymentPackage `json:"deploymentPackage"`
	}
	s.unmarshalJSON(resBody, &result)

	assert.Equal(s.T(), "test-wordpress", result.DeploymentPackage.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), "0.1.1", result.DeploymentPackage.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.NoError(s.catalogClient.DeleteDeploymentPackage(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete deployment package after upload")
	s.NoError(s.catalogClient.DeleteApplication(ctx, types.WordpressName, types.WordpressVersion, true), "Expected to delete application after upload")
	s.NoError(s.catalogClient.DeleteRegistry(ctx, types.WordpressRegistryName, true), "Expected to delete registry after upload")
}

func (s *TestSuite) TestGetCharts() {
	res, err := s.catalogClient.MakeAuthenticatedRequest("GET", "/catalog.orchestrator.apis/charts", nil, map[string]string{"registry": "harbor-helm-oci"})
	assert.NoError(s.T(), err)

	body, err := s.processResponse(res)
	assert.NoError(s.T(), err)

	// On a fresh orchestrator there should be no charts in the registry
	assert.Equal(s.T(), "null", string(body), "Expected the response body to be empty")
}
