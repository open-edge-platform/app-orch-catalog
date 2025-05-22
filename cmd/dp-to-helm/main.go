// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/open-edge-platform/app-orch-catalog/internal/dptohelm"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/verboseerror"
	"github.com/spf13/cobra"
)

type Param struct {
	name  string
	value string
}

var (
	profile      string
	listProfiles bool
	allParams    bool
	rawOverrides []string
	rootCmd      = &cobra.Command{
		Use:   "dp-to-helm <dir>",
		Short: "Convert a Deployment Package to a Helm install command",
		Long:  "This tool takes a deployment package and outputs the equivalent helm install command",
	}
)

func mainCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		err := cmd.Usage()
		verboseerror.FatalErrCheck(err)
		return
	}

	dir := args[0]

	r := &dptohelm.DpToHelm{}

	r.SetOverrides(rawOverrides)

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
		cmds, err := r.GetHelmCommands(dp, profile, allParams)
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
