// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/open-edge-platform/orch-library/go/dazl"
)

var validationLog = dazl.GetPackageLogger()

// QueryParameterValidationMiddleware creates middleware that validates query parameters
// for REST API endpoints to ensure only documented parameters are accepted
func QueryParameterValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Only validate registry endpoints (which arrive in this format after API gateway processing)
		if strings.Contains(path, "/catalog.orchestrator.apis/v3/registries") {
			validationLog.Infof("Registry endpoint detected, validating parameters for: %s %s", method, path)
			if err := validateRegistryEndpointParams(method, path, c.Request.URL.Query()); err != nil {
				validationLog.Warnf("Invalid query parameters for %s %s: %v", method, path, err)
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Bad Request",
					"message": err.Error(),
					"code":    "INVALID_QUERY_PARAMETERS",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
} // validateRegistryEndpointParams validates query parameters for registry endpoints
func validateRegistryEndpointParams(method, path string, queryParams url.Values) error {
	// Extract registry name from path to determine if this is a single registry endpoint
	// Path format is always: /catalog.orchestrator.apis/v3/registries[/{name}]
	pathParts := strings.Split(path, "/")

	// Find the "registries" part and check what comes after
	var isSingleRegistryEndpoint bool
	for i, part := range pathParts {
		if part == "registries" {
			// If there's another non-empty path segment after "registries", it's a single registry endpoint
			isSingleRegistryEndpoint = i < len(pathParts)-1 && pathParts[i+1] != ""
			break
		}
	}

	if isSingleRegistryEndpoint {
		// Single registry endpoints: GET /registries/{name}, DELETE /registries/{name}, PUT /registries/{name}
		return validateSingleRegistryParams(method, queryParams)
	}
	// List endpoint: GET /registries
	return validateListRegistriesParams(method, queryParams)
}

// validateSingleRegistryParams validates parameters for single registry operations
func validateSingleRegistryParams(method string, queryParams url.Values) error {
	switch method {
	case "GET":
		// GET /registries/{name} - only allows showSensitiveInfo
		allowedParams := map[string]bool{
			"showSensitiveInfo": true,
		}
		return validateAllowedParams(queryParams, allowedParams, "GET registry")

	case "DELETE":
		// DELETE /registries/{name} - no query parameters allowed
		if len(queryParams) > 0 {
			paramNames := make([]string, 0, len(queryParams))
			for param := range queryParams {
				paramNames = append(paramNames, param)
			}
			sort.Strings(paramNames)
			return fmt.Errorf("DELETE registry operation does not accept query parameters, received: %s",
				strings.Join(paramNames, ", "))
		}
		return nil

	case "PUT":
		// PUT /registries/{name} - no query parameters allowed
		if len(queryParams) > 0 {
			paramNames := make([]string, 0, len(queryParams))
			for param := range queryParams {
				paramNames = append(paramNames, param)
			}
			sort.Strings(paramNames)
			return fmt.Errorf("UPDATE registry operation does not accept query parameters, received: %s",
				strings.Join(paramNames, ", "))
		}
		return nil

	default:
		return nil // Allow other methods to pass through
	}
}

// validateListRegistriesParams validates parameters for list registries operation
func validateListRegistriesParams(method string, queryParams url.Values) error {
	if method != "GET" {
		return nil // Only validate GET for list operations
	}

	// GET /registries - allows specific pagination and filtering parameters
	allowedParams := map[string]bool{
		"orderBy":           true,
		"filter":            true,
		"pageSize":          true,
		"offset":            true,
		"showSensitiveInfo": true,
	}

	return validateAllowedParams(queryParams, allowedParams, "list registries")
}

// validateAllowedParams checks that only allowed parameters are present
func validateAllowedParams(queryParams url.Values, allowedParams map[string]bool, operation string) error {
	var invalidParams []string

	for param := range queryParams {
		if !allowedParams[param] {
			invalidParams = append(invalidParams, param)
		}
	}

	if len(invalidParams) > 0 {
		var allowedList []string
		for param := range allowedParams {
			allowedList = append(allowedList, param)
		}

		sort.Strings(invalidParams)
		sort.Strings(allowedList)

		return fmt.Errorf("%s operation does not support query parameters: %s. Allowed parameters: %s",
			operation, strings.Join(invalidParams, ", "), strings.Join(allowedList, ", "))
	}

	return nil
}

// ValidateQueryParameterValues validates the values of specific query parameters
func ValidateQueryParameterValues(queryParams url.Values) error {
	// Validate showSensitiveInfo parameter if present
	if showSensitive := queryParams.Get("showSensitiveInfo"); showSensitive != "" {
		if showSensitive != "true" && showSensitive != "false" {
			return fmt.Errorf("showSensitiveInfo parameter must be 'true' or 'false', received: %s", showSensitive)
		}
	}

	return nil
}
