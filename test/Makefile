# SPDX-FileCopyrightText: 2023-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SHELL = bash -e -o pipefail

PLATFORM ?= --platform linux/x86_64

export CGO_ENABLED=1
export GO111MODULE=on
GOCMD := GOPRIVATE="github.com/intel*/*" go


all: component-tests

mod-update: # @HELP update Go modules
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor


.PHONY: component-tests
component-tests:
	go test -v -count=1 ./rest/...
	go test -v -count=1 ./rest -run TestListExtensions
