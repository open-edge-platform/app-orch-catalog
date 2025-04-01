# SPDX-FileCopyrightText: 2023-present Intel Corporation
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
all: uploadAllowed uploadDeniedRo uploadDeniedNone

uploadAllowed:
	@# Help: test UploadEntitiesRequest rule as all required write roles - ALLOWED
	@cat allWriteRoles.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.UploadCatalogEntitiesRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

uploadDeniedRo:
	@# Help: test UploadEntitiesRequest rule as read role - DENIED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.UploadCatalogEntitiesRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

uploadDeniedNone:
	@# Help: test UploadEntitiesRequest rule as read role - DENIED
	@cat noRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.UploadCatalogEntitiesRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -
