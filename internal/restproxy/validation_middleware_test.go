// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryParameterValidationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		query          string
		expectedStatus int
		expectedError  string
	}{
		// GET /registries/{name} - Valid cases
		{
			name:           "GET registry - no query params",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET registry - valid showSensitiveInfo param",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "showSensitiveInfo=true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET registry - showSensitiveInfo false",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "showSensitiveInfo=false",
			expectedStatus: http.StatusOK,
		},

		// GET /registries/{name} - Invalid cases
		{
			name:           "GET registry - unknown parameter",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "unknownParam=value",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "GET registry operation does not support query parameters: unknownParam",
		},
		{
			name:           "GET registry - multiple unknown parameters",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "unknown1=value1&unknown2=value2",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "GET registry operation does not support query parameters:",
		},
		{
			name:           "GET registry - mix of valid and invalid parameters",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "showSensitiveInfo=true&invalidParam=value",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "GET registry operation does not support query parameters: invalidParam",
		},

		// DELETE /registries/{name} - Valid cases
		{
			name:           "DELETE registry - no query params",
			method:         "DELETE",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "",
			expectedStatus: http.StatusOK,
		},

		// DELETE /registries/{name} - Invalid cases
		{
			name:           "DELETE registry - any query param should fail",
			method:         "DELETE",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "someParam=value",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "DELETE registry operation does not accept query parameters, received: someParam",
		},
		{
			name:           "DELETE registry - multiple query params should fail",
			method:         "DELETE",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "param1=value1&param2=value2",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "DELETE registry operation does not accept query parameters, received: param1, param2",
		},
		{
			name:           "DELETE registry - even showSensitiveInfo should fail",
			method:         "DELETE",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "showSensitiveInfo=true",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "DELETE registry operation does not accept query parameters, received: showSensitiveInfo",
		},

		// GET /registries - Valid cases
		{
			name:           "List registries - no query params",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List registries - all valid params",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries",
			query:          "orderBy=name&filter=type%3DHELM&pageSize=10&offset=0&showSensitiveInfo=false",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List registries - some valid params",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries",
			query:          "pageSize=20&showSensitiveInfo=true",
			expectedStatus: http.StatusOK,
		},

		// GET /registries - Invalid cases
		{
			name:           "List registries - unknown parameter",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries",
			query:          "unknownParam=value",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "list registries operation does not support query parameters: unknownParam",
		},
		{
			name:           "List registries - mix of valid and invalid",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries",
			query:          "pageSize=10&invalidParam=test",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "list registries operation does not support query parameters: invalidParam",
		},

		// URL format validation
		{
			name:           "Multiple question marks in URL",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/registries/test-registry",
			query:          "param1=value1&param2=value2", // This is how it appears after Go parses the malformed URL
			expectedStatus: http.StatusBadRequest,
			expectedError:  "GET registry operation does not support query parameters",
		},

		// Non-registry endpoints should pass through
		{
			name:           "Non-registry endpoint should pass",
			method:         "GET",
			path:           "/catalog.orchestrator.apis/v3/applications",
			query:          "anyParam=value",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test router with the middleware
			router := gin.New()
			router.Use(QueryParameterValidationMiddleware())

			// Add a test handler that just returns OK if validation passes
			router.Any("/*path", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Create test request
			var req *http.Request
			var err error

			if tt.query != "" {
				req, err = http.NewRequest(tt.method, tt.path+"?"+tt.query, nil)
			} else {
				req, err = http.NewRequest(tt.method, tt.path, nil)
			}
			require.NoError(t, err)

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Unexpected status code for test: %s", tt.name)

			// If we expect an error, check the error message
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError, "Expected error message not found for test: %s", tt.name)
			}
		})
	}
}

func TestValidateRegistryEndpointParams(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		queryParams url.Values
		expectErr   bool
		errMsg      string
	}{
		// Single registry GET tests
		{
			name:        "GET single registry - no params",
			method:      "GET",
			path:        "/catalog.orchestrator.apis/v3/registries/test-registry",
			queryParams: url.Values{},
			expectErr:   false,
		},
		{
			name:   "GET single registry - valid showSensitiveInfo",
			method: "GET",
			path:   "/catalog.orchestrator.apis/v3/registries/test-registry",
			queryParams: url.Values{
				"showSensitiveInfo": []string{"true"},
			},
			expectErr: false,
		},
		{
			name:   "GET single registry - invalid param",
			method: "GET",
			path:   "/catalog.orchestrator.apis/v3/registries/test-registry",
			queryParams: url.Values{
				"invalidParam": []string{"value"},
			},
			expectErr: true,
			errMsg:    "GET registry operation does not support query parameters: invalidParam",
		},

		// Single registry DELETE tests
		{
			name:        "DELETE single registry - no params",
			method:      "DELETE",
			path:        "/catalog.orchestrator.apis/v3/registries/test-registry",
			queryParams: url.Values{},
			expectErr:   false,
		},
		{
			name:   "DELETE single registry - any param should fail",
			method: "DELETE",
			path:   "/catalog.orchestrator.apis/v3/registries/test-registry",
			queryParams: url.Values{
				"someParam": []string{"value"},
			},
			expectErr: true,
			errMsg:    "DELETE registry operation does not accept query parameters",
		},

		// List registries tests
		{
			name:        "List registries - no params",
			method:      "GET",
			path:        "/catalog.orchestrator.apis/v3/registries",
			queryParams: url.Values{},
			expectErr:   false,
		},
		{
			name:   "List registries - all valid params",
			method: "GET",
			path:   "/catalog.orchestrator.apis/v3/registries",
			queryParams: url.Values{
				"orderBy":           []string{"name"},
				"filter":            []string{"type=HELM"},
				"pageSize":          []string{"10"},
				"offset":            []string{"0"},
				"showSensitiveInfo": []string{"false"},
			},
			expectErr: false,
		},
		{
			name:   "List registries - invalid param",
			method: "GET",
			path:   "/catalog.orchestrator.apis/v3/registries",
			queryParams: url.Values{
				"invalidParam": []string{"value"},
			},
			expectErr: true,
			errMsg:    "list registries operation does not support query parameters: invalidParam",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegistryEndpointParams(tt.method, tt.path, tt.queryParams)

			if tt.expectErr {
				assert.Error(t, err, "Expected error for test: %s", tt.name)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error for test: %s", tt.name)
			}
		})
	}
}

func TestValidateQueryParameterValues(t *testing.T) {
	tests := []struct {
		name        string
		queryParams url.Values
		expectErr   bool
		errMsg      string
	}{
		{
			name:        "No params",
			queryParams: url.Values{},
			expectErr:   false,
		},
		{
			name: "Valid showSensitiveInfo true",
			queryParams: url.Values{
				"showSensitiveInfo": []string{"true"},
			},
			expectErr: false,
		},
		{
			name: "Valid showSensitiveInfo false",
			queryParams: url.Values{
				"showSensitiveInfo": []string{"false"},
			},
			expectErr: false,
		},
		{
			name: "Invalid showSensitiveInfo value",
			queryParams: url.Values{
				"showSensitiveInfo": []string{"invalid"},
			},
			expectErr: true,
			errMsg:    "showSensitiveInfo parameter must be 'true' or 'false'",
		},
		{
			name: "Valid pageSize",
			queryParams: url.Values{
				"pageSize": []string{"10"},
			},
			expectErr: false,
		},
		{
			name: "Valid offset",
			queryParams: url.Values{
				"offset": []string{"5"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueryParameterValues(tt.queryParams)

			if tt.expectErr {
				assert.Error(t, err, "Expected error for test: %s", tt.name)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error for test: %s", tt.name)
			}
		})
	}
}
