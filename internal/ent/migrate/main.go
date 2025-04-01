// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/migrate"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	// Create a local migration directory able to understand Atlas migration file format for replay.
	dir, err := atlas.NewLocalDir("internal/ent/migrate/migrations")
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}
	// Migrate diff options.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(dialect.Postgres),        // Ent dialect to use
		schema.WithFormatter(atlas.DefaultFormatter),
	}
	if len(os.Args) != 3 {
		log.Fatalln("migration name is required. Use: 'go run -mod=mod ent/migrate/main.go <name>'")
	}
	// Note: The path below does NOT include actual secrets. These are access credentials for a temporary database
	// created in order to generate migration differences and is discarded immediately afterward.
	// Generate migrations using Atlas support for MySQL (note the Ent dialect option passed above).
	err = migrate.NamedDiff(ctx, fmt.Sprintf("postgres://%s@localhost:5432/test?sslmode=disable", os.Args[2]), os.Args[1], opts...) // Not actual secrets; see note above
	if err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}
}
