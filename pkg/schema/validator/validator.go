// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-edge-platform/app-orch-catalog/pkg/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// Validator provides means to validate YAML files against Application Catalog YAML schema
type Validator struct {
	schema *jsonschema.Schema
}

const schemaProperty = `$schema: "https://schema.intel.com/catalog.orchestrator/0.1/schema"`

type ValidationResult struct {
	Path    string
	Err     error
	Message string
}

// ValidateFiles validates the specified YAML files or directories recursively against the Application Catalog YAML schema.
func ValidateFiles(paths ...string) ([]ValidationResult, error) {
	// Create a validator
	v, err := NewValidator()
	if err != nil {
		return nil, err
	}

	// Iterate over all specified files or directories and find all YAML files
	files, err := findYAMLFiles(paths)
	if err != nil {
		return nil, err
	}

	results := make([]ValidationResult, 0, len(files))
	var firstErr error
	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			results, firstErr = latchError(results, file, err, firstErr)
			continue
		}

		// Include only the YAML files with specSchema; to exclude miscellaneous values files
		if strings.Contains(string(yamlBytes), schemaProperty) {
			err = v.Validate(yamlBytes)
			if err != nil {
				results, firstErr = latchError(results, file, err, firstErr)
				continue
			}
			results = append(results, ValidationResult{Path: file})
		}
	}
	return results, firstErr
}

func latchError(results []ValidationResult, path string, err error, oldError error) ([]ValidationResult, error) {
	message := ""
	if _, ok := err.(*jsonschema.ValidationError); ok {
		message = fmt.Sprintf("%#v\n", err)
	} else {
		message = fmt.Sprintf("validation failed: %v\n", err)
	}
	if oldError == nil {
		return append(results, ValidationResult{Path: path, Err: err, Message: message}), err
	}
	return append(results, ValidationResult{Path: path, Err: err, Message: message}), oldError
}

func isDir(path string) (bool, error) {
	file, err := os.Open(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		return false, err
	}
	stat, err := file.Stat()
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func findYAMLFiles(paths []string) ([]string, error) {
	files := make([]string, 0)

	var err error
	for _, path := range paths {
		isDirectory, err := isDir(path)
		if err != nil {
			return nil, err
		}

		if isDirectory {
			dirPath := path
			err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, _ error) error {
				if !d.IsDir() && strings.HasSuffix(path, ".yaml") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			files = append(files, path)
		}
	}
	return files, err
}

// NewValidator creates a new Application Catalog YAML schema validator.
func NewValidator() (*Validator, error) {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020
	compiler.LoadURL = loadURL

	validator := &Validator{}
	var err error
	validator.schema, err = compiler.Compile("-")
	if err != nil {
		return nil, err
	}
	return validator, nil
}

// Validate validates the given YAML bytes against the Application Catalog YAML schema.
func (v *Validator) Validate(yamlBytes []byte) error {
	raw, err := decodeBytes(yamlBytes)
	if err != nil {
		return err
	}
	for _, r := range raw {
		if err = v.schema.Validate(r); err != nil {
			return filterError(err, r)
		}
	}
	return nil
}

// Filters the supplied error chain into a more consumable error
func filterError(err error, raw interface{}) error {
	if raw != nil {
		kv := raw.(map[string]interface{})
		if kind, ok := kv["specSchema"]; ok {
			msgForKind := fmt.Sprintf("doesn't validate with '/$defs/%s'", kind)

			// Scan the top-level oneOf
			for _, cause := range err.(*jsonschema.ValidationError).Causes {
				// Scan the oneOf possibilities
				for _, c1 := range cause.Causes {
					// Scan the top-level allOf
					for _, c2 := range c1.Causes {
						// Scan the allOf for each of the oneOf possibilities
						for _, c3 := range c2.Causes {
							// Here find the one that states we cannot parse our kind of entity
							if c3.Message == msgForKind {
								return c3
							}
						}
					}
				}
			}
		}
	}
	return err
}

func loadURL(s string) (io.ReadCloser, error) {
	r := io.NopCloser(strings.NewReader(schema.AppCatalogSchema))
	defer r.Close()

	v, err := decodeYAML(r, s)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

// The following has been adapted from the https://github.com/santhosh-tekuri/jsonschema/blob/master/cmd/jv/main.go

func decodeYAML(r io.Reader, name string) (interface{}, error) {
	var v interface{}
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("invalid yaml file %s: %w", name, err)
	}
	return v, nil
}

func decodeBytes(yamlBytes []byte) ([]interface{}, error) {
	raw := make([]interface{}, 0, 1)
	dec := yaml.NewDecoder(bytes.NewReader(yamlBytes))

	for {
		var r interface{}
		err := dec.Decode(&r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("invalid yaml: %w", err)
		}
		raw = append(raw, r)
	}
	return raw, nil
}
