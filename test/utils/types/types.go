// SPDX-FileCopyrightText: (C) 2025-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package types

import "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"

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

/* The reason for these Short* objects was to facilitate converting the existing
 * rest api tests to the more complex test framework that came from the mage e2e
 * tests, which have many more fields.
 */

func GetPointerString(s string) *string {
	return &s
}

func GetDeploymentPackages() []restClient.DeploymentPackage {
	pkgs := []restClient.DeploymentPackage{}
	extensioKind := restClient.DeploymentPackageKindKINDEXTENSION
	for _, dp := range []restClient.DeploymentPackage{
		{Name: "base-extensions", Description: GetPointerString("Base Extensions"), Kind: &extensioKind},
		{Name: "intel-gpu", Description: GetPointerString("Intel GPU K8S extension"), Kind: &extensioKind},
		{Name: "kubernetes-dashboard", Description: GetPointerString("kubernetes-dashboard"), Kind: &extensioKind},
		{Name: "loadbalancer", Description: GetPointerString("Enables load balancer and dns services on the edge"), Kind: &extensioKind},
		{Name: "skupper", Description: GetPointerString("Enables Skupper service on the edge"), Kind: &extensioKind},
		{Name: "sriov", Description: GetPointerString("Provisions and configures SR-IOV CNI plugin and Device plugin"), Kind: &extensioKind},
		{Name: "trusted-compute", Description: GetPointerString("Trusted Compute k8s plugin for trusted workloads. Requires cluster using a \"privilege\" template."), Kind: &extensioKind},
		{Name: "usb", Description: GetPointerString("Brings USB allocation for containers/VMs running on k8s cluster"), Kind: &extensioKind},
		{Name: "virtualization", Description: GetPointerString("Virtualization support for k8s cluster"), Kind: &extensioKind},
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
	apps := []restClient.Application{}
	extensioKind := restClient.ApplicationKindKINDEXTENSION

	for _, sa := range []restClient.Application{
		{Name: "gatekeeper-constraints", DisplayName: GetPointerString("gatekeeper-constraints"), Description: GetPointerString("Gatekeeper Constraints"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/gatekeeper-constraints", HelmRegistryName: "intel-rs-helm"},
		{Name: "ingress-nginx", DisplayName: GetPointerString("ingress-nginx"), Description: GetPointerString("Edge Orchestrator EdgeDNS"), Kind: &extensioKind, ChartName: "ingress-nginx", HelmRegistryName: "kubernetes-ingress-helm"},
		{Name: "intel-device-operator", DisplayName: GetPointerString("intel-device-operator"), Description: GetPointerString("Intel Device Plugin Operator"), Kind: &extensioKind, ChartName: "intel-device-plugins-operator", HelmRegistryName: "intel-github-io"},
		{Name: "intel-gpu-plugin", DisplayName: GetPointerString("intel-gpu-plugin"), Description: GetPointerString("Intel GPU Device Plugin"), Kind: &extensioKind, ChartName: "intel-device-plugins-gpu", HelmRegistryName: "intel-github-io"},
		{Name: "kubernetes-dashboard", DisplayName: GetPointerString("kubernetes-dashboard"), Description: GetPointerString("kubernetes-dashboard"), Kind: &extensioKind, ChartName: "kubernetes-dashboard", HelmRegistryName: "kubernetes"},
		{Name: "metallb", DisplayName: GetPointerString("metallb"), Description: GetPointerString("Load balancer for bare metal k8s clusters"), Kind: &extensioKind, ChartName: "metallb", HelmRegistryName: "bitnami-helm-oci"},
		{Name: "metallb-base", DisplayName: GetPointerString("metallb-base"), Description: GetPointerString("Metallb base configuration"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/metallb-base", HelmRegistryName: "intel-rs-helm"},
		{Name: "metallb-config", DisplayName: GetPointerString("metallb-config"), Description: GetPointerString("Load balancer configuration for bare metal k8s clusters"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/metallb-config", HelmRegistryName: "intel-rs-helm"},
		{Name: "network-policies", DisplayName: GetPointerString("network-policies"), Description: GetPointerString("Network Policies"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/network-policies", HelmRegistryName: "intel-rs-helm"},
		{Name: "cert-manager", DisplayName: GetPointerString("cert-manager"), Description: GetPointerString("Cert Manager"), Kind: &extensioKind, ChartName: "cert-manager", HelmRegistryName: "jetstack"},
		{Name: "edgedns", DisplayName: GetPointerString("edgedns"), Description: GetPointerString("Edge Orchestrator EdgeDNS"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/edgedns", HelmRegistryName: "intel-rs-helm"},
		{Name: "fluent-bit", DisplayName: GetPointerString("fluent-bit"), Description: GetPointerString("Fluent Bit"), Kind: &extensioKind, ChartName: "fluent-bit", HelmRegistryName: "fluent-bit"},
		{Name: "gatekeeper", DisplayName: GetPointerString("gatekeeper"), Description: GetPointerString("Gatekeeper"), Kind: &extensioKind, ChartName: "gatekeeper", HelmRegistryName: "gatekeeper"},
		{Name: "akri", DisplayName: GetPointerString("akri"), Description: GetPointerString("akri base application"), Kind: &extensioKind, ChartName: "akri", HelmRegistryName: "akri-helm-registry"},
		{Name: "attestation-manager", DisplayName: GetPointerString("attestation-manager"), Description: GetPointerString("Workload prptection and continus monitoring add-on for Kubernetes"), Kind: &extensioKind, ChartName: "edge-orch/trusted-compute/charts/attestation-manager", HelmRegistryName: "intel-rs-helm"},
		{Name: "attestation-verifier", DisplayName: GetPointerString("attestation-verifier"), Description: GetPointerString("attestation verifier of trusted compute"), Kind: &extensioKind, ChartName: "edge-orch/trusted-compute/charts/attestation-verifier", HelmRegistryName: "intel-rs-helm"},
		{Name: "cdi", DisplayName: GetPointerString("cdi"), Description: GetPointerString("Persistent storage management add-on for Kubernetes"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/cdi", HelmRegistryName: "intel-rs-helm"},
		{Name: "kubevirt", DisplayName: GetPointerString("kubevirt"), Description: GetPointerString("Virtual machine management add-on for Kubernetes"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/kubevirt", HelmRegistryName: "intel-rs-helm"},
		{Name: "kubevirt-helper", DisplayName: GetPointerString("kubevirt-helper"), Description: GetPointerString("Automatically restart VM when editable VM spec is updated"), Kind: &extensioKind, ChartName: "edge-orch/en/charts/kubevirt-helper", HelmRegistryName: "intel-rs-helm"},
		{Name: "nfd", DisplayName: GetPointerString("nfd"), Description: GetPointerString("NFD"), Kind: &extensioKind, ChartName: "node-feature-discovery", HelmRegistryName: "node-feature-discovery"},
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
