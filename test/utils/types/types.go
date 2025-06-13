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

func GetDeploymentPackages() []restClient.DeploymentPackage {
	var pkgs []restClient.DeploymentPackage
	extensionKind := restClient.DeploymentPackageKindKINDEXTENSION
	for _, dp := range []restClient.DeploymentPackage{
		{Name: "base-extensions", Description: GetPointerString("Base Extensions"), Kind: &extensionKind},
		{Name: "intel-gpu", Description: GetPointerString("Intel GPU K8S extension"), Kind: &extensionKind},
		{Name: "kubernetes-dashboard", Description: GetPointerString("kubernetes-dashboard"), Kind: &extensionKind},
		{Name: "loadbalancer", Description: GetPointerString("Enables load balancer and dns services on the edge"), Kind: &extensionKind},
		{Name: "skupper", Description: GetPointerString("Enables Skupper service on the edge"), Kind: &extensionKind},
		{Name: "sriov", Description: GetPointerString("Provisions and configures SR-IOV CNI plugin and Device plugin"), Kind: &extensionKind},
		{Name: "trusted-compute", Description: GetPointerString("Trusted Compute k8s plugin for trusted workloads. Requires cluster using a \"privilege\" template."), Kind: &extensionKind},
		{Name: "usb", Description: GetPointerString("Brings USB allocation for containers/VMs running on k8s cluster"), Kind: &extensionKind},
		{Name: "virtualization", Description: GetPointerString("Virtualization support for k8s cluster"), Kind: &extensionKind},
	} {
		pkgs = append(pkgs, restClient.DeploymentPackage{
			Name:        dp.Name,
			Description: dp.Description,
			Kind:        dp.Kind,
		})
	}
	return pkgs
}

func GetApplications() []restClient.Application {
	var apps []restClient.Application
	extensionKind := restClient.ApplicationKindKINDEXTENSION

	for _, sa := range []restClient.Application{
		{Name: "gatekeeper-constraints", DisplayName: GetPointerString("gatekeeper-constraints"), Description: GetPointerString("Gatekeeper Constraints"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/gatekeeper-constraints", HelmRegistryName: "intel-rs-helm"},
		{Name: "ingress-nginx", DisplayName: GetPointerString("ingress-nginx"), Description: GetPointerString("Edge Orchestrator EdgeDNS"), Kind: &extensionKind, ChartName: "ingress-nginx", HelmRegistryName: "kubernetes-ingress-helm"},
		{Name: "intel-device-operator", DisplayName: GetPointerString("intel-device-operator"), Description: GetPointerString("Intel Device Plugin Operator"), Kind: &extensionKind, ChartName: "intel-device-plugins-operator", HelmRegistryName: "intel-github-io"},
		{Name: "intel-gpu-plugin", DisplayName: GetPointerString("intel-gpu-plugin"), Description: GetPointerString("Intel GPU Device Plugin"), Kind: &extensionKind, ChartName: "intel-device-plugins-gpu", HelmRegistryName: "intel-github-io"},
		{Name: "kubernetes-dashboard", DisplayName: GetPointerString("kubernetes-dashboard"), Description: GetPointerString("kubernetes-dashboard"), Kind: &extensionKind, ChartName: "kubernetes-dashboard", HelmRegistryName: "kubernetes"},
		{Name: "metallb", DisplayName: GetPointerString("metallb"), Description: GetPointerString("Load balancer for bare metal k8s clusters"), Kind: &extensionKind, ChartName: "metallb", HelmRegistryName: "bitnami-helm-oci"},
		{Name: "metallb-base", DisplayName: GetPointerString("metallb-base"), Description: GetPointerString("Metallb base configuration"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/metallb-base", HelmRegistryName: "intel-rs-helm"},
		{Name: "metallb-config", DisplayName: GetPointerString("metallb-config"), Description: GetPointerString("Load balancer configuration for bare metal k8s clusters"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/metallb-config", HelmRegistryName: "intel-rs-helm"},
		{Name: "network-policies", DisplayName: GetPointerString("network-policies"), Description: GetPointerString("Network Policies"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/network-policies", HelmRegistryName: "intel-rs-helm"},
		{Name: "cert-manager", DisplayName: GetPointerString("cert-manager"), Description: GetPointerString("Cert Manager"), Kind: &extensionKind, ChartName: "cert-manager", HelmRegistryName: "jetstack"},
		{Name: "edgedns", DisplayName: GetPointerString("edgedns"), Description: GetPointerString("Edge Orchestrator EdgeDNS"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/edgedns", HelmRegistryName: "intel-rs-helm"},
		{Name: "fluent-bit", DisplayName: GetPointerString("fluent-bit"), Description: GetPointerString("Fluent Bit"), Kind: &extensionKind, ChartName: "fluent-bit", HelmRegistryName: "fluent-bit"},
		{Name: "gatekeeper", DisplayName: GetPointerString("gatekeeper"), Description: GetPointerString("Gatekeeper"), Kind: &extensionKind, ChartName: "gatekeeper", HelmRegistryName: "gatekeeper"},
		{Name: "akri", DisplayName: GetPointerString("akri"), Description: GetPointerString("akri base application"), Kind: &extensionKind, ChartName: "akri", HelmRegistryName: "akri-helm-registry"},
		{Name: "attestation-manager", DisplayName: GetPointerString("attestation-manager"), Description: GetPointerString("Workload prptection and continus monitoring add-on for Kubernetes"), Kind: &extensionKind, ChartName: "edge-orch/trusted-compute/charts/attestation-manager", HelmRegistryName: "intel-rs-helm"},
		{Name: "attestation-verifier", DisplayName: GetPointerString("attestation-verifier"), Description: GetPointerString("attestation verifier of trusted compute"), Kind: &extensionKind, ChartName: "edge-orch/trusted-compute/charts/attestation-verifier", HelmRegistryName: "intel-rs-helm"},
		{Name: "cdi", DisplayName: GetPointerString("cdi"), Description: GetPointerString("Persistent storage management add-on for Kubernetes"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/cdi", HelmRegistryName: "intel-rs-helm"},
		{Name: "kubevirt", DisplayName: GetPointerString("kubevirt"), Description: GetPointerString("Virtual machine management add-on for Kubernetes"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/kubevirt", HelmRegistryName: "intel-rs-helm"},
		{Name: "kubevirt-helper", DisplayName: GetPointerString("kubevirt-helper"), Description: GetPointerString("Automatically restart VM when editable VM spec is updated"), Kind: &extensionKind, ChartName: "edge-orch/en/charts/kubevirt-helper", HelmRegistryName: "intel-rs-helm"},
		{Name: "nfd", DisplayName: GetPointerString("nfd"), Description: GetPointerString("NFD"), Kind: &extensionKind, ChartName: "node-feature-discovery", HelmRegistryName: "node-feature-discovery"},
	} {
		apps = append(apps, restClient.Application{
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

func GetRegistryDefinitions(orchDomain string) []restClient.Registry {
	dockerURL := fmt.Sprintf("https://registry-oci.%s/", orchDomain)
	helmURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", orchDomain)

	return []restClient.Registry{
		{Name: "akri-helm-registry", DisplayName: GetPointerString("akri-helm-registry"), Description: GetPointerString("Public registry for akri chart"), RootUrl: "https://project-akri.github.io/akri/", Type: "HELM"},
		{Name: "bitnami-helm-oci", DisplayName: GetPointerString("bitnami-helm-oci"), Description: GetPointerString("Bitnami helm registry"), RootUrl: "oci://registry-1.docker.io/bitnamicharts", Type: "HELM"},
		{Name: "fluent-bit", DisplayName: GetPointerString("fluent-bit"), Description: GetPointerString("Public registry for fluent bit chart"), RootUrl: "https://fluent.github.io/helm-charts", Type: "HELM"},
		{Name: "gatekeeper", DisplayName: GetPointerString("gatekeeper"), Description: GetPointerString("Public registry for gatekeeper chart"), RootUrl: "https://open-policy-agent.github.io/gatekeeper/charts", Type: "HELM"},
		{Name: "harbor-docker-oci", DisplayName: GetPointerString("harbor oci docker"), Description: GetPointerString("Harbor OCI docker images registry"), RootUrl: dockerURL, Type: "IMAGE"},
		{Name: "harbor-helm-oci", DisplayName: GetPointerString("harbor oci helm"), Description: GetPointerString("Harbor OCI helm charts registry"), RootUrl: helmURL, Type: "HELM"},
		{Name: "intel-github-io", DisplayName: GetPointerString("intel-github-io"), Description: GetPointerString("Intel Public registry with device operator & plugins"), RootUrl: "https://intel.github.io/helm-charts", Type: "HELM"},
		{Name: "intel-rs-helm", DisplayName: GetPointerString("intel-rs-helm"), Description: GetPointerString("Repo on registry registry-rs.edgeorchestration.intel.com"), RootUrl: "oci://rs-proxy.orch-platform.svc.cluster.local:8443", Type: "HELM"},
		{Name: "intel-rs-images", DisplayName: GetPointerString("intel-rs-image"), Description: GetPointerString("Repo on registry registry-rs.edgeorchestration.intel.com"), RootUrl: "oci://registry-rs.edgeorchestration.intel.com", Type: "IMAGE"},
		{Name: "jetstack", DisplayName: GetPointerString("jetstack"), Description: GetPointerString("Public registry for cert manager chart"), RootUrl: "https://charts.jetstack.io", Type: "HELM"},
	}
}
