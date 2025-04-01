// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"flag"
	"github.com/open-edge-platform/app-orch-catalog/test/basic"
	"strings"
	"testing"
	"time"
)

func main() {
	keycloakServer := flag.String("keycloakServer", "", "server/port of keycloak server")
	catalogServer := flag.String("catalogServer", "", "server/port of catalog server")
	catalogRESTServer := flag.String("catalogRESTServer", "", "server/port of catalog server")
	artifactFileName := flag.String("artifactFilename", "", "file path for artifact data")
	testsToRun := flag.String("tests", "all", "comma-separated list of names of tests to run; all for all of them")
	noClear := flag.Bool("no-clear", false, "do not perform data clean-up")
	testing.Init()

	flag.Parse()

	s := basic.TestSuite{NoClear: *noClear}
	s.SetT(&testing.T{})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	s.SetContext(ctx)
	s.SetupSuite()

	if keycloakServer != nil && *keycloakServer != "" {
		s.KeycloakServer = *keycloakServer
	}

	if catalogServer != nil && *catalogServer != "" {
		s.CatalogServer = *catalogServer
	}

	if catalogRESTServer != nil && *catalogRESTServer != "" {
		s.CatalogRESTServer = *catalogRESTServer
	}

	if artifactFileName != nil && *artifactFileName != "" {
		s.ArtifactFilename = *artifactFileName
	}

	tests := map[string]func(){
		"TestBasics":         s.TestBasics,
		"TestValidateBasics": s.TestValidateBasics,
		"TestRESTBasics":     s.TestRESTBasics,
		"TestUpload":         s.TestUpload,
		"TestScale":          s.TestScale,
		"TestScaleWorkloads": s.TestScaleWorkloads,

		"TestUpdateApplicationWithDeploymentRequirements": s.TestUpdateApplicationWithDeploymentRequirements,
	}

	names := []string{}
	if testsToRun == nil || *testsToRun == "all" {
		for name := range tests {
			names = append(names, name)
		}
	} else {
		names = append(names, strings.Split(*testsToRun, ",")...)
	}

	var itests []testing.InternalTest
	for _, name := range names {
		if test, ok := tests[name]; ok {
			itests = append(
				itests,
				testing.InternalTest{
					Name: name,
					F: func(t *testing.T) {
						s.SetupTest()
						s.SetT(t)
						s.SetContext(ctx)
						s.Run(name, test)
						s.CheckStatus(name)
						s.TearDownTest(ctx)
					},
				},
			)
		}
	}

	testing.Main(func(_, _ string) (bool, error) { return true, nil }, itests, nil, nil)

	s.CheckStatus("BasicSuite")
}
