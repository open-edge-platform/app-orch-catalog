# SPDX-FileCopyrightText: 2023-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SHELL = bash -e -o pipefail

PLATFORM ?= --platform linux/x86_64

export CGO_ENABLED=1
export GO111MODULE=on
GOCMD := GOPRIVATE="github.com/intel*/*" go


all: load-catalog-test

mod-update: # @HELP update Go modules
	$(GOCMD) mod tidy
	$(GOCMD) mod vendor

catalog-test-build: mod-update
	go build -o ../build/_output/catalog-test ./cmd/catalog-test

catalog-test-docker: catalog-test-build
	cd ..; docker build ${PLATFORM} -t registry-rs.edgeorchestration.intel.com/edge-orch/app/app-orch-catalog:latest -f test/build/catalog-test/Dockerfile .

load-catalog-test: catalog-test-docker
	kind load docker-image registry-rs.edgeorchestration.intel.com/edge-orch/app/app-orch-catalog:latest

clear-previous-local:
	@test -f catalog-local-test.yaml && kubectl delete -f catalog-local-test.yaml || true; rm -f catalog-local-test.yaml

test-local: clear-previous-local
	@sed "s/\"--tests\",\"all\"/\"--tests\",\"$(TESTS)\"/" cmd/catalog-test/local.yaml >catalog-local-test.yaml
	@kubectl apply -f catalog-local-test.yaml
	@kubectl wait --for=condition=complete -n orch-app job/catalog-test --timeout=5m
	@! kubectl logs -n orch-app job/catalog-test | grep ^FAIL$

clear-previous-coder:
	@test -f catalog-coder-test.yaml && kubectl delete -f catalog-coder-test.yaml || true; rm -f catalog-coder-test.yaml
	@kubectl delete job -n orch-app catalog-test || true

test-coder: clear-previous-coder
	@sed "s/\"--tests\",\"all\"/\"--tests\",\"$(TESTS)\"/" cmd/catalog-test/coder.yaml >catalog-coder-test.yaml
	kubectl apply -f catalog-coder-test.yaml
	@kubectl wait --for=condition=complete -n orch-app job/catalog-test --timeout=5m
	@! kubectl logs -n orch-app job/catalog-test | grep ^FAIL$

basic-test:
	make test-local TESTS="TestBasics,TestRESTBasics,TestUpload,TestUpdateApplicationWithDeploymentRequirements"
basic-test-coder:
	make test-coder TESTS="TestBasics,TestRESTBasics,TestUpload,TestUpdateApplicationWithDeploymentRequirements"

scale-test:
	make test-local TESTS="TestScale"
scale-test-coder:
	make test-coder TESTS="TestScale"


.PHONY: test-catalog-api
test-catalog-api:
	go test -v ./rest -run TestSuite
