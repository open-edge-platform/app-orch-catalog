// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"context"
	"encoding/json"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasClientInterface interface {
	NewRegistry(host string) error
	Repository(ctx context.Context, name string) error
	SetUsernamePassword(username string, password string)
	SetAccessToken(password string)
	GetTags(ctx context.Context) ([]string, error)
	GetTarball(ctx context.Context, tagName string) (io.Reader, error)
}

// Abstract out all Oras client stuff, for easy mocking

type OrasClient struct {
	reg        *remote.Registry
	src        registry.Repository //*remote.Repository
	remoteHost string
}

func (oc *OrasClient) NewRegistry(host string) error {
	var err error
	oc.reg, err = remote.NewRegistry(host)
	oc.remoteHost = host
	return err
}

func (oc *OrasClient) Repository(ctx context.Context, name string) error {
	var err error
	oc.src, err = oc.reg.Repository(ctx, name)
	return err
}

func (oc *OrasClient) SetUsernamePassword(username string, password string) {
	oc.reg.Client = &auth.Client{
		Header:     auth.DefaultClient.Header,
		Credential: auth.StaticCredential(oc.remoteHost, auth.Credential{Username: username, Password: password}),
	}
}

func (oc *OrasClient) SetAccessToken(password string) {
	oc.reg.Client = &auth.Client{
		Header:     auth.DefaultClient.Header,
		Credential: auth.StaticCredential(oc.remoteHost, auth.Credential{AccessToken: password}),
	}
}

func (oc *OrasClient) GetTags(ctx context.Context) ([]string, error) {
	allTags := []string{}
	err := oc.src.Tags(ctx, "", func(tags []string) error {
		allTags = append(allTags, tags...)
		return nil
	})
	return allTags, err
}

func (oc *OrasClient) GetTarball(ctx context.Context, tagName string) (io.Reader, error) {
	/* We've fetched the Helm Chart from oras, now we need to go through a series of
	 * steps to process the oras artifact, extract the tarball that contains the helm chart,
	 * extract the Chart.yaml file from the tarball, and finally to parsed the chart.yaml
	 * file.
	 */

	ms := memory.New()
	desc, err := oras.Copy(ctx, oc.src, tagName, ms, tagName, oras.DefaultCopyOptions)
	if err != nil {
		return nil, &ExtractError{Msg: "Failed to copy content from oras to memory", Err: err}
	}

	manifestReader, err := ms.Fetch(ctx, desc)
	if err != nil {
		return nil, &ExtractError{Msg: "Failed to fetch content from memory store", Err: err}
	}

	/* The manifest contains the list of layers */

	manifestBytes, err := io.ReadAll(manifestReader)
	if err != nil {
		return nil, &ExtractError{Msg: "Failed to read manifest", Err: err}
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, &ExtractError{Msg: "Failed to unmarshal manifest", Err: err}
	}

	/* The first layer will have the helm chart tarball */

	contentReader, err := ms.Fetch(ctx, manifest.Layers[0])
	if err != nil {
		return nil, &ExtractError{Msg: "Failed to fetch tarball content", Err: err}
	}

	return contentReader, err
}
