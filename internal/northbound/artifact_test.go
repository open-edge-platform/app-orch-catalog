// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *NorthBoundTestSuite) TestCreateArtifact() {
	created, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{
			Name:        "test-artifact",
			DisplayName: "Test artifact",
			Description: "This is a Test",
			MimeType:    "image/png",
			Artifact:    asBinary(kingfisherPngB64),
		},
	})
	s.validateResponse(err, created)
	s.validateArtifact(created.Artifact, "test-artifact", "Test artifact", "This is a Test", "image/png", asBinary(kingfisherPngB64))
	s.Less(s.startTime, created.Artifact.CreateTime.AsTime())

	resp, err := s.client.GetArtifact(s.ProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "test-artifact"})
	s.validateResponse(err, resp)
	s.validateArtifact(resp.Artifact, "test-artifact", "Test artifact", "This is a Test", "image/png", asBinary(kingfisherPngB64))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{
			Name:     "test-artifact",
			MimeType: "image/png",
			Artifact: asBinary(kingfisherPngB64),
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact test-artifact already exists`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactInvalidName() {
	// Create one with invalid name
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{Artifact: &catalogv3.Artifact{Name: "Bad name"}})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact Bad name invalid: invalid Artifact.Name: value does not match regex pattern "^[a-z0-9][a-z0-9-]{0,38}[a-z0-9]{0,1}$"`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactMimeType() {
	// Create one with invalid mime-type
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/pngbad"},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.MimeType: value does not match regex pattern "^(text/plain)$|^(application/json)$|^(application/yaml)$|^(image/png)$|^(image/jpeg)$"`))

	// Create one with invalid mime-type
	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image"},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.MimeType: value does not match regex pattern "^(text/plain)$|^(application/json)$|^(application/yaml)$|^(image/png)$|^(image/jpeg)$"`))

	// Create one with no mime-type
	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: ""},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.MimeType: value length must be between 1 and 40 runes, inclusive`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactDataPlain() {
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "text/plain", Artifact: []byte{0x89, 0x50, 0x4e, 0x47, 0x0D, 0x0A, 0x1a, 0x0a}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid Plain Text`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "text/plain", Artifact: []byte("abc")},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.Artifact: value length must be between 4 and 4000000 bytes, inclusive`))

}

func (s *NorthBoundTestSuite) TestCreateArtifactDataJson() {
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/json", Artifact: []byte{0x89, 0x50, 0x4e, 0x47, 0x0D, 0x0A, 0x1a, 0x0a}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid JSON`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/json", Artifact: []byte(`{"invalid":"json":"here"`)},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid JSON`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/json", Artifact: []byte("{}")},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.Artifact: value length must be between 4 and 4000000 bytes, inclusive`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactDataYaml() {
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/yaml", Artifact: []byte{0x89, 0x50, 0x4e, 0x47, 0x0D, 0x0A, 0x1a, 0x0a}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid YAML`))

	// try with some invalid json - json is OK in yaml
	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/yaml", Artifact: []byte(`{"some": "json": "invalid"}`)},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid YAML`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/yaml", Artifact: []byte(`	yaml can't start with tab'`)},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid YAML`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "application/yaml", Artifact: []byte("{}")},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.Artifact: value length must be between 4 and 4000000 bytes, inclusive`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactDataPng() {
	// Create one with invalid data when mime type is PNG
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/png", Artifact: []byte("any old thing")},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid PNG`))

	// see http://www.libpng.org/pub/png/spec/1.2/PNG-Structure.html for reference
	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/png", Artifact: []byte{0x89, 0x50, 0x4e, 0x47, 0x0D, 0x0A, 0x1a, 0x0a}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid PNG`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/png", Artifact: []byte{0x11, 0x11, 0x11}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.Artifact: value length must be between 4 and 4000000 bytes, inclusive`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactDataJpeg() {
	// Create one with invalid data when mime type is PNG
	_, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/jpeg", Artifact: []byte("any old thing")},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid JPEG`))

	// see http://www.libpng.org/pub/png/spec/1.2/PNG-Structure.html for reference
	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/jpeg", Artifact: []byte{0xFF, 0xd8, 0x4e, 0x47, 0x0D, 0x0A, 0x1a, 0x0a}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact data is not valid JPEG`))

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "test-artifact", MimeType: "image/jpeg", Artifact: []byte{0x11, 0x11, 0x11}},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: invalid Artifact.Artifact: value length must be between 4 and 4000000 bytes, inclusive`))
}

func (s *NorthBoundTestSuite) TestCreateArtifactDisplayName() {
	data := []byte("some image data")

	// Creating two artifacts with blank display name should work
	resp, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "art1", MimeType: "text/plain", Artifact: data},
	})
	s.validateResponse(err, resp)
	resp, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "art2", MimeType: "text/plain", Artifact: data},
	})
	s.validateResponse(err, resp)

	// Creating two artifacts with the same, non-blank display name should not work
	resp, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "art3", DisplayName: "Artifact", MimeType: "text/plain", Artifact: data},
	})
	s.validateResponse(err, resp)
	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: "art4", DisplayName: "artifact", MimeType: "text/plain", Artifact: data},
	})
	s.ErrorIs(err, status.Errorf(codes.AlreadyExists, "artifact art4 already exists: display name artifact is not unique"))
}

func (s *NorthBoundTestSuite) TestListArtifacts() {
	artifacts, err := s.client.ListArtifacts(s.ProjectID(footen), &catalogv3.ListArtifactsRequest{})
	s.validateResponse(err, artifacts)
	s.Equal(2, len(artifacts.Artifacts))
}

func (s *NorthBoundTestSuite) checkListArtifacts(artifacts *catalogv3.ListArtifactsResponse,
	err error, values string, count int32, onlyLength bool) {
	s.validateResponse(err, artifacts)
	if err != nil {
		return
	}
	s.Equal(count, artifacts.GetTotalElements())
	if values == "" {
		s.Len(artifacts.Artifacts, 0)
		return
	}
	expected := strings.Split(values, ",")
	s.Equal(len(expected), len(artifacts.Artifacts))
	if !onlyLength {
		for i, name := range expected {
			app := artifacts.Artifacts[i]
			s.Equal(name, app.Name)
		}
	}
}

func (s *NorthBoundTestSuite) generateArtifacts(count int) {
	format := ""
	if count < 10 {
		format = "a%d"
	} else {
		format = "a%02d"
	}
	for i := 1; i <= count; i++ {
		app := &catalogv3.Artifact{
			Name:        fmt.Sprintf(format, i),
			Description: "XXX",
			MimeType:    "text/plain",
			Artifact:    []byte("image data"),
		}
		resp, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{Artifact: app})
		s.NoError(err)
		s.NotNil(resp)
	}
}

func (s *NorthBoundTestSuite) TestListArtifactsWithOrderBy() {
	tests := map[string]struct {
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":             {orderBy: "", wantedList: "icon,thumb,a1,a2,a3"},
		"default":          {orderBy: "description", wantedList: "thumb,icon,a1,a2,a3"},
		"asc":              {orderBy: "name asc", wantedList: "a1,a2,a3,icon,thumb"},
		"desc":             {orderBy: "name desc", wantedList: "thumb,icon,a3,a2,a1"},
		"camel case field": {orderBy: "displayName asc", wantedList: "icon,thumb,a1,a2,a3"},
		"multi":            {orderBy: "description desc, name desc", wantedList: "a3,a2,a1,icon,thumb"},
		"too many":         {orderBy: "description asc desc", wantedList: "", expectedError: "invalid:"},
		"bad direction":    {orderBy: "description ascdesc", wantedList: "", expectedError: "invalid:"},
		"bad column":       {orderBy: "descriptionXXX", wantedList: "", expectedError: "invalid:"},
	}
	s.generateArtifacts(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			artifacts, err := s.client.ListArtifacts(s.ProjectID(footen), &catalogv3.ListArtifactsRequest{OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListArtifacts(artifacts, err, testCase.wantedList, 5, testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListArtifactsWithFilter() {
	tests := map[string]struct {
		filter        string
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"none":              {filter: "", wantedList: "a1,a2,a3,icon,thumb", orderBy: "name asc"},
		"single":            {filter: "name=a1", wantedList: "a1", orderBy: "name asc"},
		"camel case field":  {filter: "displayName=a1", wantedList: "a1", orderBy: "name asc"},
		"1 wildcard":        {filter: "name=*1", wantedList: "a1", orderBy: "name asc"},
		"2 wildcard":        {filter: "name=*a*", wantedList: "a1,a2,a3", orderBy: "name asc"},
		"match all no sort": {filter: "name=*", wantedList: "icon,thumb,a1,a2,a3"},
		"match all":         {filter: "name=*", wantedList: "a1,a2,a3,icon,thumb", orderBy: "name asc"},
		"or operation":      {filter: "name=*a* OR name=*o*", wantedList: "a1,a2,a3,icon", orderBy: "name asc"},
		"contains":          {filter: "name=b", wantedList: "thumb", orderBy: "name asc"},
		"bad column":        {filter: "bad=filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
		"bad filter":        {filter: "bad filter", wantedList: "", orderBy: "name asc", expectedError: "invalid"},
	}
	s.generateArtifacts(3)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			artifacts, err := s.client.ListArtifacts(s.ProjectID(footen), &catalogv3.ListArtifactsRequest{Filter: testCase.filter, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListArtifacts(artifacts, err, testCase.wantedList, int32(len(artifacts.Artifacts)), testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestListArtifactsWithPagination() {
	tests := map[string]struct {
		pageSize      int32
		offset        int32
		orderBy       string
		wantedList    string
		expectedError string
	}{
		"first ten":         {pageSize: 10, offset: 0, wantedList: "a01,a02,a03,a04,a05,a06,a07,a08,a09,a10", orderBy: "name asc"},
		"second ten":        {pageSize: 10, offset: 10, wantedList: "a11,a12,a13,a14,a15,a16,a17,a18,a19,a20", orderBy: "name asc"},
		"last five":         {pageSize: 5, offset: 29, wantedList: "a30,icon,thumb", orderBy: "name asc"},
		"0 page size":       {offset: 29, wantedList: "a30,icon,thumb", orderBy: "name asc"},
		"default page size": {wantedList: "a01,a02,a03,a04,a05,a06,a07,a08,a09,a10,a11,a12,a13,a14,a15,a16,a17,a18,a19,a20", orderBy: "name asc"},
		"page size too big": {pageSize: 1000, expectedError: "must not exceed"},
		"negative offset":   {pageSize: 5, offset: -29, expectedError: "negative"},
		"negative pageSize": {pageSize: -5, offset: 29, expectedError: "negative"},
		"bad offset":        {pageSize: 10, offset: 41},
	}
	s.generateArtifacts(30)

	for name, testCase := range tests {
		s.T().Run(name, func(_ *testing.T) {
			artifacts, err := s.client.ListArtifacts(s.ProjectID(footen),
				&catalogv3.ListArtifactsRequest{PageSize: testCase.pageSize, Offset: testCase.offset, OrderBy: testCase.orderBy})
			if testCase.expectedError != "" {
				s.Contains(err.Error(), testCase.expectedError)
			} else {
				s.checkListArtifacts(artifacts, err, testCase.wantedList, 32, testCase.orderBy == "")
			}
		})
	}
}

func (s *NorthBoundTestSuite) TestGetArtifact() {
	resp, err := s.client.GetArtifact(s.ProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "icon"})
	s.validateResponse(err, resp)
	s.validateArtifact(resp.Artifact, "icon", "Fancy Icon", "Icon of a bird", "image/png", asBinary(kingfisherPngB64))
	s.Less(s.startTime, resp.Artifact.CreateTime.AsTime())

	// Try one that does not exist
	_, err = s.client.GetArtifact(s.ProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "test-not-present"})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "artifact test-not-present not found"))
}

func (s *NorthBoundTestSuite) TestUpdateArtifact() {
	// Try updating the 2nd artifact
	_, err := s.client.UpdateArtifact(s.ProjectID(footen), &catalogv3.UpdateArtifactRequest{
		ArtifactName: "icon",
		Artifact: &catalogv3.Artifact{
			Name:        "icon",
			DisplayName: "new display name",
			MimeType:    "text/plain",
			Artifact:    []byte("some new image data"),
		},
	})
	s.NoError(err)

	resp, err := s.client.GetArtifact(s.ProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "icon"})
	s.validateResponse(err, resp)
	s.validateArtifact(resp.Artifact, "icon", "new display name", "", "text/plain", []byte("some new image data"))
	s.Less(resp.Artifact.CreateTime.AsTime(), resp.Artifact.UpdateTime.AsTime())

	// Try one that does not exist
	_, err = s.client.UpdateArtifact(s.ProjectID(footen), &catalogv3.UpdateArtifactRequest{
		ArtifactName: "test-not-present",
		Artifact: &catalogv3.Artifact{
			Name:     "test-not-present",
			MimeType: "text/plain",
			Artifact: []byte("some irrelevant image data"),
		},
	})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "artifact test-not-present not found"))

	// Try changing artifact name
	_, err = s.client.UpdateArtifact(s.ProjectID(footen), &catalogv3.UpdateArtifactRequest{
		ArtifactName: "icon",
		Artifact: &catalogv3.Artifact{
			Name:     "some-new-name",
			MimeType: "text/plain",
			Artifact: []byte("some irrelevant image data"),
		},
	})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument, "artifact invalid: name cannot be changed icon != some-new-name"))
}

func (s *NorthBoundTestSuite) TestDeleteArtifact() {
	// Even though this is being referenced by a deployment package it is not an error to delete it, because
	// the artifact is not mandatory in the deployment package
	deleted, err := s.client.DeleteArtifact(s.ProjectID(footen), &catalogv3.DeleteArtifactRequest{ArtifactName: "icon"})
	s.validateResponse(err, deleted)

	_, err = s.client.GetArtifact(s.ProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "icon"})
	s.ErrorIs(err, status.Errorf(codes.NotFound, "artifact icon not found"))

	// Try deleting it again, i.e. non-existent - should return NotFound
	deleted, err = s.client.DeleteArtifact(s.ProjectID(footen), &catalogv3.DeleteArtifactRequest{ArtifactName: "icon"})
	s.validateNotFound(err, deleted)
}

func (s *NorthBoundTestSuite) TestDeleteDeployedArtifact() {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(footen), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
	})
	s.validateResponse(err, resp)

	resp.DeploymentPackage.Artifacts = []*catalogv3.ArtifactReference{{
		Name:    "icon",
		Purpose: "icon",
	}}
	resp.DeploymentPackage.IsDeployed = true

	updated, err := s.client.UpdateDeploymentPackage(s.ProjectID(footen), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: "ca-gigi", Version: "v0.2.1",
		DeploymentPackage: resp.DeploymentPackage,
	})
	s.validateResponse(err, updated)
	_, err = s.client.DeleteArtifact(s.ProjectID(footen), &catalogv3.DeleteArtifactRequest{ArtifactName: "icon"})
	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact icon invalid: artifact icon is in use and cannot be deleted`))

}

func (s *NorthBoundTestSuite) TestCreateArtifactMissingMalwareScanner() {
	created, err := s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{
			Name:        "test-artifact",
			DisplayName: "Test artifact",
			Description: "This is a Test",
			MimeType:    "image/png",
			Artifact:    asBinary(kingfisherPngB64),
		},
	})
	s.validateResponse(err, created)
	s.validateArtifact(created.Artifact, "test-artifact", "Test artifact", "This is a Test", "image/png", asBinary(kingfisherPngB64))
	s.Less(s.startTime, created.Artifact.CreateTime.AsTime())

	resp, err := s.client.GetArtifact(s.ProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "test-artifact"})
	s.validateResponse(err, resp)
	s.validateArtifact(resp.Artifact, "test-artifact", "Test artifact", "This is a Test", "image/png", asBinary(kingfisherPngB64))

	malware.DefaultScanner = malware.NewScanner("does.not.exist.anywhere", malware.DefaultScannerTimeout, false)

	_, err = s.client.CreateArtifact(s.ProjectID(footen), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{
			Name:     "test-artifact",
			MimeType: "image/png",
			Artifact: asBinary(kingfisherPngB64),
		},
	})
	s.ErrorIs(err, status.Errorf(codes.Unavailable,
		`artifact unavailable: malware scanner configured but not available`))

	malware.DefaultScanner = nil
}

func (s *NorthBoundTestSuite) TestArtifactEvents() {
	ctx, cancel := context.WithCancel(s.ProjectID(barten))
	stream, err := s.client.WatchArtifacts(ctx, &catalogv3.WatchArtifactsRequest{NoReplay: true})
	s.NoError(err)
	time.Sleep(100 * time.Millisecond) // Give the subscription a chance to take place

	art := s.createArtifact(barten, "newart", "Artifact", "Icon", "text/plain", []byte("content"))

	resp, err := stream.Recv()
	s.NoError(err)
	s.Equal(CreatedEvent, EventType(resp.Event.Type))
	s.validateArtifact(resp.Artifact, art.Name, art.DisplayName, art.Description, art.MimeType, art.Artifact)

	art.DisplayName = "New Artifact"
	_, err = s.client.UpdateArtifact(s.ProjectID(barten), &catalogv3.UpdateArtifactRequest{
		ArtifactName: "newart", Artifact: art,
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(UpdatedEvent, EventType(resp.Event.Type))
	s.validateArtifact(resp.Artifact, art.Name, art.DisplayName, art.Description, art.MimeType, art.Artifact)

	_, err = s.client.DeleteArtifact(s.ProjectID(barten), &catalogv3.DeleteArtifactRequest{
		ArtifactName: "newart",
	})
	s.NoError(err)

	resp, err = stream.Recv()
	s.NoError(err)
	s.Equal(DeletedEvent, EventType(resp.Event.Type))
	s.validateArtifact(resp.Artifact, "newart", "", "", "", nil)

	// Make sure we get an error back for a Recv() on a closed channel
	cancel()
	s.createArtifact(barten, "some-new-artifact", "Some New Artifact", "Icon", "text/plain", []byte("content"))
	resp, err = stream.Recv()
	s.Error(err)
	s.Nil(resp)
}

func (s *NorthBoundTestSuite) TestArtifactEventsWithReplay() {
	s.T().Skip()
	ctx, cancel := context.WithCancel(s.ProjectID(barten))
	stream, err := s.client.WatchArtifacts(ctx, &catalogv3.WatchArtifactsRequest{})
	s.NoError(err)

	existing := "icon thumb"
	for i := 0; i < 2; i++ {
		resp, err := stream.Recv()
		s.NoError(err)
		s.Equal(ReplayedEvent, EventType(resp.Event.Type))
		s.True(strings.Contains(existing, resp.Artifact.Name), "unexpected: %s", resp.Artifact.Name)
	}

	art := s.createArtifact(barten, "newart", "Artifact", "Icon", "text/plain", []byte("content"))
	resp, err := stream.Recv()
	s.NoError(err)
	s.Equal(CreatedEvent, EventType(resp.Event.Type))
	s.validateArtifact(resp.Artifact, art.Name, art.DisplayName, art.Description, art.MimeType, art.Artifact)

	// Make sure we get an error back for a Recv() on a closed channel
	cancel()
	s.createArtifact(barten, "some-new-artifact", "Some New Artifact", "Icon", "text/plain", []byte("content"))
	resp, err = stream.Recv()
	s.Error(err)
	s.Nil(resp)
}

func (s *NorthBoundDBErrTestSuite) TestArtifactWatchInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.WatchArtifactsRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			err := s.server.WatchArtifacts(testCase.req, nil)
			s.validateInvalidArgumentError(err, nil)
		})
	}
}

type catalogServiceWatchArtifactServer struct {
	testServerStream
	sendError bool
}

func (c *catalogServiceWatchArtifactServer) Send(*catalogv3.WatchArtifactsResponse) error {
	if c.sendError {
		return errors.New("no send")
	}
	return nil
}

func (s *NorthBoundDBErrTestSuite) TestWatchArtifactDatabaseErrors() {
	s.T().Skip("FIXME: Hard to troubleshoot")
	saveBase64Factory := Base64Factory
	defer func() { Base64Factory = saveBase64Factory }()
	Base64Factory = newBase64Noop

	server := &catalogServiceWatchArtifactServer{sendError: false}
	req := &catalogv3.WatchArtifactsRequest{}

	// Test unable to start a transaction
	err := s.server.WatchArtifacts(req, server)
	s.validateDBError(err, nil)

	// Test publisher existence query failure
	s.mock.ExpectBegin()
	err = s.server.WatchArtifacts(req, server)
	s.validateDBError(err, nil)

	// transaction commit error
	s.mock.ExpectBegin()
	s.addMockedEmptyQueryRows(2)
	err = s.server.WatchArtifacts(req, server)
	s.validateDBError(err, nil)

	// send failure in replay
	server.sendError = true
	s.mock.ExpectBegin()
	s.addMockedEmptyQueryRows(2)
	err = s.server.WatchArtifacts(req, server)
	s.Contains(err.Error(), "no send")

	// send failure for event
	ch := make(chan *catalogv3.WatchArtifactsResponse)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.server.watchArtifactEvents(server, ch)
		wg.Done()
	}()
	ch <- nil
	wg.Wait()
	s.Contains(err.Error(), "no send")
	close(ch)

	// end of channel
	ch = make(chan *catalogv3.WatchArtifactsResponse)
	wg = sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = s.server.watchArtifactEvents(server, ch)
		wg.Done()
	}()
	close(ch)
	wg.Wait()
	s.NoError(err)

	s.NoError(s.mock.ExpectationsWereMet())
}

// FuzzCreateArtifact - fuzz test creating a artifact
//
// In this case we are calling the Test Suite to create a Publisher through gRPC
// but calling the function-under-test directly
//
// Invoke with:
//
//	go test ./internal/northbound -fuzz FuzzCreateArtifact -fuzztime=60s
func FuzzCreateArtifact(f *testing.F) {
	f.Add("test-artifact", "Test Artifact")
	f.Add("test-artifact", " space at start")
	f.Add("test-artifact", "space at end ")
	f.Add("-", "starts with hyphen")
	f.Add("a", "Single letter OK")
	f.Add("a.", "contains .")
	f.Add("aaaaa-bbbb-cccc-dddd-eeee-ffff-gggg-hhhhh", "name too long > 40")
	f.Add("test-artifact", "display name is too long at 40 chars - here")
	f.Add("test-artifact", `display name contains
new line`)

	s := &NorthBoundTestSuite{}
	s.SetupSuite()
	defer s.TearDownSuite()

	f.Fuzz(func(t *testing.T, name string, displayName string) {
		s.SetT(t)
		s.SetupTest() // SetupTest cannot be called until here because it depends on T's Assertions
		defer s.TearDownTest()

		server := Server{
			UnimplementedCatalogServiceServer: catalogv3.UnimplementedCatalogServiceServer{},
			databaseClient:                    s.dbClient,
			listeners:                         NewEventListeners(),
		}

		// We call the function directly - not through gRPC
		created, err := server.CreateArtifact(s.ServerProjectID(footen), &catalogv3.CreateArtifactRequest{
			Artifact: &catalogv3.Artifact{
				Name:        name,
				DisplayName: displayName,
				Description: strings.Repeat(displayName, 20),
				MimeType:    "text/plain",
				Artifact:    []byte(displayName),
			},
		})

		nameSpacesPatternRE, _ := regexp.Compile(`invalid: invalid Artifact.Name: value does not match regex pattern .*`)
		nameLenRE, _ := regexp.Compile(` invalid: invalid Artifact.Name: value length must be between 1 and 40 runes, inclusive`)
		displayNameSpacesRE, _ := regexp.Compile(`rpc error: code = InvalidArgument desc = artifact invalid: display name cannot contain leading or trailing spaces`)
		displayNameLenRE, _ := regexp.Compile(`rpc error: code = InvalidArgument desc = artifact .* invalid: invalid Artifact.DisplayName: value length must be between 0 and 40 runes, inclusive`)
		displayNamePatternRE, _ := regexp.Compile(`rpc error: code = InvalidArgument desc = artifact .* invalid: invalid Artifact.DisplayName: value does not match regex pattern .*`)
		valueLenRE, _ := regexp.Compile(`rpc error: code = InvalidArgument desc = artifact [a-z0-9\-]+ invalid: invalid Artifact.Artifact: value length must be between 4 and 4000000 bytes, inclusive`)
		invalidCharRE, _ := regexp.Compile(`rpc error: code = InvalidArgument desc = artifact [a-z0-9\-]+ invalid: artifact data is not valid Plain Text.*`)

		if err != nil || created == nil {
			if !nameSpacesPatternRE.Match([]byte(err.Error())) &&
				!nameLenRE.Match([]byte(err.Error())) &&
				!displayNameSpacesRE.Match([]byte(err.Error())) &&
				!displayNameLenRE.Match([]byte(err.Error())) &&
				!displayNamePatternRE.Match([]byte(err.Error())) &&
				!valueLenRE.Match([]byte(err.Error())) &&
				!invalidCharRE.Match([]byte(err.Error())) {
				t.Errorf("%v Name: %v DisplayName: %v", err.Error(), name, displayName)
			}
		} else {
			t.Logf("created %s", created.Artifact.Name)
		}

	})
}

func (s *NorthBoundTestSuite) TestArtifactAuthErrors() {
	var err error
	server := s.newMockOPAServer()

	data := "12345"
	artifact := &catalogv3.Artifact{
		Name:     "art",
		MimeType: "text/plain",
		Artifact: []byte(data),
	}

	_, err = server.CreateArtifact(s.ServerProjectID(footen), &catalogv3.CreateArtifactRequest{Artifact: artifact})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.UpdateArtifact(s.ServerProjectID(footen), &catalogv3.UpdateArtifactRequest{ArtifactName: "art", Artifact: artifact})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.DeleteArtifact(s.ServerProjectID(footen),
		&catalogv3.DeleteArtifactRequest{ArtifactName: "art"})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.GetArtifact(s.ServerProjectID(footen), &catalogv3.GetArtifactRequest{ArtifactName: "art"})
	s.ErrorIs(err, expectedAuthError)

	_, err = server.ListArtifacts(s.ServerProjectID(footen), &catalogv3.ListArtifactsRequest{})
	s.ErrorIs(err, expectedAuthError)
}

func (s *NorthBoundDBErrTestSuite) TestArtifactCreateInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.CreateArtifactRequest
	}{
		"nil request":    {req: nil},
		"empty artifact": {req: &catalogv3.CreateArtifactRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.CreateArtifact(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestArtifactListInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.ListArtifactsRequest
	}{
		"nil request": {req: nil},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.ListArtifacts(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestArtifactGetInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.GetArtifactRequest
	}{
		"nil request": {req: nil},
		"empty name":  {req: &catalogv3.GetArtifactRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.GetArtifact(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestArtifactUpdateInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.UpdateArtifactRequest
	}{
		"nil request":    {req: nil},
		"empty name":     {req: &catalogv3.UpdateArtifactRequest{}},
		"empty artifact": {req: &catalogv3.UpdateArtifactRequest{ArtifactName: "art", Artifact: &catalogv3.Artifact{}}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.UpdateArtifact(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundDBErrTestSuite) TestArtifactDeleteInvalidArgs() {
	tests := map[string]struct {
		req *catalogv3.DeleteArtifactRequest
	}{
		"nil request": {req: nil},
		"empty name":  {req: &catalogv3.DeleteArtifactRequest{}},
	}

	for name, testCase := range tests {
		s.Run(name, func() {
			resp, err := s.server.DeleteArtifact(s.ctx, testCase.req)
			s.validateInvalidArgumentError(err, resp)
		})
	}
}

func (s *NorthBoundTestSuite) TestCreateArtifactBadDataType() {
	err := validateArtifactData("test-artifact", "unknown/type",
		[]byte{0x89, 0x50, 0x4e, 0x47, 0x0D, 0x0A, 0x1a, 0x0a})

	s.ErrorIs(err, status.Errorf(codes.InvalidArgument,
		`artifact test-artifact invalid: artifact contents do not match mime type unknown/type`))
}

func TestEmptyArtifactsDBQuery(t *testing.T) {
	s := &NorthBoundTestSuite{populateDB: false}
	s.SetT(t)
	s.SetupTest()

	artifacts, err := s.client.ListArtifacts(s.ProjectID(footen), &catalogv3.ListArtifactsRequest{})
	s.validateResponse(err, artifacts)
	s.Equal(0, len(artifacts.Artifacts))

	publisher, err := s.client.GetArtifact(s.ProjectID("nobody"), &catalogv3.GetArtifactRequest{ArtifactName: "none"})
	s.Nil(publisher)
	s.Error(err)
	s.Contains(err.Error(), "not found")
}
