// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	"gopkg.in/yaml.v2"
)

const (
	MaxExtractedFileSize = 10 * 1024 * 1024 // to limit the size of extracted files and mitigate decompression bomb lint message
	MaxTarSize           = 10 * 1024 * 1024 // to limit the size of the intermediate gzip extraction
)

var orasClient OrasClientInterface = &OrasClient{} // for mocking

/* HelmInfo contains information about the Helm Chart. */

type HelmInfo struct { // nolint:revive
	Name        string  /* Name of the Helm Chart */
	Version     string  /* Version of the Helm Chart */
	Description string  /* Description of the Helm Chart, extracted from Chart.yaml */
	OCIRegistry string  /* OCI Registry URL */
	Username    string  /* Username used to fetch chart */
	Password    string  /* Password used to fetch chart */
	Values      *[]byte /* Default values.yaml content; nil = no values.yaml file present */
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

func extractFilesFromTGZ(reader io.Reader, targetFileNames []string) (map[string][]byte, error) {
	// Create a gzip reader
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, &ExtractError{Msg: "Failed to create gzip reader while extracting", Err: err}
	}
	defer gzipReader.Close()

	limReader := io.LimitReader(gzipReader, MaxTarSize)

	// Create a tar reader
	tarReader := tar.NewReader(limReader)

	fullNames := make(map[string]string) // to track the longest full path for each target file
	extractedFiles := make(map[string][]byte)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, &ExtractError{Msg: "Failed to read tar header while extracting", Err: err}
		}

		// Check if the current file is the one we want to extract
		if header.Typeflag != tar.TypeReg {
			continue
		}

		for _, targetFileName := range targetFileNames {
			if filepath.Base(header.Name) == targetFileName {
				// Make sure we get the match with the shortest name.
				// For example, in the jupyterhub helm chart, we have:
				// - jupyterhub/Chart.yaml   								<--- we want this one
				// - jupyterhub/charts/common/Chart.yaml
				// - jupyterhub/charts/postgresql/charts/common/Chart.yaml
				fullName, ok := fullNames[targetFileName]
				if ok && (len(header.Name) > len(fullName)) {
					continue
				}
				fullNames[targetFileName] = header.Name

				limFileReader := io.LimitReader(tarReader, MaxExtractedFileSize)
				destData, err := io.ReadAll(limFileReader)
				if err != nil {
					return nil, &ExtractError{Msg: "Failed to read file contents while extracting", Filename: targetFileName, Err: err}
				}
				extractedFiles[targetFileName] = destData
			}
		}
	}

	return extractedFiles, nil //, &ExtractError{Msg: "Failed to find file while extracting", Filename: targetFileName}
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

	err = orasClient.VerifyExists(ctx, tagName)
	if err != nil {
		return HelmInfo{}, &FetchError{Msg: "Failed to verify tag exists in repository", Err: err, URL: ociurl, Host: remoteHost, Artifact: artifactName, Tag: tagName}
	}

	verboseerror.Infof("Fetching helm chart from oci://%s/%s:%s\n", remoteHost, artifactName, tagName)

	// TODO: Unless latest is used, The URL isn't actually fetched until inside GetTarball(), so we are returning ExtractError when
	// we really should be returning FetchError. See if we can do something to catch this earlier.

	contentReader, err := orasClient.GetTarball(ctx, tagName)
	if err != nil {
		return HelmInfo{}, err
	}

	/* From the tarball, we can finally extract the Chart.yaml file */

	extractedFiles, err := extractFilesFromTGZ(contentReader, []string{"Chart.yaml", "values.yaml"})
	if err != nil {
		return HelmInfo{}, err
	}

	chart, ok := extractedFiles["Chart.yaml"]
	if !ok {
		return HelmInfo{}, &ExtractError{Msg: "Failed to find file while extracting", Filename: "Chart.yaml"}
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

	values, ok := extractedFiles["values.yaml"]
	if ok {
		hi.Values = &values
	}

	return hi, nil
}
