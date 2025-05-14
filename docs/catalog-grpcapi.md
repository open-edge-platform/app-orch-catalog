<!---
  SPDX-FileCopyrightText: (C) 2022 Intel Corporation
  SPDX-License-Identifier: Apache-2.0
-->
# Protocol Documentation

<a name="top"></a>

## Table of Contents

- [catalog/v3/resources.proto](#catalog_v3_resources-proto)
  - [APIExtension](#catalog-v3-APIExtension)
  - [Application](#catalog-v3-Application)
  - [ApplicationDependency](#catalog-v3-ApplicationDependency)
  - [ApplicationReference](#catalog-v3-ApplicationReference)
  - [Artifact](#catalog-v3-Artifact)
  - [ArtifactReference](#catalog-v3-ArtifactReference)
  - [DeploymentPackage](#catalog-v3-DeploymentPackage)
  - [DeploymentPackage.DefaultNamespacesEntry](#catalog-v3-DeploymentPackage-DefaultNamespacesEntry)
  - [DeploymentProfile](#catalog-v3-DeploymentProfile)
  - [DeploymentProfile.ApplicationProfilesEntry](#catalog-v3-DeploymentProfile-ApplicationProfilesEntry)
  - [DeploymentRequirement](#catalog-v3-DeploymentRequirement)
  - [Endpoint](#catalog-v3-Endpoint)
  - [Event](#catalog-v3-Event)
  - [Namespace](#catalog-v3-Namespace)
  - [Namespace.AnnotationsEntry](#catalog-v3-Namespace-AnnotationsEntry)
  - [Namespace.LabelsEntry](#catalog-v3-Namespace-LabelsEntry)
  - [ParameterTemplate](#catalog-v3-ParameterTemplate)
  - [Profile](#catalog-v3-Profile)
  - [Registry](#catalog-v3-Registry)
  - [ResourceReference](#catalog-v3-ResourceReference)
  - [UIExtension](#catalog-v3-UIExtension)
  - [Upload](#catalog-v3-Upload)
  
  - [Kind](#catalog-v3-Kind)
  
- [catalog/v3/service.proto](#catalog_v3_service-proto)
  - [CreateApplicationRequest](#catalog-v3-CreateApplicationRequest)
  - [CreateApplicationResponse](#catalog-v3-CreateApplicationResponse)
  - [CreateArtifactRequest](#catalog-v3-CreateArtifactRequest)
  - [CreateArtifactResponse](#catalog-v3-CreateArtifactResponse)
  - [CreateDeploymentPackageRequest](#catalog-v3-CreateDeploymentPackageRequest)
  - [CreateDeploymentPackageResponse](#catalog-v3-CreateDeploymentPackageResponse)
  - [CreateRegistryRequest](#catalog-v3-CreateRegistryRequest)
  - [CreateRegistryResponse](#catalog-v3-CreateRegistryResponse)
  - [DeleteApplicationRequest](#catalog-v3-DeleteApplicationRequest)
  - [DeleteArtifactRequest](#catalog-v3-DeleteArtifactRequest)
  - [DeleteDeploymentPackageRequest](#catalog-v3-DeleteDeploymentPackageRequest)
  - [DeleteRegistryRequest](#catalog-v3-DeleteRegistryRequest)
  - [GetApplicationReferenceCountRequest](#catalog-v3-GetApplicationReferenceCountRequest)
  - [GetApplicationReferenceCountResponse](#catalog-v3-GetApplicationReferenceCountResponse)
  - [GetApplicationRequest](#catalog-v3-GetApplicationRequest)
  - [GetApplicationResponse](#catalog-v3-GetApplicationResponse)
  - [GetApplicationVersionsRequest](#catalog-v3-GetApplicationVersionsRequest)
  - [GetApplicationVersionsResponse](#catalog-v3-GetApplicationVersionsResponse)
  - [GetArtifactRequest](#catalog-v3-GetArtifactRequest)
  - [GetArtifactResponse](#catalog-v3-GetArtifactResponse)
  - [GetDeploymentPackageRequest](#catalog-v3-GetDeploymentPackageRequest)
  - [GetDeploymentPackageResponse](#catalog-v3-GetDeploymentPackageResponse)
  - [GetDeploymentPackageVersionsRequest](#catalog-v3-GetDeploymentPackageVersionsRequest)
  - [GetDeploymentPackageVersionsResponse](#catalog-v3-GetDeploymentPackageVersionsResponse)
  - [GetRegistryRequest](#catalog-v3-GetRegistryRequest)
  - [GetRegistryResponse](#catalog-v3-GetRegistryResponse)
  - [ImportRequest](#catalog-v3-ImportRequest)
  - [ImportResponse](#catalog-v3-ImportResponse)
  - [ListApplicationsRequest](#catalog-v3-ListApplicationsRequest)
  - [ListApplicationsResponse](#catalog-v3-ListApplicationsResponse)
  - [ListArtifactsRequest](#catalog-v3-ListArtifactsRequest)
  - [ListArtifactsResponse](#catalog-v3-ListArtifactsResponse)
  - [ListDeploymentPackagesRequest](#catalog-v3-ListDeploymentPackagesRequest)
  - [ListDeploymentPackagesResponse](#catalog-v3-ListDeploymentPackagesResponse)
  - [ListRegistriesRequest](#catalog-v3-ListRegistriesRequest)
  - [ListRegistriesResponse](#catalog-v3-ListRegistriesResponse)
  - [UpdateApplicationRequest](#catalog-v3-UpdateApplicationRequest)
  - [UpdateArtifactRequest](#catalog-v3-UpdateArtifactRequest)
  - [UpdateDeploymentPackageRequest](#catalog-v3-UpdateDeploymentPackageRequest)
  - [UpdateRegistryRequest](#catalog-v3-UpdateRegistryRequest)
  - [UploadCatalogEntitiesRequest](#catalog-v3-UploadCatalogEntitiesRequest)
  - [UploadCatalogEntitiesResponse](#catalog-v3-UploadCatalogEntitiesResponse)
  - [UploadMultipleCatalogEntitiesResponse](#catalog-v3-UploadMultipleCatalogEntitiesResponse)
  - [WatchApplicationsRequest](#catalog-v3-WatchApplicationsRequest)
  - [WatchApplicationsResponse](#catalog-v3-WatchApplicationsResponse)
  - [WatchArtifactsRequest](#catalog-v3-WatchArtifactsRequest)
  - [WatchArtifactsResponse](#catalog-v3-WatchArtifactsResponse)
  - [WatchDeploymentPackagesRequest](#catalog-v3-WatchDeploymentPackagesRequest)
  - [WatchDeploymentPackagesResponse](#catalog-v3-WatchDeploymentPackagesResponse)
  - [WatchRegistriesRequest](#catalog-v3-WatchRegistriesRequest)
  - [WatchRegistriesResponse](#catalog-v3-WatchRegistriesResponse)
  
  - [CatalogService](#catalog-v3-CatalogService)
  
- [Scalar Value Types](#scalar-value-types)

<a name="catalog_v3_resources-proto"></a>

## catalog/v3/resources.proto

<a name="catalog-v3-APIExtension"></a>

### APIExtension

APIExtensions represents some form of an extension to the external API provided by deployment package.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name is a human-readable unique identifier for the API extension and must be unique for all extensions of a given deployment package. |
| version | [string](#string) |  | Version of the API extension. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the API extension. When specified, it must be unique among all extensions of a given deployment package. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the API extension. Displayed on user interfaces. |
| endpoints | [Endpoint](#catalog-v3-Endpoint) | repeated | One or more API endpoints provided by the API extension. |
| ui_extension | [UIExtension](#catalog-v3-UIExtension) |  | Additional information specific to UI extensions. |

<a name="catalog-v3-Application"></a>

### Application

Application represents a Helm chart that can be deployed to one or more Kubernetes pods.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name is a human readable unique identifier for the application and must be unique for all applications of a given project. Used in network URIs. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the application. When specified, it must be unique among all applications within a project. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the application. Displayed on user interfaces. |
| version | [string](#string) |  | Version of the application. Used in combination with the name to identify a unique application within a project. |
| kind | [Kind](#catalog-v3-Kind) |  | Field designating whether the application is a system add-on, system extension, or a normal application. |
| chart_name | [string](#string) |  | Helm chart name. |
| chart_version | [string](#string) |  | Helm chart version. |
| helm_registry_name | [string](#string) |  | ID of the project's registry where the Helm chart of the application is available for download. |
| profiles | [Profile](#catalog-v3-Profile) | repeated | Set of profiles that can be used when deploying the application. |
| default_profile_name | [string](#string) |  | Name of the profile to be used by default when deploying this application. If at least one profile is available, this field must be set. |
| image_registry_name | [string](#string) |  | ID of the project's registry where the Docker image of the application is available for download. |
| ignored_resources | [ResourceReference](#catalog-v3-ResourceReference) | repeated | List of Kubernetes resources that must be ignored during the application deployment. |
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The creation time of the application. |
| update_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last update time of the application. |

<a name="catalog-v3-ApplicationDependency"></a>

### ApplicationDependency

ApplicationDependency represents the dependency of one application on another within the context of a deployment package.
This dependency is specified as the name of the application that has the dependency, and the name of the application
that is the dependency.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the application that has the dependency on the other. |
| requires | [string](#string) |  | Name of the application that is required by the other. |

<a name="catalog-v3-ApplicationReference"></a>

### ApplicationReference

ApplicationReference represents a reference to an application by its name and its version.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the referenced application. |
| version | [string](#string) |  | Version of the referenced application. |

<a name="catalog-v3-Artifact"></a>

### Artifact

Artifact represents a binary artifact that can be used for various purposes, e.g. icon or thumbnail for UI display, or
auxiliary artifacts for integration with various platform services such as Grafana dashboard and similar. An artifact may be
used by multiple deployment packages.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name is a human-readable unique identifier for the artifact and must be unique for all artifacts within a project. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the artifact. When specified, it must be unique among all artifacts within a project. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the artifact. Displayed on user interfaces. |
| mime_type | [string](#string) |  | Artifact's MIME type. Only text/plain, application/json, application/yaml, image/png, and image/jpeg are allowed at this time.

MIME types are defined and standardized in IETF's RFC 6838. |
| artifact | [bytes](#bytes) |  | Raw byte content of the artifact encoded as base64. The limits refer to the number of raw bytes. |
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The creation time of the artifact. |
| update_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last update time of the artifact. |

<a name="catalog-v3-ArtifactReference"></a>

### ArtifactReference

ArtifactReference serves as a reference to an artifact, together with the artifact's purpose within a deployment package.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the artifact. |
| purpose | [string](#string) |  | Purpose of the artifact, e.g. icon, thumbnail, Grafana dashboard, etc. |

<a name="catalog-v3-DeploymentPackage"></a>

### DeploymentPackage

DeploymentPackage represents a collection of applications (referenced by their name and a version) that are
deployed together. The package can define one or more deployment profiles that specify the individual application
profiles to be used when deploying each application. If applications need to be deployed in a particular order, the
package can also define any startup dependencies between its constituent applications as a set of dependency graph edges.

The deployment package can also refer to a set of artifacts used for miscellaneous purposes,
e.g. a thumbnail, icon, or a Grafana extension.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name is a human-readable unique identifier for the deployment package and must be unique for all packages of a given project. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the deployment package. When specified, it must be unique among all packages within a project. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the deployment package. Displayed on user interfaces. |
| version | [string](#string) |  | Version of the deployment package. |
| kind | [Kind](#catalog-v3-Kind) |  | Field designating whether the deployment package is a system add-on, system extension, or a normal package. |
| application_references | [ApplicationReference](#catalog-v3-ApplicationReference) | repeated | List of applications comprising this deployment package. Expressed as (name, version) pairs. |
| is_deployed | [bool](#bool) |  | Flag indicating whether the deployment package has been deployed. The mutability of the deployment package entity can be limited when this flag is true. For example, one may not be able to update when an application is removed from a package after it has been marked as deployed. |
| is_visible | [bool](#bool) |  | Flag indicating whether the deployment package is visible in the UI. Some deployment packages can be classified as auxiliary platform extensions and therefore are to be deployed indirectly only when specified as deployment requirements, rather than directly by the platform operator. |
| profiles | [DeploymentProfile](#catalog-v3-DeploymentProfile) | repeated | Set of deployment profiles to choose from when deploying this package. |
| default_profile_name | [string](#string) |  | Name of the default deployment profile to be used by default when deploying this package. |
| application_dependencies | [ApplicationDependency](#catalog-v3-ApplicationDependency) | repeated | Optional set of application deployment dependencies, expressed as (name, requires) pairs of edges in the deployment order dependency graph. |
| extensions | [APIExtension](#catalog-v3-APIExtension) | repeated | Optional list of API and UI extensions. |
| artifacts | [ArtifactReference](#catalog-v3-ArtifactReference) | repeated | Optional list of artifacts required for displaying or deploying this package. For example, icon or thumbnail artifacts can be used by the UI; Grafana\* dashboard definitions can be used by the deployment manager. |
| default_namespaces | [DeploymentPackage.DefaultNamespacesEntry](#catalog-v3-DeploymentPackage-DefaultNamespacesEntry) | repeated | Optional map of application-to-namespace bindings to be used as a default when deploying the applications that comprise the package. If a namespace is not defined in the set of "namespaces" in this Deployment Package, it will be inferred that it is a simple namespace with no predefined labels or annotations. |
| forbids_multiple_deployments | [bool](#bool) |  | Optional flag indicating whether multiple deployments of this package are forbidden within the same realm. |
| namespaces | [Namespace](#catalog-v3-Namespace) | repeated | Namespace definitions to be created before resources are deployed. This allows complex namespaces to be defined with predefined labels and annotations. If not defined, simple namespaces will be created as needed. |
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The creation time of the deployment package. |
| update_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last update time of the deployment package. |

<a name="catalog-v3-DeploymentPackage-DefaultNamespacesEntry"></a>

### DeploymentPackage.DefaultNamespacesEntry

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |

<a name="catalog-v3-DeploymentProfile"></a>

### DeploymentProfile

DeploymentProfile specifies which application profiles will be used for deployment of which applications.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name is a human-readable unique identifier for the profile and must be unique for all profiles of a given deployment package. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the registry. When specified, it must be unique among all profiles of a given package. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the deployment profile. Displayed on user interfaces. |
| application_profiles | [DeploymentProfile.ApplicationProfilesEntry](#catalog-v3-DeploymentProfile-ApplicationProfilesEntry) | repeated | Application profiles map application names to the names of its profile, to be used when deploying the application as part of the deployment package together with the deployment profile. |
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The creation time of the deployment profile. |
| update_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last update time of the deployment profile. |

<a name="catalog-v3-DeploymentProfile-ApplicationProfilesEntry"></a>

### DeploymentProfile.ApplicationProfilesEntry

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |

<a name="catalog-v3-DeploymentRequirement"></a>

### DeploymentRequirement

DeploymentRequirement is a reference to the deployment package that must be deployed first,
as a requirement for an application to be deployed.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name of the required deployment package. |
| version | [string](#string) |  | Version of the required deployment package. |
| deployment_profile_name | [string](#string) |  | Optional name of the deployment profile to be used. When not provided, the default deployment profile will be used. |

<a name="catalog-v3-Endpoint"></a>

### Endpoint

Endpoint represents an application service endpoint.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| service_name | [string](#string) |  | The name of the service hosted by the endpoint. |
| external_path | [string](#string) |  | Externally accessible path to the endpoint. |
| internal_path | [string](#string) |  | Internally accessible path to the endpoint. |
| scheme | [string](#string) |  | Protocol scheme provided by the endpoint. |
| auth_type | [string](#string) |  | Authentication type expected by the endpoint. |
| app_name | [string](#string) |  | The name of the application providing this endpoint. |

<a name="catalog-v3-Event"></a>

### Event

Event message carries the event type detected by the catalog service during the invocation of
the "watch" RPC.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [string](#string) |  | Type field specifies whether an entity was created, updated, or deleted. The replayed type is used to annotate entities during the replay phase of the watch RPC. |
| project_id | [string](#string) |  | ID of the project to which the subject belongs. |

<a name="catalog-v3-Namespace"></a>

### Namespace

Namespace represents a complex namespace definition with predefined labels and annotations.
They are created before any other resources in the deployment.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | namespace names must be valid RFC 1123 DNS labels. Avoid creating namespaces with the prefix `kube-`, since it is reserved for Kubernetes\* system namespaces. Avoid `default` - will already exist |
| labels | [Namespace.LabelsEntry](#catalog-v3-Namespace-LabelsEntry) | repeated |  |
| annotations | [Namespace.AnnotationsEntry](#catalog-v3-Namespace-AnnotationsEntry) | repeated |  |

<a name="catalog-v3-Namespace-AnnotationsEntry"></a>

### Namespace.AnnotationsEntry

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |

<a name="catalog-v3-Namespace-LabelsEntry"></a>

### Namespace.LabelsEntry

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |

<a name="catalog-v3-ParameterTemplate"></a>

### ParameterTemplate

ParameterTemplate describes override values for Helm chart values

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Human-readable name for the parameter template. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the template. It is used for display purposes on user interfaces. |
| default | [string](#string) |  | Default value for the parameter. |
| type | [string](#string) |  | Type of parameter: string, number, or boolean. |
| validator | [string](#string) |  | Optional validator for the parameter. Usage TBD. |
| suggested_values | [string](#string) | repeated | List of suggested values to use, to override the default value. |
| secret | [bool](#bool) |  | Optional secret flag for the parameter. |
| mandatory | [bool](#bool) |  | Optional mandatory flag for the parameter. |

<a name="catalog-v3-Profile"></a>

### Profile

Profile is a set of configuration values for customizing application deployment.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Human-readable name for the profile. Unique among all profiles of the same application. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the profile. When specified, it must be unique among all profiles of a given application. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the profile. Displayed on user interfaces. |
| chart_values | [string](#string) |  | Raw byte value containing the chart values as raw YAML bytes. |
| parameter_templates | [ParameterTemplate](#catalog-v3-ParameterTemplate) | repeated | Parameter templates available for this profile. |
| deployment_requirement | [DeploymentRequirement](#catalog-v3-DeploymentRequirement) | repeated | List of deployment requirements for this profile. |
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The creation time of the profile. |
| update_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last update time of the profile. |

<a name="catalog-v3-Registry"></a>

### Registry

Registry represents a repository from which various artifacts, such as application Docker\* images or Helm\* charts
can be retrieved. As such, the registry entity holds information used for finding and accessing the represented repository.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Name is a human-readable unique identifier for the registry and must be unique for all registries of a given project. |
| display_name | [string](#string) |  | Display name is an optional human-readable name for the registry. When specified, it must be unique among all registries within a project. It is used for display purposes on user interfaces. |
| description | [string](#string) |  | Description of the registry. Displayed on user interfaces. |
| root_url | [string](#string) |  | Root URL for retrieving artifacts, e.g. Docker images and Helm charts, from the registry. |
| username | [string](#string) |  | Optional username for accessing the registry. |
| auth_token | [string](#string) |  | Optional authentication token or password for accessing the registry. |
| type | [string](#string) |  | Type indicates whether the registry holds Docker images or Helm charts; defaults to Helm charts. |
| cacerts | [string](#string) |  | Optional CA certificates for accessing the registry using secure channels, such as HTTPS. |
| api_type | [string](#string) |  | Optional type of the API used to obtain inventory of the articles hosted by the registry. |
| inventory_url | [string](#string) |  | Optional URL of the API for accessing inventory of artifacts hosted by the registry. |
| create_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The creation time of the registry. |
| update_time | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  | The last update time of the registry. |

<a name="catalog-v3-ResourceReference"></a>

### ResourceReference

ResourceReference represents a Kubernetes resource identifier.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Kubernetes resource name. |
| kind | [string](#string) |  | Kubernetes resource kind, e.g. ConfigMap. |
| namespace | [string](#string) |  | Kubernetes namespace where the ignored resource resides. When empty, the application namespace will be used. |

<a name="catalog-v3-UIExtension"></a>

### UIExtension

UIExtension is an augmentation of an API extension.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| label | [string](#string) |  | Label is a human readable text used for display in the main UI dashboard |
| service_name | [string](#string) |  | The name of the API extension endpoint. |
| description | [string](#string) |  | Description of the API extension, used on the main UI dashboard. |
| file_name | [string](#string) |  | The name of the main file to load this specific UI extension. |
| app_name | [string](#string) |  | The name of the application corresponding to this UI extension. |
| module_name | [string](#string) |  | Name of the application module to be loaded. |

<a name="catalog-v3-Upload"></a>

### Upload

Upload represents a single file-upload record.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| file_name | [string](#string) |  | Name of the file being uploaded. |
| artifact | [bytes](#bytes) |  | Raw bytes content of the file being uploaded. |

 <!-- end messages -->

<a name="google-protobuf-Empty"></a>

### Empty

Empty is an empty Protobuf message.

<a name="google-protobuf-Timestamp"></a>

### Timestamp

Timestamp is a Protobuf message containing a timestamp.

<a name="catalog-v3-Kind"></a>

### Kind

Kind designation for applications and packages, normal (unspecified), extension, or addon.

| Name | Number | Description |
| ---- | ------ | ----------- |
| KIND_UNSPECIFIED | 0 |  |
| KIND_NORMAL | 1 |  |
| KIND_EXTENSION | 2 |  |
| KIND_ADDON | 3 |  |

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

<a name="catalog_v3_service-proto"></a>

## catalog/v3/service.proto

<a name="catalog-v3-CreateApplicationRequest"></a>

### CreateApplicationRequest

Request message for the CreateApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application | [Application](#catalog-v3-Application) |  | The registry to create. |

<a name="catalog-v3-CreateApplicationResponse"></a>

### CreateApplicationResponse

Response message for the CreateApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application | [Application](#catalog-v3-Application) |  | The application created. |

<a name="catalog-v3-CreateArtifactRequest"></a>

### CreateArtifactRequest

Request message for the CreateArtifact method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact | [Artifact](#catalog-v3-Artifact) |  | The artifact to create. |

<a name="catalog-v3-CreateArtifactResponse"></a>

### CreateArtifactResponse

Response message for the CreateArtifact method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact | [Artifact](#catalog-v3-Artifact) |  | The artifact created. |

<a name="catalog-v3-CreateDeploymentPackageRequest"></a>

### CreateDeploymentPackageRequest

Request message for the CreateDeploymentPackage method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package | [DeploymentPackage](#catalog-v3-DeploymentPackage) |  | The deployment package to create. |

<a name="catalog-v3-CreateDeploymentPackageResponse"></a>

### CreateDeploymentPackageResponse

Response message for the CreateDeploymentPackage method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package | [DeploymentPackage](#catalog-v3-DeploymentPackage) |  | The deployment package created. |

<a name="catalog-v3-CreateRegistryRequest"></a>

### CreateRegistryRequest

Request message for the CreateRegistry method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry | [Registry](#catalog-v3-Registry) |  | The registry to create. |

<a name="catalog-v3-CreateRegistryResponse"></a>

### CreateRegistryResponse

Response message for the CreateRegistry method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry | [Registry](#catalog-v3-Registry) |  | The created registry. |

<a name="catalog-v3-DeleteApplicationRequest"></a>

### DeleteApplicationRequest

Request message for the DeleteApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_name | [string](#string) |  | Name of the application. |
| version | [string](#string) |  | Version of the application. |

<a name="catalog-v3-DeleteArtifactRequest"></a>

### DeleteArtifactRequest

Request message for the DeleteArtifact method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact_name | [string](#string) |  | Name of the artifact. |

<a name="catalog-v3-DeleteDeploymentPackageRequest"></a>

### DeleteDeploymentPackageRequest

Request message for DeleteDeploymentPackage.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package_name | [string](#string) |  | Name of the DeploymentPackage. |
| version | [string](#string) |  | Version of the DeploymentPackage. |

<a name="catalog-v3-DeleteRegistryRequest"></a>

### DeleteRegistryRequest

Request message for the DeleteRegistry method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry_name | [string](#string) |  | Name of the registry. |

<a name="catalog-v3-GetApplicationReferenceCountRequest"></a>

### GetApplicationReferenceCountRequest

Request message for the GetApplicationReferenceCount method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_name | [string](#string) |  | Name of the application. |
| version | [string](#string) |  | Version of the application. |

<a name="catalog-v3-GetApplicationReferenceCountResponse"></a>

### GetApplicationReferenceCountResponse

Response message for the GetApplicationReferenceCount method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reference_count | [uint32](#uint32) |  |  |

<a name="catalog-v3-GetApplicationRequest"></a>

### GetApplicationRequest

Request message for the GetApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_name | [string](#string) |  | Name of the application. |
| version | [string](#string) |  | Version of the application. |

<a name="catalog-v3-GetApplicationResponse"></a>

### GetApplicationResponse

Response message for the GetApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application | [Application](#catalog-v3-Application) |  | The requested application. |

<a name="catalog-v3-GetApplicationVersionsRequest"></a>

### GetApplicationVersionsRequest

Request message for the GetApplicationVersions method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_name | [string](#string) |  | Name of the application. |

<a name="catalog-v3-GetApplicationVersionsResponse"></a>

### GetApplicationVersionsResponse

Response message for the GetApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application | [Application](#catalog-v3-Application) | repeated | A list of applications with the same project and name.

TODO rename to 'applications' |

<a name="catalog-v3-GetArtifactRequest"></a>

### GetArtifactRequest

Request message for the GetArtifact method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact_name | [string](#string) |  | Name of the artifact. |

<a name="catalog-v3-GetArtifactResponse"></a>

### GetArtifactResponse

Response message for the GetArtifact method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact | [Artifact](#catalog-v3-Artifact) |  | The requested artifact. |

<a name="catalog-v3-GetDeploymentPackageRequest"></a>

### GetDeploymentPackageRequest

Request message for the GetDeploymentPackage method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package_name | [string](#string) |  | Name of the DeploymentPackage. |
| version | [string](#string) |  | Version of the DeploymentPackage. |

<a name="catalog-v3-GetDeploymentPackageResponse"></a>

### GetDeploymentPackageResponse

Response message for the GetDeploymentPackage method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package | [DeploymentPackage](#catalog-v3-DeploymentPackage) |  | The DeploymentPackage requested. |

<a name="catalog-v3-GetDeploymentPackageVersionsRequest"></a>

### GetDeploymentPackageVersionsRequest

Request message for the GetDeploymentPackageVersions method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package_name | [string](#string) |  | Name of the DeploymentPackage. |

<a name="catalog-v3-GetDeploymentPackageVersionsResponse"></a>

### GetDeploymentPackageVersionsResponse

Response message for the GetDeploymentPackageVersions method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_packages | [DeploymentPackage](#catalog-v3-DeploymentPackage) | repeated | A list of DeploymentPackages with the same project and name. |

<a name="catalog-v3-GetRegistryRequest"></a>

### GetRegistryRequest

Request message for the GetRegistry method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry_name | [string](#string) |  | Name of the registry. |
| show_sensitive_info | [bool](#bool) |  | Request that sensitive information, such as username, auth_token, and CA certificates are included in the response. |

<a name="catalog-v3-GetRegistryResponse"></a>

### GetRegistryResponse

Response message for the GetRegistry method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry | [Registry](#catalog-v3-Registry) |  |  |

<a name="catalog-v3-ImportRequest"></a>

### ImportRequest

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| url | [string](#string) |  | Required URL of Helm Chart to import |
| username | [string](#string) |  | Optional username for downloading from the URL |
| auth_token | [string](#string) |  | Optional authentication token or password for downloading from the URL |
| chart_values | [string](#string) |  | Optional raw byte value containing the chart values as raw YAML bytes. |
| include_auth | [bool](#bool) |  | If true and a username/auth_token is specified then they will be included in the generated Registry object. |
| generate_default_values | [bool](#bool) |  | If true and chart_values is not set, then the values.yaml will be extracted and used to generate default profile values. |
| generate_default_parameters | [bool](#bool) |  | Generates default parameters from the values, from chart_values or from generate_default_values as appropriate. |

<a name="catalog-v3-ImportResponse"></a>

### ImportResponse

Response message for the Import method

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| error_messages | [string](#string) | repeated | Any error messages encountered either during chart parsing or entity creation or update. |

<a name="catalog-v3-ListApplicationsRequest"></a>

### ListApplicationsRequest

Request message for the ListApplications method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Names the field to be used for ordering the returned results. |
| filter | [string](#string) |  | Expression to use for filtering the results. |
| page_size | [int32](#int32) |  | Maximum number of items to return. |
| offset | [int32](#int32) |  | Index of the first item to return. |
| kinds | [Kind](#catalog-v3-Kind) | repeated | List of application kinds to be returned; empty list means all kinds. |

<a name="catalog-v3-ListApplicationsResponse"></a>

### ListApplicationsResponse

Response message for the ListApplications method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| applications | [Application](#catalog-v3-Application) | repeated | A list of applications. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |

<a name="catalog-v3-ListArtifactsRequest"></a>

### ListArtifactsRequest

Request message for the ListArtifacts method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Names the field to be used for ordering the returned results. |
| filter | [string](#string) |  | Expression to use for filtering the results. |
| page_size | [int32](#int32) |  | Maximum number of items to return. |
| offset | [int32](#int32) |  | Index of the first item to return. |

<a name="catalog-v3-ListArtifactsResponse"></a>

### ListArtifactsResponse

Response message for the ListArtifacts method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifacts | [Artifact](#catalog-v3-Artifact) | repeated | A list of artifacts. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |

<a name="catalog-v3-ListDeploymentPackagesRequest"></a>

### ListDeploymentPackagesRequest

Request message for the ListDeploymentPackages method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Names the field to be used for ordering the returned results. |
| filter | [string](#string) |  | Expression to use for filtering the results. |
| page_size | [int32](#int32) |  | Maximum number of items to return. |
| offset | [int32](#int32) |  | Index of the first item to return. |
| kinds | [Kind](#catalog-v3-Kind) | repeated | List of deployment package kinds to be returned; empty list means all kinds. |

<a name="catalog-v3-ListDeploymentPackagesResponse"></a>

### ListDeploymentPackagesResponse

Response message for the ListDeploymentPackages method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_packages | [DeploymentPackage](#catalog-v3-DeploymentPackage) | repeated | A list of DeploymentPackages. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |

<a name="catalog-v3-ListRegistriesRequest"></a>

### ListRegistriesRequest

Request message for the ListRegistries method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| order_by | [string](#string) |  | Names the field to be used for ordering the returned results. |
| filter | [string](#string) |  | Expression to use for filtering the results. |
| page_size | [int32](#int32) |  | Maximum number of items to return. |
| offset | [int32](#int32) |  | Index of the first item to return. |
| show_sensitive_info | [bool](#bool) |  | Request that sensitive information, such as username, auth_token, and CA certificates are included in the response. |

<a name="catalog-v3-ListRegistriesResponse"></a>

### ListRegistriesResponse

Response message for the ListRegistries method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registries | [Registry](#catalog-v3-Registry) | repeated | A list of registries. |
| total_elements | [int32](#int32) |  | Count of items in the entire list, regardless of pagination. |

<a name="catalog-v3-UpdateApplicationRequest"></a>

### UpdateApplicationRequest

Request message for the UpdateApplication method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| application_name | [string](#string) |  | Name of the application. |
| version | [string](#string) |  | Version of the application. |
| application | [Application](#catalog-v3-Application) |  | The application update. |

<a name="catalog-v3-UpdateArtifactRequest"></a>

### UpdateArtifactRequest

Request message for the UpdateArtifact method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| artifact_name | [string](#string) |  | Name of the artifact. |
| artifact | [Artifact](#catalog-v3-Artifact) |  | The artifact update. |

<a name="catalog-v3-UpdateDeploymentPackageRequest"></a>

### UpdateDeploymentPackageRequest

Request message for the UpdateDeploymentPackage method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| deployment_package_name | [string](#string) |  | Name of the DeploymentPackage. |
| version | [string](#string) |  | Version of the DeploymentPackage. |
| deployment_package | [DeploymentPackage](#catalog-v3-DeploymentPackage) |  | The DeploymentPackage update. |

<a name="catalog-v3-UpdateRegistryRequest"></a>

### UpdateRegistryRequest

Request message for the UpdateRegistry method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| registry_name | [string](#string) |  | Name of the Registry. |
| registry | [Registry](#catalog-v3-Registry) |  | The Registry update. |

<a name="catalog-v3-UploadCatalogEntitiesRequest"></a>

### UploadCatalogEntitiesRequest

Request message for the UploadCatalogItems method

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [string](#string) |  | First upload request in the batch must not specify session ID. Subsequent upload requests must copy the session ID from the previously issued response. |
| upload_number | [uint32](#uint32) |  | Deprecated: Upload number must increase sequentially, starting with 1. |
| last_upload | [bool](#bool) |  | Must be set to 'true' to perform load of all entity files uploaded as part of this session. |
| upload | [Upload](#catalog-v3-Upload) |  | Upload record containing the file name and file contents being uploaded. |

<a name="catalog-v3-UploadCatalogEntitiesResponse"></a>

### UploadCatalogEntitiesResponse

Response message for the UploadCatalogItems method

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| session_id | [string](#string) |  | Session ID, generated by the server after the first upload request has been processed. |
| upload_number | [uint32](#uint32) |  | Deprecated: Next expected upload number or total number of uploads on the last upload request. |
| error_messages | [string](#string) | repeated | Any error messages encountered either during YAML parsing or entity creation or update. |

<a name="catalog-v3-UploadMultipleCatalogEntitiesResponse"></a>

### UploadMultipleCatalogEntitiesResponse

Response message when multiple files are uploaded at the same time through rest-proxy.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| responses | [UploadCatalogEntitiesResponse](#catalog-v3-UploadCatalogEntitiesResponse) | repeated |  |

<a name="catalog-v3-WatchApplicationsRequest"></a>

### WatchApplicationsRequest

Request message for the WatchApplications method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | ID of the project. |
| no_replay | [bool](#bool) |  | Indicates whether replay of existing entities will be performed. |
| kinds | [Kind](#catalog-v3-Kind) | repeated | Application kinds to be watched; empty list means all kinds. |

<a name="catalog-v3-WatchApplicationsResponse"></a>

### WatchApplicationsResponse

Response message for the WatchApplications method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#catalog-v3-Event) |  |  |
| application | [Application](#catalog-v3-Application) |  |  |

<a name="catalog-v3-WatchArtifactsRequest"></a>

### WatchArtifactsRequest

Request message for the WatchArtifacts method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | ID of the project. |
| no_replay | [bool](#bool) |  | Indicates whether replay of existing entities will be performed. |

<a name="catalog-v3-WatchArtifactsResponse"></a>

### WatchArtifactsResponse

Response message for the WatchArtifacts method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#catalog-v3-Event) |  |  |
| artifact | [Artifact](#catalog-v3-Artifact) |  |  |

<a name="catalog-v3-WatchDeploymentPackagesRequest"></a>

### WatchDeploymentPackagesRequest

Request message for the WatchDeploymentPackages method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | ID of the project. |
| no_replay | [bool](#bool) |  | Indicates whether replay of existing entities will be performed. |
| kinds | [Kind](#catalog-v3-Kind) | repeated | Deployment package kinds to be watched; empty list means all kinds. |

<a name="catalog-v3-WatchDeploymentPackagesResponse"></a>

### WatchDeploymentPackagesResponse

Response message for the WatchDeploymentPackages method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#catalog-v3-Event) |  |  |
| deployment_package | [DeploymentPackage](#catalog-v3-DeploymentPackage) |  |  |

<a name="catalog-v3-WatchRegistriesRequest"></a>

### WatchRegistriesRequest

Request message for the WatchRegistries method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| project_id | [string](#string) |  | ID of the project. |
| no_replay | [bool](#bool) |  | Indicates whether replay of existing entities will be performed. |
| show_sensitive_info | [bool](#bool) |  | Request that sensitive information, such as username, auth_token, and CA certificates are included in the response. |

<a name="catalog-v3-WatchRegistriesResponse"></a>

### WatchRegistriesResponse

Response message for the WatchRegistries method.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| event | [Event](#catalog-v3-Event) |  |  |
| registry | [Registry](#catalog-v3-Registry) |  |  |

 <!-- end messages -->

<a name="google-protobuf-Empty"></a>

### Empty

Empty is an empty Protobuf message.

<a name="google-protobuf-Timestamp"></a>

### Timestamp

Timestamp is a Protobuf message containing a timestamp.

 <!-- end enums -->

 <!-- end HasExtensions -->

<a name="catalog-v3-CatalogService"></a>

### CatalogService

CatalogService provides API to manage the inventory of applications, deployment packages, and other resources related
to deployment of applications at the network edge.

The principal resources managed by the application catalog service are as follows:
- [Application](catalog.v3.Application) represents a Helm\* chart that can be deployed to one or more Kubernetes\* pods.

- [DeploymentPackage](catalog.v3.DeploymentPackage) represents a collection of applications (referenced by their name and a version) that are
   deployed together. The package can define one or more deployment profiles that specify the individual application
   profiles to be used when deploying each application. If applications need to be deployed in a particular order, the
   package can also define any startup dependencies between its constituent applications as a set of dependency graph edges.

- [Registry](catalog.v3.Registry) represents a repository from which various artifacts, such as application Docker\* images or Helm charts,
   can be retrieved. As such, registry entity holds information used for finding and accessing the represented repository.

- [Artifact](catalog.v3.Artifact) represents a binary artifact that can be used for various purposes, e.g. icon or thumbnail for UI display, or
   auxiliary artifacts for integration with various platform services such as Grafana\* dashboard and similar. An artifact may be
   used by multiple deployment packages.

The API provides Create, Get, List, Update, Delete, and Watch operations for each of the above resources.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| UploadCatalogEntities | [UploadCatalogEntitiesRequest](#catalog-v3-UploadCatalogEntitiesRequest) | [UploadCatalogEntitiesResponse](#catalog-v3-UploadCatalogEntitiesResponse) | Allows uploading of a YAML file containing various application catalog entities. Multiple RPC invocations tagged with the same upload session ID can be used to upload multiple files and to create or update several catalog entities as a single transaction. |
| Import | [ImportRequest](#catalog-v3-ImportRequest) | [ImportResponse](#catalog-v3-ImportResponse) | Allows importing a deployment package from a Helm Chart. This is done as a single invocation with the URL of the asset to be imported. |
| CreateRegistry | [CreateRegistryRequest](#catalog-v3-CreateRegistryRequest) | [CreateRegistryResponse](#catalog-v3-CreateRegistryResponse) | Creates a new registry. |
| ListRegistries | [ListRegistriesRequest](#catalog-v3-ListRegistriesRequest) | [ListRegistriesResponse](#catalog-v3-ListRegistriesResponse) | Gets a list of registries. |
| GetRegistry | [GetRegistryRequest](#catalog-v3-GetRegistryRequest) | [GetRegistryResponse](#catalog-v3-GetRegistryResponse) | Gets a specific registry. |
| UpdateRegistry | [UpdateRegistryRequest](#catalog-v3-UpdateRegistryRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Updates a registry. |
| DeleteRegistry | [DeleteRegistryRequest](#catalog-v3-DeleteRegistryRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Deletes a registry. |
| WatchRegistries | [WatchRegistriesRequest](#catalog-v3-WatchRegistriesRequest) | [WatchRegistriesResponse](#catalog-v3-WatchRegistriesResponse) stream | Watches inventory of registries for changes. |
| CreateDeploymentPackage | [CreateDeploymentPackageRequest](#catalog-v3-CreateDeploymentPackageRequest) | [CreateDeploymentPackageResponse](#catalog-v3-CreateDeploymentPackageResponse) | Creates a new deployment package. |
| ListDeploymentPackages | [ListDeploymentPackagesRequest](#catalog-v3-ListDeploymentPackagesRequest) | [ListDeploymentPackagesResponse](#catalog-v3-ListDeploymentPackagesResponse) | Gets a list of deployment packages. |
| GetDeploymentPackage | [GetDeploymentPackageRequest](#catalog-v3-GetDeploymentPackageRequest) | [GetDeploymentPackageResponse](#catalog-v3-GetDeploymentPackageResponse) | Gets a specific deployment package. |
| GetDeploymentPackageVersions | [GetDeploymentPackageVersionsRequest](#catalog-v3-GetDeploymentPackageVersionsRequest) | [GetDeploymentPackageVersionsResponse](#catalog-v3-GetDeploymentPackageVersionsResponse) | Gets all versions of a named deployment package. |
| UpdateDeploymentPackage | [UpdateDeploymentPackageRequest](#catalog-v3-UpdateDeploymentPackageRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Updates a deployment package. |
| DeleteDeploymentPackage | [DeleteDeploymentPackageRequest](#catalog-v3-DeleteDeploymentPackageRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Deletes a deployment package. |
| WatchDeploymentPackages | [WatchDeploymentPackagesRequest](#catalog-v3-WatchDeploymentPackagesRequest) | [WatchDeploymentPackagesResponse](#catalog-v3-WatchDeploymentPackagesResponse) stream | Watches inventory of deployment packages for changes. |
| CreateApplication | [CreateApplicationRequest](#catalog-v3-CreateApplicationRequest) | [CreateApplicationResponse](#catalog-v3-CreateApplicationResponse) | Creates a new application. |
| ListApplications | [ListApplicationsRequest](#catalog-v3-ListApplicationsRequest) | [ListApplicationsResponse](#catalog-v3-ListApplicationsResponse) | Gets a list of applications. |
| GetApplication | [GetApplicationRequest](#catalog-v3-GetApplicationRequest) | [GetApplicationResponse](#catalog-v3-GetApplicationResponse) | Gets a specific application. |
| GetApplicationReferenceCount | [GetApplicationReferenceCountRequest](#catalog-v3-GetApplicationReferenceCountRequest) | [GetApplicationReferenceCountResponse](#catalog-v3-GetApplicationReferenceCountResponse) | Gets application reference count - the number of deployment packages using this application. |
| GetApplicationVersions | [GetApplicationVersionsRequest](#catalog-v3-GetApplicationVersionsRequest) | [GetApplicationVersionsResponse](#catalog-v3-GetApplicationVersionsResponse) | Gets all versions of a named application. |
| UpdateApplication | [UpdateApplicationRequest](#catalog-v3-UpdateApplicationRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Updates an application. |
| DeleteApplication | [DeleteApplicationRequest](#catalog-v3-DeleteApplicationRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Deletes an application. |
| WatchApplications | [WatchApplicationsRequest](#catalog-v3-WatchApplicationsRequest) | [WatchApplicationsResponse](#catalog-v3-WatchApplicationsResponse) stream | Watches inventory of applications for changes. |
| CreateArtifact | [CreateArtifactRequest](#catalog-v3-CreateArtifactRequest) | [CreateArtifactResponse](#catalog-v3-CreateArtifactResponse) | Creates a new artifact. |
| ListArtifacts | [ListArtifactsRequest](#catalog-v3-ListArtifactsRequest) | [ListArtifactsResponse](#catalog-v3-ListArtifactsResponse) | Gets a list of artifacts. |
| GetArtifact | [GetArtifactRequest](#catalog-v3-GetArtifactRequest) | [GetArtifactResponse](#catalog-v3-GetArtifactResponse) | Gets a specific artifact. |
| UpdateArtifact | [UpdateArtifactRequest](#catalog-v3-UpdateArtifactRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Updates an artifact. |
| DeleteArtifact | [DeleteArtifactRequest](#catalog-v3-DeleteArtifactRequest) | [.google.protobuf.Empty](#google-protobuf-Empty) | Deletes an artifact. |
| WatchArtifacts | [WatchArtifactsRequest](#catalog-v3-WatchArtifactsRequest) | [WatchArtifactsResponse](#catalog-v3-WatchArtifactsResponse) stream | Watches inventory of artifacts for changes. |

 <!-- end services -->

## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
