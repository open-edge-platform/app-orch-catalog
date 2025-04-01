// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"bytes"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func (s *ProxyTestSuite) TestUploadNoBody() {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:6942/catalog.orchestrator.apis/upload", http.NoBody)
	s.NoError(err)

	resp, err := s.httpClient.Do(req)
	s.NoError(err)
	if s.NotNil(resp) {
		s.Equal(500, resp.StatusCode)
	}
}

func (s *ProxyTestSuite) TestUploadFile() {
	resp, err := uploadMultipartFile(&s.httpClient, "http://localhost:6942/catalog.orchestrator.apis/upload",
		[]string{
			"../northbound/testdata/artifact.yaml",
		})
	s.NoError(err)
	if s.NotNil(resp) {
		_, err = io.ReadAll(resp.Body)
		s.NoError(err)
		s.Equal(200, resp.StatusCode)
	}
}

func (s *ProxyTestSuite) TestUploadBadYAMLFiles() {
	resp, _ := uploadMultipartFile(&s.httpClient, "http://localhost:6942/catalog.orchestrator.apis/upload",
		[]string{
			"../northbound/testdata/badyaml/registry-intel.yaml",
		})
	if s.NotNil(resp) {
		s.Equal(400, resp.StatusCode)
	}
}

func uploadMultipartFile(client *http.Client, url string, files []string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for _, path := range files {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		part, _ := w.CreateFormFile("files", path)
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}
		_ = file.Close()
	}
	_ = w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set(ActiveProjectID, "project")
	return client.Do(req)
}
