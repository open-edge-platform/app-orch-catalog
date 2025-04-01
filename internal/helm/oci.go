// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"github.com/blang/semver/v4"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	"gopkg.in/yaml.v2"
	"io"
	"net/url"
	"sort"
	"strings"
)

const (
	MaxExtractedFileSize = 10 * 1024 * 1024 // to limit the size of extracted files and mitigate decompression bomb lint message
)

var orasClient OrasClientInterface = &OrasClient{} // for mocking

/* HelmInfo contains information about the Helm Chart. */

type HelmInfo struct { // nolint:revive
	Name        string /* Name of the Helm Chart */
	Version     string /* Version of the Helm Chart */
	Description string /* Description of the Helm Chart, extracted from Chart.yaml */
	OCIRegistry string /* OCI Registry URL */
	Username    string /* Username used to fetch chart */
	Password    string /* Password used to fetch chart */
}

func parseOrasURL(ociurl string) (string, string, string, string, error) {
	var tag string
	parsedURL, err := url.Parse(ociurl)
	if err != nil {
		return "", "", "", "", &ParseError{URL: ociurl, Msg: "Failed to parse URL", Err: err}
	}
	if parsedURL.Scheme != "oci" {
		return "", "", "", "", &ParseError{URL: ociurl, Msg: "Scheme is not oci:// in URL"}
	}
	if parsedURL.Host == "" {
		return "", "", "", "", &ParseError{URL: ociurl, Msg: "Missing host in URL"}
	}
	if parsedURL.Path == "" {
		return "", "", "", "", &ParseError{URL: ociurl, Msg: "Missing path in URL"}
	}

	path := parsedURL.Path
	if idx := strings.LastIndex(path, ":"); idx != -1 {
		tag = path[idx+1:]
		path = path[:idx]
	} else {
		tag = "latest"
	}

	if path[0] == '/' {
		path = path[1:]
	}

	fullName := path
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		path = strings.Join(parts[:len(parts)-1], "/")
	}

	return parsedURL.Host, path, fullName, tag, nil
}

func extractFileFromTGZ(reader io.Reader, targetFileName string) ([]byte, error) {
	// Create a gzip reader
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, &ExtractError{Msg: "Failed to create gzip reader while extracting", Filename: targetFileName, Err: err}
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, &ExtractError{Msg: "Failed to read tar header while extracting", Filename: targetFileName, Err: err}
		}

		// Check if the current file is the one we want to extract
		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, targetFileName) {
			// Read the file content into a buffer
			var buf strings.Builder
			if _, err := io.CopyN(&buf, tarReader, MaxExtractedFileSize); err != nil && err != io.EOF {
				return nil, &ExtractError{Msg: "Failed to copy file content while extracting", Filename: targetFileName, Err: err}
			}

			verboseerror.Infof("Extracted file: %s\n", targetFileName)
			return []byte(buf.String()), nil
		}
	}

	return nil, &ExtractError{Msg: "Failed to find file while extracting", Filename: targetFileName}
}

// FetchHelmChartOCI fetches a Helm Chart from an OCI registry and extracts some useful info

func FetchHelmChartOCI(ociurl string, user string, password string) (HelmInfo, error) {
	remoteHost, path, artifactName, tagName, err := parseOrasURL(ociurl)
	if err != nil {
		return HelmInfo{}, err
	}

	err = orasClient.NewRegistry(remoteHost)
	if err != nil {
		return HelmInfo{}, &FetchError{Msg: "Failed to create registry object", Err: err, URL: ociurl, Host: remoteHost}
	}

	if user != "" && password != "" {
		verboseerror.Infof("Using username/password authentication\n")
		orasClient.SetUsernamePassword(user, password)
	} else if password != "" {
		verboseerror.Infof("Using token authentication\n")
		orasClient.SetAccessToken(password)
	}

	ctx := context.Background()
	err = orasClient.Repository(ctx, artifactName)
	if err != nil {
		return HelmInfo{}, &FetchError{Msg: "Failed to get repository using oras", Err: err, URL: ociurl, Host: remoteHost, Artifact: artifactName}
	}

	if tagName == "latest" {
		allTags, err := orasClient.GetTags(ctx)
		if err != nil {
			return HelmInfo{}, &FetchError{Msg: "Failed to get tags using oras", Err: err, URL: ociurl, Host: remoteHost, Artifact: artifactName}
		}
		validTags := []string{}
		for _, t := range allTags {
			if _, err := semver.Parse(t); err == nil {
				validTags = append(validTags, t)
			}
		}
		sort.Slice(validTags, func(i, j int) bool {
			vi, _ := semver.Parse(validTags[i])
			vj, _ := semver.Parse(validTags[j])
			return vi.LT(vj)
		})
		tagName = validTags[len(validTags)-1]
		verboseerror.Infof("Found latest tag %s\n", tagName)
	}

	verboseerror.Infof("Fetching helm chart from oci://%s/%s:%s\n", remoteHost, artifactName, tagName)

	contentReader, err := orasClient.GetTarball(ctx, tagName)
	if err != nil {
		return HelmInfo{}, err
	}

	/* From the tarball, we can finally extract the Chart.yaml file */

	chart, err := extractFileFromTGZ(contentReader, "Chart.yaml")
	if err != nil {
		return HelmInfo{}, err
	}

	var chartData map[string]interface{}

	err = yaml.Unmarshal(chart, &chartData)
	if err != nil {
		return HelmInfo{}, &ExtractError{Msg: "Failed to parse the chart yaml", Err: err}
	}

	hi := HelmInfo{
		Name:        chartData["name"].(string),
		Version:     chartData["version"].(string),
		Description: chartData["description"].(string),
		OCIRegistry: strings.Join([]string{"oci:/", remoteHost, path}, "/"),
	}

	if user != "" {
		hi.Username = user
	}
	if password != "" {
		hi.Password = password
	}

	return hi, nil
}
