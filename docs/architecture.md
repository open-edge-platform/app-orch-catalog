<!---
  SPDX-FileCopyrightText: (C) 2025 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->
# Architecture

Application Orchestrator objects and their relationships are depicted in the following UML class diagram.
![image](./architecture.png)

## Design Decisions

- The project follows the [Golang Standard Project Layout] for its directory structure.
- The source of truth for the API is [protobuf] models found in the [api] directory.
- Code generation is driven by [buf], which relies on `protoc` and `protoc plugins`
found in [buf.gen.yaml](../buf.gen.yaml).
- [gRPC-Gateway] is used as a reverse proxy that acts as a RESTful/JSON
application to the client.
- The catalog uses [PostgreSQL] as the database backend. The database schema is generated using [ent].

## Security Design

### Enforcing the Principle of the Least Privilege

The Application Catalog enforces the principle of the least privilege throughout its design:

1. **Restrict Access to Others at Deployment**  
   The Application Catalog is deployed in the `orch-app` namespace and only has access to other services in that
namespace. However, it does not use most of them.  
The only services it relies on are:
    - the Malware Scanner,  
    - the Vault Service (in the `orch-platform` namespace through a service account), to a minimal level, and  
    - a PostgreSQL database external to the cluster (AWS Aurora RDS).

> Note: The Malware Scanner is disabled by default, but the code is available to enable it if needed.

1. **Restricted Access to Others**  
   The Application Catalog restricts access to its two endpoints: the gRPC interface and the REST interface.
   - Only the Application Deployment Manager is allowed to access the gRPC interface. When doing so, it only
   has access to write to the Deployment Package to update the `isDeployed` flag. It is allowed to read all resources.
   - Through the REST interface, clients must first present a valid JWT token. The "roles" listed within the
   token determine the level of access control (RBAC). These access rules are written as Open Policy Agent REGO rules
   that define which role has access to which resources.

## Multi-Tenancy

The Application Catalog is designed to support multi-tenancy, allowing multiple tenants to coexist within
the same instance of the service. Each tenant has its own isolated environment, ensuring that data and resources
are not shared between tenants. For more information, see the [Multi-Tenancy] document.

## Authentication and Authorization

The details of the Authentication and Authorization implementation are described in the [Authorization] document.

[buf]: https://docs.buf.build/introduction
[protobuf]: https://developers.google.com/protocol-buffers
[grpc-gateway]: https://grpc-ecosystem.github.io/grpc-gateway/
[ent]: https://entgo.io/
[PostgreSQL]: https://www.postgresql.org/about/
[Golang Standard Project Layout]: https://github.com/golang-standards/project-layout
[api]: ../api
[Authorization]: ./authorization.md
[Multi-Tenancy]: ./tenants.md
