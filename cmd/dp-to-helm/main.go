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
)

var (
	profile       string
	listProfiles  bool
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
		_ = appProfile
		valuesFileName := fmt.Sprintf("%s-%s.yaml", app.Name, profileName)
		err = os.WriteFile(valuesFileName, []byte(appProfile.ChartValues), 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write to values file %s: %v", valuesFileName, err)
		}
		fmt.Printf("# created values file %s for app %s profile %s\n", valuesFileName, app.Name, appProfileName)
		url := fmt.Sprintf("%s/%s", reg.RootUrl, app.ChartName)
		helmCmd := fmt.Sprintf("helm install %s %s --version %s --namespace %s -f %s", app.Name, url, app.ChartVersion, namespace, valuesFileName)
		cmds = append(cmds, helmCmd)
	}

	return cmds, nil
}

func mainCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		err := cmd.Usage()
		verboseerror.FatalErrCheck(err, "Failed to print usage: %v", err)
	}

	dir := args[0]

	r := &yamlreader.YamlReader{}
	fileSet, err := r.ReadYamlFilesFromDir(dir)
	verboseerror.FatalErrCheck(err, "Failed to read YAML files from directory: %v", err)

	fileSets, err := r.ExpandFileSet(fileSet)
	verboseerror.FatalErrCheck(err, "Failed to expand file set: %v", err)

	for _, fileSet := range fileSets {
		err := r.ProcessFiles(fileSet)
		verboseerror.FatalErrCheck(err, "Failed to load YAML specs: %v", err)
	}

	if len(r.DeploymentPackages) == 0 {
		verboseerror.FatalErrCheck(fmt.Errorf("multiple deployment packages found"), "Make sure there is a deployment package in the directory")
	}

	if len(r.DeploymentPackages) > 1 {
		verboseerror.FatalErrCheck(fmt.Errorf("multiple deployment packages found"), "Please use only one deployment paackage")
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
		verboseerror.FatalErrCheck(err, "Failed to print helm commands: %v", err)
		for _, cmd := range cmds {
			fmt.Printf("%s\n", cmd)
		}
	}
}

func main() {
	rootCmd.PersistentFlags().BoolVarP(&verboseerror.Quiet, "quiet", "q", false, "enable quiet mode, suppressing info level messages")
	rootCmd.PersistentFlags().BoolVarP(&listProfiles, "listprofiles", "L", false, "List the available deployment package profiles")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "set which deployment package profile to use")
	rootCmd.Run = mainCommand

	err := rootCmd.Execute()
	verboseerror.FatalErrCheck(err, "Failed to execute root command: %v", err)
}
