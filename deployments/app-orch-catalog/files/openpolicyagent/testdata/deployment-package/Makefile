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
all: createAllowed createAllowedM2m createDeniedWriteOld createDeniedWriteOldM2m createDenied createDeniedOld getAllowed getAllowedM2m getDeniedOldM2m getDeniedOld getDeniedOnRead getDpVerAllowed listAllowed updateAllowed updateDenied

createAllowed:
	@# Help: test CreateDeploymentPackageRequest rule as write role - ALLOWED
	@cat writeRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

createAllowedM2m:
	@# Help: test CreateDeploymentPackageRequest rule as write role - ALLOWED
	@cat writeRoleM2m.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

createDeniedWriteOld:
	@# Help: test CreateDeploymentPackageRequest rule as write role - ALLOWED
	@cat writeRoleOld.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

createDeniedWriteOldM2m:
	@# Help: test CreateDeploymentPackageRequest rule as write role - ALLOWED
	@cat writeRoleOldM2m.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

createDenied:
	@# Help: test CreateDeploymentPackageRequest rule as read role - DENIED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

createDeniedOld:
	@# Help: test CreateDeploymentPackageRequest rule as read role - DENIED
	@cat readRoleOld.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.CreateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

getAllowed:
	@# Help: test GetDeploymentPackageRequest rule as read role - ALLOWED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

getDeniedOld:
	@# Help: test GetDeploymentPackageRequest rule as read role - ALLOWED
	@cat readRoleOld.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

getAllowedM2m:
	@# Help: test GetDeploymentPackageRequest rule as read role - ALLOWED
	@cat readRoleM2m.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

getDeniedOldM2m:
	@# Help: test GetDeploymentPackageRequest rule as read role - ALLOWED
	@cat readRoleOldM2m.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

getDeniedOnRead:
	@# Help: test GetDeploymentPackageRequest rule as read role - UNDEFINED
	@cat noRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -

getDpVerAllowed:
	@# Help: test GetDeploymentPackageVersionsRequest versions rule as read role - ALLOWED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.GetDeploymentPackageVersionsRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

listAllowed:
	@# Help: test ListDeploymentPackageRequest rule as read-only role - ALLOWED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.ListDeploymentPackagesRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

updateAllowed:
	@# Help: test UpdateDeploymentPackageRequest rule as read-write role - ALLOWED
	@cat writeRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.UpdateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value TRUE) | diff -u ${TMP_DIR}/opa-result -

updateDenied:
	@# Help: test UpdateDeploymentPackageRequest rule as read-only role - DENIED
	@cat readRole.json | ${OPA} eval ${PRETTY} -b ${BUNDLE} -I data.catalogv3.UpdateDeploymentPackageRequest > ${TMP_DIR}/opa-result
	@echo $(value UNDEFINED) | diff -u ${TMP_DIR}/opa-result -
