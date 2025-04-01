// SPDX-FileCopyrightText: (C) 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package upload

import (
	"reflect"
)

const (
	DeploymentPackageType       = "DeploymentPackage"
	DeploymentPackageLegacyType = "Deployment-Package"
	ApplicationType             = "Application"
	ArtifactType                = "Artifact"
	RegistryType                = "Registry"
	Quot                        = `'`
)

// ParameterTemplate is a structure for loading ignores resources of an application.
type ParameterTemplate struct {
	Name            string   `yaml:"name,omitempty"`
	DisplayName     string   `yaml:"displayName,omitempty"`
	Type            string   `yaml:"type,omitempty"`
	Default         string   `yaml:"default,omitempty"`
	Validator       string   `yaml:"validator,omitempty"`
	SuggestedValues []string `yaml:"suggestedValues,omitempty"`
	Secret          bool     `yaml:"secret,omitempty"`
	Mandatory       bool     `yaml:"mandatory,omitempty"`
}

// Profile is a structure for loading application deployment values
type Profile struct {
	Name                   string                  `yaml:"name,omitempty"`
	DisplayName            string                  `yaml:"displayName,omitempty"`
	Description            string                  `yaml:"description,omitempty"`
	ValuesFileName         string                  `yaml:"valuesFileName,omitempty"`
	ParameterTemplates     []ParameterTemplate     `yaml:"parameterTemplates,omitempty"`
	DeploymentRequirements []DeploymentRequirement `yaml:"deploymentRequirements,omitempty"`
}

// Application is a structure for loading application references, e.g. name/version pairs
// when defining a DeploymentPackages
type Application struct {
	Publisher string `yaml:"publisher,omitempty"`
	Name      string `yaml:"name,omitempty"`
	Version   string `yaml:"version,omitempty"`
}

// APIExtension is a structure for loading API extensions
type APIExtension struct {
	Name        string       `yaml:"name,omitempty"`
	Version     string       `yaml:"version,omitempty"`
	DisplayName string       `yaml:"displayName,omitempty"`
	Description string       `yaml:"description,omitempty"`
	Endpoints   []*Endpoint  `yaml:"endpoints,omitempty"`
	UIExtension *UIExtension `yaml:"uiExtension,omitempty"`
}

// Endpoint is a structure for loading extension endpoints
type Endpoint struct {
	ServiceName  string `yaml:"serviceName,omitempty"`
	ExternalPath string `yaml:"externalPath,omitempty"`
	InternalPath string `yaml:"internalPath,omitempty"`
	Scheme       string `yaml:"scheme,omitempty"`
	AuthType     string `yaml:"authType,omitempty"`
	AppName      string `yaml:"appName,omitempty"`
}

// UIExtension holds label and service information for UI extensions
type UIExtension struct {
	ServiceName string  `yaml:"serviceName,omitempty"`
	Description string  `yaml:"description,omitempty"`
	Label       *string `yaml:"label,omitempty"`
	FileName    string  `yaml:"fileName,omitempty"`
	AppName     string  `yaml:"appName,omitempty"`
	ModuleName  string  `yaml:"moduleName,omitempty"`
}

// ArtifactReference holds label and service information for UI extensions
type ArtifactReference struct {
	Publisher string `yaml:"publisher,omitempty"`
	Name      string `yaml:"name,omitempty"`
	Purpose   string `yaml:"purpose,omitempty"`
}

// ApplicationDependency holds dependencies to applications
type ApplicationDependency struct {
	Name     string `yaml:"name"`
	Requires string `yaml:"requires"`
}

// ApplicationProfile is a structure for loading application profiles, i.e. application/profile
// bindings when defining DeploymentProfiles
type ApplicationProfile struct {
	ApplicationName string `yaml:"application,omitempty"`
	ProfileName     string `yaml:"profile,omitempty"`
}

// DeploymentProfile is a structure for loading deployment package profiles.
type DeploymentProfile struct {
	Name                string               `yaml:"name,omitempty"`
	DisplayName         string               `yaml:"displayName,omitempty"`
	Description         string               `yaml:"description,omitempty"`
	ApplicationProfiles []ApplicationProfile `yaml:"applicationProfiles,omitempty"`
}

// ResourceReference is a structure for loading ignores resources of an application.
type ResourceReference struct {
	Name      string `yaml:"name,omitempty"`
	Kind      string `yaml:"kind,omitempty"`
	Namespace string `yaml:"namespace,omitempty"`
}

// DeploymentRequirement is a structure for specifying deployment requirements of an application profile.
type DeploymentRequirement struct {
	Publisher         string `yaml:"publisher,omitempty"`
	Name              string `yaml:"name,omitempty"`
	Version           string `yaml:"version,omitempty"`
	DeploymentProfile string `yaml:"deploymentProfileName,omitempty"`
}

// Namespace is a structure for loading deployment package required namespaces.
type Namespace struct {
	Name        string            `yaml:"name,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// YamlSpec is a main structure for loading application catalog entities.
type YamlSpec struct {
	FileName                   string                  `yaml:"filename,omitempty"`
	SpecSchema                 string                  `yaml:"specSchema,omitempty"`
	SchemaVersion              string                  `yaml:"schemaVersion,omitempty"`
	Name                       string                  `yaml:"name,omitempty"`
	DisplayName                string                  `yaml:"displayName,omitempty"`
	Version                    string                  `yaml:"version,omitempty"`
	Kind                       string                  `yaml:"kind,omitempty"`
	Description                string                  `yaml:"description,omitempty"`
	HelmRegistry               string                  `yaml:"helmRegistry,omitempty"`
	Type                       string                  `yaml:"type,omitempty"`
	ChartName                  string                  `yaml:"chartName,omitempty"`
	ChartVersion               string                  `yaml:"chartVersion,omitempty"`
	ImageRegistry              string                  `yaml:"imageRegistry,omitempty"`
	MimeType                   string                  `yaml:"mimeType,omitempty"`
	Artifact                   string                  `yaml:"artifact,omitempty"`
	RootURL                    string                  `yaml:"rootUrl,omitempty"`
	InventoryURL               string                  `yaml:"inventoryUrl,omitempty"`
	APIType                    string                  `yaml:"apiType,omitempty"`
	UserName                   string                  `yaml:"userName,omitempty"`
	AuthToken                  string                  `yaml:"authToken,omitempty"`
	CACerts                    string                  `yaml:"caCerts,omitempty"`
	Profiles                   []Profile               `yaml:"profiles,omitempty"`
	DeploymentProfiles         []DeploymentProfile     `yaml:"deploymentProfiles,omitempty"`
	DefaultProfile             string                  `yaml:"defaultProfile,omitempty"`
	Applications               []Application           `yaml:"applications,omitempty"`
	ApplicationDependencies    []ApplicationDependency `yaml:"applicationDependencies,omitempty"`
	Extensions                 []APIExtension          `yaml:"extensions,omitempty"`
	Artifacts                  []ArtifactReference     `yaml:"artifacts,omitempty"`
	DefaultNamespaces          map[string]string       `yaml:"defaultNamespaces,omitempty"`
	Namespaces                 []Namespace             `yaml:"namespaces,omitempty"`
	IgnoredResources           []ResourceReference     `yaml:"ignoredResources,omitempty"`
	IsVisible                  bool                    `yaml:"isVisible,omitempty"`
	IsDeployed                 bool                    `yaml:"isDeployed,omitempty"`
	ForbidsMultipleDeployments bool                    `yaml:"forbidsMultipleDeployments,omitempty"`

	Registry           string              `yaml:"registry,omitempty"`           // deprecated in lieu of 'helmRegistry'
	ArtifactReferences []ArtifactReference `yaml:"artifactReferences,omitempty"` // deprecated in lieu of 'artifacts'
	RegistryType       string              `yaml:"registryType,omitempty"`       // deprecated in lieu of 'type'
}

// GetHelmRegistry retrieves the Helm registry in a backwards compatible manner, drawing on deprecated Registry field if
// the HelmRegistry field is empty
func (ys YamlSpec) GetHelmRegistry() string {
	if ys.HelmRegistry == "" {
		return ys.Registry
	}
	return ys.HelmRegistry
}

// GetRegistryType retrieves the registry type in a backwards compatible manner, drawing on deprecated RegistryType field if
// the Type field is empty
func (ys YamlSpec) GetRegistryType() string {
	if ys.Type == "" {
		return ys.RegistryType
	}
	return ys.Type
}

// GetArtifacts retrieves the artifact references in a backwards compatible manner, drawing on deprecated ArtifactReferences field if
// the Artifacts field is empty
func (ys YamlSpec) GetArtifacts() []ArtifactReference {
	if len(ys.Artifacts) == 0 && len(ys.ArtifactReferences) > 0 {
		return ys.ArtifactReferences
	}
	return ys.Artifacts
}

// YamlSpecs is a collection of YamlSpec structures
type YamlSpecs []YamlSpec

// Len is the number of elements in the collection.
func (cas YamlSpecs) Len() int {
	return len(cas)
}

// Less reports whether the element with index i should sort before the element with index j.
// Results in the following sort order:
//
//	1 publisher(s)
//	2 registry(s)
//	3 application(s)
//	4 artifact(s)
//	5 deploymentPackage(s)
func (cas YamlSpecs) Less(i, j int) bool {
	iSchema := cas[i].SpecSchema
	jSchema := cas[j].SpecSchema

	switch iSchema {
	case RegistryType:
		switch jSchema {
		case RegistryType:
			return cas[i].Name < cas[j].Name
		default:
			return true
		}
	case ApplicationType:
		switch jSchema {
		case DeploymentPackageType:
			return true
		case ApplicationType:
			return cas[i].Name < cas[j].Name
		default:
			return false
		}
	case ArtifactType:
		switch jSchema {
		case DeploymentPackageType:
			return true
		case ArtifactType:
			return cas[i].Name < cas[j].Name
		default:
			return false
		}
	case DeploymentPackageType:
		switch jSchema {
		case DeploymentPackageType:
			return cas[i].Name < cas[j].Name
		default:
			return false // Always after others
		}
	case DeploymentPackageLegacyType:
		switch jSchema {
		case DeploymentPackageLegacyType:
			return cas[i].Name < cas[j].Name
		default:
			return false // Always after others
		}
	default:
		return cas[i].Name < cas[j].Name
	}
}

// Swap swaps the elements with indexes i and j.
func (cas YamlSpecs) Swap(i, j int) {
	swap := reflect.Swapper(cas)
	swap(i, j)
}
