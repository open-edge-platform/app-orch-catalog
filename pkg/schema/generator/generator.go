// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package generator

import (
	"bytes"
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	schemaFile      = "pkg/schema/catalog-schema"
	schemaBaseFile  = "pkg/schema/generator/catalog-schema-base.yaml"
	openapiSpecFile = "api/spec/openapi.yaml"
)

// GenerateSchema generates Application Catalog YAML schema from the OpenAPI spec.
func GenerateSchema() error {
	return generateSchema(schemaBaseFile, openapiSpecFile, schemaFile)
}

func generateSchema(schemaBasePath string, apiSpecPath string, schemaPath string) error {
	apiSpec, err := loadRawYAML(apiSpecPath)
	if err != nil {
		return err
	}
	schema, err := generateSchemaDefs(apiSpec)
	if err != nil {
		return err
	}
	return saveRawYAML(schemaBasePath, schemaPath, schema)
}

// Generates schema from the given openapi spec.
func generateSchemaDefs(spec interface{}) (interface{}, error) {
	componentsNode := spec.(map[string]interface{})["components"]
	schemasNode := componentsNode.(map[string]interface{})["schemas"]

	defs := make(map[string]interface{})
	for name, node := range schemasNode.(map[string]interface{}) {
		if isRelevant(name) {
			defs[name] = augmentNode(name, node)
		}
	}
	return map[string]interface{}{"$defs": defs}, nil
}

// Augments the given named node to be backwards compatible with the existing YAML schema
func augmentNode(name string, node interface{}) interface{} {
	// Rename any names in the required properties list
	requiredNode, ok := node.(map[string]interface{})["required"]
	if ok {
		nm := node.(map[string]interface{})
		nm["required"] = augmentRequiredFields(name, requiredNode)
	}

	// Rename any property nodes
	propertiesNode, ok := node.(map[string]interface{})["properties"]
	properties := propertiesNode.(map[string]interface{})
	if ok {
		augmentProperties(name, properties)
	}
	return node
}

// Field name mappings for backwards compatibility with the existing YAML schema
var renames = map[string]string{
	"imageRegistryName":     "imageRegistry",
	"helmRegistryName":      "helmRegistry",
	"defaultProfileName":    "defaultProfile",
	"applicationReferences": "applications",
	"username":              "userName",
	"cacerts":               "caCerts",
}

// Augments the required fields list of a given node
func augmentRequiredFields(name string, requiredNode interface{}) interface{} {
	required := requiredNode.([]interface{})
	var newRequired []string
	for _, n := range required {
		field := n.(string)

		if !isNotRequiredField(name, field) {
			// If the field is not marked as not required, see if it needs to be renamed; otherwise include as is
			if nn, ok := renames[field]; ok {
				newRequired = append(newRequired, nn)
			} else {
				newRequired = append(newRequired, field)
			}
		}
	}
	return newRequired
}

// Augments the properties of a given node
func augmentProperties(name string, properties map[string]interface{}) {
	for n, pn := range properties {
		// If the property needs to be renamed remap the node to the new name
		if nn, ok := renames[n]; ok {
			properties[nn] = pn   // insert the node with the new name
			delete(properties, n) // delete the node under the old name
		}

		// Node specific transformations
		if name == "DeploymentPackage" && n == "profiles" {
			// deployment profiles node needs to be renamed from profiled to deploymentProfiles to avoid collision
			properties["deploymentProfiles"] = pn // insert the node with the new name
			delete(properties, n)                 // delete the node under the old name
		}

		// Field specific augmentations
		nm := pn.(map[string]interface{})
		if n == "createTime" || n == "updateTime" || n == "artifact" {
			delete(nm, "format") // format attribute for these properties is not supported
		}
		if n == "displayName" || n == "purpose" || n == "label" {
			delete(nm, "pattern") // pattern attribute for these properties is not supported
		}

		// Application profiles needs to have a completely custom structure for backwards compatibility
		if n == "applicationProfiles" {
			// insert node with different structure
			properties["applicationProfiles"] = map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"application": map[string]interface{}{"type": "string"},
						"profile":     map[string]interface{}{"type": "string"},
					},
					"required":              []string{"application", "profile"},
					"unevaluatedProperties": false,
				},
				"description": nm["description"],
			}
		}
	}
}

// Returns true if the field is indeed not required for the given node.
func isNotRequiredField(name string, field string) bool {
	return name == "DeploymentPackage" && (field == "artifacts" || field == "extensions")
}

// Returns true if the given named node is relevant to the YAML schema.
func isRelevant(name string) bool {
	return !strings.HasSuffix(name, "Request") && !strings.HasSuffix(name, "Response") && name != "Upload"
}

// Loads the specified YAML file as raw structure.
func loadRawYAML(path string) (interface{}, error) {
	yamlBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw interface{}
	if err = yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Saves the YAML schema using the schema base file and appended generated $defs node.
func saveRawYAML(schemaBasePath string, schemaPath string, raw interface{}) error {
	baseBytes, err := os.ReadFile(schemaBasePath)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	err = yamlEncoder.Encode(raw)
	if err != nil {
		return err
	}

	// Replace '#/components/schemas/' with '#/$defs/' references
	schema := strings.ReplaceAll(b.String(), "#/components/schemas/", "#/$defs/")
	schema = strings.ReplaceAll(schema, "`", "")

	// Replace kind values
	schema = strings.ReplaceAll(schema, "- KIND_NORMAL", "- normal")
	schema = strings.ReplaceAll(schema, "- KIND_ADDON", "- addon")
	schema = strings.ReplaceAll(schema, "- KIND_EXTENSION", "- extension")

	// Strip the $defs: line from the generated output since we're appending to the base schema file
	if len(schema) < 8 {
		return errors.New("generated schema too short")
	}

	// Append generated schema defs to the base schema for semi-ordered output
	schema = string(baseBytes) + schema[7:]

	// Save the YAML file
	err = os.WriteFile(schemaPath+".yaml", []byte(schema), 0600)
	if err != nil {
		return err
	}

	// Save the corresponding Go file containing YAML schema as a string constant
	schemaGoFile := "// DO NOT EDIT: Autogenerated by 'schema generate'\n" +
		"\npackage schema\n\n" +
		"// AppCatalogSchema contains auto-generated Application Catalog YAML schema\n" +
		"const AppCatalogSchema = `\n" + schema + "\n`\n"
	return os.WriteFile(schemaPath+".go", []byte(schemaGoFile), 0600)
}
