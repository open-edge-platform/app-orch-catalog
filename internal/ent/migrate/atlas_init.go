// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// InitializeAtlasTracking creates the atlas_schema_revisions table and marks
// migrations as applied for databases migrating from pre-Atlas versions.
// This handles the migration from Ent auto-migrate to Atlas-based migrations.
// For fresh installations, this function does nothing and lets Atlas handle everything.
// Returns (isFreshInstall bool, error)
func InitializeAtlasTracking(dbPath string) (bool, error) {
	// Open a direct database connection for Atlas tracking initialization
	db, err := sql.Open("postgres", dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		return false, fmt.Errorf("failed to ping database: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if atlas_schema_revisions already exists
	var exists bool
	err = db.QueryRowContext(ctx, `
SELECT EXISTS (
SELECT FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name = 'atlas_schema_revisions'
)
`).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check for atlas_schema_revisions table: %w", err)
	}

	if exists {
		log.Infof("Atlas migration tracking already initialized")
		return false, nil
	}

	// Check if this is a fresh installation by looking for catalog tables
	// If no catalog tables exist, this is a fresh install and we should let Atlas handle everything
	var hasCatalogTables bool
	err = db.QueryRowContext(ctx, `
SELECT EXISTS (
SELECT FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('registries', 'applications', 'deployment_packages', 'artifacts')
)
`).Scan(&hasCatalogTables)

	if err != nil {
		return false, fmt.Errorf("failed to check for existing catalog tables: %w", err)
	}

	if !hasCatalogTables {
		log.Infof("Fresh installation detected - Atlas will handle all migrations from scratch")
		return true, nil
	}

	log.Infof("Initializing Atlas migration tracking for pre-0.15.5 database")

	// Check if the 15th migration (ignoredResources) was actually applied
	// by verifying the namespace column is NOT NULL
	var namespaceIsNullable string
	err = db.QueryRowContext(ctx, `
SELECT is_nullable 
FROM information_schema.columns 
WHERE table_name = 'ignored_resources' 
AND column_name = 'namespace'
`).Scan(&namespaceIsNullable)

	migration15Applied := false
	if err == nil && namespaceIsNullable == "NO" {
		// Namespace is NOT NULL, meaning the 15th migration was applied
		migration15Applied = true
		log.Infof("Detected that migration 20250507105755 (ignoredResources) was already applied")
	} else if err != nil {
		log.Warnf("Could not verify namespace column state: %v", err)
	}

	// Create the tracking table
	_, err = db.ExecContext(ctx, `
CREATE TABLE atlas_schema_revisions (
version character varying NOT NULL,
description character varying NOT NULL,
type smallint NOT NULL DEFAULT 2,
applied integer NOT NULL DEFAULT 0,
total integer NOT NULL DEFAULT 0,
executed_at timestamptz NOT NULL,
execution_time bigint NOT NULL,
error text,
error_stmt text,
hash character varying NOT NULL,
partial_hashes jsonb,
operator_version character varying NOT NULL,
PRIMARY KEY (version)
)
`)

	if err != nil {
		return false, fmt.Errorf("failed to create atlas_schema_revisions table: %w", err)
	}

	// First 14 migrations that exist in all versions since v0.11.30
	migrations := []struct {
		version     string
		description string
	}{
		{"20230713224447", "base"},
		{"20230814153600", "uiextension"},
		{"20230907033412", "appname"},
		{"20231003230017", "ignored_resources"},
		{"20231207183234", "template_parameters"},
		{"20231214020503", "dependencies"},
		{"20240116191116", "chart-inventory"},
		{"20240117200903", "deployment-requirements"},
		{"20240306184024", "deployment-requirements-profile"},
		{"20240313214642", "deployment-requirements-cascade"},
		{"20240620170212", "kind"},
		{"20240711172122", "resource-namespace"},
		{"20240906060507", "v3"},
		{"20240906164744", "namespaces"},
	}

	// Only add 15th migration if it was actually applied
	if migration15Applied {
		migrations = append(migrations, struct {
			version     string
			description string
		}{"20250507105755", "ignoredResources"})
	}

	stmt, err := db.PrepareContext(ctx, `
INSERT INTO atlas_schema_revisions 
(version, description, type, applied, total, executed_at, execution_time, hash, operator_version)
VALUES ($1, $2, 2, 0, 0, $3, 0, $4, 'Atlas CLI v1.0.0')
`)

	if err != nil {
		return false, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, m := range migrations {
		hash := fmt.Sprintf("h1:%s", m.description)
		_, err := stmt.ExecContext(ctx, m.version, m.description, now, hash)
		if err != nil {
			return false, fmt.Errorf("failed to mark migration %s as applied: %w", m.version, err)
		}
	}

	log.Infof("Successfully initialized Atlas tracking with %d pre-applied migrations", len(migrations))
	return false, nil
}
