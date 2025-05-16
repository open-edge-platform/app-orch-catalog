// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	"github.com/open-edge-platform/app-orch-catalog/internal/yamlreader"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/spf13/cobra"
	"os"
	"bufio"
	"strings"
)

type Param struct {
	name string
	value string
}

var (
	profile       string
	listProfiles  bool
	allParams	 bool
	rawOverrides 	 []string
	Overrides map[string]string
	rootCmd = &cobra.Command{
		Use:   "dp-to-helm <dir>",
		Short: "Convert a Deployment Package to a Helm install command",
		Long:  "This tool takes a deployment package and outputs the equivalent helm install command",
	}
)

func FindApp(r *yamlreader.YamlReader, name, version string) (*catalogv3.Application, error) {
	for _, app := range r.Applications {
		if app.Name == name && app.Version == version {
			return app, nil
		}
	}
	return nil, fmt.Errorf("application %s with version %s not found", name, version)
}

func FindRegistry(r *yamlreader.YamlReader, name string) (*catalogv3.Registry, error) {
	for _, reg := range r.Registries {
		if reg.Name == name {
			return reg, nil
		}
	}
	return nil, fmt.Errorf("registry %s not found", name)
}

func FindDeploymentProfile(dp *catalogv3.DeploymentPackage, name string) (*catalogv3.DeploymentProfile, error) {
	for _, profile := range dp.Profiles {
		if profile.Name == name {
			return profile, nil
		}
	}
	return nil, fmt.Errorf("deployment profile %s not found", name)
}

func FindAppProfile(app *catalogv3.Application, name string) (*catalogv3.Profile, error) {
	for _, appProfile := range app.Profiles {
		if appProfile.Name == name {
			return appProfile, nil
		}
	}
	return nil, fmt.Errorf("application profile %s not found", name)
}

func ApplyParameters(appProfile *catalogv3.Profile) ([]Param, error) {
	namedParams := make([]Param, 0)
	for _, param := range appProfile.ParameterTemplates {
		// if param was overridden on the command line, use that value
		override, ok := Overrides[param.Name]
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

func GetHelmCommands(r *yamlreader.YamlReader, dp *catalogv3.DeploymentPackage, profileName string) ([]string, error) {
	if profileName == "" {
		profileName = dp.DefaultProfileName
	}
	profile, err := FindDeploymentProfile(dp, profileName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("# using deployment package profile: %s\n", profileName)

	cmds := make([]string, 0)
	for _, app := range dp.ApplicationReferences {
		app, err := FindApp(r, app.Name, app.Version)
		if err != nil {
			return nil, err
		}
		reg, err := FindRegistry(r, app.HelmRegistryName)
		if err != nil {
			return nil, err
		}
		var namespace string
		namespace, okay := dp.DefaultNamespaces[app.Name]
		if !okay {
			namespace = "default"
		}
		appProfileName := profile.ApplicationProfiles[app.Name]
		appProfile, err := FindAppProfile(app, appProfileName)
		if err != nil {
			return nil, err
		}
		namedParams, err := ApplyParameters(appProfile)
		if err != nil {
			return nil, err
		}
		valuesFileName := fmt.Sprintf("%s-%s.yaml", app.Name, profileName)
		err = os.WriteFile(valuesFileName, []byte(appProfile.ChartValues), 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write to values file %s: %v", valuesFileName, err)
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

func mainCommand(cmd *cobra.Command, args []string) {
	Overrides = make(map[string]string)
    for _, override := range rawOverrides {
		parts := strings.Split(override, "=")
		if len(parts) != 2 {
			verboseerror.FatalErrCheck(fmt.Errorf("invalid --set override format: %s, expected <key>=<value>", override))
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if name == "" || value == "" {
			verboseerror.FatalErrCheck(fmt.Errorf("invalid --set override format: %s, expected <key>=<value>", override))
		}
		Overrides[name] = value
	}

	if len(args) != 1 {
		err := cmd.Usage()
		verboseerror.FatalErrCheck(err)
	}

	dir := args[0]

	r := &yamlreader.YamlReader{}
	fileSet, err := r.ReadYamlFilesFromDir(dir)
	verboseerror.FatalErrCheck(err)

	fileSets, err := r.ExpandFileSet(fileSet)
	verboseerror.FatalErrCheck(err)

	for _, fileSet := range fileSets {
		err := r.ProcessFiles(fileSet)
		verboseerror.FatalErrCheck(err)
	}

	if len(r.DeploymentPackages) == 0 {
		verboseerror.FatalErrCheck(fmt.Errorf("no deployment packages found"))
	}

	if len(r.DeploymentPackages) > 1 {
		verboseerror.FatalErrCheck(fmt.Errorf("multiple deployment packages found"))
	}

	if listProfiles {
		for _, dp := range r.DeploymentPackages {
			for _, profile := range dp.Profiles {
				fmt.Printf("%s\n", profile.Name)
			}
		}
		return
	}

	for _, dp := range r.DeploymentPackages {
		cmds, err := GetHelmCommands(r, dp, profile)
		verboseerror.FatalErrCheck(err)
		for _, cmd := range cmds {
			fmt.Printf("%s\n", cmd)
		}
	}
}

func main() {
	rootCmd.PersistentFlags().BoolVarP(&verboseerror.Quiet, "quiet", "q", false, "enable quiet mode, suppressing info level messages")
	rootCmd.PersistentFlags().BoolVarP(&listProfiles, "listprofiles", "L", false, "List the available deployment package profiles")
	rootCmd.PersistentFlags().BoolVarP(&allParams, "allparams", "A", false, "Ask for all parameters, not just mandatory ones")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "set which deployment package profile to use")
	rootCmd.PersistentFlags().StringArrayVarP(&rawOverrides, "set", "S", nil, "Set a parameter values using <key>=<value> format")
	rootCmd.Run = mainCommand

	err := rootCmd.Execute()
	verboseerror.FatalErrCheck(err)
}
