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
	"github.com/open-edge-platform/app-orch-catalog/pkg/exporter"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
	"gopkg.in/yaml.v2"
)

const (
	SchemaVersion         = "0.1"
	DollarSchema          = "https://schema.intel.com/catalog.orchestrator/0.1/schema"
	DefaultFilePermission = 0600
)

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

/*
func appendHeader(yaml []byte, kind string) []byte {
	header := fmt.Sprintf("---\nspecSchema: \"%s\"\nschemaVersion: \"%s\"\n$schema: \"%s\"\n\n", kind, SchemaVersion, DollarSchema)
	return append([]byte(header), yaml...)
}
*/

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
	//yamlData = appendHeader(yamlData, spec.SpecSchema)

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

	var values string
	if valuesFile != "" {
		content, err := os.ReadFile(valuesFile)
		if err != nil {
			return &InputError{Helm: helm, InputFile: valuesFile, Msg: "Failed to read values file", Err: err}
		}

		values = string(content)
	} else {
		values = "# this file intentionally left blank\n"
	}

	name, dp, app, registry, err := GenerateDeploymentPackageResources(helm, values, namespace, includeAuth)
	if err != nil {
		return err
	}

	e := exporter.NewExporter()
	err = e.ExportRegistry(*registry, fmt.Sprintf("%s/%s-registry.yaml", outputDir, registry.Name))
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: fmt.Sprintf("%s/%s-registry.yaml", outputDir, registry.Name), Msg: "Failed to export registry", Err: err}
	}
	err = e.ExportApplication(*app, fmt.Sprintf("%s/%s-application.yaml", outputDir, name), outputDir)
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: fmt.Sprintf("%s/%s-application.yaml", outputDir, name), Msg: "Failed to export application", Err: err}
	}
	err = e.ExportDeploymentPackage(*dp, fmt.Sprintf("%s/%s-deployment-package.yaml", outputDir, dp.Name))
	if err != nil {
		return &OutputError{Helm: helm, OutputDir: outputDir, OutputFile: fmt.Sprintf("%s/%s-deployment-package.yaml", outputDir, dp.Name), Msg: "Failed to export deployment package", Err: err}
	}

	return nil
}
