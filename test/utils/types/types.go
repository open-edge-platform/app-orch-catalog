// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"

	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
)

const (
	DeploymentPackagesEndpoint = "/catalog.orchestrator.apis/v3/deployment_packages"
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

func GetPointerString(s string) *string {
	return &s
}

func GetDeploymentPackages() []restClient.CatalogV3DeploymentPackage {
	var pkgs []restClient.CatalogV3DeploymentPackage
	extensionKind := restClient.KINDEXTENSION
	for _, dp := range []restClient.CatalogV3DeploymentPackage{
		{Name: "cert-manager", Description: GetPointerString("cert-manager Deployment Package"), Kind: &extensionKind},
			{Name: "headlamp", Description: GetPointerString("headlamp is a web-based UI for Kubernetes cluster management."), Kind: &extensionKind},
			{Name: "intel-gpu", Description: GetPointerString("Intel GPU K8S extension"), Kind: &extensionKind},
		{Name: "loadbalancer", Description: GetPointerString("Enables load balancer and ingress controller on the edge"), Kind: &extensionKind},
		{Name: "nfd", Description: GetPointerString("Node Feature Discovery (NFD) Deployment Package"), Kind: &extensionKind},
		{Name: "nvidia-gpu-operator", Description: GetPointerString("NVIDIA GPU Operator deployment package"), Kind: &extensionKind},
		{Name: "observability", Description: GetPointerString("observability Stack"), Kind: &extensionKind},
		{Name: "trusted-compute", Description: GetPointerString("Trusted Compute k8s plugin for trusted workloads.\n"), Kind: &extensionKind},
		{Name: "virtualization", Description: GetPointerString("Virtualization support for k8s cluster"), Kind: &extensionKind},
	} {
		pkgs = append(pkgs, restClient.CatalogV3DeploymentPackage{
			Name:        dp.Name,
			Description: dp.Description,
			Kind:        dp.Kind,
		})
	}
	return pkgs
}

func GetApplications() []restClient.CatalogV3Application {
	var apps []restClient.CatalogV3Application
	extensionKind := restClient.KINDEXTENSION

	for _, sa := range []restClient.CatalogV3Application{
		{Name: "attestation-manager", DisplayName: GetPointerString("attestation-manager"), Description: GetPointerString("Workload prptection and continus monitoring add-on for Kubernetes"), Kind: &extensionKind, ChartName: "edge-orch/trusted-compute/charts/attestation-manager", HelmRegistryName: "intel-rs-helm"},
		{Name: "attestation-verifier", DisplayName: GetPointerString("attestation-verifier"), Description: GetPointerString("attestaion verifier of trusted compute"), Kind: &extensionKind, ChartName: "edge-orch/trusted-compute/charts/attestation-verifier", HelmRegistryName: "intel-rs-helm"},
		{Name: "cdi", DisplayName: GetPointerString("cdi"), Description: GetPointerString("Persistent storage management add-on for Kubernetes"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/cdi", HelmRegistryName: "intel-rs-helm"},
		{Name: "cert-manager", DisplayName: GetPointerString("cert-manager"), Description: GetPointerString("Cert Manager"), Kind: &extensionKind, ChartName: "cert-manager", HelmRegistryName: "jetstack"},
		{Name: "intel-device-operator", DisplayName: GetPointerString("intel-device-operator"), Description: GetPointerString("Intel Device Plugin Operator"), Kind: &extensionKind, ChartName: "intel-device-plugins-operator", HelmRegistryName: "intel-github-io"},
		{Name: "intel-gpu-plugin", DisplayName: GetPointerString("intel-gpu-plugin"), Description: GetPointerString("Intel GPU Device Plugin"), Kind: &extensionKind, ChartName: "intel-device-plugins-gpu", HelmRegistryName: "intel-github-io"},
			{Name: "headlamp", DisplayName: GetPointerString("headlamp"), Description: GetPointerString("headlamp is a web-based UI for Kubernetes cluster management."), Kind: &extensionKind, ChartName: "edge-orch/en/charts/headlamp", HelmRegistryName: "intel-rs-helm"},
		{Name: "kubevirt", DisplayName: GetPointerString("kubevirt"), Description: GetPointerString("Virtual machine management add-on for Kubernetes"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/kubevirt", HelmRegistryName: "intel-rs-helm"},
		{Name: "kubevirt-helper", DisplayName: GetPointerString("kubevirt-helper"), Description: GetPointerString("Automatically restart VM when editable VM spec is updated"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/kubevirt-helper", HelmRegistryName: "intel-rs-helm"},
		{Name: "metallb", DisplayName: GetPointerString("metallb"), Description: GetPointerString("Load balancer for bare metal k8s clusters"), Kind: &extensionKind, ChartName: "metallb", HelmRegistryName: "metallb-helm"},
		{Name: "metallb-base", DisplayName: GetPointerString("metallb-base"), Description: GetPointerString("Metallb base configuration"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/metallb-base", HelmRegistryName: "intel-rs-helm"},
		{Name: "metallb-config", DisplayName: GetPointerString("metallb-config"), Description: GetPointerString("Load balancer configuration for bare metal k8s clusters"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/metallb-config", HelmRegistryName: "intel-rs-helm"},
		{Name: "nfd", DisplayName: GetPointerString("nfd"), Description: GetPointerString("NFD"), Kind: &extensionKind, ChartName: "node-feature-discovery", HelmRegistryName: "node-feature-discovery"},
		{Name: "nvidia-gpu-operator-app", DisplayName: GetPointerString("nvidia-gpu-operator-app"), Description: GetPointerString("nvidia gpu operator application"), Kind: &extensionKind, ChartName: "gpu-operator", HelmRegistryName: "nvidia-ncg"},
		{Name: "fluent-bit", DisplayName: GetPointerString("fluent-bit"), Description: GetPointerString("Fluent Bit"), Kind: &extensionKind, ChartName: "fluent-bit", HelmRegistryName: "fluent-bit"},
		{Name: "node-exporter", DisplayName: GetPointerString("node-exporter"), Description: GetPointerString("Node Exporter"), Kind: &extensionKind, ChartName: "prometheus-node-exporter", HelmRegistryName: "node-exporter"},
		{Name: "observability-config", DisplayName: GetPointerString("observability-config"), Description: GetPointerString("Observability Config"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/observability-config", HelmRegistryName: "intel-rs-helm"},
		{Name: "prometheus", DisplayName: GetPointerString("prometheus"), Description: GetPointerString("Prometheus"), Kind: &extensionKind, ChartName: "kube-prometheus-stack", HelmRegistryName: "prometheus"},
		{Name: "telegraf", DisplayName: GetPointerString("telegraf"), Description: GetPointerString("Telegraf"), Kind: &extensionKind, ChartName: "telegraf", HelmRegistryName: "telegraf"},
		{Name: "trust-agent", DisplayName: GetPointerString("trust-agent"), Description: GetPointerString("Automatically restart VM when editable VM spec is updated"), Kind: &extensionKind, ChartName: "edge-orch/trusted-compute/charts/trustagent", HelmRegistryName: "intel-rs-helm"},
		{Name: "trusted-workload", DisplayName: GetPointerString("trusted-workload"), Description: GetPointerString("Deploys the necessary CRD and runtime class to enable trusted compute workloads within virtual machines."), Kind: &extensionKind, ChartName: "edge-orch/trusted-compute/charts/trusted-workload", HelmRegistryName: "intel-rs-helm"},
	} {
		apps = append(apps, restClient.CatalogV3Application{
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

func GetRegistryDefinitions(orchDomain string) []restClient.CatalogV3Registry {
	dockerURL := fmt.Sprintf("https://registry-oci.%s/", orchDomain)
	helmURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", orchDomain)

	return []restClient.CatalogV3Registry{
		{Name: "bitnami-helm-oci", DisplayName: GetPointerString("bitnami-helm-oci"), Description: GetPointerString("Bitnami helm registry"), RootUrl: "oci://registry-1.docker.io/bitnamicharts", Type: "HELM"},
		{Name: "fluent-bit", DisplayName: GetPointerString("fluent-bit"), Description: GetPointerString("Public registry for fluent bit chart"), RootUrl: "https://fluent.github.io/helm-charts", Type: "HELM"},
		{Name: "harbor-docker-oci", DisplayName: GetPointerString("harbor oci docker"), Description: GetPointerString("Harbor OCI docker images registry"), RootUrl: dockerURL, Type: "IMAGE"},
		{Name: "harbor-helm-oci", DisplayName: GetPointerString("harbor oci helm"), Description: GetPointerString("Harbor OCI helm charts registry"), RootUrl: helmURL, Type: "HELM"},
		{Name: "intel-github-io", DisplayName: GetPointerString("intel-github-io"), Description: GetPointerString("Intel Public registry with device operator & plugins"), RootUrl: "https://intel.github.io/helm-charts", Type: "HELM"},
		{Name: "intel-rs-helm", DisplayName: GetPointerString("intel-rs-helm"), Description: GetPointerString("Repo on registry registry-rs.edgeorchestration.intel.com"), RootUrl: "oci://rs-proxy.orch-platform.svc.cluster.local:8443", Type: "HELM"},
		{Name: "intel-rs-images", DisplayName: GetPointerString("intel-rs-image"), Description: GetPointerString("Repo on registry registry-rs.edgeorchestration.intel.com"), RootUrl: "oci://registry-rs.edgeorchestration.intel.com", Type: "IMAGE"},
		{Name: "jetstack", DisplayName: GetPointerString("jetstack"), Description: GetPointerString("Public registry for cert manager chart"), RootUrl: "https://charts.jetstack.io", Type: "HELM"},
			{Name: "headlamp", DisplayName: GetPointerString("headlamp"), Description: GetPointerString("Headlamp registry"), RootUrl: "https://kubernetes-sigs.github.io/headlamp/", Type: "HELM"},
		{Name: "kubernetes-ingress-helm", DisplayName: GetPointerString("kubernetes-ingress-helm"), Description: GetPointerString("Kubernetes Github helm registry for ingress-nginx"), RootUrl: "https://kubernetes.github.io/ingress-nginx", Type: "HELM"},
	}
}
