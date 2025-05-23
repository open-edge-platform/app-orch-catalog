# SPDX-FileCopyrightText: 2023-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SHELL = bash -e -o pipefail

PLATFORM ?= --platform linux/x86_64

export CGO_ENABLED=1
export GO111MODULE=on
GOCMD := GOPRIVATE="github.com/intel*/*" go

GOTESTREPORT := $(shell which go-testreport || echo "")

all: component-tests

mod-update: # @HELP update Go modules
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor

install-go-test-report: # @HELP install go-test-report if not already installed
	@if [ -z "$(GOTESTREPORT)" ]; then \
		echo "Installing go-testreport..."; \
		go install github.com/becheran/go-testreport@latest; \
	else \
		echo "go-testreport is already installed"; \
	fi

.PHONY: component-tests
<<<<<<< HEAD
component-tests: install-go-test-report
	@echo "Running component tests..."
	#$(GOCMD) test -json -v -count=1  ./restapi/... | $(shell go env GOPATH)/bin/go-test-report -v -o test-report.html
	$(GOCMD) test -json -v -count=1  ./restapi/... > test-report.json


=======
component-tests:
	go test -v -count=1 ./restapi/... ./cmdline/...
>>>>>>> 4d9aa73 (add component tests for cmdline tools)
