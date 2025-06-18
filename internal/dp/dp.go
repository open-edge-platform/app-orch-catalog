// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dp

import (
	"fmt"
	"os"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/internal/helm"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/exporter"
	"gopkg.in/yaml.v2"
)

const (
	SchemaVersion         = "0.1"
	DollarSchema          = "https://schema.intel.com/catalog.orchestrator/0.1/schema"
	DefaultFilePermission = 0600
)

func GenerateDeploymentPackageResources(helm helm.HelmInfo, values string, namespace string, includeAuth bool) (string, *catalogv3.DeploymentPackage, *catalogv3.Application, *catalogv3.Registry, error) {
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

	// Ensure values contains yaml
	var yamlContent map[string]interface{}
	err := yaml.Unmarshal([]byte(values), &yamlContent)
	if err != nil {
		return "", nil, nil, nil, &InputError{Helm: helm, Msg: "Invalid YAML content in values file", Err: err}
	}

	registryName := name + "-registry"

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
		DefaultProfileName: "default",
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
				ApplicationProfiles: map[string]string{name: "default"},
			},
		},
		DefaultProfileName: "default",
	}

	if namespace != "" {
		dp.DefaultNamespaces = map[string]string{name: namespace}
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

// GetValuesFromFile reads a values.yaml file and returns its content as a string.
// It performs validation to ensure the content is valid YAML.
func GetValuesFromFile(valuesFile string) (string, error) {
	content, err := os.ReadFile(valuesFile)
	if err != nil {
		return "", &InputError{InputFile: valuesFile, Msg: "Failed to read values file", Err: err}
	}

	// Ensure the content is valid YAML
	var yamlContent map[string]interface{}
	err = yaml.Unmarshal(content, &yamlContent)
	if err != nil {
		return "", &InputError{InputFile: valuesFile, Msg: "Invalid YAML content in values file", Err: err}
	}

	return string(content), nil
}

// GetValuesFromChart uses the values.yaml from a chart, performing validation to ensure it is valid YAML.
func GetValuesFromChart(helm helm.HelmInfo) (string, error) {
	if helm.Values == nil {
		return "", &InputError{Helm: helm, Msg: "Default values requested but no values provided in Helm chart"}
	}

	// Ensure the content is valid YAML
	var yamlContent map[string]interface{}
	err := yaml.Unmarshal(*helm.Values, &yamlContent)
	if err != nil {
		return "", &InputError{Helm: helm, Msg: "Invalid YAML content in Helm chart values", Err: err}
	}

	return string(*helm.Values), nil
}

// GenerateDefaultParametersFromYaml is the helper function for GenerateDefaultParameters. It works on a yaml tree that
// has already been parsed into a map[interface{}]interface{}.
func GenerateDefaultParametersFromYaml(parent string, yamlContent map[interface{}]interface{}) ([]*catalogv3.ParameterTemplate, error) {
	pts := make([]*catalogv3.ParameterTemplate, 0)
	for keyInterface := range yamlContent {
		key, ok := keyInterface.(string)
		if !ok {
			continue
		}
		var fullName string
		if parent != "" {
			fullName = fmt.Sprintf("%s.%s", parent, key)
		} else {
			fullName = key
		}
		if svalue, ok := yamlContent[key].(string); ok {
			if len(svalue) >= 100 || strings.Contains(svalue, "{{") || strings.Contains(svalue, "\n") {
				svalue = ""
			}
			pt := &catalogv3.ParameterTemplate{
				Name:        fullName,
				DisplayName: fullName,
				Default:     svalue,
				Type:        "string",
				Secret:      false,
				Mandatory:   false,
			}
			pts = append(pts, pt)
		} else if mvalue, ok := yamlContent[key].(map[interface{}]interface{}); ok {
			// Recursively process nested maps
			thisPts, err := GenerateDefaultParametersFromYaml(fullName, mvalue)
			if err != nil {
				return nil, err
			}
			pts = append(pts, thisPts...)
		}
	}

	return pts, nil
}

// GenerateDefaultParametersFromYaml generates parameter templates from a values.yaml file.
// It does this by recursively parsing the YAML. It builds up names in dotted notations and also supplies the default value.
// If a default value is problematic (e.g., too long, contains templating syntax, or has newlines), then no default will be used.
// For some charts, like Bitnami Charts, this may result in a very large number of parameters, so this should be used with care.
func GenerateDefaultParameters(values string) ([]*catalogv3.ParameterTemplate, error) {
	var yamlContent map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(values), &yamlContent)
	if err != nil {
		// This probably can't happen, because we probably already ensured it was valid YAML
		return nil, fmt.Errorf("Invalid YAML content in values: %v", err)
	}
	return GenerateDefaultParametersFromYaml("", yamlContent)
}

// GenerateDeploytmentPackage generates a deployment package from a Helm chart.
func GenerateDeploymentPackage(helm helm.HelmInfo, valuesFile string, outputDir string, namespace string, includeAuth bool, includeDefaultValues bool, includeDefaultParameters bool) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, Msg: "Failed to create output directory", Err: err}
	}

	var values string
	if includeDefaultValues && valuesFile != "" {
		return &InputError{Helm: helm, InputFile: valuesFile, Msg: "Cannot specify both to use chart values and a values file"}
	}

	if includeDefaultValues {
		values, err = GetValuesFromChart(helm)
		if err != nil {
			return err
		}
	} else if valuesFile != "" {
		values, err = GetValuesFromFile(valuesFile)
		if err != nil {
			return err
		}
	} else {
		values = "# this file intentionally left blank\n"
	}

	name, dp, app, registry, err := GenerateDeploymentPackageResources(helm, values, namespace, includeAuth)
	if err != nil {
		return err
	}

	if includeDefaultParameters {
		pts, err := GenerateDefaultParameters(values)
		if err != nil {
			return err
		}
		app.Profiles[0].ParameterTemplates = pts
	}

	e := exporter.NewExporter()
	fileName := fmt.Sprintf("%s/%s.yaml", outputDir, registry.Name)
	data, err := e.ExportRegistry(registry) // note: registry already has "-registry" in the name
	if err == nil {
		err = os.WriteFile(fileName, data, DefaultFilePermission)
	}
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: fileName, Msg: "Failed to export registry", Err: err}
	}

	fileName = fmt.Sprintf("%s/%s-application.yaml", outputDir, name)
	data, profileData, err := e.ExportApplication(app)
	if err == nil {
		err = os.WriteFile(fileName, data, DefaultFilePermission)
	}
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: fileName, Msg: "Failed to export application", Err: err}
	}

	for profileFileName, profileContent := range profileData {
		profileFilePathName := fmt.Sprintf("%s/%s", outputDir, profileFileName)
		err = os.WriteFile(profileFilePathName, profileContent, DefaultFilePermission)
		if err != nil {
			return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: profileFilePathName, Msg: "Failed to export profile", Err: err}
		}
	}

	fileName = fmt.Sprintf("%s/%s-deployment-package.yaml", outputDir, dp.Name)
	data, err = e.ExportDeploymentPackage(dp)
	if err == nil {
		err = os.WriteFile(fileName, data, DefaultFilePermission)
	}
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: fileName, Msg: "Failed to export deployment package", Err: err}
	}

	verboseerror.Infof("Created deployment package for %s in directory '%s'\n", name, outputDir)

	return nil
}
