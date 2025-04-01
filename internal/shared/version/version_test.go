// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"
)

func TestGetVersion(_ *testing.T) {
	// TODO: dazl logger does have easy support for output capture for unit testing
	// at this time. Just make sure it doesn't segfault.
	LogVersion("  ")
}
