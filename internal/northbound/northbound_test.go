// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"encoding/base64"
	"fmt"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/malware"
	"google.golang.org/grpc/metadata"
	"strings"
	"testing"
	"time"

	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/enttest"
	"github.com/open-edge-platform/orch-library/go/pkg/openpolicyagent"
	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Suite of northbound tests
type NorthBoundTestSuite struct {
	suite.Suite

	startTime time.Time
	ctx       context.Context
	cancel    context.CancelFunc

	dbClient             *generated.Client
	conn                 *grpc.ClientConn
	client               catalogv3.CatalogServiceClient
	opa                  openpolicyagent.ClientWithResponsesInterface
	populateDB           bool
	malwareDelayInterval time.Duration
	malwareMaxRetries    int
}

func (s *NorthBoundTestSuite) SetupSuite() {
	s.populateDB = true
	s.malwareDelayInterval = malware.ErrorRetryInterval
	s.malwareMaxRetries = malware.MaxErrorRetries
	malware.ErrorRetryInterval = 1
	malware.MaxErrorRetries = 1
}

func (s *NorthBoundTestSuite) TearDownSuite() {
	malware.ErrorRetryInterval = s.malwareDelayInterval
	malware.MaxErrorRetries = s.malwareMaxRetries
}

func (s *NorthBoundTestSuite) SetupTest() {
	mockController := gomock.NewController(s.T())
	s.dbClient = enttest.Open(s.T(), "sqlite3", "file:ent?mode=memory&_fk=1")
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	opaMock := openpolicyagent.NewMockClientWithResponsesInterface(mockController)
	result := openpolicyagent.OpaResponse_Result{}
	err := result.FromOpaResponseResult1(true)
	s.NoError(err)
	opaMock.EXPECT().PostV1DataPackageRuleWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&openpolicyagent.PostV1DataPackageRuleResponse{
			JSON200: &openpolicyagent.OpaResponse{
				DecisionId: nil,
				Metrics:    nil,
				Result:     result,
			},
		}, nil,
	).AnyTimes()
	s.opa = opaMock

	s.conn = createServerConnection(s.T(), s.dbClient, s.opa)
	s.client = catalogv3.NewCatalogServiceClient(s.conn)
	s.startTime = time.Now()
	s.createSomeTestEntities()
}

func (s *NorthBoundTestSuite) TearDownTest() {
	if s.conn != nil {
		_ = s.conn.Close()
		_ = s.dbClient.Close()
		s.cancel()
	}
	s.conn = nil
}

func TestNorthBound(t *testing.T) {
	suite.Run(t, &NorthBoundTestSuite{})
}

const (
	footen    = "footen"
	fooreg    = "fooreg"
	fooregalt = "fooregalt"
	barten    = "barten"
	barreg    = "barreg"
	barregalt = "barregalt"
	axeten    = "axeten"
	axereg    = "axereg"

	genten = "genten"
	genreg = "genreg"
)

func (s *NorthBoundTestSuite) createSomeTestEntities() {
	if !s.populateDB {
		return
	}
	s.createRegistry(footen, fooreg, helmType)
	s.createRegistry(footen, fooregalt, imageType)

	s.createRegistry(barten, barreg, helmType)
	s.createRegistry(barten, barregalt, imageType)

	s.createRegistry(axeten, axereg, helmType)
	s.createRegistry(genten, genreg, helmType)

	s.createArtifact(footen, "icon", "Fancy Icon", "Icon of a bird", "image/png", asBinary(kingfisherPngB64))
	s.createArtifact(footen, "thumb", "Formal Thumbnail", "Glyph of a serious avian", "image/jpeg", asBinary(falconJpegB64))
	s.createArtifact(barten, "icon", "Silly Icon", "Icon of a dopey bird", "application/json", []byte(sillyIconJSON))
	s.createArtifact(barten, "thumb", "Silly Thumbnail", "Glyph of a goofy avian", "application/yaml", []byte(goofyAvianYAML))

	// Create a set of app for footen tenant
	s.createApp(footen, fooreg, "foo", "v0.1.0", 2)
	s.createApp(footen, fooreg, "goo", "v0.1.2", 3)
	s.createApp(footen, fooreg, "bar", "v0.2.0", 3)
	s.createApp(footen, fooreg, "bar", "v0.2.1", 4)

	// And a set of identical apps for barten tenant
	s.createApp(barten, barreg, "foo", "v0.1.0", 2)
	s.createApp(barten, barreg, "goo", "v0.1.2", 3)
	s.createApp(barten, barreg, "bar", "v0.2.0", 3)
	s.createApp(barten, barreg, "bar", "v0.2.1", 4)

	s.createDeploymentPkg(footen, "ca-gigi", "v0.2.1", "foo:v0.1.0", "bar:v0.2.1")
	s.createDeploymentProfile(footen, "ca-gigi", "v0.2.1", "cp-1", map[string]string{"foo": "p2", "bar": "p1"})
	s.createDeploymentProfile(footen, "ca-gigi", "v0.2.1", "cp-2", map[string]string{"foo": "p1", "bar": "p3"})

	s.createDeploymentPkg(footen, "ca-gigi", "v0.3.4", "foo:v0.1.0", "bar:v0.2.1", "goo:v0.1.2")
	s.createDeploymentProfile(footen, "ca-gigi", "v0.3.4", "cp-1", map[string]string{"foo": "p2", "bar": "p1", "goo": "p2"})
	s.createDeploymentProfile(footen, "ca-gigi", "v0.3.4", "cp-2", map[string]string{"foo": "p1", "bar": "p3", "goo": "p3"})

	// Create some barten tenant packages
	s.createDeploymentPkg(barten, "ca-fifi", "v0.2.0", "foo:v0.1.0", "bar:v0.2.0", "goo:v0.1.2")
	s.createDeploymentProfile(barten, "ca-fifi", "v0.2.0", "cp-1", map[string]string{"foo": "p2", "bar": "p2", "goo": "p1"})
	s.createDeploymentProfile(barten, "ca-fifi", "v0.2.0", "cp-2", map[string]string{"foo": "p2", "bar": "p3", "goo": "p1"})

	// Duplicate barten packages for the footen tenant.
	s.createDeploymentPkg(footen, "ca-fifi", "v0.2.0", "foo:v0.1.0", "bar:v0.2.0", "goo:v0.1.2")
	s.createDeploymentProfile(footen, "ca-fifi", "v0.2.0", "cp-1", map[string]string{"foo": "p2", "bar": "p2", "goo": "p1"})
	s.createDeploymentProfile(footen, "ca-fifi", "v0.2.0", "cp-2", map[string]string{"foo": "p2", "bar": "p3", "goo": "p1"})
}

func (s *NorthBoundTestSuite) validateResponse(err error, r interface{}) {
	s.NoError(err)
	s.NotNil(r)
}

func (s *NorthBoundTestSuite) validateNotFound(err error, r interface{}) {
	s.validateError(err, codes.NotFound, r)
}

func (s *NorthBoundTestSuite) validateFailedPrecondition(err error, r interface{}) {
	s.validateError(err, codes.FailedPrecondition, r)
}

func (s *NorthBoundTestSuite) validateError(err error, code codes.Code, r any) {
	s.Error(err)
	s.Equal(code, status.Code(err))
	s.Nil(r)
}

// Appends ActiveProjectID metadata to the outgoing context
func (s *NorthBoundTestSuite) ProjectID(projectUUID string) context.Context {
	return metadata.AppendToOutgoingContext(s.ctx, ActiveProjectID, projectUUID)
}

// Appends ActiveProjectID metadata to a new (server) incoming context
func (s *NorthBoundTestSuite) ServerProjectID(projectUUID string) context.Context {
	return metadata.NewIncomingContext(s.ctx, metadata.New(map[string]string{ActiveProjectID: projectUUID}))
}

// Creates a test registry with the given name
func (s *NorthBoundTestSuite) createRegistry(project string, name string, registryType string) *catalogv3.Registry {
	reg := &catalogv3.Registry{
		Name:        name,
		DisplayName: fmt.Sprintf("Registry %s", name),
		Description: fmt.Sprintf("Registry that holds %s", name),
		RootUrl:     fmt.Sprintf("http://%s.com/%s", project, name),
		Username:    "admin",
		AuthToken:   "token",
		Cacerts:     "cacerts",
		Type:        registryType,
	}
	resp, err := s.client.CreateRegistry(s.ProjectID(project), &catalogv3.CreateRegistryRequest{Registry: reg})
	s.validateResponse(err, resp)
	s.validateRegistry(resp.Registry, reg.Name, reg.DisplayName, reg.Description, reg.RootUrl, reg.Username, reg.AuthToken, reg.Cacerts)
	return resp.Registry
}

func (s *NorthBoundTestSuite) validateRegistry(reg *catalogv3.Registry, name string, display string, description string, url string, user string, token string, certs string) {
	s.Equal(name, reg.Name)
	s.Equal(display, reg.DisplayName)
	s.Equal(description, reg.Description)
	s.Equal(url, reg.RootUrl)
	s.Equal(user, reg.Username)
	s.Equal(token, reg.AuthToken)
	if strings.HasSuffix(certs, "...") {
		s.True(strings.HasPrefix(reg.Cacerts, strings.TrimSuffix(certs, "...")))
	} else {
		s.Equal(certs, reg.Cacerts)
	}
}

func (s *NorthBoundTestSuite) createArtifact(project string, name string, display string, description string, mimeType string, value []byte) *catalogv3.Artifact {
	r, err := s.client.CreateArtifact(s.ProjectID(project), &catalogv3.CreateArtifactRequest{
		Artifact: &catalogv3.Artifact{Name: name, DisplayName: display, Description: description,
			MimeType: mimeType, Artifact: value},
	})
	s.validateResponse(err, r)
	s.validateArtifact(r.Artifact, name, display, description, mimeType, value)
	return r.Artifact
}

func (s *NorthBoundTestSuite) validateArtifact(art *catalogv3.Artifact, name string, display string, description string, mimeType string, value []byte) {
	s.Equal(name, art.Name)
	s.Equal(display, art.DisplayName)
	s.Equal(description, art.Description)
	s.Equal(mimeType, art.MimeType)
	if value != nil {
		s.Equal(value, art.Artifact)
	}
}

// Creates a test application with the given name, version and with a specified number of test application profiles.
func (s *NorthBoundTestSuite) createApp(project string, registry string, name string, ver string, profileCount int) *catalogv3.Application {
	profiles := make([]*catalogv3.Profile, 0, profileCount)
	for pi := 1; pi <= profileCount; pi++ {
		profiles = append(profiles, &catalogv3.Profile{
			Name:        fmt.Sprintf("p%d", pi),
			DisplayName: fmt.Sprintf("Profile %d for %s", pi, name),
			Description: fmt.Sprintf("This is a profile #%d for %s", pi, name),
			ChartValues: fmt.Sprintf("something: nothing\nanything: everything\n%s: %s\n", name, ver),
		})
	}

	app := &catalogv3.Application{
		Name:             name,
		Version:          ver,
		Kind:             catalogv3.Kind_KIND_NORMAL,
		DisplayName:      fmt.Sprintf("Application %s", name),
		Description:      fmt.Sprintf("This is application %s", name),
		ChartName:        fmt.Sprintf("%s-chart", name),
		ChartVersion:     ver,
		HelmRegistryName: registry,
		Profiles:         profiles,
	}
	if len(profiles) > 0 {
		app.DefaultProfileName = "p1"
	}

	r, err := s.client.CreateApplication(s.ProjectID(project), &catalogv3.CreateApplicationRequest{Application: app})
	s.validateResponse(err, r)
	s.validateApp(r.Application, app.Name, app.Version, app.DisplayName, app.Description, len(app.Profiles),
		app.DefaultProfileName, app.ChartName, app.ChartVersion, app.HelmRegistryName)
	return r.Application
}

func (s *NorthBoundTestSuite) validateApp(app *catalogv3.Application, name string, ver string, display string,
	description string, profileCount int, defaultProfile string, chart string, chartVer string, registry string) {
	s.Equal(name, app.Name)
	s.Equal(ver, app.Version)
	s.Equal(display, app.DisplayName)
	s.Equal(description, app.Description)
	s.Equal(chart, app.ChartName)
	s.Equal(chartVer, app.ChartVersion)
	s.Len(app.Profiles, profileCount)
	s.Equal(defaultProfile, app.DefaultProfileName)
	s.Equal(registry, app.HelmRegistryName)
}

//func (s *NorthBoundTestSuite) validateProfile(profile *catalogv3.Profile, name string, display string, description string, values string) {
//	s.Equal(name, profile.Name)
//	s.Equal(display, profile.DisplayName)
//	s.Equal(description, profile.Description)
//	s.Equal(values, profile.ChartValues)
//}

func (s *NorthBoundTestSuite) validateParameterTemplate(templ *catalogv3.ParameterTemplate, name string, typ string, deflt string, suggestedValues []string) {
	s.Len(templ.SuggestedValues, len(suggestedValues))
	s.Equal(name, templ.Name)
	s.Equal(typ, templ.Type)
	s.Equal(deflt, templ.Default)
	for i, sv := range suggestedValues {
		s.Equal(templ.SuggestedValues[i], sv)
	}
}

// Creates a test deployment package of the given name and version, consisting of the specified app references.
func (s *NorthBoundTestSuite) createDeploymentPkg(project string, name string, ver string, appRefs ...string) *catalogv3.DeploymentPackage {
	app := catalogv3.DeploymentPackage{
		Name:                  name,
		Kind:                  catalogv3.Kind_KIND_NORMAL,
		DisplayName:           fmt.Sprintf("Deployment Package %s", name),
		Description:           fmt.Sprintf("This is deployment package %s", name),
		Version:               ver,
		ApplicationReferences: appReferences(appRefs...),
	}
	resp, err := s.client.CreateDeploymentPackage(s.ProjectID(project), &catalogv3.CreateDeploymentPackageRequest{DeploymentPackage: &app})
	s.validateResponse(err, resp)
	s.validateDeploymentPkg(resp.DeploymentPackage, app.Name, app.Version, app.DisplayName, app.Description,
		"", "", len(app.ApplicationReferences), 0, 0, "", 0, false)
	return resp.DeploymentPackage
}

func (s *NorthBoundTestSuite) validateDeploymentPkg(app *catalogv3.DeploymentPackage, name string, ver string, display string, description string, _ string, _ string,
	refCount int, depCount int, profileCount int, defaultProfile string, nsCount int, deployed bool) {
	s.Equal(name, app.Name)
	s.Equal(ver, app.Version)
	s.Equal(display, app.DisplayName)
	s.Equal(description, app.Description)
	s.Len(app.ApplicationReferences, refCount)
	s.Len(app.ApplicationDependencies, depCount)
	s.Len(app.Profiles, profileCount)
	s.Equal(defaultProfile, app.DefaultProfileName)
	s.Len(app.DefaultNamespaces, nsCount)
	s.Equal(app.IsDeployed, deployed)
}

//func (s *NorthBoundTestSuite) validateApplicationReferences(a []*catalogv3.ApplicationReference, b []*catalogv3.ApplicationReference) {
//	sort.SliceStable(a, func(i int, j int) bool { return a[i].Name < a[j].Name })
//	sort.SliceStable(b, func(i int, j int) bool { return b[i].Name < b[j].Name })
//	s.Len(b, len(a))
//	for i := range a {
//		s.Equal(a[i].Name, b[i].Name)
//		s.Equal(a[i].Version, b[i].Version)
//	}
//}

// Creates a test deployment profile with the given name, version and consisting of given application profiles
func (s *NorthBoundTestSuite) createDeploymentProfile(project string, appName string, appVer string, name string, profiles map[string]string) {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(project), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: appName,
		Version:               appVer,
	})
	s.NoError(err)

	// Remove the implicit default, if there is one
	if len(resp.DeploymentPackage.Profiles) == 1 && resp.DeploymentPackage.Profiles[0].Name == "implicit-default" {
		resp.DeploymentPackage.Profiles = []*catalogv3.DeploymentProfile{}
		resp.DeploymentPackage.DefaultProfileName = name
	}

	resp.DeploymentPackage.Profiles = append(resp.DeploymentPackage.Profiles, &catalogv3.DeploymentProfile{
		Name:                name,
		DisplayName:         fmt.Sprintf("Deployment Profile %s", name),
		Description:         fmt.Sprintf("This is a deployment profile %s for %s:%s", name, appName, appVer),
		ApplicationProfiles: profiles,
	})

	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(project), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: appName, Version: appVer, DeploymentPackage: resp.DeploymentPackage,
	})
	s.NoError(err)
}

// Deletes test deployment profile with the given names
func (s *NorthBoundTestSuite) deleteDeploymentProfiles(project string, appName string, appVer string, names ...string) {
	resp, err := s.client.GetDeploymentPackage(s.ProjectID(project), &catalogv3.GetDeploymentPackageRequest{
		DeploymentPackageName: appName, Version: appVer,
	})
	s.NoError(err)

	// Remove the implicit default, if there is one
	if len(resp.DeploymentPackage.Profiles) == 1 && resp.DeploymentPackage.Profiles[0].Name == "implicit-default" {
		resp.DeploymentPackage.Profiles = []*catalogv3.DeploymentProfile{}
		resp.DeploymentPackage.DefaultProfileName = ""
	}

	for _, name := range names {
		for i, profile := range resp.DeploymentPackage.Profiles {
			if name == profile.Name {
				resp.DeploymentPackage.Profiles = append(resp.DeploymentPackage.Profiles[:i], resp.DeploymentPackage.Profiles[i+1:]...)
				break
			}
		}
		if name == resp.DeploymentPackage.DefaultProfileName && len(resp.DeploymentPackage.Profiles) > 0 {
			resp.DeploymentPackage.DefaultProfileName = resp.DeploymentPackage.Profiles[0].Name
		}
	}
	if len(resp.DeploymentPackage.Profiles) == 0 {
		resp.DeploymentPackage.DefaultProfileName = ""
	}

	_, err = s.client.UpdateDeploymentPackage(s.ProjectID(project), &catalogv3.UpdateDeploymentPackageRequest{
		DeploymentPackageName: appName, Version: appVer, DeploymentPackage: resp.DeploymentPackage,
	})
	s.NoError(err)
}

// Creates application references from an array of string references, e.g. "goo:v0.1.0" or "goo:v0.1.0:goopub"
func appReferences(refs ...string) []*catalogv3.ApplicationReference {
	references := make([]*catalogv3.ApplicationReference, 0, len(refs))
	for _, ref := range refs {
		f := strings.Split(ref, ":")
		if len(f) >= 2 {
			references = append(references, &catalogv3.ApplicationReference{Name: f[0], Version: f[1]})
		}
	}
	return references
}

func asBinary(b64 string) []byte {
	decodeString, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}
	return decodeString
}

const sillyIconJSON = `{"icon": "silly"}`

const goofyAvianYAML = `# a comment
avian: 
  glyph: goofy
`

const kingfisherPngB64 = `iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhk
iAAAAAlwSFlzAAAAsQAAALEBxi1JjQAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3Nj
YXBlLm9yZ5vuPBoAAALzSURBVEiJrVVLSFRRGP7+c+/cOzpNZmFFuCgtXAwZQaSV
UGYuiqLHMqIoqKAWLcrKFjKbIIioIFpEC6UkQoIeYJklPSnC6UFUU6JTYkqo4wud
mXvnnr+FOk7O3Jkp+nfnP9/j/+49nEP4y6p6xmujplkgsvjVudW6Px2eMhX2fg2V
DQe0u+EhkTvZU7PkiHBFD16o0G/Y8URm4qObJPNTyZwb34+GhNsc0OqPtkR2/LPB
ifZgjmS6NomVZhiDnxpj+2yBzBFR6/F4tGR8NZ2BbmQdAvHsybUR7MTPB2eQU7QB
pI5rWibc7rGB0dKCBd+IySdZ+iDIJwzhS/kPvK2cLWeEAgDmAsBwQIMxrPwJIqCn
6RR6WuoS+ATqSJnAcoX20YS4XQlhck9LXWxQBgaJ8I4YPglutjU40MoOotCxVOIk
GNnOkNy2c2Xj7fo39aQI3+v2rvZxn1jA5FXzJbQHxLXxvfhPtD/vCDxoGU+hO76r
a94tSpowqTozQaDKzlyoMiYOANKILuSPJfMyNvD6Q9vB7LEzmOXsnT4QrFFZmbGB
JFTbiQPAcmdTQo8ZRRkZ1HwOVwJYYSeu6RFspIsJfYLSmQyfeIqEtJ2eiLF15lmo
MKZvQHE5nyflxC+8n8dKpcCr5EDG6mAdysOXE/c0Negoez8nbQJLoDrZudUQxjac
RxE3Th3w+FIcia7TE5z2B4sN1j8wTVmoMLEUT1COesxEP6z+QXAkMs1dD2hlvgI7
g1iCLXTJ241CmHDChUHkoROL8RY6xuy4ELrWrRjuYlvAZAJuW1/IFvmR5naNJVBI
Cod+U1nVuosIMhVHBQBp0UlKd3Uzs1AVPwZG88WPrlt02NybEj9RgtvW5ROw225o
gF9yOHw9+qsPxpixSXQEriBqbmZv+rcEAFQpxXEC4l+jCAjNBGqAijtU+HiIXxS5
LemYDwUEwlUwliHD55akv6IPgAughwRugIPuUeGjoUzIGSUghUsAvZeW3B/+X6Lx
9Rt4NR9AKH0xsAAAAABJRU5ErkJggg==`

const falconJpegB64 = `/9j/4AAQSkZJRgABAQIAAQABAAD/4QeaRXhpZgAASUkqAAgAAAAHABIBAwABAAAA
AQAAABoBBQABAAAAYgAAABsBBQABAAAAagAAACgBAwABAAAAAwAAADEBAgANAAAA
cgAAADIBAgAUAAAAgAAAAGmHBAABAAAAlAAAAKYAAACaAAAAVwAAAJoAAABXAAAA
R0lNUCAyLjEwLjMwAAAyMDIzOjA0OjI3IDE0OjM4OjE0AAEAAaADAAEAAAABAAAA
AAAAAAkA/gAEAAEAAAABAAAAAAEEAAEAAAAAAQAAAQEEAAEAAAAAAQAAAgEDAAMA
AAAYAQAAAwEDAAEAAAAGAAAABgEDAAEAAAAGAAAAFQEDAAEAAAADAAAAAQIEAAEA
AAAeAQAAAgIEAAEAAABzBgAAAAAAAAgACAAIAP/Y/+AAEEpGSUYAAQEAAAEAAQAA
/9sAQwAIBgYHBgUIBwcHCQkICgwUDQwLCwwZEhMPFB0aHx4dGhwcICQuJyAiLCMc
HCg3KSwwMTQ0NB8nOT04MjwuMzQy/9sAQwEJCQkMCwwYDQ0YMiEcITIyMjIyMjIy
MjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIy/8AAEQgB
AAEAAwEiAAIRAQMRAf/EAB8AAAEFAQEBAQEBAAAAAAAAAAABAgMEBQYHCAkKC//E
ALUQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEV
UtHwJDNicoIJChYXGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVm
Z2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrC
w8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+v/EAB8BAAMBAQEB
AQEBAQEAAAAAAAABAgMEBQYHCAkKC//EALURAAIBAgQEAwQHBQQEAAECdwABAgMR
BAUhMQYSQVEHYXETIjKBCBRCkaGxwQkjM1LwFWJy0QoWJDThJfEXGBkaJicoKSo1
Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoKDhIWGh4iJipKT
lJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uLj5OXm
5+jp6vLz9PX29/j5+v/aAAwDAQACEQMRAD8A+f6KKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiig
AooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKAP/9kA
/+EMeWh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8APD94cGFja2V0IGJlZ2lu
PSLvu78iIGlkPSJXNU0wTXBDZWhpSHpyZVN6TlRjemtjOWQiPz4gPHg6eG1wbWV0
YSB4bWxuczp4PSJhZG9iZTpuczptZXRhLyIgeDp4bXB0az0iWE1QIENvcmUgNC40
LjAtRXhpdjIiPiA8cmRmOlJERiB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3Jn
LzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiPiA8cmRmOkRlc2NyaXB0aW9uIHJk
ZjphYm91dD0iIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAv
MS4wL21tLyIgeG1sbnM6c3RFdnQ9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEu
MC9zVHlwZS9SZXNvdXJjZUV2ZW50IyIgeG1sbnM6ZGM9Imh0dHA6Ly9wdXJsLm9y
Zy9kYy9lbGVtZW50cy8xLjEvIiB4bWxuczpHSU1QPSJodHRwOi8vd3d3LmdpbXAu
b3JnL3htcC8iIHhtbG5zOnhtcD0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4w
LyIgeG1wTU06RG9jdW1lbnRJRD0iZ2ltcDpkb2NpZDpnaW1wOjZjODQ5NTMwLTgx
NDAtNDRhYS1iMGM0LTMwYjQyMTI0NmEwNSIgeG1wTU06SW5zdGFuY2VJRD0ieG1w
LmlpZDowMzI5ZDgwNC1lMzg0LTQ5MDctODE2Ni1lM2E2MjVjMWRmY2YiIHhtcE1N
Ok9yaWdpbmFsRG9jdW1lbnRJRD0ieG1wLmRpZDpiYjZjOGZjMS01OTA2LTRkNWQt
YmUwMC0yYzk5ZGQzN2JhYWQiIGRjOkZvcm1hdD0iaW1hZ2UvanBlZyIgR0lNUDpB
UEk9IjIuMCIgR0lNUDpQbGF0Zm9ybT0iTWFjIE9TIiBHSU1QOlRpbWVTdGFtcD0i
MTY4MjYwMjcwODI4Mzg4MiIgR0lNUDpWZXJzaW9uPSIyLjEwLjMwIiB4bXA6Q3Jl
YXRvclRvb2w9IkdJTVAgMi4xMCI+IDx4bXBNTTpIaXN0b3J5PiA8cmRmOlNlcT4g
PHJkZjpsaSBzdEV2dDphY3Rpb249InNhdmVkIiBzdEV2dDpjaGFuZ2VkPSIvIiBz
dEV2dDppbnN0YW5jZUlEPSJ4bXAuaWlkOmQzOTQyMGM4LWNhODgtNDMwOS1iODlh
LTZmYjc0YmJjNWQxYiIgc3RFdnQ6c29mdHdhcmVBZ2VudD0iR2ltcCAyLjEwIChN
YWMgT1MpIiBzdEV2dDp3aGVuPSIyMDIzLTA0LTI3VDE0OjM4OjI4KzAxOjAwIi8+
IDwvcmRmOlNlcT4gPC94bXBNTTpIaXN0b3J5PiA8L3JkZjpEZXNjcmlwdGlvbj4g
PC9yZGY6UkRGPiA8L3g6eG1wbWV0YT4gICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAg
ICAgICAgICA8P3hwYWNrZXQgZW5kPSJ3Ij8+/+ICsElDQ19QUk9GSUxFAAEBAAAC
oGxjbXMEMAAAbW50clJHQiBYWVogB+cABAAbAA0AJQA4YWNzcEFQUEwAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAPbWAAEAAAAA0y1sY21zAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANZGVzYwAAASAAAABAY3By
dAAAAWAAAAA2d3RwdAAAAZgAAAAUY2hhZAAAAawAAAAsclhZWgAAAdgAAAAUYlhZ
WgAAAewAAAAUZ1hZWgAAAgAAAAAUclRSQwAAAhQAAAAgZ1RSQwAAAhQAAAAgYlRS
QwAAAhQAAAAgY2hybQAAAjQAAAAkZG1uZAAAAlgAAAAkZG1kZAAAAnwAAAAkbWx1
YwAAAAAAAAABAAAADGVuVVMAAAAkAAAAHABHAEkATQBQACAAYgB1AGkAbAB0AC0A
aQBuACAAcwBSAEcAQm1sdWMAAAAAAAAAAQAAAAxlblVTAAAAGgAAABwAUAB1AGIA
bABpAGMAIABEAG8AbQBhAGkAbgAAWFlaIAAAAAAAAPbWAAEAAAAA0y1zZjMyAAAA
AAABDEIAAAXe///zJQAAB5MAAP2Q///7of///aIAAAPcAADAblhZWiAAAAAAAABv
oAAAOPUAAAOQWFlaIAAAAAAAACSfAAAPhAAAtsRYWVogAAAAAAAAYpcAALeHAAAY
2XBhcmEAAAAAAAMAAAACZmYAAPKnAAANWQAAE9AAAApbY2hybQAAAAAAAwAAAACj
1wAAVHwAAEzNAACZmgAAJmcAAA9cbWx1YwAAAAAAAAABAAAADGVuVVMAAAAIAAAA
HABHAEkATQBQbWx1YwAAAAAAAAABAAAADGVuVVMAAAAIAAAAHABzAFIARwBC/9sA
QwADAgIDAgIDAwMDBAMDBAUIBQUEBAUKBwcGCAwKDAwLCgsLDQ4SEA0OEQ4LCxAW
EBETFBUVFQwPFxgWFBgSFBUU/9sAQwEDBAQFBAUJBQUJFA0LDRQUFBQUFBQUFBQU
FBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQU/8IAEQgAGAAY
AwERAAIRAQMRAf/EABgAAAMBAQAAAAAAAAAAAAAAAAQCBQMH/8QAGAEAAwEBAAAA
AAAAAAAAAAAAAgMEBQb/2gAMAwEAAhADEAAAAeQdxxAU1WYGQ1IqXoBOYz4662hn
/wD/xAAcEAADAQACAwAAAAAAAAAAAAACAwEABBESFCP/2gAIAQEAAQUCzW/Rpetj
O0meCd1T5FRQy+NIQEK9/8QAJBEAAgIAAwkBAAAAAAAAAAAAAQIAAxESQRMhIjEy
YYGxweH/2gAIAQMBAT8BldfCbGEQbbEZYiADO/L3Ez2Yt4m5aiun2C0NoMe8e4kE
Kw/IwZ+phP/EACIRAAIBAwIHAAAAAAAAAAAAAAECAAMREhNBIjEyYYHB4f/aAAgB
AgEBPwGVH4hTUx20bHKMxJwTnHwp2XzLFqobf1NIrubdolAAgsD9ilV6Vn//xAAk
EAABAgQFBQAAAAAAAAAAAAACAQARA1ExEiJBYoETITJCYf/aAAgBAQAGPwJjKEoK
t1o0XqR2lq8AeWq0YDGy4lVgXtSgtcxqOy/LFSlzYw7r9eWUScP/xAAfEAACAgIC
AwEAAAAAAAAAAAABEQAxcSFBYVHB0fD/2gAIAQEAAT8hjfEOvDMadazNOV3PWOD9
hmHs3Xk9moCfp3jG8k/lEImCSBptyXMNeUxL73BSZ2eeZ//aAAwDAQACAAMAAAAQ
oDGP/wD/xAAhEQEAAQQCAQUAAAAAAAAAAAABEUEAITFRYYGRscHR8f/aAAgBAwEB
PxC4JoGjnO+YKxVC4+D2UaTSH1vwAFV9cvzYFzZApTB0GXiu7UVmETzMcdAR+3Ao
ACdjFAdcvdx1M4IPDVGLclUd+1//xAAhEQEAAQQCAQUAAAAAAAAAAAABEUEAITFR
YYFxkbHR8f/aAAgBAgEBPxC5You3jGuJaTQbmZE0alYrJ7UvzgNA++C0humTX1e1
wc01clYlMcQTPas/lskUUxo55TfB1bVbGWXy3UmzYQnr5v8A/8QAHRABAQACAwEB
AQAAAAAAAAAAAREAIVExQXFhwf/aAAgBAQABPxDG3doUAUF1QKXwWZsfoAlYFJQW
8PU9x8goqlV68rw/mCUAL0AOz1EB21nWAS38wnOCBn4YBO+VoqhbHR1rnHaOxcIQ
lbF+TGBp2irlWv1z/9k=`

var expectedAuthError = status.Errorf(codes.InvalidArgument, "invalid: authentication failed")

func (s *NorthBoundTestSuite) newMockOPAServer() *Server {
	mockController := gomock.NewController(s.T())
	opaMock := openpolicyagent.NewMockClientWithResponsesInterface(mockController)
	opaMock.EXPECT().PostV1DataPackageRuleWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, expectedAuthError).AnyTimes()
	return NewServer(s.dbClient, opaMock)
}
