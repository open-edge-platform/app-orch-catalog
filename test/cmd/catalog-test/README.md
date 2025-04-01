<!--
SPDX-FileCopyrightText: 2023-present Intel Corporation
SPDX-License-Identifier: Apache-2.0
-->
Database scale test
---------------------------

The database scale test executes a series of operations using the Catalog gRPC API and generates performance metrics
for the requests. The test runs on an existing cluster and should not affect any running payloads on the cluster.

## Running on a local developer cluster
To build and run the scale test image on a local `kind` development cluster, pull the latest app catalog sources 
and execute the command:

`make kind-run-scale-cmd`

Then to get the output of the test:

``kubectl logs -n orch-app `kubectl get pods -A | grep scale | awk '{print $2}'` ``

To remove the job:

`kubectl delete job -n orch-app catalog-scale-test`

## Running on a deployed cluster
To build and run the scale test image on deployed cluster, pull the latest app catalog sources
and execute the command:

`make coder-run-scale-cmd`

Then to get the output of the test:

`kubectl logs -n orch-app `kubectl get pods -A | grep scale | awk '{print $2}'` -f`

To remove the job:

`kubectl delete job -n orch-app catalog-scale-test`
