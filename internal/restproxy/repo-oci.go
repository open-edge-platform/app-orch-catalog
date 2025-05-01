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
	"strings"
)

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

	if registry.Username != "" {
		// If we have a username, inject it together with the password into the request as a basic authorization header
		req.SetBasicAuth(registry.Username, registry.AuthToken)
	} else if registry.AuthToken != "" {
		// Otherwise, if the registry specified just an auth token, inject it into the request as a bearer authorization header
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", registry.AuthToken))
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

func fetchOCIChartsList(c *gin.Context, registry *catalogv3.Registry, chartName string) {
	inventoryURL := strings.Replace(registry.InventoryUrl, "oci://", "https://", 1)
	if strings.HasSuffix(inventoryURL, "/") {
		inventoryURL = fmt.Sprintf("%srepositories", inventoryURL)
	} else {
		inventoryURL = fmt.Sprintf("%s/repositories", inventoryURL)
	}

	// If chart parameter is specified, modify the URL obtain the versions
	if chartName != "" {
		inventoryURL = fmt.Sprintf("%s/%s/artifacts", inventoryURL, chartName)
	}

	resBody, ok := callHarborAPI(c, registry, inventoryURL)
	if !ok {
		return
	}

	items, err := parseOCIRecords(resBody, chartName)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	writeResponse(c, items)
}

func parseOCIRecords(response []byte, chartName string) ([]string, error) {
	if chartName == "" {
		return parseChartNames(response)
	}
	return parseChartVersions(response)
}

type OCIChartNameRecord struct {
	Name string `json:"name"`
}

func parseChartNames(response []byte) ([]string, error) {
	var records []OCIChartNameRecord
	var err error
	if err = json.Unmarshal(response, &records); err != nil {
		log.Errorf("Unable to parse chart name records from registry: %+v", err)
		return nil, err
	}

	var items []string
	for _, record := range records {
		names := strings.SplitN(record.Name, "/", 2)
		if len(names) > 1 {
			items = append(items, names[1])
		} else {
			items = append(items, names[0])
		}
	}
	return items, nil
}

type OCIChartVersionRecord struct {
	ExtraAttrs struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"extra_attrs"`
}

func parseChartVersions(response []byte) ([]string, error) {
	var records []OCIChartVersionRecord
	var err error
	if err = json.Unmarshal(response, &records); err != nil {
		log.Errorf("Unable to parse chart version records from registry: %+v", err)
		return nil, err
	}

	var items []string
	for _, record := range records {
		fmt.Printf("%+v\n", record)
		items = append(items, record.ExtraAttrs.Version)
	}
	return items, nil
}
