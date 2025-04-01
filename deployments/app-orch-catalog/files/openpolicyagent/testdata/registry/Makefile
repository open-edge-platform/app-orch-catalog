# SPDX-FileCopyrightText: 2024-present Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

OPA          ?= opa
BUNDLE       ?= ../..
PRETTY       ?= -f pretty
TMP_DIR      ?= /tmp
TESTDATA_DIR ?= testdata
TRUE         ?= true
UNDEFINED    ?= undefined

.PHONY: all
all: createAllowed deleteDenied deleteAllowed getDenied getAllowed

createAllowed:
	@# Help: test CreateRegistryRequest rule as write role - ALLOWED
	@cat writeRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateRegistryRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

deleteDenied:
	@# Help: test DeleteRegistryRequest rule as read-only - DENIED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.DeleteRegistryRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

deleteAllowed:
	@# Help: test DeleteRegistryRequest rule as read-only - DENIED
	@cat writeRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.DeleteRegistryRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

getDenied:
	@# Help: test GetRegistryRequest rule as restricted read role - DENIED
	@cat noReadRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetArtifactRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

getAllowed:
	@# Help: test GetArtifactRequest rule as read role - ALLOWED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetArtifactRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -
