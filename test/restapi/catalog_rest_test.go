// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	// Third-party imports
	"github.com/stretchr/testify/assert"

	// Standard library imports
	"log"

	// Project-specific imports
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
)

func init() {
	log.SetOutput(io.Discard) // Discard default logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(io.Writer(os.Stdout)) // Redirect log output to console
}

type Registry struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	RootURL     string `json:"rootUrl"`
	Type        string `json:"type"`
}

func (s *TestSuite) getRegistries() []Registry {
	return []Registry{
		{"akri-helm-registry", "akri-helm-registry", "Public registry for akri chart", "https://project-akri.github.io/akri/", "HELM"},
		{"bitnami-helm-oci", "bitnami-helm-oci", "Bitnami helm registry", "oci://registry-1.docker.io/bitnamicharts", "HELM"},
		{"fluent-bit", "fluent-bit", "Public registry for fluent bit chart", "https://fluent.github.io/helm-charts", "HELM"},
		{"gatekeeper", "gatekeeper", "Public registry for gatekeeper chart", "https://open-policy-agent.github.io/gatekeeper/charts", "HELM"},
		{"harbor-docker-oci", "harbor oci docker", "Harbor OCI docker images registry", "https://registry-oci.kind.internal/", "IMAGE"},
		{"harbor-helm-oci", "harbor oci helm", "Harbor OCI helm charts registry", "oci://registry-oci.kind.internal/catalog-apps-sample-org-sample-project", "HELM"},
		{"intel-github-io", "intel-github-io", "Intel Public registry with device operator & plugins", "https://intel.github.io/helm-charts", "HELM"},
		{"intel-rs-helm", "intel-rs-helm", "Repo on registry registry-rs.edgeorchestration.intel.com", "oci://rs-proxy.orch-platform.svc.cluster.local:8443", "HELM"},
		{"intel-rs-images", "intel-rs-image", "Repo on registry registry-rs.edgeorchestration.intel.com", "oci://registry-rs.edgeorchestration.intel.com", "IMAGE"},
		{"jetstack", "jetstack", "Public registry for cert manager chart", "https://charts.jetstack.io", "HELM"},
	}
}

type Application struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	Description      string `json:"description"`
	Version          string `json:"version"`
	Kind             string `json:"kind"`
	ChartName        string `json:"chartName"`
	ChartVersion     string `json:"chartVersion"`
	HelmRegistryName string `json:"helmRegistryName"`
}

/* If there is change in the versions, you can verify the list by executiong the function TestListBootStrapExtensions and
update the version information here */

func (s *TestSuite) getApplications() []Application {
	return []Application{
		{"gatekeeper-constraints", "gatekeeper-constraints", "Gatekeeper Constraints", "1.0.15", "KIND_EXTENSION", "edge-orch/en/charts/gatekeeper-constraints", "1.0.15", "intel-rs-helm"},
		{"ingress-nginx", "ingress-nginx", "Edge Orchestrator EdgeDNS", "5.1.1", "KIND_EXTENSION", "ingress-nginx", "4.12.0", "kubernetes-ingress-helm"},
		{"intel-device-operator", "intel-device-operator", "Intel Device Plugin Operator", "0.29.0", "KIND_EXTENSION", "intel-device-plugins-operator", "0.29.0", "intel-github-io"},
		{"intel-gpu-plugin", "intel-gpu-plugin", "Intel GPU Device Plugin", "0.29.0", "KIND_EXTENSION", "intel-device-plugins-gpu", "0.29.0", "intel-github-io"},
		{"kubernetes-dashboard", "kubernetes-dashboard", "kubernetes-dashboard", "0.0.5", "KIND_EXTENSION", "kubernetes-dashboard", "7.10.0", "kubernetes"},
		{"metallb", "metallb", "Load balancer for bare metal k8s clusters", "1.1.1", "KIND_EXTENSION", "metallb", "6.4.5", "bitnami-helm-oci"},
		{"metallb-base", "metallb-base", "Metallb base configuration", "0.15.2", "KIND_EXTENSION", "edge-orch/en/charts/metallb-base", "0.15.2", "intel-rs-helm"},
		{"metallb-config", "metallb-config", "Load balancer configuration for bare metal k8s clusters", "0.1.4", "KIND_EXTENSION", "edge-orch/en/charts/metallb-config", "0.1.4", "intel-rs-helm"},
		{"network-policies", "network-policies", "Network Policies", "0.1.13", "KIND_EXTENSION", "edge-orch/en/charts/network-policies", "0.1.13", "intel-rs-helm"},
		{"cert-manager", "cert-manager", "Cert Manager", "1.16.2", "KIND_EXTENSION", "cert-manager", "1.16.2", "jetstack"},
		{"edgedns", "edgedns", "Edge Orchestrator EdgeDNS", "2.0.8", "KIND_EXTENSION", "edge-orch/en/charts/edgedns", "1.3.31", "intel-rs-helm"},
		{"fluent-bit", "fluent-bit", "Fluent Bit", "0.48.9", "KIND_EXTENSION", "fluent-bit", "0.48.9", "fluent-bit"},
		{"gatekeeper", "gatekeeper", "Gatekeeper", "3.17.1", "KIND_EXTENSION", "gatekeeper", "3.17.1", "gatekeeper"},
		{"akri", "akri", "akri base application", "0.13.8", "KIND_EXTENSION", "akri", "0.13.8", "akri-helm-registry"},
		{"attestation-manager", "attestation-manager", "Workload prptection and continus monitoring add-on for Kubernetes", "1.0.0", "KIND_EXTENSION", "edge-orch/trusted-compute/charts/attestation-manager", "1.4.0", "intel-rs-helm"},
		{"attestation-verifier", "attestation-verifier", "attestaion verifier of trusted compute", "1.0.0", "KIND_EXTENSION", "edge-orch/trusted-compute/charts/attestation-verifier", "1.4.0", "intel-rs-helm"},
		{"cdi", "cdi", "Persistent storage management add-on for Kubernetes", "2.0.3", "KIND_EXTENSION", "edge-orch/en/charts/cdi", "1.60.4", "intel-rs-helm"},
		{"kubevirt", "kubevirt", "Virtual machine management add-on for Kubernetes", "1.2.7", "KIND_EXTENSION", "edge-orch/en/charts/kubevirt", "1.2.7", "intel-rs-helm"},
		{"kubevirt-helper", "kubevirt-helper", "Automatically restart VM when editable VM spec is updated", "1.4.5", "KIND_EXTENSION", "edge-orch/en/charts/kubevirt-helper", "1.4.5", "intel-rs-helm"},
		{"nfd", "nfd", "NFD", "0.17.0", "KIND_EXTENSION", "node-feature-discovery", "0.17.0", "node-feature-discovery"},
	}
}

type DeploymentPackages struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Kind        string `json:"kind"`
}

func (s *TestSuite) getDeploymentPackages() []DeploymentPackages {
	return []DeploymentPackages{
		{"base-extensions", "Base Extensions", "0.7.9", "KIND_EXTENSION"},
		{"intel-gpu", "Intel GPU K8S extension", "1.2.4", "KIND_EXTENSION"},
		{"kubernetes-dashboard", "kubernetes-dashboard", "0.0.6", "KIND_EXTENSION"},
		{"loadbalancer", "Enables load balancer and dns services on the edge", "0.4.5", "KIND_EXTENSION"},
		{"skupper", "Enables Skupper service on the edge", "0.1.7", "KIND_EXTENSION"},
		{"sriov", "Provisions and configures SR-IOV CNI plugin and Device plugin", "0.3.4", "KIND_EXTENSION"},
		{"trusted-compute", "Trusted Compute k8s plugin for trusted workloads. Requires cluster using a \"privilege\" template.", "0.4.4", "KIND_EXTENSION"},
		{"usb", "Brings USB allocation for containers/VMs running on k8s cluster", "0.3.3", "KIND_EXTENSION"},
		{"virtualization", "Virtualization support for k8s cluster", "0.3.7", "KIND_EXTENSION"},
	}
}

func (s *TestSuite) TestListBootStrapExtensions() {
	// Form the request URL
	const applicationsEndpoint = "/catalog.orchestrator.apis/v3/applications"
	requestURL := fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, applicationsEndpoint)

	// Make the curl request using the access token and format the output with jq
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)
	// defer res.Body.Close() (remove duplicate calls)

	auth.AddRestAuthHeader(req, s.token, s.projectID)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	assert.NoError(s.T(), err)
	s.Equal("200 OK", res.Status)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)

	var result struct {
		Applications []Application `json:"applications"`
	}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Assert that the size of the result.Applications matches the size of getApplications
	assert.Equal(s.T(), len(s.getApplications()), len(result.Applications), "Mismatch in the number of applications")
	// Log application details for debugging purposes
	for _, app := range result.Applications {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, Version: %s, Kind: %s, ChartName: %s",
			app.Name, app.DisplayName, app.Description, app.Version, app.Kind, app.ChartName)
	}

}

func (s *TestSuite) TestListBootStrapDeploymentPackages() {
	// Form the request URL
	requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/v3/deployment_packages", s.CatalogRESTServerUrl)

	// Make the HTTP GET request
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	// Add query parameters
	query := req.URL.Query()
	query.Add("orderBy", "name")
	query.Add("pageSize", "10")
	query.Add("offset", "0")
	query.Add("kinds", "KIND_EXTENSION")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	var result struct {
		DeploymentPackages []DeploymentPackages `json:"deploymentPackages"`
	}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Assert that the size of the result.DeploymentPackages matches the size of getDeploymentPackages
	assert.Equal(s.T(), len(s.getDeploymentPackages()), len(result.DeploymentPackages), "Mismatch in the number of deployment packages")
	// Log deployment package details for debugging purposes
	for _, pkg := range result.DeploymentPackages {
		log.Printf("Name: %s, Description: %s, Version: %s, Kind: %s",
			pkg.Name, pkg.Description, pkg.Version, pkg.Kind)
	}

}

func (s *TestSuite) TestListBootStrapRegistries() {
	// Form the request URL
	requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/v3/registries", s.CatalogRESTServerUrl)

	// Make the HTTP GET request
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	// Add query parameters
	query := req.URL.Query()
	query.Add("orderBy", "name")
	query.Add("pageSize", "10")
	query.Add("offset", "0")
	query.Add("showSensitiveInfo", "true")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)

	var result struct {
		Registries []Registry `json:"registries"`
	}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Assert that the size of the result.Registries matches the size of getRegistries
	assert.Equal(s.T(), len(s.getRegistries()), len(result.Registries), "Mismatch in the number of registries")
	// Log registry details for debugging purposes
	for _, registry := range result.Registries {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, RootURL: %s, Type: %s",
			registry.Name, registry.DisplayName, registry.Description, registry.RootURL, registry.Type)
	}
}
func (s *TestSuite) TestVerifyBootstrappedRegistriesExist() {
	for _, registry := range s.getRegistries() {
		requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/v3/registries/%s", s.CatalogRESTServerUrl, registry.Name)

		req, err := http.NewRequest("GET", requestURL, nil)
		assert.NoError(s.T(), err)

		auth.AddRestAuthHeader(req, s.token, s.projectID)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(s.T(), err)
		s.Equal("200 OK", res.Status)

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			Registry Registry `json:"registry"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		assert.Equal(s.T(), registry.Name, result.Registry.Name, "Mismatch in 'Name' for registry: %s", registry.Name)
		assert.Equal(s.T(), registry.DisplayName, result.Registry.DisplayName, "Mismatch in 'DisplayName' for registry: %s", registry.Name)
		assert.Equal(s.T(), registry.RootURL, result.Registry.RootURL, "Mismatch in 'RootURL' for registry: %s", registry.Name)
		assert.Equal(s.T(), registry.Type, result.Registry.Type, "Mismatch in 'Type' for registry: %s", registry.Name)
		// assert.Equal(s.T(), registry.Description, result.Registry.Description)
	}
}

func (s *TestSuite) TestVerifyBootstrappedExtensionsExist() {
	for _, app := range s.getApplications() {
		requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/v3/applications/%s/versions/%s", s.CatalogRESTServerUrl,
			app.Name, app.Version)

		req, err := http.NewRequest("GET", requestURL, nil)
		assert.NoError(s.T(), err)

		auth.AddRestAuthHeader(req, s.token, s.projectID)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(s.T(), err)
		s.Equal("200 OK", res.Status)

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			Application Application `json:"application"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		assert.Equalf(s.T(), app.Name, result.Application.Name, "Mismatch in 'Name' for application: %s", app.Name)
		assert.Equalf(s.T(), app.DisplayName, result.Application.DisplayName, "Mismatch in 'DisplayName' for application: %s", app.Name)
		assert.Equalf(s.T(), app.ChartName, result.Application.ChartName, "Mismatch in 'ChartName' for application: %s", app.Name)
		assert.Equalf(s.T(), app.ChartVersion, result.Application.ChartVersion, "Mismatch in 'ChartVersion' for application: %s", app.Name)
		assert.Equalf(s.T(), app.Version, result.Application.Version, "Mismatch in 'Version' for application: %s", app.Name)
		assert.Equalf(s.T(), app.Kind, result.Application.Kind, "Mismatch in 'Kind' for application: %s", app.Name)
		assert.Equalf(s.T(), app.HelmRegistryName, result.Application.HelmRegistryName, "Mismatch in 'HelmRegistryName' for application: %s", app.Name)
		//assert.Equal(s.T(), app.Description, result.Application.Description)
	}
}

func (s *TestSuite) TestVerifyBootstrappedDeploymentPackagesExist() {
	for _, pkg := range s.getDeploymentPackages() {
		requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/v3/deployment_packages/%s/versions/%s", s.CatalogRESTServerUrl,
			pkg.Name, pkg.Version)

		req, err := http.NewRequest("GET", requestURL, nil)
		assert.NoError(s.T(), err)

		auth.AddRestAuthHeader(req, s.token, s.projectID)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(s.T(), err)
		s.Equal("200 OK", res.Status)

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			DeploymentPackage DeploymentPackages `json:"deploymentPackage"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		assert.Equalf(s.T(), pkg.Name, result.DeploymentPackage.Name, "Mismatch in 'Name' for deployment package: %s", pkg.Name)
		assert.Equalf(s.T(), pkg.Version, result.DeploymentPackage.Version, "Mismatch in 'Version' for deployment package: %s", pkg.Name)
		assert.Equalf(s.T(), pkg.Kind, result.DeploymentPackage.Kind, "Mismatch in 'Kind' for deployment package: %s", pkg.Name)
	}
}
