// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"net/http"
	"strings"
)

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
