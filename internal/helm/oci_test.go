// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

type MockOrasClient struct {
}

func (oc *MockOrasClient) NewRegistry(host string) error {
	_ = host
	return nil
}

func (oc *MockOrasClient) Repository(ctx context.Context, name string) error {
	_ = ctx
	_ = name
	return nil
}

func (oc *MockOrasClient) SetUsernamePassword(username string, password string) {
	_ = username
	_ = password
}

func (oc *MockOrasClient) SetAccessToken(password string) {
	_ = password
}

func (oc *MockOrasClient) GetTags(ctx context.Context) ([]string, error) {
	_ = ctx
	return []string{"1.0.0", "1.0.1", "1.0.2"}, nil
}

func (oc *MockOrasClient) GetTarball(ctx context.Context, tagName string) (io.Reader, error) {
	_ = ctx
	_ = tagName
	file, err := os.Open("testdata/chart.tgz")
	if err != nil {
		return nil, err
	}
	return file, nil
}

func TestFetchHelmChartOCI(t *testing.T) {
	orasClient = &MockOrasClient{}

	h, err := FetchHelmChartOCI("oci://foo/bar", "testuser", "testpassword")
	assert.NoError(t, err)

	_ = h
}
