// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package rest is a suite of REST API functionality tests for the catalog service
package restapi

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const ()

// TestSuite is the basic test suite
type TestSuite struct {
	suite.Suite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
