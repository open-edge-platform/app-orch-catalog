# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

allow_k8s_contexts('tilt-remotify')
load('ext://helm_remote', 'helm_remote')

# ------------ postgresql ------------
helm_remote(
  'postgresql',
  repo_url='https://charts.bitnami.com/bitnami',
  set=["postgresqlDatabase=remotify","postgresqlPassword=" + dbPassword]
)

k8s_resource('postgresql-postgresql', port_forwards=[5432])

# ------------ other services... ------------
