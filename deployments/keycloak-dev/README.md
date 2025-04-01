<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Keycloak Helm Chart configuration for Development

[Keycloak] is Open Source Identity and Access Management for Modern Applications and
Services.

It can also act as a Federated [OpenID Connect] provider. It can connect to a variety of backends.
In this deployment it is not connected to a backend, and just uses its own internal format
persisted to a local Postgres DB.

> This chart can be deployed alongside Application Catalog, App Deployment Manager, Web UI or
> any other microservice that requires an OpenID provider.

## Helm install

To install the standalone Keycloak server in to a namespace e.g. `orch-app` use:

```shell
make keycloak-install-kind
```

To access this use a port-forward in the cluster
```shell
kubectl -n orch-app port-forward service/keycloak 8090:80
```

> To test it, browse to http://localhost:8090/realms/master/.well-known/openid-configuration to see the configuration.
>
> Verify the login details at http://localhost:8090/realms/master/account/

See [Authorization](../../docs/authorization.md) for details of how to use with the orchestrator

> Note here that the connection of Application Catalog to keycloak is inside the cluster for the backend services at `http://keycloak`
> whereas the GUI connects to `http://localhost:8090`

## Administration
The Keycloak Admin console can be reached at http://localhost:8090 `admin/ChangeMeOn1stLogin!`

## Users
Browse the Admin console for usernames. The default password for all accounts is the same as `admin` above.

## Get Token Directly
To get a token directly for development purposes use:

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
