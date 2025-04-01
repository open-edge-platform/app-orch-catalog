// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package migrate

import (
	"context"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"os/exec"
	"time"
)

var log = dazl.GetPackageLogger()

// RunAtlasMigrations attempts to migrate the given database to the latest schema with the provided migration files.
func RunAtlasMigrations(dbPath, migrationsDir string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if migrationsDir == "" {
		msg := "No migrations directory given. Skipping."
		log.Infof(msg)
		return []byte(msg), nil
	}

	args := []string{
		"migrate",
		"apply",
		"--url", dbPath,
		"--dir", "file://" + migrationsDir,
		//"--baseline", "20230713211509",
	}
	log.Debugf("Prepared Atlas command: %v %v", "atlas", args)
	out, err := exec.CommandContext(ctx, "atlas", args...).CombinedOutput()
	log.Infof("Atlas output: %s", string(out))

	if err != nil {
		log.Errorf("Atlas command failed: %w", err)
	} else {
		log.Infof("Migration successful")
	}

	return out, err
}
