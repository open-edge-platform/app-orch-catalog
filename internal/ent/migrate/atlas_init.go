// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// InitializeAtlasTracking creates the atlas_schema_revisions table and marks
// migrations as applied for databases migration
// This handles the migration from Ent auto-migrate to Atlas-based migrations.
func InitializeAtlasTracking(dbPath string) error {
	// Open a direct database connection for Atlas tracking initialization
	db, err := sql.Open("postgres", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
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
		return fmt.Errorf("failed to check for atlas_schema_revisions table: %w", err)
	}

	if exists {
		log.Infof("Atlas migration tracking already initialized")
		return nil
	}

	log.Infof("Initializing Atlas migration tracking for pre-0.15.5 database")

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
		return fmt.Errorf("failed to create atlas_schema_revisions table: %w", err)
	}

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
		{"20250507105755", "ignoredResources"},
	}

	stmt, err := db.PrepareContext(ctx, `
		INSERT INTO atlas_schema_revisions 
		(version, description, type, applied, total, executed_at, execution_time, hash, operator_version)
		VALUES ($1, $2, 2, 0, 0, $3, 0, $4, 'Atlas CLI v0.38.1')
	`)

	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, m := range migrations {
		hash := fmt.Sprintf("h1:%s", m.description)
		_, err := stmt.ExecContext(ctx, m.version, m.description, now, hash)
		if err != nil {
			return fmt.Errorf("failed to mark migration %s as applied: %w", m.version, err)
		}
	}

	log.Infof("Successfully initialized Atlas tracking with %d pre-applied migrations", len(migrations))
	return nil
}
