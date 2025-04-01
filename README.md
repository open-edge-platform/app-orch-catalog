<!---
  SPDX-FileCopyrightText: (C) 2022 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Application Orchestration Catalog Service

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview

The Application Orchestration Catalog Service (App Catalog) is a cloud native application on the Edge Orchestrator that
provides a repository for end user **Application** definitions and **Deployment Packages** (collections of Applications)
that can be deployed to Edge Node clusters in the Open Edge Platform.

An App Catalog Application is an application definition that points to a Helm Chart (ultimately stored in a Helm Registry),
which points to zero or more Container images (ultimately stored in Container registries). A Deployment Package is a definition
of a group of Applications that are deployed together to an Edge Node cluster.

App Catalog is also used to store [Cluster Extensions] as Application definitions that provide many of the packages
needed to secure Edge Node clusters, along with others that package commonly used Cloud Native applications such as
Virtualization, Observability etc.

App Catalog presents a REST API to manage the lifecycle of these Deployment Packages. This API is used by the Edge
Orchestrator Web UI to present this functionality to the end user. App Catalog provides a YAML schema
that defines the structure of Deployment Packages, Applications and Registry information.

[Application Orchestration Deployment] updates App Catalog with the deployment status of deployed applications.

App Catalog is multi-tenant capable and the [Tenant Controller] populates App Catalog as new multi-tenancy Projects
are created and deleted.

App Catalog depends on the Edge Orchestrator [Foundational Platform] for many support functions such as API Gateway,
Authorization, Authentication etc.

The overall architecture of the Application Orchestration environment is explained in the
Edge Orchestrator [Application Orchestration Developer Guide](https://literate-adventure-7vjeyem.pages.github.io/developer_guide/application_orchestration/application_orchestration_main.html).

## Get Started

Many parts of the App Catalog are described the following sub documents

- [API](docs/api.md)
- [CLI](docs/cli.md)
- [gRPC through Postman](docs/grpc-postman.md)
- [Architecture](docs/architecture.md)
- [Authorization](docs/authorization.md)
- [Developer Guide](docs/developer.md)
- [Deploy on KinD](kind/README.md)
- [Versioned Schema Migrations](docs/migrations.md)

## Develop

App Catalog is developed in the **Go** language and is built as a Docker image, through a `Dockerfile`
in its `build` folder. The CI integration for this repository will publish the container image to the Edge Orchestrator
Release Service OCI registry upon merge to the `main` branch.

App Catalog has a corresponding Helm chart in its `deployment` folder. The CI integration for this repository will
publish this Helm charts to the Edge Orchestrator Release Service OCI registry upon merge to `main` branch.
App Catalog is deployed to the Edge Orchestrator using this Helm chart, whose lifecycle is in turn managed by
Argo CD (see [Foundational Platform]).

App Catalog API is defined first in Protobuf format in the `api` folder and then the Go code and the REST API defintion
and implementation is generated from this.

App Catalog uses a SQL database to store the Deployment Packages, Applications and Registry information, utilizing the
ENT Object Relational Mapping framework to manage the database schema. The database technology used by default is
PostgreSQL, but other databases could be used by changing the configuration.

## Contribute

We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed
and merged into the main. We encourage you to add appropriate unit tests and e2e tests if your contribution introduces
a new feature. See the [CONTRIBUTING.md](CONTRIBUTING.md) file for more information.

Additionally, ensure the following commands are successful:

```shell
make test
make lint
make license
```

## Community and Support

To learn more about the project, its community, and governance, visit the Edge Orchestrator Community.
For support, start with Troubleshooting or contact us.

## License

Application Orchestration Catalog is licensed under Apache 2.0.

[Application Orchestration Deployment]: https://github.com/open-edge-platform/app-orch-deployment
[Tenant Controller]: https://github.com/open-edge-platform/app-orch-tenant-controller
[Cluster Extensions]: https://github.com/open-edge-platform/cluster-extensions
[Foundational Platform]: https://literate-adventure-7vjeyem.pages.github.io/developer_guide/foundational_platform/foundational_platform_main.html
