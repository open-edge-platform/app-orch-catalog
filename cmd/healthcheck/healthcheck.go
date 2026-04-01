// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// healthcheck is a minimal binary used as a Docker HEALTHCHECK for the
// application-catalog container. It calls the gRPC Health Checking Protocol
// endpoint served by the application-catalog process on localhost:8080.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("localhost:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: dial error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx,
		&grpc_health_v1.HealthCheckRequest{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: health check error: %v\n", err)
		os.Exit(1)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		fmt.Fprintf(os.Stderr, "healthcheck: not serving: %v\n", resp.GetStatus())
		os.Exit(1)
	}
	os.Exit(0)
}
