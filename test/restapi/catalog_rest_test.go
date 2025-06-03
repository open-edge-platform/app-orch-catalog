// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restapi

import (
	// Standard library imports
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	// Third-party imports

	"github.com/stretchr/testify/assert"

	// Project-specific imports
	"github.com/open-edge-platform/app-orch-catalog/test/auth"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

const applicationsEndpoint = "/catalog.orchestrator.apis/v3/applications"
const deploymentPackagesEndpoint = "/catalog.orchestrator.apis/v3/deployment_packages"
const registriesEndpoint = "/catalog.orchestrator.apis/v3/registries"
const uploadEndpoint = "/catalog.orchestrator.apis/upload"

/* The reason for these Short* objects was to facilitate converting the existing
 * rest api tests to the more complex test framework that came from the mage e2e
 * tests, which have many more fields.
 */

type ShortRegistry struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	RootURL     string `json:"rootUrl"`
	Type        string `json:"type"`
}

func (s *TestSuite) getRegistries() []Registry {
	dockerURL := fmt.Sprintf("https://registry-oci.%s/", s.orchDomain)
	helmURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", s.orchDomain)

	regs := []Registry{}
	for _, ra := range []ShortRegistry{
		{"akri-helm-registry", "akri-helm-registry", "Public registry for akri chart", "https://project-akri.github.io/akri/", "HELM"},
		{"bitnami-helm-oci", "bitnami-helm-oci", "Bitnami helm registry", "oci://registry-1.docker.io/bitnamicharts", "HELM"},
		{"fluent-bit", "fluent-bit", "Public registry for fluent bit chart", "https://fluent.github.io/helm-charts", "HELM"},
		{"gatekeeper", "gatekeeper", "Public registry for gatekeeper chart", "https://open-policy-agent.github.io/gatekeeper/charts", "HELM"},
		{"harbor-docker-oci", "harbor oci docker", "Harbor OCI docker images registry", dockerURL, "IMAGE"},
		{"harbor-helm-oci", "harbor oci helm", "Harbor OCI helm charts registry", helmURL, "HELM"},
		{"intel-github-io", "intel-github-io", "Intel Public registry with device operator & plugins", "https://intel.github.io/helm-charts", "HELM"},
		{"intel-rs-helm", "intel-rs-helm", "Repo on registry registry-rs.edgeorchestration.intel.com", "oci://rs-proxy.orch-platform.svc.cluster.local:8443", "HELM"},
		{"intel-rs-images", "intel-rs-image", "Repo on registry registry-rs.edgeorchestration.intel.com", "oci://registry-rs.edgeorchestration.intel.com", "IMAGE"},
		{"jetstack", "jetstack", "Public registry for cert manager chart", "https://charts.jetstack.io", "HELM"},
	} {
		regs = append(regs, Registry{
			Name:        ra.Name,
			DisplayName: ra.DisplayName,
			Description: ra.Description,
			RootURL:     ra.RootURL,
			Type:        ra.Type,
		})
	}

	return regs
}

type ShortApplication struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	Description      string `json:"description"`
	Kind             string `json:"kind"`
	ChartName        string `json:"chartName"`
	HelmRegistryName string `json:"helmRegistryName"`
}

/*
If there is change in the versions, you can verify the list by executing the function TestListBootStrapExtensions and
update the version information here
*/
func (s *TestSuite) getApplications() []Application {
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

type ShortDeploymentPackage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

/*
If there is change in the versions, you can verify the list by executing the function TestListBootStrapDeploymentPackages and
update the version information here
*/
func (s *TestSuite) getDeploymentPackages() []DeploymentPackage {
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

func (s *TestSuite) TestListBootStrapExtensions() {
	requestURL := fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, applicationsEndpoint)
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)
	auth.AddRestAuthHeader(req, s.token, s.projectID)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	assert.NoError(s.T(), err)
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)

	var result struct {
		Applications []Application `json:"applications"`
	}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), len(s.getApplications()), len(result.Applications), "Mismatch in the number of applications")
	// Log application details for debugging purposes
	log.Printf("Extensions:")
	for _, app := range result.Applications {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, Version: %s, Kind: %s, ChartName: %s, ChartVersion: %s, HelmRegistryName: %s",
			app.Name, app.DisplayName, app.Description, app.Version, app.Kind, app.ChartName, app.ChartVersion, app.HelmRegistryName)
	}

}

func (s *TestSuite) TestListBootStrapDeploymentPackages() {
	requestURL := fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, deploymentPackagesEndpoint)
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
		DeploymentPackages []DeploymentPackage `json:"deploymentPackages"`
	}
	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), len(s.getDeploymentPackages()), len(result.DeploymentPackages), "Mismatch in the number of deployment packages")

	// Log deployment package details for debugging purposes
	log.Printf("Deployment Packages:")
	for _, pkg := range result.DeploymentPackages {
		log.Printf("Name: %s, Description: %s, Version: %s, Kind: %s",
			pkg.Name, pkg.Description, pkg.Version, pkg.Kind)
	}

}

func (s *TestSuite) TestListBootStrapRegistries() {
	requestURL := fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, registriesEndpoint)
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
	defer res.Body.Close()

	if res.Status != "200 OK" {
		s.Equal("200 OK", res.Status)
		return // Everything else is going to fail...
	}

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
	log.Printf("Registries:")
	for _, registry := range result.Registries {
		log.Printf("Name: %s, DisplayName: %s, Description: %s, RootURL: %s, Type: %s",
			registry.Name, registry.DisplayName, registry.Description, registry.RootURL, registry.Type)
	}
}

func (s *TestSuite) TestVerifyBootstrappedRegistriesExist() {
	for _, registry := range s.getRegistries() {
		requestURL := fmt.Sprintf("%s%s/%s", s.CatalogRESTServerUrl, registriesEndpoint, registry.Name)
		req, err := http.NewRequest("GET", requestURL, nil)
		assert.NoError(s.T(), err)
		auth.AddRestAuthHeader(req, s.token, s.projectID)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(s.T(), err)
		defer res.Body.Close()
		if res.Status != "200 OK" {
			assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for Registry: %s", registry.Name)
			continue
		}

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			Registry Registry `json:"registry"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		switch {
		case registry.Name != result.Registry.Name:
			assert.Equal(s.T(), registry.Name, result.Registry.Name, "Mismatch in 'Name' for registry: %s", registry.Name)
		case registry.DisplayName != result.Registry.DisplayName:
			assert.Equal(s.T(), registry.DisplayName, result.Registry.DisplayName, "Mismatch in 'DisplayName' for registry: %s", registry.Name)
		case registry.RootURL != result.Registry.RootURL:
			oldDockerURL := fmt.Sprintf("https://registry-oci.%s/", s.orchDomain)
			newDockerURL := fmt.Sprintf("oci://registry-oci.%s/catalog-apps-sample-org-sample-project", s.orchDomain)
			// Docker Registry URL was changed recently. Avoid throwing errors in a development environment that's using the new URL
			// TODO: remove this special case when component-tests are moved forward
			if registry.RootURL != oldDockerURL || result.Registry.RootURL != newDockerURL {
				assert.Equal(s.T(), registry.RootURL, result.Registry.RootURL, "Mismatch in 'RootURL' for registry: %s", registry.Name)
			}
		case registry.Type != result.Registry.Type:
			assert.Equal(s.T(), registry.Type, result.Registry.Type, "Mismatch in 'Type' for registry: %s", registry.Name)
		}
		// assert.Equal(s.T(), registry.Description, result.Registry.Description)
	}
}

func (s *TestSuite) TestVerifyBootstrappedExtensionsExist() {
	for _, app := range s.getApplications() {
		requestURL := fmt.Sprintf("%s%s/%s/versions", s.CatalogRESTServerUrl,
			applicationsEndpoint, app.Name)

		req, err := http.NewRequest("GET", requestURL, nil)
		assert.NoError(s.T(), err)

		auth.AddRestAuthHeader(req, s.token, s.projectID)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(s.T(), err)
		defer res.Body.Close()
		if res.Status != "200 OK" {
			assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for application: %s - %s", app.Name, requestURL)
			continue
		}

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			Application []Application `json:"application"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		s.True(len(result.Application) > 0, "Expected at least one application for %s", app.Name)

		if len(result.Application) > 0 {
			gotApp := result.Application[0]

			switch {
			case app.Name != gotApp.Name:
				assert.Equalf(s.T(), app.Name, gotApp.Name, "Mismatch in 'Name' for application: %s", app.Name)
			case app.DisplayName != gotApp.DisplayName:
				assert.Equalf(s.T(), app.DisplayName, gotApp.DisplayName, "Mismatch in 'DisplayName' for application: %s", app.Name)
			case app.ChartName != gotApp.ChartName:
				assert.Equalf(s.T(), app.ChartName, gotApp.ChartName, "Mismatch in 'ChartName' for application: %s", app.Name)
			case app.Kind != gotApp.Kind:
				assert.Equalf(s.T(), app.Kind, gotApp.Kind, "Mismatch in 'Kind' for application: %s", app.Name)
			case app.HelmRegistryName != gotApp.HelmRegistryName:
				assert.Equalf(s.T(), app.HelmRegistryName, gotApp.HelmRegistryName, "Mismatch in 'HelmRegistryName' for application: %s", app.Name)
			}
			//assert.Equal(s.T(), app.Description, result.Application.Description)
		}
	}
}

func (s *TestSuite) TestVerifyBootstrappedDeploymentPackagesExist() {
	for _, pkg := range s.getDeploymentPackages() {
		requestURL := fmt.Sprintf("%s%s/%s/versions", s.CatalogRESTServerUrl,
			deploymentPackagesEndpoint, pkg.Name)

		req, err := http.NewRequest("GET", requestURL, nil)
		assert.NoError(s.T(), err)

		auth.AddRestAuthHeader(req, s.token, s.projectID)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(s.T(), err)
		defer res.Body.Close()
		if res.Status != "200 OK" {
			assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for Package: %s", pkg.Name)
			continue // Everything else is going to fail...
		}

		body, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)

		var result struct {
			DeploymentPackage []DeploymentPackage `json:"deploymentPackages"`
		}
		err = json.Unmarshal(body, &result)
		assert.NoError(s.T(), err)

		s.True(len(result.DeploymentPackage) > 0, "Expected at least one deployment package for %s", pkg.Name)
		if len(result.DeploymentPackage) > 0 {
			gotPkg := result.DeploymentPackage[0]

			switch {
			case pkg.Name != gotPkg.Name:
				assert.Equalf(s.T(), pkg.Name, gotPkg.Name, "Mismatch in 'Name' for deployment package: %s", pkg.Name)
			case pkg.Kind != gotPkg.Kind:
				assert.Equalf(s.T(), pkg.Kind, gotPkg.Kind, "Mismatch in 'Kind' for deployment package: %s", pkg.Name)
			}
		}
	}
}

func (s *TestSuite) Delete(url string) {
	req, err := http.NewRequest("DELETE", url, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	if res.Status != "200 OK" {
		assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for delete on url %s", url)
	}
}

func (s *TestSuite) TestUploadTarball() {

	file, err := os.Open("../testdata/wordpress.tar.gz")
	assert.NoError(s.T(), err)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", "wordpress.tar.gz")
	_, err = io.Copy(part, file)
	assert.NoError(s.T(), err)
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, uploadEndpoint), body)
	assert.NoError(s.T(), err)

	req.Header.Add("Content-Type", writer.FormDataContentType())

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for upload")
	if res.Status != "200 OK" {
		// print response message if something has gone wrong, for debugging
		bodyBytes, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)
		log.Printf("Response Body: %s", string(bodyBytes))
	}

	// Make sure the wordpress DP was created

	requestURL := fmt.Sprintf("%s%s/test-wordpress/versions/0.1.1", s.CatalogRESTServerUrl, deploymentPackagesEndpoint)
	req, err = http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	// Add query parameters
	query := req.URL.Query()
	query.Add("orderBy", "name")
	query.Add("pageSize", "10")
	query.Add("offset", "0")
	req.URL.RawQuery = query.Encode()

	res, err = http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	s.Equal("200 OK", res.Status)

	resBody, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	var result struct {
		DeploymentPackage DeploymentPackage `json:"deploymentPackage"`
	}
	err = json.Unmarshal(resBody, &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test-wordpress", result.DeploymentPackage.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), "0.1.1", result.DeploymentPackage.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.Delete(fmt.Sprintf("%s%s/test-wordpress/versions/0.1.1", s.CatalogRESTServerUrl, deploymentPackagesEndpoint))
	s.Delete(fmt.Sprintf("%s%s/test-wordpress/versions/0.1.1", s.CatalogRESTServerUrl, applicationsEndpoint))
	s.Delete(fmt.Sprintf("%s%s/test-bitnami", s.CatalogRESTServerUrl, registriesEndpoint))
}

func (s *TestSuite) TestUploadSeparateFiles() {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	pathNames := []string{"../testdata/wordpress/app-wordpress-0.1.1.yaml",
		"../testdata/wordpress/dp-wordpress-0.1.1.yaml",
		"../testdata/wordpress/registry-bitnami.yaml",
		"../testdata/wordpress/values-wordpress-0.1.1.yaml",
	}

	for _, pathName := range pathNames {
		file, err := os.Open(pathName)
		assert.NoError(s.T(), err)
		defer file.Close()

		fileName := pathName[strings.LastIndex(pathName, "/")+1:]

		part, _ := writer.CreateFormFile("files", fileName)
		_, err = io.Copy(part, file)
		assert.NoError(s.T(), err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", s.CatalogRESTServerUrl, uploadEndpoint), body)
	assert.NoError(s.T(), err)

	req.Header.Add("Content-Type", writer.FormDataContentType())

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	assert.Equalf(s.T(), "200 OK", res.Status, "Mismatch in 'Response' for upload")
	if res.Status != "200 OK" {
		// print response message if something has gone wrong, for debugging
		bodyBytes, err := io.ReadAll(res.Body)
		assert.NoError(s.T(), err)
		log.Printf("Response Body: %s", string(bodyBytes))
	}

	// Make sure the wordpress DP was created

	requestURL := fmt.Sprintf("%s%s/test-wordpress/versions/0.1.1", s.CatalogRESTServerUrl, deploymentPackagesEndpoint)
	req, err = http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	// Add query parameters
	query := req.URL.Query()
	query.Add("orderBy", "name")
	query.Add("pageSize", "10")
	query.Add("offset", "0")
	req.URL.RawQuery = query.Encode()

	res, err = http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	s.Equal("200 OK", res.Status)

	resBody, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)
	var result struct {
		DeploymentPackage DeploymentPackage `json:"deploymentPackage"`
	}
	err = json.Unmarshal(resBody, &result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test-wordpress", result.DeploymentPackage.Name, "Mismatch in the name of the deployment package")
	assert.Equal(s.T(), "0.1.1", result.DeploymentPackage.Version, "Mismatch in the version of the deployment package")

	// Note: Not verifying the application or registry, as the DP would fail without them

	// Cleanup
	s.Delete(fmt.Sprintf("%s%s/test-wordpress/versions/0.1.1", s.CatalogRESTServerUrl, deploymentPackagesEndpoint))
	s.Delete(fmt.Sprintf("%s%s/test-wordpress/versions/0.1.1", s.CatalogRESTServerUrl, applicationsEndpoint))
	s.Delete(fmt.Sprintf("%s%s/test-bitnami", s.CatalogRESTServerUrl, registriesEndpoint))
}

func (s *TestSuite) TestGetCharts() {
	requestURL := fmt.Sprintf("%s/catalog.orchestrator.apis/charts?registry=harbor-helm-oci", s.CatalogRESTServerUrl)
	req, err := http.NewRequest("GET", requestURL, nil)
	assert.NoError(s.T(), err)

	auth.AddRestAuthHeader(req, s.token, s.projectID)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(s.T(), err)
	defer res.Body.Close()
	s.Equal("200 OK", res.Status)

	body, err := io.ReadAll(res.Body)
	assert.NoError(s.T(), err)

	// On a fresh orchestrator there should be no charts in the registry

	assert.Equal(s.T(), "null", string(body), "Expected the response body to be empty")
}
