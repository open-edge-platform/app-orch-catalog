<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Keycloak Helm Chart Configuration for Development

[Keycloak] is an Open Source Identity and Access Management solution for modern applications and
services.

It can also act as a federated [OpenID Connect] provider and connect to a variety of backends.
In this deployment, it is not connected to a backend and uses its own internal format
persisted to a local Postgres database.

> This chart can be deployed alongside the Application Catalog, App Deployment Manager, Web UI, or
> any other microservice that requires an OpenID provider.

## Helm Install

To install the standalone Keycloak server into a namespace, e.g., `orch-app`, use:

```shell
make keycloak-install-kind
```

To access this, use a port-forward in the cluster:

```shell
kubectl -n orch-app port-forward service/keycloak 8090:80
```

> To test it, browse to <http://localhost:8090/realms/master/.well-known/openid-configuration> to see the configuration.
>
> Verify the login details at <http://localhost:8090/realms/master/account/>.

See [Authorization](../../docs/authorization.md) for details on how to use it with the orchestrator.

> Note that the connection of the Application Catalog to Keycloak is inside the cluster for the backend services at `http://keycloak`,
> whereas the GUI connects to `http://localhost:8090`.

## Administration

The Keycloak Admin Console can be reached at <http://localhost:8090> with the credentials `admin/ChangeMeOn1stLogin!`.

## Users

Browse the Admin Console for usernames. The default password for all accounts is the same as the `admin` password above.

## Get Token Directly

To get a token directly for development purposes, use:

```shell
USER_NAME='sample-project-edge-mgr'
PASSWORD='ChangeMeOn1stLogin!'
curl --location --request POST 'http://localhost:8090/realms/master/protocol/openid-connect/token' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'grant_type=password' \
--data-urlencode 'client_id=system-client' \
--data-urlencode "username=$USER_NAME" \
--data-urlencode "password=$PASSWORD" \
--data-urlencode 'scope=openid profile email groups'
```

## Update Keycloak Config

To update the Keycloak configuration, regenerate the `values.yaml` file using the following command:

```shell
make keycloak-config-generate
```

[Keycloak]: https://www.keycloak.org/
[OpenID Connect]: https://openid.net/connect/
