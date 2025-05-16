// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package yamlreader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	b64 "encoding/base64"
	"errors"
	"path/filepath"
	"fmt"
	nberrors "github.com/open-edge-platform/app-orch-catalog/internal/northbound/errors"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/upload"
	"github.com/open-edge-platform/app-orch-catalog/pkg/schema/validator"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/open-edge-platform/orch-library/go/dazl"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	yaml "gopkg.in/yaml.v3"
)

const (
	MaxExtractedFileSize = 10 * 1024 * 1024 // to limit the size of extracted files and mitigate decompression bomb lint message
	kindNormal    = "normal"
	kindAddon     = "addon"
	kindExtension = "extension"
	helmType  = "HELM"
)

type YamlReader struct {
	Applications		[]*catalogv3.Application
	DeploymentPackages 	[]*catalogv3.DeploymentPackage
	Artifacts			[]*catalogv3.Artifact
	Registries			[]*catalogv3.Registry
}

// fileSet is a map of file names to their contents. It is a set of files that should be evaluated
// together. For example, it may contain applications and their profiles.
type fileSet map[string][]byte

var log dazl.Logger

func kindFromDB(kind string) catalogv3.Kind {
	switch kind {
	case kindNormal:
			return catalogv3.Kind_KIND_NORMAL
	case kindAddon:
			return catalogv3.Kind_KIND_ADDON
	case kindExtension:
			return catalogv3.Kind_KIND_EXTENSION
	}
	return catalogv3.Kind_KIND_NORMAL
}

func (u *YamlReader) ReadYamlFilesFromDir(dir string) (fileSet, error) {
	files := make(fileSet)

	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("directory %s does not exist", dir)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to access directory %s: %v", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path %s is not a directory", dir)
	}

	direntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %v", dir, err)
	}

	for _, direntry := range direntries {
		if !direntry.IsDir() && (filepath.Ext(direntry.Name()) == ".yaml" || filepath.Ext(direntry.Name()) == ".yml") {
			filename := filepath.Join(dir, direntry.Name())
			file, err := os.Open(filename)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
			}
			defer file.Close()
			content, err := io.ReadAll(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
			}
			files[path.Base(filename)] = content

		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML files found in directory %s", dir)
	}

	return files, nil
}

func (u *YamlReader) ProcessFiles(files fileSet) error {
	// Process each fileset, load its yaml, sort, and add to the catalog.

	orderedSpecs, err := u.loadYamlSpecs(files)
	if err != nil {
		return err
	}

	for _, d := range orderedSpecs {
		switch d.SpecSchema {
		case upload.DeploymentPackageType:
			err = u.readDeploymentPackage(d)
		case upload.DeploymentPackageLegacyType:
			err = u.readDeploymentPackage(d)
		case upload.ApplicationType:
			err = u.readApplication(d, files) // application uses the fileSet to lookup profiles
		case upload.RegistryType:
			err = u.readRegistry(d)
		case upload.ArtifactType:
			err = u.readArtifact(d)
		default:
			return nberrors.NewInvalidArgument(nberrors.WithMessage("uploaded file %s: unhandled type %s", d.FileName, d.SpecSchema))
		}
		if err != nil {
			return nberrors.NewInvalidArgument(nberrors.WithError(err), nberrors.WithMessage("uploaded file %s: %v", d.FileName, err))
		}
	}


	return nil
}

// shouldValidateYAMLSchema determines if the schema checker should run on the given
// artifact. If the YAML can be unmarshaled or contains the schema directive,
// it should be validated. Values files containing {{ }} markers should
// not be validated.
func shouldValidateYAMLSchema(fileBytes []byte) bool {
	var raw interface{}
	err := yaml.Unmarshal(fileBytes, &raw)
	if err != nil {
		re, _ := regexp.Compile(`\$schema:\s*"https://schema\.intel\.com/catalog\.orchestrator/0\.1/schema`)

		if !re.MatchString(string(fileBytes)) {
			return false
		}
	}
	return true
}

func (u *YamlReader) extractTarball(fileBytes []byte) (fileSet, error) {
	contents := make(fileSet, 0)

	// Convert fileBytes into an io.Reader
	reader := bytes.NewReader(fileBytes)

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("Failed to create gzip reader while extracting: %w", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, fmt.Errorf("Failed to read tar header while extracting: %w", err)
		}

		// Check if the current file is the one we want to extract
		if header.Typeflag == tar.TypeReg {
			// Read the file content into a buffer
			var buf bytes.Buffer
			if _, err := io.CopyN(&buf, tarReader, MaxExtractedFileSize); err != nil && err != io.EOF {
				return nil, fmt.Errorf("Failed to copy file content while extracting: %w", err)
			}

			//verboseerror.Infof("Extracted file: %s\n", targetFileName)
			contents[header.Name] = buf.Bytes()
		}
	}

	return contents, nil
}

// loadYamlSpecs loads the contents of the specified uploads and returns them as an ordered
// collection of YamlSpecs
func (u *YamlReader) ExpandFileSet(srcFiles fileSet) ([]fileSet, error) {
	var fileSets []fileSet

	baseFileSet := make(fileSet, 0)

	for fileName, fileContents := range srcFiles {
		if strings.HasSuffix(fileName, ".tgz") || strings.HasSuffix(fileName, ".tar.gz") {
			// The file is a tarvball. Extract the individual files from it and add them to a fileset
			extractedFileSet, err := u.extractTarball(fileContents)
			if err != nil {
				return nil, fmt.Errorf("failed to extract tarball %s: %w", fileName, err)
			}
			fileSets = append(fileSets, extractedFileSet)
		} else {
			// The file is a single YAML file. Add it to the base fileset
			baseFileSet[fileName] = fileContents
		}
	}

	if len(baseFileSet) > 0 {
		// If we have a base file set, add it to the list of file sets
		fileSets = append(fileSets, baseFileSet)
	}

	// We will validate the yaml contents inside of loadYamlSpecs()

	return fileSets, nil
}

func (u *YamlReader) loadYamlSpecs(files fileSet) (upload.YamlSpecs, error) {
	orderedSpecs := make(upload.YamlSpecs, 0)

	for fileName, fileBytes := range files {
		if shouldValidateYAMLSchema(fileBytes) {
			// Deal only with files that can be successfully unmarshalled; value.yaml files with templates can't be for example
			decoder := yaml.NewDecoder(bytes.NewBuffer(fileBytes))
			for {
				var d upload.YamlSpec
				if err := decoder.Decode(&d); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return nil, fmt.Errorf("document decode failed: %w", err)
				}
				d.FileName = fileName
				if d.SpecSchema != "" {
					// check that the uploaded YAML complies with the schema
					v, err := validator.NewValidator()
					if err != nil {
						return nil, err
					}

					err = v.Validate(fileBytes)
					if err != nil {
						log.Infof("YAML validation failed for %s:%s", fileName, err)
						return nil, nberrors.NewInvalidArgument(
							nberrors.WithMessage("uploaded file %s is invalid YAML: %+v", fileName, err),
							nberrors.WithError(err))
					}
					orderedSpecs = append(orderedSpecs, d)
				}
			}
		}
	}

	sort.Sort(orderedSpecs)
	return orderedSpecs, nil
}

func valueOrDefault(val string, def string) string {
	if val == "" {
		return def
	}
	return val
}

func (u *YamlReader) readRegistry(d upload.YamlSpec) error {
	reg := &catalogv3.Registry{
		Name:         d.Name,
		DisplayName:  d.DisplayName,
		Description:  d.Description,
		RootUrl:      d.RootURL,
		InventoryUrl: d.InventoryURL,
		Username:     d.UserName,
		AuthToken:    d.AuthToken,
		Type:         valueOrDefault(d.GetRegistryType(), helmType),
		ApiType:      d.APIType,
		Cacerts:      d.CACerts,
	}

	u.Registries = append(u.Registries, reg)
	return nil
}

func (u *YamlReader) readArtifact(d upload.YamlSpec) error {
	artifactBinary, err := b64.StdEncoding.DecodeString(d.Artifact)
	if err != nil {
		return nberrors.NewInvalidArgument(
			nberrors.WithResourceType(nberrors.ArtifactType),
			nberrors.WithMessage("error decoding base64 from file %s %v", d.FileName, err))
	}
	if malware.DefaultScanner != nil {
		// Note: Artifact Size is already limited by protobuf validator
		okay, res, err := malware.DefaultScanner.ScanBytes(artifactBinary)
		if err != nil {
			if malware.DefaultScanner.IsPermissive() {
				log.Warn("Malware scanner is not available. Skipping scan due to permissive mode.")
			} else {
				log.Warn("Malware scanner returned error %s", err)
				return nberrors.NewUnavailable(
					nberrors.WithResourceType(nberrors.ArtifactType),
					nberrors.WithMessage("malware scanner configured but not available"))
			}
		} else if !okay {
			return nberrors.NewInvalidArgument(
				nberrors.WithResourceType(nberrors.ArtifactType),
				nberrors.WithMessage("malware detected: %s", res))
		}
	}

	art := &catalogv3.Artifact{
		Name:        d.Name,
		DisplayName: d.DisplayName,
		Description: d.Description,
		MimeType:    d.MimeType,
		Artifact:    artifactBinary,
	}

	u.Artifacts = append(u.Artifacts, art)
	return nil
}

func (u *YamlReader) readApplication(d upload.YamlSpec, f fileSet) error {
	app := &catalogv3.Application{
		Name:               d.Name,
		Version:            d.Version,
		Kind:               kindFromDB(d.Kind),
		DisplayName:        d.DisplayName,
		Description:        d.Description,
		ChartName:          d.ChartName,
		ChartVersion:       d.ChartVersion,
		HelmRegistryName:   d.GetHelmRegistry(),
		ImageRegistryName:  d.ImageRegistry,
		DefaultProfileName: d.DefaultProfile,
		Profiles:           make([]*catalogv3.Profile, 0, len(d.Profiles)),
		IgnoredResources:   make([]*catalogv3.ResourceReference, 0, len(d.IgnoredResources)),
	}

	if len(d.Profiles) > 0 {
		for _, p := range d.Profiles {
			prof, err := u.readProfile(d.FileName, p, f)
			if err != nil {
				return err
			}
			app.Profiles = append(app.Profiles, prof)
		}
		if app.DefaultProfileName == "" {
			app.DefaultProfileName = d.Profiles[0].Name
		}
	}

	if len(d.IgnoredResources) > 0 {
		for _, r := range d.IgnoredResources {
			app.IgnoredResources = append(app.IgnoredResources, &catalogv3.ResourceReference{
				Name:      r.Name,
				Kind:      r.Kind,
				Namespace: r.Namespace,
			})
		}
	}

	u.Applications = append(u.Applications, app)
	return nil
}

func (u *YamlReader) readProfile(appFileName string, p upload.Profile, f fileSet) (*catalogv3.Profile, error) {
	fileBytes, ok := f[p.ValuesFileName]
	if !ok {
		fileBytes, ok = f[fmt.Sprintf("%s/%s", path.Dir(appFileName), p.ValuesFileName)]
		if !ok {
			return nil, nberrors.NewInvalidArgument(
				nberrors.WithMessage("chart values file %s not found in uploads", p.ValuesFileName))
		}
	}

	requirements := make([]*catalogv3.DeploymentRequirement, 0)
	for _, dr := range p.DeploymentRequirements {
		newRequirement := &catalogv3.DeploymentRequirement{
			Name:                  dr.Name,
			Version:               dr.Version,
			DeploymentProfileName: dr.DeploymentProfile,
		}
		requirements = append(requirements, newRequirement)
	}

	parameterTemplates := make([]*catalogv3.ParameterTemplate, 0)
	for _, pt := range p.ParameterTemplates {
		newParameterTemplate := &catalogv3.ParameterTemplate{
			Name:            pt.Name,
			DisplayName:     pt.DisplayName,
			Default:         pt.Default,
			Type:            pt.Type,
			Validator:       pt.Validator,
			SuggestedValues: pt.SuggestedValues,
			Secret:          pt.Secret,
			Mandatory:       pt.Mandatory,
		}
		parameterTemplates = append(parameterTemplates, newParameterTemplate)
	}

	yamlString := string(fileBytes)
	return &catalogv3.Profile{
		Name:                  p.Name,
		DisplayName:           p.DisplayName,
		Description:           p.Description,
		ChartValues:           yamlString,
		DeploymentRequirement: requirements,
		ParameterTemplates:    parameterTemplates,
	}, nil
}

func (u *YamlReader) readDeploymentPackage(d upload.YamlSpec) error {
	pkg := &catalogv3.DeploymentPackage{
		Name:                    d.Name,
		DisplayName:             d.DisplayName,
		Description:             d.Description,
		Version:                 d.Version,
		Kind:                    kindFromDB(d.Kind),
		DefaultProfileName:      d.DefaultProfile,
		Profiles:                make([]*catalogv3.DeploymentProfile, 0, len(d.DeploymentProfiles)),
		ApplicationReferences:   make([]*catalogv3.ApplicationReference, 0, len(d.Applications)),
		ApplicationDependencies: make([]*catalogv3.ApplicationDependency, 0, len(d.ApplicationDependencies)),
		Extensions:              make([]*catalogv3.APIExtension, 0, len(d.Extensions)),
		Artifacts:               make([]*catalogv3.ArtifactReference, 0, len(d.ArtifactReferences)),
		DefaultNamespaces:       make(map[string]string, len(d.DefaultNamespaces)),
		Namespaces:              make([]*catalogv3.Namespace, 0, len(d.Namespaces)),
	}

	for _, a := range d.Applications {
		pkg.ApplicationReferences = append(pkg.ApplicationReferences,
			&catalogv3.ApplicationReference{
				Name:    a.Name,
				Version: a.Version,
			})
	}

	for _, e := range d.Extensions {
		var endpoints []*catalogv3.Endpoint
		for _, ep := range e.Endpoints {
			endpoints = append(endpoints,
				&catalogv3.Endpoint{
					ServiceName:  ep.ServiceName,
					ExternalPath: ep.ExternalPath,
					InternalPath: ep.InternalPath,
					Scheme:       ep.Scheme,
					AuthType:     ep.AuthType,
					AppName:      ep.AppName,
				})
		}
		var uiExtension *catalogv3.UIExtension
		if e.UIExtension != nil {
			uiExtension =
				&catalogv3.UIExtension{
					ServiceName: e.UIExtension.ServiceName,
					Label:       *e.UIExtension.Label,
					Description: e.UIExtension.Description,
					FileName:    e.UIExtension.FileName,
					AppName:     e.UIExtension.AppName,
					ModuleName:  e.UIExtension.ModuleName,
				}
		}
		pkg.Extensions = append(pkg.Extensions,
			&catalogv3.APIExtension{
				Name:        e.Name,
				Version:     e.Version,
				DisplayName: e.DisplayName,
				Description: e.Description,
				Endpoints:   endpoints,
				UiExtension: uiExtension,
			})
	}

	for _, a := range d.GetArtifacts() {
		pkg.Artifacts = append(pkg.Artifacts,
			&catalogv3.ArtifactReference{
				Name:    a.Name,
				Purpose: a.Purpose,
			})
	}

	if len(d.ApplicationDependencies) > 0 {
		for _, a := range d.ApplicationDependencies {
			pkg.ApplicationDependencies = append(pkg.ApplicationDependencies,
				&catalogv3.ApplicationDependency{
					Name:     a.Name,
					Requires: a.Requires,
				})
		}
	}

	if len(d.DefaultNamespaces) != 0 {
		for k, v := range d.DefaultNamespaces {
			pkg.DefaultNamespaces[k] = v
		}
	}

	if len(d.Namespaces) != 0 {
		for _, ns := range d.Namespaces {
			pkg.Namespaces = append(pkg.Namespaces,
				&catalogv3.Namespace{
					Name:        ns.Name,
					Labels:      ns.Labels,
					Annotations: ns.Annotations,
				})
		}
	}

	if len(d.DeploymentProfiles) > 0 {
		for _, profile := range d.DeploymentProfiles {
			pkg.Profiles = append(pkg.Profiles, u.deploymentProfile(profile))
		}
		if pkg.DefaultProfileName == "" {
			pkg.DefaultProfileName = d.DeploymentProfiles[0].Name
		}
	}

	pkg.IsVisible = d.IsVisible
	pkg.IsDeployed = d.IsDeployed
	pkg.ForbidsMultipleDeployments = d.ForbidsMultipleDeployments

	u.DeploymentPackages = append(u.DeploymentPackages, pkg)
	return nil
}

func (u *YamlReader) deploymentProfile(deploymentProfile upload.DeploymentProfile) *catalogv3.DeploymentProfile {
	prof := &catalogv3.DeploymentProfile{
		Name:                deploymentProfile.Name,
		DisplayName:         deploymentProfile.DisplayName,
		Description:         deploymentProfile.Description,
		ApplicationProfiles: make(map[string]string, len(deploymentProfile.ApplicationProfiles)),
	}

	for _, p := range deploymentProfile.ApplicationProfiles {
		prof.ApplicationProfiles[p.ApplicationName] = p.ProfileName
	}
	return prof
}
