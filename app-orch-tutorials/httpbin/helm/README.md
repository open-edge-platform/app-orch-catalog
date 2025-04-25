<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# httpbin-chart

Helm chart to install [httpbingo.org](https://httpbingo.org) on Kubernetes

This chart adds an Envoy proxy sidecar that sets the `Authorization` header removed by Kube API proxy.

