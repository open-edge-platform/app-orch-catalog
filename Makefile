# SPDX-FileCopyrightText: (C) 2023 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

SHELL := bash -eu -o pipefail

# Code Versions
VERSION            := $(shell cat VERSION)
CHART_VERSION      := $(shell cat VERSION)
VERSION_DEV_SUFFIX := -dev
GIT_COMMIT         ?= $(shell git rev-parse --short HEAD)

ifeq ($(patsubst %$(VERSION_DEV_SUFFIX),,$(lastword $(VERSION))),)
    DOCKER_VERSION ?= $(VERSION)-$(GIT_COMMIT)
else
    DOCKER_VERSION ?= $(VERSION)
endif

PLATFORM                       ?= --platform linux/x86_64
PUBLISH_REPOSITORY             ?= edge-orch
PUBLISH_REGISTRY               ?= 080137407410.dkr.ecr.us-west-2.amazonaws.com
PUBLISH_SUB_PROJ               ?= app
PUBLISH_CHART_PREFIX           ?= charts
CHART_NAME                     ?= app-orch-catalog
APPLICATION_CATALOG_IMAGE_NAME ?= app-orch-catalog
APPLICATION_CATALOG_VERSION    ?= ${VERSION}
INSTALL_PATH                   ?= /usr/local/bin
FUZZ_SECONDS                   ?= 60

OIE_CI_TESTING                   = rrp-devops/oie_ci_testing
OIE_CI_TESTING_VER               = 2.9.34
GOLANG_COVER_VERSION             = v0.2.0
GOLANG_GOCOVER_COBERTURA_VERSION = v1.2.0
GOPATH                           := $(shell go env GOPATH)
GOCMD                            := GOPRIVATE="github.com/open-edge-platform/*" go
PKG                              := github.com/open-edge-platform/app-orch-catalog

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL        ?= $(shell git remote get-url $(shell git remote | head -n 1))
DOCKER_LABEL_VCS_REF        = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE     ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE    = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_EXTRA_ARGS           ?=
DOCKER_BUILD_ARGS ?= \
	${DOCKER_EXTRA_ARGS} \
	--build-arg org_label_schema_version="${APPLICATION_CATALOG_VERSION}" \
	--build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
	--build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
	--build-arg org_opencord_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \
	--build-arg org_opencord_vcs_dirty="${DOCKER_LABEL_VCS_DIRTY}"

DOCKER_BUILD_COMMAND    := docker build

## CHART_NAMEs are specified in Chart.yaml
CHART_NAME					?= app-orch-catalog

## CHART_PATHs is given based on repo structure
CHART_PATH					?= "./deployments/app-orch-catalog"

## MIGRATION_BASE_VERSION is the base version from which migration should be tested
MIGRATION_BASE_VERSION   	?= 0.5.5

## POSTGRESS_VERSION is specified in Chart.yaml
POSTGRES_VERSION				?= $(shell yq -r .version ./deployments/postgres/Chart.yaml)
## CHART_TEST is specified in test-connection.yaml
CHART_TEST					?= test-connection
## CHART_BUILD_DIR is given based on repo structure
CHART_BUILD_DIR				?= ./build/_output/
## CHART_APP_VERSION is modified on every commit
CHART_APP_VERSION			?= "${APPLICATION_CATALOG_VERSION}"
## CHART_NAMESPACE can be modified here
CHART_NAMESPACE				?= orch-app
## CHART_RELEASE can be modified here
CHART_RELEASE				?= catalog-service

POSTGRES_CHART_PATH			?= "./deployments/postgres"
POSTGRES_CHART_VERSION		?= 12.12.10

OAPI_CODEGEN_VERSION ?= v2.2.0

# The endpoint URL of a keycloak server e.g. http://keycloak/realms/master refers to a keycloak server in the cluster
OIDC_SERVER                 ?= http://keycloak.$(CHART_NAMESPACE).svc/realms/master
# The endpoint URL of a keycloak server e.g. http://localhost:8090/realms/master refers to a keycloak server in the cluster
# by it's externally visible address
OIDC_SERVER_EXTERNAL        ?= http://localhost:8090/realms/master

YQ_QUERY_SCHEMA_OBJ=select(.key != "Create*" and .key != "List*" and .key != "Update*" and .key != "Get*" and .key != "Status" and .key != "GoogleProtobufAny" and .key != "ApplicationReference" and .key != "ApplicationDependency" and .key != "ArtifactReference" and .key != "Endpoint" and .key != "UIExtension")
API_DIR                     = api
TMP_DIR                     = /tmp

PGUSER						?= $(shell grep PGUSER deployments/application-catalog/templates/postgres-secrets.yaml  | cut -d'"' -f2)
PGPASSWORD					?= $(shell grep PGPASSWORD deployments/application-catalog/templates/postgres-secrets.yaml  | cut -d'"' -f2)

HELM_REPOSITORY    		?=
HELM_REGISTRY      		?=

DOCKER_TAG              := $(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(APPLICATION_CATALOG_IMAGE_NAME):$(DOCKER_VERSION)

RELEASE_DIR     ?= release
RELEASE_OS_ARCH ?= linux-amd64 linux-arm64 windows-amd64 darwin-amd64 darwin-arm64

SCHEMA_CMD_DIR         ?= ./cmd/schema
SCHEMA_RELEASE_NAME    ?= catalog-schema
SCHEMA_RELEASE_BINS    := $(foreach rel,$(RELEASE_OS_ARCH),$(RELEASE_DIR)/$(SCHEMA_RELEASE_NAME)-$(rel))

HELM_TO_DP_CMD_DIR         ?= ./cmd/helm-to-dp
HELM_TO_DP_RELEASE_NAME    ?= helm-to-dp
HELM_TO_DP_RELEASE_BINS    := $(foreach rel,$(RELEASE_OS_ARCH),$(RELEASE_DIR)/$(HELM_TO_DP_RELEASE_NAME)-$(rel))

## coder env variables
MGMT_NAME        ?= kind
MGMT_CLUSTER     ?= kind-${MGMT_NAME}
CODER_DIR 		 ?= ~/orch-deploy
CATALOG_HELM_PKG ?= ${CHART_BUILD_DIR}${CHART_NAME}-${CHART_VERSION}.tgz

SAMPLE_ORG_ID := "11111111-1111-1111-1111-111111111111"
SAMPLE_PROJECT_ID := "11111111-1111-1111-1111-222222222222"
PLATFORM_NS := "orch-platform"
KEYCLOAK_HELM_VERSION := 24.4.11
BUF_VERSION := 1.52.1

# Functions to extract the OS/ARCH
schema_rel_os    = $(word 3, $(subst -, ,$(notdir $@)))
schema_rel_arch  = $(word 4, $(subst -, ,$(notdir $@)))

helm_to_dp_rel_os    = $(word 4, $(subst -, ,$(notdir $@)))
helm_to_dp_rel_arch  = $(word 5, $(subst -, ,$(notdir $@)))

# Exclude these packages from coverage analysis
EXCLUDE_PKGS_TEST := grep -v $(PKG)/pkg/restClient | grep -v $(PKG)/pkg/api | grep -v $(PKG)/internal/ent | grep -v $(PKG)/internal/testing | grep -v $(PKG)/pkg/restProxy | grep -v $(PKG)/internal/testing


.PHONY: build lint test all

all: build lint test ## Runs build, lint, test stages

.PHONY: ent-generate
ent-generate: ## Regenerate ENT assets from schema.go
	go generate ./internal/ent/generate.go

.PHONY: migration-generate
migration-generate: ## Generate DB schema migration "make migration-generate MIGRATION=<migration-name>"
	@if test -z $(MIGRATION); then echo "Please specify migration name" && exit 1; fi
	@docker run --name migration --rm -p 5432:5432 -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=test -d postgres # temporary DB
	@sleep 3
	@go run -mod=mod internal/ent/migrate/main.go $(MIGRATION) postgres:pass # use credentials of the temporary DB; not actual secrets
	@docker container kill migration

.PHONY: migration-lint
migration-lint: ## Validate the DB schema migration
	@atlas migrate lint \
      --dev-url="docker://postgres/15/test?search_path=public" \
      --dir="file://internal/ent/migrate/migrations" \
      --latest=1

.PHONY: ent-describe
ent-describe: ## Describe ENT assets
	go run -mod=mod entgo.io/ent/cmd/ent describe ./internal/ent/schema


# Define the target for installing all plugins
install-protoc-plugins:
	@echo "Installing protoc-gen-doc..."
	@go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	@echo "Installing protoc-gen-validate..."
	@go install github.com/envoyproxy/protoc-gen-validate@latest
	@echo "Installing protoc-gen-go-grpc..."
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Installing protoc-gen-openapi"
	@go install github.com/kollalabs/protoc-gen-openapi@latest
	echo "Installing oapi-codegen"
	# for the binary install
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}
	@echo "Installing buf..."
	go install github.com/bufbuild/buf/cmd/buf@v${BUF_VERSION}
	# for the binary installation
	@echo "Adding Go bin directory to PATH..."
	@export PATH=$(PATH):$(GOBIN)
	@echo "All plugins installed successfully."

# Define a target to verify the installation of all plugins
verify-protoc-plugins:
	@echo "Verifying protoc-gen-doc installation..."
	@command -v protoc-gen-doc >/dev/null 2>&1 && echo "protoc-gen-doc is installed." || echo "protoc-gen-doc is not installed."
	@echo "Verifying protoc-gen-validate installation..."
	@command -v protoc-gen-validate >/dev/null 2>&1 && echo "protoc-gen-validate is installed." || echo "protoc-gen-validate is not installed."
	@echo "Verifying protoc-gen-go-grpc installation..."
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 && echo "protoc-gen-go-grpc is installed." || echo "protoc-gen-go-grpc is not installed."
	@echo "Verifying protoc-gen-openapi installation..."
	@command -v protoc-gen-openapi >/dev/null 2>&1 && echo "protoc-gen-openapi is installed." || echo "protoc-gen-openapi is not installed."
	echo "Verifying oapi-codegen installation..."
	@command -v oapi-codegen >/dev/null 2>&1 && echo "oapi-codegen is installed." || echo "oapi-codegen is not installed."
	@echo "Verifying buf installation..."
	@command -v buf >/dev/null 2>&1 && echo "buf is installed." || echo "buf is not installed."

#### Python Targets ####

VENV_NAME = venv-env
$(VENV_NAME): requirements.txt ## Create Python venv
	python3 -m venv $@ ;\
  set +u; . ./$@/bin/activate; set -u ;\
  python -m pip install --upgrade pip ;\
  python -m pip install openapi-spec-validator;\
  python -m pip install -r requirements.txt

.PHONY: proto-generate
proto-generate: proto-generate-local schema-generate ## generate language files from proto definitions


.PHONY: proto-generate-local ## Generate Openapi, customize it and generate rest client
proto-generate-local: buf-generate customise-openapi openapi-spec-validate rest-client-gen buf-format
	@echo "Dependencies installed and virtual environment activated."

.PHONY: buf-format
buf-format: ## Format protobuf file for consistent layout
	buf format -w

.PHONY: buf-generate
buf-generate: $(VENV_NAME)  ## Format protobuf file for consistent layout
	set +u; . ./$</bin/activate; set -u ;\
            buf --version ;\
            buf generate
	# suppress multiple blank lines created by protoc-gen-docs
	cat -s docs/catalog-grpcapi.md > docs/catalog-grpcapi.tmp && mv docs/catalog-grpcapi.tmp docs/catalog-grpcapi.md

.PHONY: schema-generate
schema-generate: ## Generate YAML schema from OpenAPI Spec
	go run cmd/schema/schema.go generate

.PHONY: customise-openapi
customise-openapi: ## Customize Openapi Spec after generation
	@echo "openapi.yaml Add required true to projectId query parameter"
	@yq -i '(.paths.*.*.parameters[] | select(.name=="projectId") |.required) = true' api/spec/openapi.yaml
	@echo "openapi.yaml required false for projectId query parameter in Lists"
	@yq -i '(.paths.*.get | select(.operationId=="CatalogService_List*") | .parameters[] | select(.name=="projectId") |.required) = false' api/spec/openapi.yaml
	@# TODO: Replace the following remedy with yq-based one; both of the previous yq commands wrongly inject "get: null" for the upload path.
	@echo "openapi.yaml removing upload path get"
	@grep -v ' get: null' api/spec/openapi.yaml > api/spec/openapi.yaml.aux; mv api/spec/openapi.yaml.aux api/spec/openapi.yaml

.PHONY: openapi-spec-validate
openapi-spec-validate: $(VENV_NAME) ## Install openapi-spec-validator
	set +u; . ./$</bin/activate; set -u ;\
	openapi-spec-validator api/spec/openapi.yaml

.PHONY: oapi-codegen
oapi-codegen: ## Install oapi-codegen
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}

.PHONY: rest-client-gen
rest-client-gen: ## Generate Rest client from the generated openapi spec.
	oapi-codegen -generate client -old-config-style -package restClient -o pkg/restClient/client.go api/spec/openapi.yaml
	oapi-codegen -generate types -old-config-style -package restClient -o pkg/restClient/types.go api/spec/openapi.yaml

.PHONY: mod-update
mod-update: ## Update Go modules
	$(GOCMD) mod tidy

.PHONY: vendor
vendor: ## Build vendor directory of module dependencies
	$(GOCMD) mod vendor

.PHONY: build
build: mod-update ## Runs build stage
	go build -o build/_output/application-catalog ./cmd/application-catalog
	go build -o build/_output/rest-proxy ./cmd/rest-proxy
	go build -o build/_output/catalog-schema ./cmd/schema

.PHONY: install
install: ## Installs the application-catalog server and the schema generation/validation tool
	cp build/_output/application-catalog ${INSTALL_PATH}
	cp build/_output/catalog-schema ${INSTALL_PATH}



.PHONY: license
license: $(VENV_NAME) ## Check licensing with the reuse tool.
	. ./$</bin/activate; set -u;\
	reuse --version;\
	reuse --root . lint

.PHONY: yamllint
YAML_FILES         := $(shell find . -type f \( -name '*.yaml' -o -name '*.yml' \) -print )
YAML_IGNORE        := .cache, vendor, ci, .github/workflows, test, pkg/schema, api/spec/openapi.yaml, kind, internal/northbound/testdata, deployments/app-orch-catalog/templates, $(VENV_NAME)

yamllint: $(VENV_NAME) ## Lint YAML files
	. ./$</bin/activate; set -u ;\
  yamllint --version ;\
  yamllint -d '{extends: default, rules: {line-length: {max: 500}, braces: {min-spaces-inside: 0, max-spaces-inside: 5}, brackets: {min-spaces-inside: 0, max-spaces-inside: 5},colons: {max-spaces-before: 1, max-spaces-after: 5}}, ignore: [$(YAML_IGNORE)]}' -s $(YAML_FILES)


.PHONY: lint
lint: rego-service-write-rule-match yamllint mdlint shelllint ## Runs lint stage
	helm lint ${CHART_PATH}
	buf lint
	golangci-lint run --timeout 10m
	hadolint --ignore DL3059 build/Dockerfile
	opa check ${CHART_PATH}/files/openpolicyagent/*.rego

.PHONY: mdlint ## lint markdown files
mdlint:
	@echo "---MAKEFILE LINT README---"
	@markdownlint --version
	@markdownlint "*.md"
	@echo "---END MAKEFILE LINT README---"

SHELL_FILES := $(shell find . -not \( -path ./ci -prune \) -not \( -path ./trivy -prune \) -not \( -path ./vendor -prune \) -type f -name \*.sh;)
.PHONY: shelllint ## lint shell files
shelllint:
	@echo "---MAKEFILE LINT SCRIPTS---"
	@shellcheck --version
	set -e ;\
	$(foreach file,$(SHELL_FILES),\
		shellcheck $(file) ;\
	)
	@echo "---END MAKEFILE LINT SCRIPTS---"

.PHONY: rego-service-write-rule-match
rego-service-write-rule-match: ## For every service request in Proto we expect a corresponding REGO rule
	@egrep -oh "\((Create|Update|Delete|List|Get|Watch|Upload).*Request" ${API_DIR}/catalog/v3/service.proto | awk -F'(' '{print $$2}' | sort > ${TMP_DIR}/list_service_requests_out;
	@egrep -oh "(Create|Update|Delete|List|Get|Watch|Upload).*Request {" ${CHART_PATH}/files/openpolicyagent/*.rego | grep -v "WithSensitiveInfo" | awk '{print $$1}' | sort > ${TMP_DIR}/list_rego_rules_out;
	@diff ${TMP_DIR}/list_service_requests_out ${TMP_DIR}/list_rego_rules_out;

.PHONY: rego-rule-test
rego-rule-test: ## test the REGO rules
	@make -C deployments/app-orch-catalog/files/openpolicyagent/testdata/artifact all
	@make -C deployments/app-orch-catalog/files/openpolicyagent/testdata/deployment-package all
	@make -C deployments/app-orch-catalog/files/openpolicyagent/testdata/upload all
	@make -C deployments/app-orch-catalog/files/openpolicyagent/testdata/registry all

.PHONY: go-cover-dependency
go-cover-dependency: ## install the gocover tool
	go tool cover -V || go install golang.org/x/tools/cmd/cover@${GOLANG_COVER_VERSION}
	go install github.com/boumenot/gocover-cobertura@${GOLANG_GOCOVER_COBERTURA_VERSION}

.PHONY: go-format
go-format: ## Formats go source files
	@go fmt $(shell sh -c "go list ./...")

.PHONY: test
test: mod-update rego-rule-test ## Runs test stage
	$(GOCMD) test -race -gcflags=-l `go list $(PKG)/cmd/... $(PKG)/pkg/... $(PKG)/internal/...`

FUZZ_FUNCS ?= FuzzCreateRegistry FuzzCreateArtifact FuzzCreateDeploymentPackage
FUZZ_FUNC_PATH := ./internal/northbound

.PHONY: go-fuzz
go-fuzz: ## GO fuzz tests
	for func in $(FUZZ_FUNCS); do \
		$(GOCMD) test $(FUZZ_FUNC_PATH) -fuzz $$func -fuzztime=${FUZZ_SECONDS}s -v; \
	done

.PHONY: coverage
coverage: go-cover-dependency ## Runs coverage stage
	@echo "---MAKEFILE COVERAGE---"
	$(GOCMD) test -gcflags=-l -race -coverpkg=$$(go list ./... | ${EXCLUDE_PKGS_TEST} | tr '\n' ,) -coverprofile=coverage.txt -covermode atomic `go list $(PKG)/cmd/... $(PKG)/pkg/... $(PKG)/internal/... | ${EXCLUDE_PKGS_TEST}`
	${GOPATH}/bin/gocover-cobertura < coverage.txt > coverage.xml
	#$(GOCMD) tool cover -html=coverage.txt -o cover.html
	#$(GOCMD) tool cover -func cover.out -o cover.function-coverage.log
	@echo "---END MAKEFILE COVERAGE---"

.PHONY: docker-build
docker-build: mod-update vendor ##Builds the docker image
	@echo "---MAKEFILE DOCKER BUILD---"
	$(DOCKER_BUILD_COMMAND) . -f build/Dockerfile \
	$(PLATFORM) \
	-t $(APPLICATION_CATALOG_IMAGE_NAME):$(DOCKER_VERSION) \
	$(DOCKER_BUILD_ARGS)
	docker tag $(APPLICATION_CATALOG_IMAGE_NAME):$(DOCKER_VERSION) $(DOCKER_TAG)

.PHONY: docker-push
docker-push: docker-build ## Push the docker image to the target registry
	aws ecr create-repository --region us-west-2 --repository-name $(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(APPLICATION_CATALOG_IMAGE_NAME) || true

	docker tag $(APPLICATION_CATALOG_IMAGE_NAME):$(VERSION) $(DOCKER_TAG)
	docker push $(DOCKER_TAG)

docker-list: ## Print name of docker container image
	@echo "images:"
	@echo "  $(APPLICATION_CATALOG_IMAGE_NAME):"
	@echo "    name: '$(APPLICATION_CATALOG_IMAGE_NAME):$(DOCKER_VERSION)'"
	@echo "    version: '$(DOCKER_VERSION)'"
	@echo "    gitTagPrefix: 'v'"
	@echo "    buildTarget: 'docker-build'"

.PHONY: kind-delete
kind-delete: ## Deletes kind cluster
	kind delete cluster

export KIND_CONFIG_FILE_NAME=kind.config.yaml
## Create file definition for the kind cluster
define get_kind_config_file
# Remove config file
rm -rf ${KIND_CONFIG_FILE_NAME}
# Define config file
cat << EOF >> ${KIND_CONFIG_FILE_NAME}
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
endef
export KIND_CLUSTER_FILE_CREATOR = $(value get_kind_config_file)

kind-config-file:; @ eval "$$KIND_CLUSTER_FILE_CREATOR"

.PHONY: kind-config-file kind
kind: kind-config-file ## Creates kind cluster
	kind create cluster --image kindest/node:v1.23.4 --config=${KIND_CONFIG_FILE_NAME}
	# Remove config file
	rm -rf ${KIND_CONFIG_FILE_NAME}
	kubectl cluster-info --context kind-kind
	# Add Ingress NGINX
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	# Add Cert-Manager
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml

.PHONY: chart-clean
chart-clean: ## Cleans the build directory of the helm chart
	rm -rf ${CHART_BUILD_DIR}/*.tgz
	yq eval -i 'del(.annotations.revision)' ${CHART_PATH}/Chart.yaml
	yq eval -i 'del(.annotations.created)' ${CHART_PATH}/Chart.yaml

.PHONY: helm-build
helm-build: chart ## builds the helm charts release

.PHONY: chart
chart: chart-clean ## Builds the application catalog helm chart
	@echo "---MAKEFILE CHART---"
	yq eval -i '.version = "${CHART_VERSION}"' ${CHART_PATH}/Chart.yaml; \
	yq eval -i '.appVersion = "${DOCKER_VERSION}"' ${CHART_PATH}/Chart.yaml; \
	yq eval -i '.annotations.revision = "${DOCKER_LABEL_REVISION}"' ${CHART_PATH}/Chart.yaml; \
	yq eval -i '.annotations.created = "${DOCKER_LABEL_BUILD_DATE}"' ${CHART_PATH}/Chart.yaml; \
	helm package --app-version=${DOCKER_VERSION} --version=${CHART_VERSION} --dependency-update --destination ${CHART_BUILD_DIR} ${CHART_PATH}
	@echo "---END MAKEFILE CHART---"

helmlint: ## Lint Helm charts.
	@# Help: Lint Helm charts
	helm lint ${CHART_PATH}

helm-push: ## Push helm charts.
	@# Help: Pushes the helm chart
	@echo "---MAKEFILE HELM PUSH---"
	aws ecr create-repository --region us-west-2 --repository-name $(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(PUBLISH_CHART_PREFIX)/$(CHART_NAME) || true
	helm push ${CHART_BUILD_DIR}${CHART_NAME}-[0-9]*.tgz oci://$(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(PUBLISH_CHART_PREFIX)
	@echo "---END MAKEFILE HELM PUSH---"

helm-list: ## List helm charts, tag format, and versions in YAML format
	@echo "charts:" ;\
  echo "  $(CHART_NAME):" ;\
  echo -n "    "; grep "^version" "${CHART_PATH}/Chart.yaml"  ;\
  echo "    gitTagPrefix: 'v'" ;\
  echo "    outDir: '${CHART_BUILD_DIR}'" ;\

.PHONY: chart-install-kind
chart-install-kind: docker-build kind-load chart postgres-install-kind keycloak-install-kind catalog-install-kind ## install the catalog app in a local kind cluster

.PHONY: catalog-install-kind
catalog-install-kind: ## Installs the catalog helm chart in the kind cluster
	@echo "---MAKEFILE CHART-INSTALL-KIND---"
	helm upgrade --install -n ${CHART_NAMESPACE} ${CHART_RELEASE} \
			--wait --timeout 300s \
			--values ./kind/kind-ingress-values.yaml \
			--set fullnameOverride=${CHART_RELEASE} \
			--set postgres.local.secrets=true \
			--set openidc.issuer=${OIDC_SERVER} \
			--set openidc.external=${OIDC_SERVER_EXTERNAL} \
			--set postgres.ssl=true \
			--set logging.rootLogger.level=DEBUG \
			${CHART_BUILD_DIR}${CHART_NAME}-${CHART_VERSION}.tgz
	@echo "---END MAKEFILE CHART-INSTALL-KIND---"

.PHONY: catalog-migration-base-install-kind
catalog-migration-base-install-kind: ## Installs the catalog helm chart in the kind cluster (for manual migration testing)
	@echo "Installing application-catalog:${MIGRATION_BASE_VERSION}..."
	helm pull oci://registry-rs.edgeorchestration.intel.com/edge-orch/${CHART_NAME} --version ${MIGRATION_BASE_VERSION}
	helm upgrade --install -n ${CHART_NAMESPACE} ${CHART_RELEASE} \
			--wait --timeout 300s \
			--values ./kind/kind-ingress-values.yaml \
			--set fullnameOverride=${CHART_RELEASE} \
			--set postgres.local.secrets=true \
			--set postgres.secrets="application-catalog-postgres-dev-config" \
			--set openidc.issuer=${OIDC_SERVER} \
			--set openidc.external=${OIDC_SERVER_EXTERNAL} \
			--set postgres.ssl=true \
			--set logging.rootLogger.level=DEBUG \
			${CHART_NAME}-${MIGRATION_BASE_VERSION}.tgz

.PHONY: postgres-install-kind
postgres-install-kind: ## Installs the postgres helm chart in the kind cluster
	helm install ${CHART_RELEASE}-db oci://registry-1.docker.io/bitnamicharts/postgresql \
		--wait --timeout 300s \
		--version ${POSTGRES_CHART_VERSION} \
		--create-namespace -n ${CHART_NAMESPACE} \
		--set fullnameOverride=${CHART_RELEASE}-db-postgres \
		--set auth.username=${PGUSER} --set auth.password=${PGPASSWORD} \
		--set service.ports.postgresql=5432 \
		--set tls.enabled=true --set tls.autoGenerated=true

.PHONY: keycloak-install-kind
keycloak-install-kind: ## Installs the keycloak server chart in the kind cluster
	@echo "---MAKEFILE KEYCLOAK-INSTALL-KIND---"
	helm -n ${CHART_NAMESPACE} install --create-namespace keycloak oci://registry-1.docker.io/bitnamicharts/keycloak --version ${KEYCLOAK_HELM_VERSION} -f deployments/keycloak-dev/dev-values.yaml -f deployments/keycloak-dev/values.yaml --timeout 8m
	@echo "---END MAKEFILE KEYCLOAK-INSTALL-KIND---"

keycloak-config-generate: ## Generates Keycloak configuration file from a running cluster e.g. Coder
	@SAMPLE_ORG_ID=$(SAMPLE_ORG_ID) \
	SAMPLE_PROJECT_ID=$(SAMPLE_PROJECT_ID) \
	PLATFORM_NS=$(PLATFORM_NS) \
	KEYCLOAK_HELM_VERSION=$(KEYCLOAK_HELM_VERSION) \
 	DEBUG=false scripts/extract-keycloak-config.sh

.PHONY: kind-migration-base-load
kind-migration-base-load: ## Load base versions of catalog image into the kind cluster (for migration testing)
	docker pull ${PUBLISH_REGISTRY}${PUBLISH_REPOSITORY}/$(APPLICATION_CATALOG_IMAGE_NAME):${MIGRATION_BASE_VERSION}
	kind load docker-image ${PUBLISH_REGISTRY}${PUBLISH_REPOSITORY}/$(APPLICATION_CATALOG_IMAGE_NAME):${MIGRATION_BASE_VERSION}


.PHONY: migration-test
migration-test: delete-namespace kind-migration-base-load postgres-install-kind keycloak-install-kind ## Executes a deployment sequence to test DB schema migration
	@sleep 10
	@echo "Installing initial catalog service version..."
	make catalog-migration-base-install-kind
	@sleep 15
	@echo "Running initial basic tests..."
	make -C test test-local TESTS=TestBasics
	kubectl -n ${CHART_NAMESPACE} logs deploy/$(CHART_RELEASE) application-catalog-server

	@sleep 10
	kubectl -n ${CHART_NAMESPACE} logs job/catalog-test
	make -C test clear-previous-local

	@echo "Upgrading catalog service version..."
	make catalog-install-kind
	@sleep 15
	@echo "Validating data after migration..."
	make -C test test-local NO_CLEAR=--no-clear TESTS=TestValidateBasics
	kubectl -n ${CHART_NAMESPACE} logs deploy/$(CHART_RELEASE) application-catalog-server
	kubectl -n ${CHART_NAMESPACE} logs job/catalog-test

.PHONY: delete-namespace
delete-namespace: ## delete namspace in a development deployment
	@kubectl delete namespace $(CHART_RELEASE) || echo "Namespace already deleted"

.PHONY: chart-test
chart-test: ## Performs smoketest of the deployment
	docker pull busybox:1.36.0
	docker tag busybox:1.36.0 docker.io/library/busybox:1.36.0
	kind load docker-image docker.io/library/busybox:1.36.0
	helm test ${CHART_RELEASE} -n ${CHART_NAMESPACE}
	kubectl -n ${CHART_NAMESPACE} logs ${CHART_RELEASE}-${CHART_NAME}-${CHART_TEST} --all-containers

.PHONY: chart-test-delete
chart-test-delete: ## Deletes the pod that executed smoketest
	kubectl delete pod ${CHART_RELEASE}-${CHART_NAME}-${CHART_TEST} -n ${CHART_NAMESPACE}

.PHONY: chart-uninstall
chart-uninstall: catalog-uninstall postgres-uninstall ## Uninstalls the helm charts

.PHONY: catalog-uninstall
catalog-uninstall: ## Uninstalls the catalog helm chart
	helm uninstall -n ${CHART_NAMESPACE} ${CHART_RELEASE}

.PHONY: postgres-uninstall
postgres-uninstall: ## Uninstalls the postgres helm chart
	helm uninstall -n ${CHART_NAMESPACE} ${CHART_RELEASE}-db

.PHONY: integration-tests
integration-tests: ## run integration tests locally
	make -C test

.PHONY: list
list: help ## displays make targets

.PHONY: load-catalog-test
load-catalog-test: ## loads catalog test images
	make -C test load-catalog-test

.PHONY: run-basic-test-coder run-basic-test-kind run-scale-test-coder run-scale-test-kind
run-basic-test-kind: load-catalog-test ## runs basic integration test in a kind environment
	make -C test basic-test

run-basic-test-coder: load-catalog-test ## runs basic integration test in a coder environment
	make -C test basic-test-coder

run-scale-test-kind: load-catalog-test ## runs scale integration test in a kind environment
	make -C test scale-test

run-scale-test-coder: load-catalog-test ## runs basic integration test in a coder environment
	make -C test scale-test-coder

use-orch-context:
	kubectl config use-context ${MGMT_CLUSTER}


.PHONY: kind-load
kind-load: docker-build
	# Override PUBLISH_REGISTRY for this target
	$(eval PUBLISH_REGISTRY := registry-rs.edgeorchestration.intel.com)
	# Update DOCKER_TAG to reflect the overridden registry
	$(eval DOCKER_TAG := $(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(APPLICATION_CATALOG_IMAGE_NAME):$(DOCKER_VERSION))
	# Explicitly tag the image with the correct registry
	docker tag $(APPLICATION_CATALOG_IMAGE_NAME):$(VERSION) $(DOCKER_TAG)
	# Load the Docker image into the kind cluster
	kind load docker-image -n ${MGMT_NAME} $(DOCKER_TAG)

.PHONY: coder-rebuild
coder-rebuild: kind-load ## Rebuild the application catalog from source and redeploy
	kubectl config use-context ${MGMT_CLUSTER}
	kubectl -n ${CHART_NAMESPACE} delete pod -l app.kubernetes.io/instance=${APPLICATION_CATALOG_IMAGE_NAME}

.PHONY: coder-redeploy
coder-redeploy: coder-rebuild chart ## Installs the helm chart in the kind cluster
	@echo "---MAKEFILE CHART-INSTALL-KIND---"
	kubectl config use-context ${MGMT_CLUSTER}
	kubectl patch application -n dev root-app --type=merge -p '{"spec":{"syncPolicy":{"automated":{"selfHeal":false}}}}'
	kubectl delete application -n dev app-orch-catalog --ignore-not-found=true
	helm upgrade --install -n orch-app app-orch-catalog -f $(CODER_DIR)/argocd/applications/configs/app-orch-catalog.yaml  $(CATALOG_HELM_PKG)
	helm -n orch-app ls
	@echo "---END MAKEFILE CHART-INSTALL-KIND---"

$(SCHEMA_RELEASE_BINS):
	export GOOS=$(schema_rel_os) ;\
	export GOARCH=$(schema_rel_arch) ;\
	GOPRIVATE=$(GOPRIVATE) go build -o "$@" $(SCHEMA_CMD_DIR)

$(HELM_TO_DP_RELEASE_BINS):
	export GOOS=$(helm_to_dp_rel_os) ;\
	export GOARCH=$(helm_to_dp_rel_arch) ;\
	GOPRIVATE=$(GOPRIVATE) go build -o "$@" $(HELM_TO_DP_CMD_DIR)

release: $(SCHEMA_RELEASE_BINS) $(HELM_TO_DP_RELEASE_BINS) ## Builds releasable binaries for multiple architectures. test

.PHONY: clean ## removes go artifacts and test results
clean:
	go clean -testcache
	rm -rf vendor release test/vendor build/_output/* $(VENV_NAME)

.PHONY: dependency-check
dependency-check: ## Unsupported target
	echo '"make $@" is unsupported'

.PHONY: help
help: ## Print help for each target
	@echo $(PROJECT_NAME) make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'

