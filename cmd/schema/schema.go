// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/generator"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/validator"
	_ "github.com/open-edge-platform/orch-library/go/dazl/zap"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := getRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "catalog-schema {generate, validate} [flags]",
		Short:         "App Catalog schema generation and validation utility",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.AddCommand(
		getGenerateSchemaCommand(),
		getValidateSchemaCommand(),
	)
	return rootCmd
}

func getValidateSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Args:  cobra.MinimumNArgs(1),
		Short: "Validate YAML files against Application Catalog YAML schema",
		RunE:  runValidateSchemaCommand,
	}
	cmd.Flags().BoolP("verbose", "v", false, "Emit verbose output")
	return cmd
}

func runValidateSchemaCommand(cmd *cobra.Command, args []string) error {
	results, err := validator.ValidateFiles(args...)
	verbose, _ := cmd.Flags().GetBool("verbose")
	for _, result := range results {
		if result.Err != nil {
			fmt.Printf("%s: %s\n", result.Path, result.Message)
		} else if verbose {
			fmt.Printf("%s: OK\n", result.Path)
		}
	}
	return err
}

func getGenerateSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Args:  cobra.ExactArgs(0),
		Short: "Generate Application Catalog YAML schema from OpenAPI specs",
		RunE:  runGenerateSchemaCommand,
	}
	return cmd
}

func runGenerateSchemaCommand(_ *cobra.Command, _ []string) error {
	return generator.GenerateSchema()
}
