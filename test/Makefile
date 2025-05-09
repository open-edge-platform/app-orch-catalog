# SPDX-FileCopyrightText: 2023-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SHELL = bash -e -o pipefail

PLATFORM ?= --platform linux/x86_64

export CGO_ENABLED=1
export GO111MODULE=on
GOCMD := GOPRIVATE="github.com/intel*/*" go

GOTESTREPORT := $(shell which go-test-report || echo "")

all: component-tests

mod-update: # @HELP update Go modules
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor

install-go-test-report: # @HELP install go-test-report if not already installed
	@if [ -z "$(GOTESTREPORT)" ]; then \
		echo "Installing go-test-report..."; \
		go install github.com/vakenbolt/go-test-report@latest; \
		export PATH=$(shell go env GOPATH)/bin:$$PATH; \
		hash go-test-report 2>/dev/null || { echo "go-test-report installation failed"; exit 1; }; \
	else \
		echo "go-test-report is already installed"; \
	fi

.PHONY: component-tests
component-tests: install-go-test-report
	@echo "Running component tests..."
	$(GOCMD) test -v -count=1 -json ./restapi/... | tee test-output.json | $(shell go env GOPATH)/bin/go-test-report -o test-report.html

