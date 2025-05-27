// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports

	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
	// Third-party imports
)

const (
	doToHelmTool = "../../build/_output/dp-to-helm"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func (s *TestSuite) runDpToHelm(args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Executing command: %s %v", doToHelmTool, args)

	c := exec.CommandContext(ctx, doToHelmTool, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()

	outStr := stdout.String()
	errStr := stderr.String()
	return outStr, errStr, err
}

func (s *TestSuite) TestDpToHelmGoodPackage() {
	tempDir, err := os.MkdirTemp("", "catalog_test_*")
	s.Require().NoError(err)
	defer os.RemoveAll(tempDir)

	stdout, stderr, err := s.runDpToHelm(wordpressDir, "-o", tempDir)
	s.NoError(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stderr, "Expected no error output when running catalog-schema on a good package")

	_, err = os.Stat(fmt.Sprintf("%s/test-wordpress-testing.yaml", tempDir))
	s.NoError(err, "Expected test-wordpress-testing.yaml to exist")

	expected := fmt.Sprintf("helm install test-wordpress https://charts.bitnami.com/bitnami/wordpress --version 19.4.3 --namespace default -f %s/test-wordpress-testing.yaml", tempDir)
	s.Contains(stdout, expected, "Expected stdout to contain helm install command")
}

func (s *TestSuite) TestDpToHelmBadPackageMissingAppName() {
	stdout, stderr, err := s.runDpToHelm(wordpressMissingAppDir)
	s.Error(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stdout, "Expected no stdout output when running catalog-schema on a good package")
	s.Contains(stderr, "InvalidArgument", "Expected stdout to contain 'InvalidArgument' when running catalog-schema on a bad package")
}

func (s *TestSuite) TestDpToHelmComplexPackage() {
	tempDir, err := os.MkdirTemp("", "catalog_test_*")
	s.Require().NoError(err)
	defer os.RemoveAll(tempDir)

	stdout, stderr, err := s.runDpToHelm(complexDir, "-o", tempDir, "--set", "password=1234")
	s.NoError(err, "Expected no error when running catalog-schema on a good package")
	s.Equal("", stderr, "Expected no error output when running catalog-schema on a good package")

	_, err = os.Stat(fmt.Sprintf("%s/one-default.yaml", tempDir))
	s.NoError(err, "Expected one-default.yaml to exist")

	_, err = os.Stat(fmt.Sprintf("%s/two-default.yaml", tempDir))
	s.NoError(err, "Expected two-default.yaml to exist")

	expectedOne := fmt.Sprintf("helm install one https://charts.bitnami.com/bitnami/one --version 19.4.3 --namespace default -f %s/one-default.yaml", tempDir)
	s.Contains(stdout, expectedOne, "Expected stdout to contain helm install command")

	expectedTwo := fmt.Sprintf("helm install two https://charts.some-other-registry.com/charts/two --version 2.3.4-alpga --namespace default -f %s/two-default.yaml --set password=\"1234\"", tempDir)
	s.Contains(stdout, expectedTwo, "Expected stdout to contain helm install command")
}
