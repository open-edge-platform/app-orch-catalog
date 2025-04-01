// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	// Registries with this value for inventory URL are allowed to use dynamically loaded admin credentials
	dynamicAdminAuthDomain = `harbor-core.*\.svc\.cluster\.local`
)

func fetchLegacyChartsList(c *gin.Context, registry *catalogv3.Registry, chartName string) {
	inventoryURL := registry.InventoryUrl
	if chartName != "" {
		// If chart parameter is specified, append it to the URL to obtain the versions
		if strings.HasSuffix(inventoryURL, "/") {
			inventoryURL = fmt.Sprintf("%s%s", inventoryURL, chartName)
		} else {
			inventoryURL = fmt.Sprintf("%s/%s", inventoryURL, chartName)
		}
	}

	/* TODO: temporary change to avoid breaking GUI Harbor link */
	if registry.Name == "harbor-helm" {
		parsedURL, err := url.Parse(inventoryURL)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		parsedURL.Scheme = "http"
		parsedURL.Host = dynamicAdminAuthDomain
		inventoryURL = parsedURL.String()
	}

	resBody, ok := callHarborAPI(c, registry, inventoryURL)
	if !ok {
		return
	}

	items, err := parseLegacyRecords(resBody, chartName)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	writeResponse(c, items)
}

func callHarborAPI(c *gin.Context, registry *catalogv3.Registry, inventoryURL string) ([]byte, bool) {
	log.Infof("Directing request to [%s]", inventoryURL)

	caCert := []byte(registry.Cacerts)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caCertPool,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, inventoryURL, nil)
	if err != nil {
		log.Errorf("Unable to create request for charts from registry: %+v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return nil, false
	}

	// Assume that credentials come from the registry entity by default
	username := registry.Username
	tokenOrPassword := registry.AuthToken

	// If the inventoryURL matches a sanctioned domain, fetch the admin credentials dynamically
	dynamicAdminAuthDomainRE := regexp.MustCompile(dynamicAdminAuthDomain)
	if dynamicAdminAuthDomainRE.MatchString(inventoryURL) {
		username, tokenOrPassword, err = readAdminSecret()
		if err != nil {
			log.Errorf("Unable to fetch admin secret: %+v", err)
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return nil, false
		}
	}

	if username != "" {
		// If we have a username, inject it together with the password into the request as a basic authorization header
		req.SetBasicAuth(username, tokenOrPassword)
	} else if tokenOrPassword != "" {
		// Otherwise, if the registry specified just an auth token, inject it into the request as a bearer authorization header
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenOrPassword))
	}

	res, err := client.Do(req)
	if err != nil {
		log.Errorf("Unable to fetch charts from registry: %+v", err)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return nil, false
	}
	if res.StatusCode != http.StatusOK {
		// Pass through the status code to our caller.
		c.AbortWithStatus(res.StatusCode)
		return nil, false
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Unable to read chart records from registry: %+v", err)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return nil, false
	}
	return resBody, true
}

func writeResponse(c *gin.Context, items []string) {
	response, err := json.Marshal(items)
	if err != nil {
		log.Errorf("Unable to marshal chart names: %+v", err)
		return
	}

	if _, err := c.Writer.Write(response); err != nil {
		log.Errorf("Unable to write response: %+v", err)
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	}
}

type LegacyChartRecord struct {
	Name       string
	Version    string
	Deprecated bool
}

func parseLegacyRecords(response []byte, chartName string) ([]string, error) {
	var records []LegacyChartRecord
	var err error
	if err = json.Unmarshal(response, &records); err != nil {
		log.Errorf("Unable to parse chart records from registry: %+v", err)
		return nil, err
	}

	var items []string
	for _, record := range records {
		if !record.Deprecated {
			if chartName == "" {
				items = append(items, record.Name)
			} else {
				items = append(items, record.Version)
			}
		}
	}
	return items, err
}
