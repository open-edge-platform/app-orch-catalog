// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"time"
	// Third-party imports
)

const (
	catalogSchemaTool      = "../../build/_output/catalog-schema"
	testdataDir            = "../testdata"
	wordpressDir           = testdataDir + "/wordpress"
	wordpressMissingAppDir = testdataDir + "/wordpress-missing-app-name"
	complexDir             = testdataDir + "/complex"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func (s *TestSuite) runCatalogSchema(arg string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, catalogSchemaTool, "validate", arg)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	outStr := stdout.String()
	errStr := stderr.String()
	return outStr, errStr, err
}

func (s *TestSuite) TestSchemaGoodPackage() {
	stdout, stderr, err := s.runCatalogSchema(wordpressDir)
	s.NoError(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stderr, "Expected no error output when running catalog-schema on a good package")
	s.Equal("", stdout, "Expected no output when running catalog-schema on a good package")
}

func (s *TestSuite) TestSchemaBadPackageMissingAppName() {
	stdout, stderr, err := s.runCatalogSchema(wordpressMissingAppDir)
	s.Error(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stderr, "Expected no error output when running catalog-schema on a good package")
	s.Contains(stdout, "does not validate", "Expected stdout to contain 'does not validate' when running catalog-schema on a bad package")
}

func (s *TestSuite) TestSchemaComplexPackage() {
	stdout, stderr, err := s.runCatalogSchema(wordpressDir)
	s.NoError(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stderr, "Expected no error output when running catalog-schema on a good package")
	s.Equal("", stdout, "Expected no output when running catalog-schema on a good package")
}
