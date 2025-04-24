// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"github.com/open-edge-platform/app-orch-catalog/internal/restproxy"
	"github.com/open-edge-platform/orch-library/go/dazl"
	_ "github.com/open-edge-platform/orch-library/go/dazl/zap"
)

// logger
var log = dazl.GetPackageLogger()

func main() {
	allowedCorsOrigins := flag.String("allowedCorsOrigins", "",
		"Comma separated list of allowed CORS origins")
	basePath := flag.String("basePath", "",
		"The rest server base Path")
	specFilePath := flag.String("spec-file-path", "/usr/local/etc/openapi.yaml",
		"The location of the spec file")
	port := flag.Int("rest-port", 8081,
		"port that REST service runs on")
	gRPCEndpoint := flag.String("grpc-endpoint", "localhost:8080",
		"The endpoint of the gRPC server")
	oidcExternal := flag.String("openidc-external", "",
		"URL of external OIDC server")
	flag.Parse()

	cfg := &restproxy.Config{
		AllowedCorsOrigins: *allowedCorsOrigins,
		BasePath:           *basePath,
		SpecFilePath:       *specFilePath,
		Port:               *port,
		GRPCEndpoint:       *gRPCEndpoint,
		OIDCExternal:       *oidcExternal,
	}

	rp, err := restproxy.NewRESTProxy(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err = rp.Run(); err != nil {
		log.Fatal(err)
	}
}
