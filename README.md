<!---
  SPDX-FileCopyrightText: (C) 2022 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->

# Application Orchestration Catalog Service

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Component Test](https://github.com/open-edge-platform/app-orch-catalog/actions/workflows/component-test.yml/badge.svg)](https://github.com/open-edge-platform/app-orch-catalog/actions/workflows/component-test.yml)

## Overview

The Application Orchestration Catalog Service (App Catalog) is a cloud-native application on the Edge Orchestrator that
provides a repository for end-user **Application** definitions and **Deployment Packages** (collections of Applications)
that can be deployed to Edge Node clusters in the Open Edge Platform.

An App Catalog Application is an application definition that points to a Helm Chart (ultimately stored in a Helm Registry),
which points to zero or more Container images (ultimately stored in Container registries). A Deployment Package is a definition
of a group of Applications that are deployed together to an Edge Node cluster.

The App Catalog is also used to store [Cluster Extensions] as Application definitions that provide many of the packages
needed to secure Edge Node clusters, along with others that package commonly used cloud-native applications such as
Virtualization, Observability, etc.

The App Catalog presents a REST API to manage the lifecycle of these Deployment Packages. This API is used by the Edge
Orchestrator Web UI to present this functionality to the end-user. The App Catalog provides a YAML schema
that defines the structure of Deployment Packages, Applications, and Registry information.

[Application Orchestration Deployment] updates the App Catalog with the deployment status of deployed applications.

The App Catalog is multi-tenant capable, and the [Tenant Controller] populates the App Catalog as new multi-tenancy Projects
are created and deleted.

The App Catalog depends on the Edge Orchestrator [Foundational Platform] for many support functions such as API Gateway,
Authorization, Authentication, etc.

The overall architecture of the Application Orchestration environment is explained in the
Edge Orchestrator [Application Orchestration Developer Guide](https://literate-adventure-7vjeyem.pages.github.io/developer_guide/application_orchestration/application_orchestration_main.html).

## Get Started

Many parts of the App Catalog are described in the following sub-documents:

- [Architecture](docs/architecture.md)
- [API](docs/api.md)
- [Developer Guide](docs/developer.md)
- [Versioned Schema Migrations](docs/migrations.md)

## Develop

The App Catalog is developed in the **Go** language and is built as a Docker image, through a `Dockerfile`
in its `build` folder. The CI integration for this repository will publish the container image to the Edge Orchestrator
Release Service OCI registry upon merging to the `main` branch.

The App Catalog has a corresponding Helm chart in its [deployments](deployments) folder.
The CI integration for this repository will
publish this Helm chart to the Edge Orchestrator Release Service OCI registry upon merging to the `main` branch.
The App Catalog is deployed to the Edge Orchestrator using this Helm chart, whose lifecycle is in turn managed by
Argo CD (see [Foundational Platform]).

The App Catalog API is defined first in Protobuf format in the [api](api) folder, and then the Go code and the REST API definition
and implementation are generated from this.

The App Catalog uses a SQL database to store the Deployment Packages, Applications, and Registry information, utilizing the
ENT Object Relational Mapping framework to manage the database schema. The database technology used by default is
PostgreSQL, but other databases could be used by changing the configuration.

## Contribute
We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed
and merged into the `main` branch. We encourage you to add appropriate unit tests and end-to-end tests if
your contribution introduces a new feature. See [Contributor Guide] for information on how to contribute to the project.

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

The Application Orchestration Catalog is licensed under Apache 2.0.

[Application Orchestration Deployment]: https://github.com/open-edge-platform/app-orch-deployment
[Tenant Controller]: https://github.com/open-edge-platform/app-orch-tenant-controller
[Cluster Extensions]: https://github.com/open-edge-platform/cluster-extensions
[Foundational Platform]: https://literate-adventure-7vjeyem.pages.github.io/developer_guide/foundational_platform/foundational_platform_main.html
[Contributor Guide]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html
