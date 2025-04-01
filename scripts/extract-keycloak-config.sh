#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

LOCALFILE=${LOCALFILE:-"false"}
PLATFORM_NS=${PLATFORM_NS:-"orch-platform"}
SAMPLE_ORG_ID=${SAMPLE_ORG_ID:-"11111111-1111-1111-1111-111111111111"}
SAMPLE_PROJECT_ID=${SAMPLE_PROJECT_ID:-"11111111-1111-1111-1111-222222222222"}
KEYCLOAK_HELM_VERSION=${KEYCLOAK_HELM_VERSION:-"24.4.11"}
DEBUG=${DEBUG:-"false"}
HELM_VERSION="24.4.11"

if [ "${DEBUG}" == "true" ]; then
  set -o xtrace
fi


echo "Extract new Keycloak config from a cluster. Expects sample-org and sample-project to be present."
echo "Checking cluster is accessible and Keycloak is running"
if [ "${LOCALFILE}" == "false" ]; then
	KEYCLOAK_POD=$(kubectl -n "${PLATFORM_NS}" get pod -l "app.kubernetes.io/name=keycloak" -l "app.kubernetes.io/component=keycloak" -o name | yq 'split("/") | .[1]')
	if [ -z "${KEYCLOAK_POD}" ]; then
	  echo "Keycloak pod not found in ${PLATFORM_NS} namespace" && exit 1;
	fi
	HELM_VERSION=$(kubectl -n "${PLATFORM_NS}" get pod/"${KEYCLOAK_POD}" -o yaml | yq '.metadata.labels."helm.sh/chart" | split("-") | .[1]')
	if [ "${HELM_VERSION}" != "${KEYCLOAK_HELM_VERSION}" ]; then
	  echo "Helm version ${HELM_VERSION} does not match input value ${KEYCLOAK_HELM_VERSION}. Please update input." && exit 1;
  fi
	kubectl -n "${PLATFORM_NS}" exec -it "${KEYCLOAK_POD}" -- /opt/bitnami/keycloak/bin/kc.sh export --realm master  --file /tmp/keycloak-config.json
	kubectl -n "${PLATFORM_NS}" cp "${KEYCLOAK_POD}":/tmp/keycloak-config.json /tmp/keycloak-config.json
fi;

SAMPLE_ORG_ID_OLD=$(jq -r '.users[] | select(.username == "sample-org-admin").groups[] | select(endswith("Project-Manager-Group"))  | ltrimstr("/") | rtrimstr("_Project-Manager-Group")' /tmp/keycloak-config.json);
if [ "${SAMPLE_ORG_ID_OLD}" == "" ]; then
  echo "sample-org-admin user not found or does not have group" && exit 1;
fi;
SAMPLE_PROJECT_ID_OLD=$(jq -r ".users[] | select(.username == \"sample-project-edge-mgr\").realmRoles[] | select(startswith(\"${SAMPLE_ORG_ID_OLD}\")) | split(\"_\")[1]" /tmp/keycloak-config.json);
if [ "${SAMPLE_PROJECT_ID_OLD}" == "" ]; then
  echo "sample-project-edge-mgr user not found or does not have realmRoles" && exit 1;
fi;

jq . /tmp/keycloak-config.json > /tmp/keycloak-config.json.tmp;
sed '/\"id\":/d' /tmp/keycloak-config.json.tmp > /tmp/keycloak-config.json;
jq 'del(.users[] | select(.username=="admin"))' /tmp/keycloak-config.json > /tmp/keycloak-config.json.tmp;
jq 'del(.components, .displayNameHtml, .attributes, .keycloakVersion, .userManagedAccessAllowed, .organizationsEnabled, .verifiableCredentialsEnabled, .adminPermissionsEnabled, .clientProfiles, .clientPolicies)' /tmp/keycloak-config.json.tmp > /tmp/keycloak-config.json;

echo "Replacing sample org ID ${SAMPLE_ORG_ID_OLD} with ${SAMPLE_ORG_ID}";
sed "s/${SAMPLE_ORG_ID_OLD}/${SAMPLE_ORG_ID}/g" /tmp/keycloak-config.json > /tmp/keycloak-config.json.tmp

echo "Replacing sample project ID ${SAMPLE_PROJECT_ID_OLD} with ${SAMPLE_PROJECT_ID}"
sed "s/${SAMPLE_PROJECT_ID_OLD}/${SAMPLE_PROJECT_ID}/g" /tmp/keycloak-config.json.tmp > /tmp/keycloak-config.json

yq -o json '.keycloakConfigCli.configuration."realm-master.json" = load("/tmp/keycloak-config.json")' deployments/keycloak-dev/values.yaml > /tmp/values.yaml

{ printf "# SPDX-FileCopyrightText: (C) 2025 Intel Corporation\n#\n# SPDX-License-Identifier: Apache-2.0\n\n---\n"; cat /tmp/values.yaml;} > deployments/keycloak-dev/values.yaml

rm /tmp/keycloak-config.json.tmp
rm /tmp/keycloak-config.json
rm /tmp/values.yaml

echo ""
echo "Keycloak config updated in deployments/keycloak-dev/values.yaml"
echo Try it out on a new cluster with:
echo "helm -n test-keycloak install --create-namespace keycloak oci://registry-1.docker.io/bitnamicharts/keycloak --version ${HELM_VERSION} -f deployments/keycloak-dev/values.yaml"
echo "kubectl -n test-keycloak port-forward service/keycloak 8090:80"
echo "And open your browser at http://localhost:8090 and login with admin/$(yq -r '.auth.adminPassword' deployments/keycloak-dev/values.yaml)"