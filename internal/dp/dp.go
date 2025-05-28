// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dp

import (
	"fmt"
	"os"

	"github.com/open-edge-platform/app-orch-catalog/internal/helm"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
	"gopkg.in/yaml.v2"
)

const (
	SchemaVersion         = "0.1"
	DollarSchema          = "https://schema.intel.com/catalog.orchestrator/0.1/schema"
	DefaultFilePermission = 0600
)

// Registry is a Catalog Registry Object

type Registry struct {
	Name        string
	Description string
	Type        string
	RootURL     string `yaml:"rootUrl"`
	Username    string `yaml:"userName"`
	Authtoken   string `yaml:"authToken"`
}

// Profile is a Catalog profile in an application

type Profile struct {
	Name           string
	ValuesFileName string `yaml:"valuesFileName"`
}

// Application is a Catalog Application

type Application struct {
	Name         string
	Version      string
	Description  string
	HelmRegistry string `yaml:"helmRegistry"`
	ChartName    string `yaml:"chartName"`
	ChartVersion string `yaml:"chartVersion"`
	Profiles     []Profile
}

// AppProfileLink is the link between an Application and a Profile

type AppProfileLink struct {
	Application string
	Profile     string
}

// DeploymentProfile is a Deployment Package Profile

type DeploymentProfile struct {
	Name                string
	ApplicationProfiles []AppProfileLink `yaml:"applicationProfiles"`
}

// ApplicationLink is the link between a Deployment Package and an Application

type ApplicationLink struct {
	Name    string
	Version string
}

// DeploymentPackage is a Catalog Deployment Package

type DeploymentPackage struct {
	Name               string
	Version            string
	Description        string
	Applications       []ApplicationLink
	DeploymentProfiles []DeploymentProfile `yaml:"deploymentProfiles"`
	DefaultNamespaces  map[string]string   `yaml:"defaultNamespaces,omitempty"`
}

func appendHeader(yaml []byte, kind string) []byte {
	header := fmt.Sprintf("---\nspecSchema: \"%s\"\nschemaVersion: \"%s\"\n$schema: \"%s\"\n\n", kind, SchemaVersion, DollarSchema)
	return append([]byte(header), yaml...)
}

func GenerateDeploymentPackageResources(helm helm.HelmInfo, valuesFile string, outputDir string, namespace string, includeAuth bool) (string, *catalogv3.DeploymentPackage, *catalogv3.Application, *catalogv3.Registry, error) {
	name := helm.Name
	if len(name) > 30 {
		newName := name[:25]
		// make sure it does not end in an illegal character
		for len(newName) > 0 && name[len(newName)-1] == '-' {
			newName = newName[:len(newName)-1]
		}
		verboseerror.Infof("Truncating deployment package name from %s to %s\n", name, newName)
		name = newName
	}

	registryName := name + "-registry"

	var values string
	if valuesFile != "" {
		content, err := os.ReadFile(valuesFile)
		if err != nil {
			return "", nil, nil, nil, &InputError{Helm: helm, InputFile: valuesFile, Msg: "Failed to read values file", Err: err}
		}

		var yamlContent map[string]interface{}
		err = yaml.Unmarshal(content, &yamlContent)
		if err != nil {
			return "", nil, nil, nil, &InputError{Helm: helm, InputFile: valuesFile, Msg: "Invalid YAML content in values file", Err: err}
		}

		values = string(content)
	} else {
		values = "# this file intentionally left blank\n"
	}

	app := &catalogv3.Application{
		Name:             name,
		Version:          helm.Version,
		Description:      helm.Description,
		HelmRegistryName: registryName,
		ChartName:        helm.Name,
		ChartVersion:     helm.Version,
		Profiles: []*catalogv3.Profile{
			{
				Name:        "default",
				ChartValues: values,
			},
		},
	}

	dp := &catalogv3.DeploymentPackage{
		Name:    name,
		Version: helm.Version,
		ApplicationReferences: []*catalogv3.ApplicationReference{
			{
				Name:    name,
				Version: helm.Version,
			},
		},
		Profiles: []*catalogv3.DeploymentProfile{
			{
				Name:                "default",
				ApplicationProfiles: map[string]string{"default": name},
			},
		},
	}

	registry := &catalogv3.Registry{
		Name:        registryName,
		Description: "OCI registry for " + name,
		Type:        "HELM",
		RootUrl:     helm.OCIRegistry,
	}
	if includeAuth && helm.Username != "" && helm.Password != "" {
		verboseerror.Infof("NOTE: Username and password have been added to registry object.\n")
		verboseerror.Infof("      Please ensure that the deployment package is stored securely.\n")
		registry.Username = helm.Username
		registry.AuthToken = helm.Password
	}

	return name, dp, app, registry, nil
}

func SaveSpec(spec *upload.YamlSpec, outputFile string) error {
	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		return &GenerationError{Msg: "Failed to marshal spec to YAML", Err: err}
	}
	yamlData = appendHeader(yamlData, spec.SpecSchema)

	err = os.WriteFile(outputFile, yamlData, DefaultFilePermission)
	if err != nil {
		return &OutputError{OutputFile: outputFile, Msg: "Failed to write spec YAML to file", Err: err}
	}

	return nil
}

func GenerateDeploymentPackage(helm helm.HelmInfo, valuesFile string, outputDir string, namespace string, includeAuth bool) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to create output directory", Err: err}
	}

	name, dp, app, registry, err := GenerateDeploymentPackageResources(helm, valuesFile, outputDir, namespace, includeAuth)
	if err != nil {
		return err
	}

	valuesFileName := name + "-values.yaml"

	appSpec := &upload.YamlSpec{
		SpecSchema:   upload.RegistryType,
		Name:         app.Name,
		Version:      app.Version,
		Description:  app.Description,
		HelmRegistry: app.HelmRegistryName,
		ChartName:    app.ChartName,
		ChartVersion: app.ChartVersion,
	}
	outputFile := fmt.Sprintf("%s/%s-application.yaml", outputDir, name)
	err = SaveSpec(appSpec, outputFile)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: outputFile, Msg: "Failed to write application YAML to file", Err: err}
	}

	appRefs := make([]upload.Application, 0, len(dp.ApplicationReferences))
	for _, appRef := range dp.ApplicationReferences {
		appRefs = append(appRefs, upload.Application{
			Name:    appRef.Name,
			Version: appRef.Version,
		})
	}

	dpProfiles := make([]upload.DeploymentProfile, 0, len(dp.Profiles))
	for _, profile := range dp.Profiles {
		appProfileLinks := make([]upload.ApplicationProfile, 0, len(profile.ApplicationProfiles))
		for appName, profileName := range profile.ApplicationProfiles {
			appProfileLinks = append(appProfileLinks, upload.ApplicationProfile{
				ApplicationName: appName,
				ProfileName:     profileName,
			})
		}
		dpProfiles = append(dpProfiles, upload.DeploymentProfile{
			Name:                profile.Name,
			ApplicationProfiles: appProfileLinks,
		})
	}

	dpSpec := &upload.YamlSpec{
		SpecSchema:         upload.DeploymentPackageType,
		Name:               dp.Name,
		Version:            dp.Version,
		Description:        dp.Description,
		Applications:       appRefs,
		DeploymentProfiles: dpProfiles,
		DefaultNamespaces:  dp.DefaultNamespaces,
	}
	outputFile = fmt.Sprintf("%s/%s-deployment-package.yaml", outputDir, dp.Name)
	err = SaveSpec(dpSpec, outputFile)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: outputFile, Msg: "Failed to write deployment package YAML to file", Err: err}
	}

	registrySpec := &upload.YamlSpec{
		SpecSchema:  upload.RegistryType,
		Name:        registry.Name,
		Description: registry.Description,
		Type:        registry.Type,
		RootURL:     registry.RootUrl,
		UserName:    registry.Username,
		AuthToken:   registry.AuthToken,
	}
	outputFile = fmt.Sprintf("%s/%s-registry.yaml", outputDir, registry.Name)
	err = SaveSpec(registrySpec, outputFile)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: outputFile, Msg: "Failed to write registry YAML to file", Err: err}
	}

	outputFile = fmt.Sprintf("%s/%s", outputDir, valuesFileName)
	err = os.WriteFile(outputFile, []byte(app.Profiles[0].ChartValues), DefaultFilePermission)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: outputFile, Msg: "Failed to write values file to output directory", Err: err}
	}

	return nil
}

// Generate the Deployment Package and write it to the output directory

func GenerateDeploymentPackageOld(helm helm.HelmInfo, valuesFile string, outputDir string, namespace string, includeAuth bool) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to create output directory", Err: err}
	}

	name := helm.Name
	if len(name) > 30 {
		newName := name[:25]
		// make sure it does not end in an illegal character
		for len(newName) > 0 && name[len(newName)-1] == '-' {
			newName = newName[:len(newName)-1]
		}
		verboseerror.Infof("Truncating deployment package name from %s to %s\n", name, newName)
		name = newName
	}

	registryName := name + "-registry"
	valuesFileName := name + "-values.yaml"

	app := Application{
		Name:         name,
		Version:      helm.Version,
		Description:  helm.Description,
		HelmRegistry: registryName,
		ChartName:    helm.Name,
		ChartVersion: helm.Version,
		Profiles: []Profile{
			{
				Name:           "default",
				ValuesFileName: valuesFileName,
			},
		},
	}
	appYaml, err := yaml.Marshal(app)
	if err != nil {
		return &GenerationError{Helm: helm, Msg: "Failed to marshal application to YAML", Err: err}
	}
	appYaml = appendHeader(appYaml, "Application")

	outputFile := fmt.Sprintf("%s/%s-application.yaml", outputDir, name)
	err = os.WriteFile(outputFile, appYaml, DefaultFilePermission)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to create output directory", Err: err}
	}

	dp := DeploymentPackage{
		Name:    name,
		Version: helm.Version,
		Applications: []ApplicationLink{
			{
				Name:    name,
				Version: helm.Version,
			},
		},
		DeploymentProfiles: []DeploymentProfile{
			{
				Name: "default",
				ApplicationProfiles: []AppProfileLink{
					{
						Application: name,
						Profile:     "default",
					},
				},
			},
		},
	}

	if namespace != "" {
		dp.DefaultNamespaces = map[string]string{name: namespace}
	}

	dpYaml, err := yaml.Marshal(dp)
	if err != nil {
		return &GenerationError{Helm: helm, Msg: "Failed to marshal application to YAML", Err: err}
	}
	dpYaml = appendHeader(dpYaml, "DeploymentPackage")

	outputFile = fmt.Sprintf("%s/%s-deployment-package.yaml", outputDir, name)
	err = os.WriteFile(outputFile, dpYaml, DefaultFilePermission)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to write deployment package YAML to file", Err: err}
	}

	registry := Registry{
		Name:        registryName,
		Description: "OCI registry for " + name,
		Type:        "HELM",
		RootURL:     helm.OCIRegistry,
	}
	if includeAuth && helm.Username != "" && helm.Password != "" {
		verboseerror.Infof("NOTE: Username and password have been added to registry object.\n")
		verboseerror.Infof("      Please ensure that the deployment package is stored securely.\n")
		registry.Username = helm.Username
		registry.Authtoken = helm.Password
	}
	registryYaml, err := yaml.Marshal(registry)
	if err != nil {
		return &GenerationError{Helm: helm, Msg: "Failed to marshal registry to YAML", Err: err}
	}
	registryYaml = appendHeader(registryYaml, "Registry")

	outputFile = fmt.Sprintf("%s/%s-registry.yaml", outputDir, name)
	err = os.WriteFile(outputFile, registryYaml, DefaultFilePermission)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to write registry YAML to file", Err: err}
	}

	if valuesFile != "" {
		content, err := os.ReadFile(valuesFile)
		if err != nil {
			return &InputError{Helm: helm, InputFile: valuesFile, Msg: "Failed to read values file", Err: err}
		}

		var yamlContent map[string]interface{}
		err = yaml.Unmarshal(content, &yamlContent)
		if err != nil {
			return &InputError{Helm: helm, InputFile: valuesFile, Msg: "Invalid YAML content in values file", Err: err}
		}

		outputFile = fmt.Sprintf("%s/%s", outputDir, valuesFileName)
		err = os.WriteFile(outputFile, content, DefaultFilePermission)
		if err != nil {
			return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to write values file to output directory", Err: err}
		}
	} else {
		outputFile = fmt.Sprintf("%s/%s", outputDir, valuesFileName)
		err = os.WriteFile(outputFile, []byte("# this file intentionally left blank\n"), DefaultFilePermission)
		if err != nil {
			return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: outputFile, Msg: "Failed to write values file to output directory", Err: err}
		}
	}

	verboseerror.Infof("Deployment package saved to %s\n", outputDir)

	return nil
}
