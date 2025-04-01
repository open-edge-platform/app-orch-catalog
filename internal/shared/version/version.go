// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package version is used for reporting version and environment.
package version

import (
	"github.com/open-edge-platform/orch-library/go/dazl"
)

var log = dazl.GetPackageLogger()

// Default build-time variable.
// These values can (should) be overridden via ldflags when built with
// `make`
var (
	Version   = "unknown-version"
	GoVersion = "unknown-goversion"
	GitCommit = "unknown-gitcommit"
	GitDirty  = "unknown-gitdirty"
	Os        = "unknown-os"
	Arch      = "unknown-arch"
)

// LogVersion logs the version info
func LogVersion(indent string) {
	log.Infof("%sVersion:      %s\n", indent, Version)
	log.Infof("%sGoVersion:    %s\n", indent, GoVersion)
	log.Infof("%sGit Commit:   %s\n", indent, GitCommit)
	log.Infof("%sGit Dirty:    %s\n", indent, GitDirty)
	log.Infof("%sOS/Arch:      %s/%s\n", indent, Os, Arch)
}
