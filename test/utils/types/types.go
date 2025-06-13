// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package types

const (
	ApplicationsEndpoint       = "/catalog.orchestrator.apis/v3/applications"
	DeploymentPackagesEndpoint = "/catalog.orchestrator.apis/v3/deployment_packages"
	RegistriesEndpoint         = "/catalog.orchestrator.apis/v3/registries"
	UploadEndpoint             = "/catalog.orchestrator.apis/upload"

	WordpressTarballPathName = "../testdata/wordpress.tar.gz"
	WordpressName            = "test-wordpress"
	WordpressVersion         = "0.1.1"
	WordpressRegistryName    = "test-bitnami"
)

const (
	RestAddressPortForward = "127.0.0.1"

	PortForwardServiceNamespace = "orch-app"
	PortForwardService          = "svc/app-orch-catalog-rest-proxy"
	PortForwardLocalPort        = "8081"
	PortForwardAddress          = "0.0.0.0"
	PortForwardRemotePort       = "8081"
)

const (
	SampleOrg     = "sample-org"
	SampleProject = "sample-project"
)

const ImportEndpoint = "/catalog.orchestrator.apis/v3/import"

/* The reason for these Short* objects was to facilitate converting the existing
 * rest api tests to the more complex test framework that came from the mage e2e
 * tests, which have many more fields.
 */

type ShortApplication struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	Description      string `json:"description"`
	Kind             string `json:"kind"`
	ChartName        string `json:"chartName"`
	HelmRegistryName string `json:"helmRegistryName"`
}

type ShortDeploymentPackage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

type (
	DeploymentPackage struct {
		ApplicationDependencies *[]ApplicationDependency `json:"applicationDependencies,omitempty"`
		ApplicationReferences   []ApplicationReference   `json:"applicationReferences"`
		Artifacts               []ArtifactReference      `json:"artifacts"`
		DefaultNamespaces       *map[string]string       `json:"defaultNamespaces,omitempty"`
		DefaultProfileName      string                   `json:"defaultProfileName,omitempty"`
		Description             string                   `json:"description,omitempty"`
		DisplayName             string                   `json:"displayName,omitempty"`
		Extensions              []APIExtension           `json:"extensions"`
		IsDeployed              bool                     `json:"isDeployed,omitempty"`
		IsVisible               bool                     `json:"isVisible,omitempty"`
		Name                    string                   `json:"name"`
		Profiles                []Profile                `json:"profiles,omitempty"`
		Version                 string                   `json:"version"`
		Kind                    string                   `json:"kind"`
	}

	DeploymentPackages struct {
		DeploymentPackages []DeploymentPackage `json:"DeploymentPackages"`
	}

	DeploymentPackageGetResponse struct {
		DeploymentPackage DeploymentPackage `json:"deploymentPackage"`
	}

	Profile struct {
		ChartValues string `json:"chartValues,omitempty"`
		Description string `json:"description,omitempty"`
		DisplayName string `json:"displayName,omitempty"`
		Name        string `json:"name"`
	}

	Application struct {
		ChartName          string    `json:"chartName"`
		ChartVersion       string    `json:"chartVersion"`
		DefaultProfileName string    `json:"defaultProfileName,omitempty"`
		Description        string    `json:"description,omitempty"`
		DisplayName        string    `json:"displayName,omitempty"`
		HelmRegistryName   string    `json:"helmRegistryName"`
		ImageRegistryName  string    `json:"imageRegistryName,omitempty"`
		Name               string    `json:"name"`
		Profiles           []Profile `json:"profiles,omitempty"`
		Version            string    `json:"version"`
		Kind               string    `json:"kind"`
	}

	ApplicationGetResponse struct {
		Application Application `json:"application"`
	}

	ApplicationDependency struct{}
	ApplicationReference  struct{}
	ArtifactReference     struct{}
	Endpoint              struct {
		AuthType     string `json:"authType"`
		ExternalPath string `json:"externalPath"`
		InternalPath string `json:"internalPath"`
		Scheme       string `json:"scheme"`
		ServiceName  string `json:"serviceName"`
	}

	UIExtension struct{}

	APIExtension struct {
		Description string      `json:"description,omitempty"`
		DisplayName string      `json:"displayName,omitempty"`
		Endpoints   []Endpoint  `json:"endpoints,omitempty"`
		Name        string      `json:"name"`
		UiExtension UIExtension `json:"uiExtension,omitempty"`
		Version     string      `json:"version"`
	}

	Applications struct {
		Applications []Application `json:"applications"`
	}

	Registry struct {
		AuthToken   string  `json:"authToken,omitempty"`
		Cacerts     string  `json:"cacerts,omitempty"`
		Description string  `json:"description,omitempty"`
		DisplayName string  `json:"displayName,omitempty"`
		Name        string  `json:"name"`
		RootURL     string  `json:"rootUrl"`
		SecretID    *string `json:"secretId,omitempty"`
		Type        string  `json:"type"`
		Username    string  `json:"username,omitempty"`
	}

	RegistryGetResponse struct {
		Registry Registry `json:"registry"`
	}
)

func GetDeploymentPackages() []DeploymentPackage {
	pkgs := []DeploymentPackage{}
	for _, dp := range []ShortDeploymentPackage{
		{"base-extensions", "Base Extensions", "KIND_EXTENSION"},
		{"intel-gpu", "Intel GPU K8S extension", "KIND_EXTENSION"},
		{"kubernetes-dashboard", "kubernetes-dashboard", "KIND_EXTENSION"},
		{"loadbalancer", "Enables load balancer and dns services on the edge", "KIND_EXTENSION"},
		{"skupper", "Enables Skupper service on the edge", "KIND_EXTENSION"},
		{"sriov", "Provisions and configures SR-IOV CNI plugin and Device plugin", "KIND_EXTENSION"},
		{"trusted-compute", "Trusted Compute k8s plugin for trusted workloads. Requires cluster using a \"privilege\" template.", "KIND_EXTENSION"},
		{"usb", "Brings USB allocation for containers/VMs running on k8s cluster", "KIND_EXTENSION"},
		{"virtualization", "Virtualization support for k8s cluster", "KIND_EXTENSION"},
	} {
		pkgs = append(pkgs, DeploymentPackage{
			Name:        dp.Name,
			Description: dp.Description,
			Kind:        dp.Kind,
		})
	}
	return pkgs
}

func GetApplications() []Application {
	apps := []Application{}
	for _, sa := range []ShortApplication{
		{"gatekeeper-constraints", "gatekeeper-constraints", "Gatekeeper Constraints", "KIND_EXTENSION", "edge-orch/en/charts/gatekeeper-constraints", "intel-rs-helm"},
		{"ingress-nginx", "ingress-nginx", "Edge Orchestrator EdgeDNS", "KIND_EXTENSION", "ingress-nginx", "kubernetes-ingress-helm"},
		{"intel-device-operator", "intel-device-operator", "Intel Device Plugin Operator", "KIND_EXTENSION", "intel-device-plugins-operator", "intel-github-io"},
		{"intel-gpu-plugin", "intel-gpu-plugin", "Intel GPU Device Plugin", "KIND_EXTENSION", "intel-device-plugins-gpu", "intel-github-io"},
		{"kubernetes-dashboard", "kubernetes-dashboard", "kubernetes-dashboard", "KIND_EXTENSION", "kubernetes-dashboard", "kubernetes"},
		{"metallb", "metallb", "Load balancer for bare metal k8s clusters", "KIND_EXTENSION", "metallb", "bitnami-helm-oci"},
		{"metallb-base", "metallb-base", "Metallb base configuration", "KIND_EXTENSION", "edge-orch/en/charts/metallb-base", "intel-rs-helm"},
		{"metallb-config", "metallb-config", "Load balancer configuration for bare metal k8s clusters", "KIND_EXTENSION", "edge-orch/en/charts/metallb-config", "intel-rs-helm"},
		{"network-policies", "network-policies", "Network Policies", "KIND_EXTENSION", "edge-orch/en/charts/network-policies", "intel-rs-helm"},
		{"cert-manager", "cert-manager", "Cert Manager", "KIND_EXTENSION", "cert-manager", "jetstack"},
		{"edgedns", "edgedns", "Edge Orchestrator EdgeDNS", "KIND_EXTENSION", "edge-orch/en/charts/edgedns", "intel-rs-helm"},
		{"fluent-bit", "fluent-bit", "Fluent Bit", "KIND_EXTENSION", "fluent-bit", "fluent-bit"},
		{"gatekeeper", "gatekeeper", "Gatekeeper", "KIND_EXTENSION", "gatekeeper", "gatekeeper"},
		{"akri", "akri", "akri base application", "KIND_EXTENSION", "akri", "akri-helm-registry"},
		{"attestation-manager", "attestation-manager", "Workload prptection and continus monitoring add-on for Kubernetes", "KIND_EXTENSION", "edge-orch/trusted-compute/charts/attestation-manager", "intel-rs-helm"},
		{"attestation-verifier", "attestation-verifier", "attestation verifier of trusted compute", "KIND_EXTENSION", "edge-orch/trusted-compute/charts/attestation-verifier", "intel-rs-helm"},
		{"cdi", "cdi", "Persistent storage management add-on for Kubernetes", "KIND_EXTENSION", "edge-orch/en/charts/cdi", "intel-rs-helm"},
		{"kubevirt", "kubevirt", "Virtual machine management add-on for Kubernetes", "KIND_EXTENSION", "edge-orch/en/charts/kubevirt", "intel-rs-helm"},
		{"kubevirt-helper", "kubevirt-helper", "Automatically restart VM when editable VM spec is updated", "KIND_EXTENSION", "edge-orch/en/charts/kubevirt-helper", "intel-rs-helm"},
		{"nfd", "nfd", "NFD", "KIND_EXTENSION", "node-feature-discovery", "node-feature-discovery"},
	} {
		apps = append(apps, Application{
			Name:             sa.Name,
			DisplayName:      sa.DisplayName,
			Description:      sa.Description,
			Kind:             sa.Kind,
			ChartName:        sa.ChartName,
			HelmRegistryName: sa.HelmRegistryName,
		})
	}
	return apps
}
