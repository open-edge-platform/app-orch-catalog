// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package dptohelm

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/internal/yamlreader"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
)

type Param struct {
	name  string
	value string
}

type DpToHelm struct {
	yamlreader.YamlReader
	Applications       []*catalogv3.Application
	DeploymentPackages []*catalogv3.DeploymentPackage
	Registries         []*catalogv3.Registry
	overrides          map[string]string
}

// ProcessFiles processes the given FileSet, loading its YAML files and
// adding the resulting objects to the YamlReader object.

func (d *DpToHelm) ProcessFiles(files yamlreader.FileSet) error {
	// Process each FileSet, load its yaml, sort, and add to the catalog.

	orderedSpecs, err := d.LoadYamlSpecs(files)
	if err != nil {
		return err
	}

	for _, spec := range orderedSpecs {
		switch spec.SpecSchema {
		case upload.DeploymentPackageType:
			dp, err := d.ReadDeploymentPackage(spec)
			if err != nil {
				return err
			}
			d.DeploymentPackages = append(d.DeploymentPackages, dp)
		case upload.DeploymentPackageLegacyType:
			dp, err := d.ReadDeploymentPackage(spec)
			if err != nil {
				return err
			}
			d.DeploymentPackages = append(d.DeploymentPackages, dp)
		case upload.ApplicationType:
			app, err := d.ReadApplication(spec, files) // application uses the FileSet to lookup profiles
			if err != nil {
				return err
			}
			d.Applications = append(d.Applications, app)
		case upload.RegistryType:
			reg, err := d.ReadRegistry(spec)
			if err != nil {
				return err
			}
			d.Registries = append(d.Registries, reg)
		case upload.ArtifactType:
			_, err = d.ReadArtifact(spec)
			if err != nil {
				return err
			}
			// we don't use these for anything
		default:
			return &InputError{Msg: fmt.Sprintf("unhandled type %s", spec.SpecSchema), InputFile: spec.FileName}
		}
	}

	return nil
}

func (d *DpToHelm) SetOverrides(rawOverrides []string) error {
	d.overrides = make(map[string]string)
	for _, override := range rawOverrides {
		parts := strings.Split(override, "=")
		if len(parts) != 2 {
			return &UsageError{Input: override, Msg: "invalid --set override format, expected <key>=<value>"}
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if name == "" || value == "" {
			return &UsageError{Input: override, Msg: "invalid --set override format, expected <key>=<value>"}
		}
		d.overrides[name] = value
	}
	return nil
}

func (d *DpToHelm) FindApp(name, version string) (*catalogv3.Application, error) {
	for _, app := range d.Applications {
		if app.Name == name && app.Version == version {
			return app, nil
		}
	}
	return nil, &NotFoundError{Msg: "Not Found", ObjectKind: "application", ObjectName: name, ObjectVersion: version}
}

func (d *DpToHelm) FindRegistry(name string) (*catalogv3.Registry, error) {
	for _, reg := range d.Registries {
		if reg.Name == name {
			return reg, nil
		}
	}
	return nil, &NotFoundError{Msg: "Not Found", ObjectKind: "registry", ObjectName: name}
}

func (d *DpToHelm) FindDeploymentProfile(dp *catalogv3.DeploymentPackage, name string) (*catalogv3.DeploymentProfile, error) {
	for _, profile := range dp.Profiles {
		if profile.Name == name {
			return profile, nil
		}
	}
	return nil, &NotFoundError{Msg: "Not Found", ObjectKind: "deployment profile", ObjectName: name}
}

func (d *DpToHelm) FindAppProfile(app *catalogv3.Application, name string) (*catalogv3.Profile, error) {
	for _, appProfile := range app.Profiles {
		if appProfile.Name == name {
			return appProfile, nil
		}
	}
	return nil, &NotFoundError{Msg: "Not Found", ObjectKind: "application profile", ObjectName: name, ApplicationName: app.Name}
}

func (d *DpToHelm) ApplyParameters(appProfile *catalogv3.Profile, allParams bool) ([]Param, error) {
	namedParams := make([]Param, 0)
	for _, param := range appProfile.ParameterTemplates {
		// if param was overridden on the command line, use that value
		override, ok := d.overrides[param.Name]
		if ok {
			namedParam := Param{
				name:  param.Name,
				value: override,
			}
			namedParams = append(namedParams, namedParam)
			continue
		}

		// Mandatory parameters only, unless the user asked for everything
		if !param.Mandatory && !allParams {
			continue
		}

		for {
			if param.Mandatory {
				fmt.Printf("(mandatory) ")
			}
			fmt.Printf("Parameter %s [%s]: ", param.Name, param.Default)
			reader := bufio.NewReader(os.Stdin)
			value, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("failed to read parameter %s: %v\n", param.Name, err)
				continue
			}
			value = value[:len(value)-1] // Trim the newline character
			if value == "" {
				value = param.Default
			}
			if value == "" && param.Mandatory {
				fmt.Printf("Parameter %s is mandatory, please provide a value\n", param.Name)
				continue
			}
			namedParam := Param{
				name:  param.Name,
				value: value,
			}
			namedParams = append(namedParams, namedParam)
			break
		}
	}
	return namedParams, nil
}

func (d *DpToHelm) GetHelmCommands(dp *catalogv3.DeploymentPackage, profileName string, allParams bool, outputDir string) ([]string, error) {
	if profileName == "" {
		profileName = dp.DefaultProfileName
	}
	profile, err := d.FindDeploymentProfile(dp, profileName)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			return nil, &OutputError{Msg: "failed to create output directory", OutputFile: outputDir, Err: err}
		}
	}

	fmt.Printf("# using deployment package profile: %s\n", profileName)

	cmds := make([]string, 0)
	for _, app := range dp.ApplicationReferences {
		app, err := d.FindApp(app.Name, app.Version)
		if err != nil {
			return nil, err
		}
		reg, err := d.FindRegistry(app.HelmRegistryName)
		if err != nil {
			return nil, err
		}
		var namespace string
		namespace, okay := dp.DefaultNamespaces[app.Name]
		if !okay {
			namespace = "default"
		}
		appProfileName := profile.ApplicationProfiles[app.Name]
		appProfile, err := d.FindAppProfile(app, appProfileName)
		if err != nil {
			return nil, err
		}
		namedParams, err := d.ApplyParameters(appProfile, allParams)
		if err != nil {
			return nil, err
		}
		valuesFileName := fmt.Sprintf("%s/%s-%s.yaml", outputDir, app.Name, profileName)
		err = os.WriteFile(valuesFileName, []byte(appProfile.ChartValues), 0600)
		if err != nil {
			return nil, &OutputError{Msg: "failed to write values file", OutputFile: valuesFileName, Err: err}
		}
		fmt.Printf("# created values file %s for app %s profile %s\n", valuesFileName, app.Name, appProfileName)
		url := fmt.Sprintf("%s/%s", reg.RootUrl, app.ChartName)
		helmCmd := fmt.Sprintf("helm install %s %s --version %s --namespace %s -f %s", app.Name, url, app.ChartVersion, namespace, valuesFileName)
		for _, param := range namedParams {
			helmCmd += fmt.Sprintf(" --set %s=\"%s\"", param.name, param.value)
		}
		if namespace != "default" {
			cmds = append(cmds, fmt.Sprintf("kubectl create namespace %s || true", namespace))
		}
		cmds = append(cmds, helmCmd)
	}

	return cmds, nil
}
