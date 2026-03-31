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

			// Create short aliases for main entity types
			switch name {
			case "catalog.v3.Application":
				defs["Application"] = augmentNode(name, node)
			case "catalog.v3.Artifact":
				defs["Artifact"] = augmentNode(name, node)
			case "catalog.v3.DeploymentPackage":
				defs["DeploymentPackage"] = augmentNode(name, node)
			case "catalog.v3.Registry":
				defs["Registry"] = augmentNode(name, node)
			}
		}
	}
	return map[string]interface{}{"$defs": defs}, nil
}

// Augments the given named node to be backwards compatible with the existing YAML schema
func augmentNode(name string, node interface{}) interface{} {
	nodeMap := node.(map[string]interface{})

	// For main entity types, remove additionalProperties entirely
	// The unevaluatedProperties: false is handled at the oneOf level in the base schema
	if isMainEntityType(name) {
		delete(nodeMap, "additionalProperties")
	}

	// Rename any names in the required properties list
	requiredNode, ok := nodeMap["required"]
	if ok {
		nodeMap["required"] = augmentRequiredFields(name, requiredNode)
	}

	// Rename any property nodes
	propertiesNode, ok := nodeMap["properties"]
	if ok && propertiesNode != nil {
		properties := propertiesNode.(map[string]interface{})
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
	var required []interface{}

	// Handle both []interface{} and []string types
	switch v := requiredNode.(type) {
	case []interface{}:
		required = v
	case []string:
		required = make([]interface{}, len(v))
		for i, s := range v {
			required[i] = s
		}
	default:
		return requiredNode
	}

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
		if (name == "DeploymentPackage" || name == "catalog.v3.DeploymentPackage") && n == "profiles" {
			// deployment profiles node needs to be renamed from profiles to deploymentProfiles to avoid collision
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

	// Add backwards compatibility properties for Profile schema
	if name == "Profile" || name == "catalog.v3.Profile" {
		// Add valuesFileName as an optional property for backwards compatibility
		properties["valuesFileName"] = map[string]interface{}{
			"type":        "string",
			"description": "(OPTIONAL) Legacy property: name of the values file. Use chartValues instead.",
		}
		// Rename deploymentRequirement to deploymentRequirements for backwards compatibility
		if req, ok := properties["deploymentRequirement"]; ok {
			properties["deploymentRequirements"] = req
		}
	}
}

// Returns true if the given named node is a main entity type that should allow metadata properties.
func isMainEntityType(name string) bool {
	return name == "catalog.v3.Application" ||
		name == "catalog.v3.Artifact" ||
		name == "catalog.v3.DeploymentPackage" ||
		name == "catalog.v3.Registry"
}

// Returns true if the field is indeed not required for the given node.
func isNotRequiredField(name string, field string) bool {
	return (name == "DeploymentPackage" || name == "catalog.v3.DeploymentPackage") &&
		(field == "artifacts" || field == "extensions")
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
	err = os.WriteFile(schemaPath+".yaml", []byte(schema), 0600) //nolint:gosec // G703: schemaPath is from trusted generator configuration, not user input
	if err != nil {
		return err
	}

	// Save the corresponding Go file containing YAML schema as a string constant
	schemaGoFile := "// DO NOT EDIT: Autogenerated by 'schema generate'\n" +
		"\npackage schema\n\n" +
		"// AppCatalogSchema contains auto-generated Application Catalog YAML schema\n" +
		"const AppCatalogSchema = `\n" + schema + "\n`\n"
	return os.WriteFile(schemaPath+".go", []byte(schemaGoFile), 0600) //nolint:gosec // G703: schemaPath is from trusted generator configuration, not user input
}
