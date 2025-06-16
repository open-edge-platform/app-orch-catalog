# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
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
PUBLISH_REGISTRY               ?= ghcr.io
PUBLISH_REPOSITORY             ?= edge-orch
PUBLISH_SUB_PROJ               ?= app
PUBLISH_CHART_PREFIX           ?= charts
TUTORIAL_IMAGE_NAME_PREFIX     ?= tutorial
TUTORIAL_IMAGE_NAME_SUFFIX     ?= image
TUTORIAL_VERSION               ?= ${VERSION}

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL        ?= $(shell git remote get-url $(shell git remote | head -n 1))
DOCKER_LABEL_VCS_REF        = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE     ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE    = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_EXTRA_ARGS           ?=
DOCKER_BUILD_ARGS ?= \
	${DOCKER_EXTRA_ARGS} \
	--build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
	--build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
	--build-arg org_label_build_date="${DOCKER_LABEL_BUILD_DATE}" \
	--build-arg org_label_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \

DOCKER_BUILD_COMMAND    := docker build

## CHART_NAMEs are specified in Chart.yaml
CHART_NAME					?= httpbin

## CHART_PATHs is given based on repo structure
CHART_PATH					?= "./helm"

## CHART_TEST is specified in test-connection.yaml
CHART_TEST					?= test-connection
## CHART_BUILD_DIR is given based on repo structure
CHART_BUILD_DIR				?= ./build/_output/

API_DIR                     = api
TMP_DIR                     = /tmp

HELM_REPOSITORY    		?=
HELM_REGISTRY      		?=

DOCKER_TAG_SERVER       := $(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(TUTORIAL_IMAGE_NAME_PREFIX)-server-$(TUTORIAL_IMAGE_NAME_SUFFIX):$(DOCKER_VERSION)
DOCKER_TAG_UI           := $(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)/$(TUTORIAL_IMAGE_NAME_PREFIX)-web-ui-$(TUTORIAL_IMAGE_NAME_SUFFIX):$(DOCKER_VERSION)

.PHONY: build lint test all

all: build lint test ## Runs build, lint, test stages

#### Python Targets ####

VENV_NAME = venv-env
$(VENV_NAME): tutorial-server/requirements.txt ## Create Python venv
	python3 -m venv $@ ;\
  set +u; . ./$@/bin/activate; set -u ;\
  python -m pip install --upgrade pip ;\
  python -m pip install openapi-spec-validator;\
  python -m pip install yamllint;\
  python -m pip install -r tutorial-server/requirements.txt

.PHONY: license
license: $(VENV_NAME) ## Check licensing with the reuse tool.
	. ./$</bin/activate; set -u;\
	reuse --version;\
	reuse --root . lint

.PHONY: yamllint
yamllint: $(VENV_NAME) ## Lint YAML files
	. ./$</bin/activate; set -u ;\
  yamllint --version ;\
  yamllint -s .

.PHONY: lint
lint: yamllint mdlint shelllint helmlint hadolint validate-dp ## Runs lint stage

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


trivyfsscan: ## run Trivy scan locally
	@echo "Running Trivy scan on the filesystem"
	trivy --version ;\
	trivy fs --scanners vuln,misconfig,secret -s HIGH,CRITICAL .

ENVOY_FILES := app-orch-tutorials/httpbin/helm/files/envoy-config.yaml
PHONY: envoy-lint
envoy-lint: ## Lint envoy config files
	@echo "---MAKEFILE LINT ENVOY---"
	set -e ;\
	$(foreach file,$(ENVOY_FILES),\
		docker run -v $(shell pwd):/config --rm envoyproxy/envoy:${ENVOY_VERSION} \
			--mode validate -c /config/$(file) ;\
	)
	@echo "---END MAKEFILE LINT ENVOY---"


.PHONY: test
test: mod-update opa-test go-test ## Runs test stage

PHONY: validate-dp
validate-dp: ## Validate the deployment package
	@echo "---MAKEFILE VALIDATE DP---"
	set -e ;\
	$(foreach folder,$(DP_FOLDERS),\
		$(GOCMD) run cmd/schema/schema.go validate $(folder) ;\
	)
	@echo "---END MAKEFILE VALIDATE DP---"

DOCKERFILES := $(shell find . -type f -name 'Dockerfile' | grep -v vendor/;)
.PHONY: hadolint
hadolint: ## lint Dockerfiles
	@echo "Linting Dockerfiles"
	set -e ;\
	$(foreach file,$(DOCKERFILES),\
		hadolint --ignore DL3059 $(file) ;\
    )

.PHONY: docker-build-server
docker-build-server: ## Builds the docker image for the tutorial server
	@echo "---MAKEFILE DOCKER BUILD TUTORIAL SERVER---"
	$(DOCKER_BUILD_COMMAND) $(TUTORIAL_IMAGE_NAME_PREFIX)-server -f tutorial-server/Dockerfile \
	$(PLATFORM) \
	-t $(TUTORIAL_IMAGE_NAME_PREFIX)-server:$(DOCKER_VERSION) \
	$(DOCKER_BUILD_ARGS)
	docker tag $(TUTORIAL_IMAGE_NAME_PREFIX)-server:$(DOCKER_VERSION) $(DOCKER_TAG_SERVER)

docker-build-web-ui: ## Builds the docker image for the tutorial web UI
	@echo "---MAKEFILE DOCKER BUILD TUTORIAL WEB UI---"
	$(DOCKER_BUILD_COMMAND) $(TUTORIAL_IMAGE_NAME_PREFIX)-web-ui -f tutorial-web-ui/Dockerfile \
	$(PLATFORM) \
	-t $(TUTORIAL_IMAGE_NAME_PREFIX)-web-ui:$(DOCKER_VERSION) \
	$(DOCKER_BUILD_ARGS)
	docker tag $(TUTORIAL_IMAGE_NAME_PREFIX)-web-ui:$(DOCKER_VERSION) $(DOCKER_TAG_UI)


.PHONY: docker-build
docker-build: docker-build-server docker-build-web-ui ## Builds the docker images

.PHONY: docker-push
docker-push: docker-build ## Push the docker image to ghcr.io
	docker tag $(TUTORIAL_IMAGE_NAME_PREFIX)-server:$(DOCKER_VERSION) $(DOCKER_TAG_SERVER)
	docker push $(DOCKER_TAG_SERVER)
	docker tag $(TUTORIAL_IMAGE_NAME_PREFIX)-web-ui:$(DOCKER_VERSION) $(DOCKER_TAG_WEB_UI)
	docker push $(DOCKER_TAG_WEB_UI)

docker-list: ## Print name of docker container image
	@echo "images:"
	@echo "  $(APPLICATION_CATALOG_IMAGE_NAME):"
	@echo "    name: '$(APPLICATION_CATALOG_IMAGE_NAME):$(DOCKER_VERSION)'"
	@echo "    version: '$(DOCKER_VERSION)'"
	@echo "    gitTagPrefix: 'v'"
	@echo "    buildTarget: 'docker-build'"


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

HELM_CHARTS := $(shell find . -type f -name 'Chart.yaml' -exec dirname {} \;)
.PHONY: helmlint
helmlint: ## lint helm charts
	@echo "Linting helm charts"
	set -e ;\
	$(foreach file,$(HELM_CHARTS),\
		helm lint $(file) ;\
    )

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


ghcr-publish: docker-push helm-push ## Publishes the docker image and helm chart to ghcr.io
	@echo "Published docker image and helm chart to $(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(PUBLISH_SUB_PROJ)"

.PHONY: list
list: help ## displays make targets

.PHONY: clean ## removes go artifacts and test results
clean:
	rm -rf $(VENV_NAME)

.PHONY: help
help: ## Print help for each target
	@echo $(PROJECT_NAME) make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'

