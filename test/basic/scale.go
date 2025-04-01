// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"context"
	"fmt"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math"
	"sync"
	"time"
)

var (
	registryMetric     = NewLatency("Creating Registries")
	artifactMetric     = NewLatency("Creating Artifacts")
	packageMetric      = NewLatency("Creating Packages")
	applicationsMetric = NewLatency("Creating Artifacts")
)

func (s *TestSuite) generateRegistries(projectUUID string, registries countRange) []*catalogv3.Registry {
	list := make([]*catalogv3.Registry, 0, registries.getCount())
	for i := 1; i <= cap(list); i++ {
		start := time.Now()
		registry := s.createRegistry(projectUUID, fmt.Sprintf("%s-reg%d", projectUUID, i),
			fmt.Sprintf("Registry %d of project %s", i, projectUUID),
			fmt.Sprintf("This is project %s registry %d", projectUUID, i),
			fmt.Sprintf("https://%s-reg-%d.com/artifacts", projectUUID, i))
		registryMetric.Add(time.Since(start))
		list = append(list, registry)
	}
	return list
}

func (s *TestSuite) generateArtifacts(projectUUID string, artifacts countRange) {
	count := artifacts.getCount()
	for i := 1; i <= count; i++ {
		start := time.Now()
		s.createArtifact(projectUUID, fmt.Sprintf("art%05d", i),
			fmt.Sprintf("Artifact %05d of %s", i, projectUUID),
			fmt.Sprintf("Some artifact %05d of %s", i, projectUUID),
			"image/png", s.ArtifactFilename)
		artifactMetric.Add(time.Since(start))
	}
}

func (s *TestSuite) generateApps(projectUUID string, registries int, apps countRange, spec *appSpec) []*catalogv3.Application {
	list := make([]*catalogv3.Application, 0, apps.getCount())
	for ai := 1; ai <= cap(list); ai++ {
		versions := spec.versions.getCount()
		registryName := fmt.Sprintf("%s-reg%d", projectUUID, oneOf(registries))
		for vi := 1; vi < versions; vi++ {
			appName := fmt.Sprintf("app%03d", ai)
			profiles := generateAppProfiles(appName, spec.profiles)
			defaultProfile := fmt.Sprintf("p%d", oneOf(len(profiles)))
			start := time.Now()
			app := s.createApplication(projectUUID, registryName, appName, fmt.Sprintf("%d.0", vi),
				fmt.Sprintf("App %d v%d.0.0 of %s", ai, vi, projectUUID),
				fmt.Sprintf("Some app %d v%d.0.0 of %s", ai, vi, projectUUID),
				profiles, defaultProfile)
			applicationsMetric.Add(time.Since(start))
			if app != nil {
				list = append(list, app)
			}
		}
	}
	return list
}

func (s *TestSuite) generatePackages(projectUUID string, apps []*catalogv3.Application, packages countRange, spec *appPackageSpec) []*catalogv3.DeploymentPackage {
	list := make([]*catalogv3.DeploymentPackage, 0, packages.getCount())
	for ai := 1; ai <= cap(list); ai++ {
		versions := spec.versions.getCount()
		for vi := 1; vi < versions; vi++ {
			appName := fmt.Sprintf("package%03d", ai)
			packageApps := selectPackageApps(apps, spec.apps)
			references := generateAppReferences(packageApps)
			extensions := generatePackageExtensions(appName, spec.extensions, spec.endpoints)
			artifacts := generatePackageArtifacts(spec.artifacts)
			profiles := generatePackageProfiles(appName, spec.profiles, packageApps)
			defaultProfile := fmt.Sprintf("p%d", oneOf(len(profiles)))
			start := time.Now()
			app := s.createPackage(projectUUID, appName, fmt.Sprintf("%d.0", vi),
				fmt.Sprintf("App %d v%d.0.0 of %s", ai, vi, projectUUID),
				fmt.Sprintf("Some app %d v%d.0.0 of %s", ai, vi, projectUUID),
				references, profiles, defaultProfile, extensions, artifacts)
			packageMetric.Add(time.Since(start))
			list = append(list, app)
		}
	}
	return list
}

func generateAppProfiles(appName string, profiles countRange) []*catalogv3.Profile {
	list := make([]*catalogv3.Profile, 0, profiles.getCount())
	for i := 1; i <= cap(list); i++ {
		list = append(list, &catalogv3.Profile{
			Name:        fmt.Sprintf("p%d", i),
			DisplayName: fmt.Sprintf("Profile %d of %s", i, appName),
			Description: fmt.Sprintf("%s profile %d", appName, i),
			ChartValues: fmt.Sprintf("app: %s\nprofile: %d", appName, i),
		})
	}
	return list
}

func selectPackageApps(apps []*catalogv3.Application, applications countRange) []*catalogv3.Application {
	list := make([]*catalogv3.Application, 0, int(math.Min(float64(applications.getCount()), float64(len(apps)/2))))
	indexes := make(map[int]*catalogv3.Application, cap(list))
	names := make(map[string]*catalogv3.Application, cap(list))

	for len(list) < cap(list) {
		// Randomly select some apps from the given list - making sure to select only one version of the same app
		i := oneOf(len(apps)) - 1
		if _, ok := indexes[i]; !ok { // if we did not pick one we already did
			app := apps[i]
			if _, ok = names[app.Name]; !ok { // if we did not pick the same app name
				list = append(list, apps[i])
				indexes[i] = apps[i]
				names[apps[i].Name] = apps[i]
			}
		}
	}
	return list
}

func generateAppReferences(apps []*catalogv3.Application) []*catalogv3.ApplicationReference {
	list := make([]*catalogv3.ApplicationReference, 0, len(apps))
	for _, app := range apps {
		list = append(list, &catalogv3.ApplicationReference{Name: app.Name, Version: app.Version})
	}
	return list
}

func generatePackageExtensions(appName string, extensions countRange, endpoints countRange) []*catalogv3.APIExtension {
	list := make([]*catalogv3.APIExtension, 0, extensions.getCount())
	for i := 1; i <= cap(list); i++ {
		extName := fmt.Sprintf("ext%d", i)
		endpoints := generateEndpoints(appName, extName, endpoints)
		list = append(list, &catalogv3.APIExtension{
			Name:        extName,
			Version:     fmt.Sprintf("v%d.0", i),
			DisplayName: fmt.Sprintf("Extension %d of %s", i, appName),
			Description: fmt.Sprintf("App %s extension %d", appName, i),
			Endpoints:   endpoints,
			UiExtension: nil,
		})
	}
	return list
}

func generateEndpoints(appName string, extName string, endpoints countRange) []*catalogv3.Endpoint {
	list := make([]*catalogv3.Endpoint, 0, endpoints.getCount())
	for i := 1; i <= cap(list); i++ {
		list = append(list, &catalogv3.Endpoint{
			ServiceName:  fmt.Sprintf("svc%d", i),
			ExternalPath: fmt.Sprintf("external/%s/%s/svc%d", appName, extName, i),
			InternalPath: fmt.Sprintf("internal/%s/%s/svc%d", appName, extName, i),
			Scheme:       "http",
			AuthType:     "insecure",
		})
	}
	return list
}

func generatePackageArtifacts(artifacts countRange) []*catalogv3.ArtifactReference {
	list := make([]*catalogv3.ArtifactReference, 0, artifacts.getCount())
	return list
}

func generatePackageProfiles(appName string, profiles countRange, _ []*catalogv3.Application) []*catalogv3.DeploymentProfile {
	list := make([]*catalogv3.DeploymentProfile, 0, profiles.getCount())
	for i := 1; i <= cap(list); i++ {
		list = append(list, &catalogv3.DeploymentProfile{
			Name:        fmt.Sprintf("p%d", i),
			DisplayName: fmt.Sprintf("Profile %d of %s", i, appName),
			Description: fmt.Sprintf("%s profile %d", appName, i),
		})
	}
	return list
}

const tenant = "tenant"

func (s *TestSuite) generateProject(projectUUID string, spec *projectSpec) {
	registries := s.generateRegistries(projectUUID, spec.registries)
	s.generateArtifacts(projectUUID, spec.artifacts)
	apps := s.generateApps(projectUUID, len(registries), spec.apps, spec.appSpec)
	s.generatePackages(projectUUID, apps, spec.packages, spec.appPackageSpec)
}
func (s *TestSuite) TestScale() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	s.SetContext(ctx)
	fmt.Printf("Scale test starting\n")

	pSpec := &projectSpec{
		registries: countRange{1, 3},
		artifacts:  countRange{20, 50},
		apps:       countRange{10, 100},
		appSpec: &appSpec{
			versions: countRange{1, 3},
			profiles: countRange{1, 4},
		},
		packages: countRange{10, 40},
		appPackageSpec: &appPackageSpec{
			versions:   countRange{1, 3},
			profiles:   countRange{1, 3},
			apps:       countRange{1, 4},
			deps:       countRange{0, 2},
			artifacts:  countRange{2, 5},
			extensions: countRange{0, 5},
			endpoints:  countRange{1, 3},
			namespaces: countRange{1, 4},
		},
	}

	// Populate the catalog with some root entities structured according to the given specs.
	start := time.Now()
	s.generateProject(tenant, pSpec)
	duration := time.Since(start)

	// Count all objects first
	reresp, _ := s.client.ListRegistries(s.AddHeaders(tenant), &catalogv3.ListRegistriesRequest{})
	arresp, _ := s.client.ListArtifacts(s.AddHeaders(tenant), &catalogv3.ListArtifactsRequest{})
	apresp, _ := s.client.ListApplications(s.AddHeaders(tenant), &catalogv3.ListApplicationsRequest{})
	paresp, _ := s.client.ListDeploymentPackages(s.AddHeaders(tenant), &catalogv3.ListDeploymentPackagesRequest{})

	fmt.Printf("Registries: %d\n", len(reresp.Registries))
	fmt.Printf("Artifacts: %d\n", len(arresp.Artifacts))
	fmt.Printf("Applications: %d\n", len(apresp.Applications))
	fmt.Printf("App Packages: %d\n", len(paresp.DeploymentPackages))

	fmt.Printf("Creation of All: %s\n", duration)

	fmt.Println(printer.Sprintf("%20s%14s%14s%14s%14s", "Metric", "Iterations", "Average(ms)", "Shortest(ms)", "Longest(ms)"))
	fmt.Println(registryMetric)
	fmt.Println(artifactMetric)
	fmt.Println(applicationsMetric)
	fmt.Println(packageMetric)

	// Measure performance of various operations with the repo having been populated
	listIterations := 10
	oneItemIterations := 100

	// Get: Publisher, Registry, Artifact, App, CompositeApp
	registry := oneItem(reresp.Registries)
	s.measureOperation("Get Registry", oneItemIterations, func() {
		_, err := s.client.GetRegistry(s.AddHeaders(tenant), &catalogv3.GetRegistryRequest{RegistryName: registry.Name})
		s.NoError(err)
	})

	art := oneItem(arresp.Artifacts)
	s.measureOperation("Get Artifact", oneItemIterations, func() {
		_, err := s.client.GetArtifact(s.AddHeaders(tenant), &catalogv3.GetArtifactRequest{ArtifactName: art.Name})
		s.NoError(err)
	})

	app := oneItem(apresp.Applications)
	s.measureOperation("Get Application", oneItemIterations, func() {
		_, err := s.client.GetApplication(s.AddHeaders(tenant), &catalogv3.GetApplicationRequest{ApplicationName: app.Name, Version: app.Version})
		s.NoError(err)
	})

	pkg := oneItem(paresp.DeploymentPackages)
	s.measureOperation("Get Deployment Package", oneItemIterations, func() {
		_, err := s.client.GetDeploymentPackage(s.AddHeaders(tenant), &catalogv3.GetDeploymentPackageRequest{DeploymentPackageName: pkg.Name, Version: pkg.Version})
		s.NoError(err)
	})

	// List (for tenant): Registries, Artifacts, Apps, DeploymentPackages
	s.measureOperation("List Registries", listIterations, func() { s.listRegistries(tenant) })
	s.measureOperation("List Artifacts", listIterations, func() { s.listArtifacts(tenant) })
	s.measureOperation("List Applications", listIterations, func() { s.listApplications(tenant) })
	s.measureOperation("List Deployment Packages", listIterations, func() { s.listDeploymentPackages(tenant) })

	fmt.Println("Scale test done!")
}

var printer = message.NewPrinter(language.English)

func (s *TestSuite) measureOperation(name string, iterations int, op func()) {
	start := time.Now()
	longestRun := time.Nanosecond
	shortestRun := time.Hour
	for i := 0; i < iterations; i++ {
		thisRunStart := time.Now()
		op()
		thisRunDuration := time.Since(thisRunStart)
		if thisRunDuration > longestRun {
			longestRun = thisRunDuration
		}
		if thisRunDuration < shortestRun {
			shortestRun = thisRunDuration
		}
	}
	duration := time.Since(start)
	avg := float64(duration.Milliseconds()) / float64(iterations)
	fmt.Println(printer.Sprintf("%30s%20d%20.2f%20d%20d", name, iterations, avg, shortestRun.Milliseconds(), longestRun.Milliseconds()))
}

func (s *TestSuite) listRegistries(projectUUID string) {
	_, err := s.client.ListRegistries(s.AddHeaders(projectUUID), &catalogv3.ListRegistriesRequest{})
	s.NoError(err)
}

func (s *TestSuite) listArtifacts(projectUUID string) {
	_, err := s.client.ListArtifacts(s.AddHeaders(projectUUID), &catalogv3.ListArtifactsRequest{})
	s.NoError(err)
}
func (s *TestSuite) listApplications(projectUUID string) {
	_, err := s.client.ListApplications(s.AddHeaders(projectUUID), &catalogv3.ListApplicationsRequest{})
	s.NoError(err)
}

func (s *TestSuite) listDeploymentPackages(projectUUID string) {
	_, err := s.client.ListDeploymentPackages(s.AddHeaders(projectUUID), &catalogv3.ListDeploymentPackagesRequest{})
	s.NoError(err)
}

func (s *TestSuite) TestScaleWorkloads() {
	fmt.Printf("Scale workloads test starting\n")

	pSpec := &projectSpec{
		registries: countRange{1, 3},
		artifacts:  countRange{20, 50},
		apps:       countRange{10, 100},
		appSpec: &appSpec{
			versions: countRange{1, 3},
			profiles: countRange{1, 4},
		},
		packages: countRange{10, 40},
		appPackageSpec: &appPackageSpec{
			versions:   countRange{1, 3},
			profiles:   countRange{1, 3},
			apps:       countRange{1, 4},
			deps:       countRange{0, 2},
			artifacts:  countRange{2, 5},
			extensions: countRange{0, 5},
			endpoints:  countRange{1, 3},
			namespaces: countRange{1, 4},
		},
	}

	// Populate the catalog with some root entities structured according to the given specs.
	s.generateProject(tenant, pSpec)

	// Create and start the simulators
	uiSims := make([]*UISimulator, 70)
	for i := range uiSims {
		sim := NewUISimulator(fmt.Sprintf("ui%d", i+1), s.restClient)
		uiSims[i] = sim
		sim.Start()
	}

	// Wait for a bit
	time.Sleep(2 * time.Minute)

	// Stop all the simulators
	wg := sync.WaitGroup{}
	wg.Add(len(uiSims))
	for _, sim := range uiSims {
		sim.Stop(&wg)
	}
	wg.Wait()

	// Aggregate performance stats
	stats := make(map[string]*Latency, 8)
	for _, sim := range uiSims {
		for name, simStat := range sim.latencies {
			stat, ok := stats[name]
			if !ok {
				stats[name] = NewLatency(name).Combine(simStat)
			} else {
				stat.Combine(simStat)
			}
		}
	}

	// Print out the stats
	fmt.Println(printer.Sprintf("%-20s%14s%14s%14s%14s", "Metric", "Iterations", "Average(ms)", "Shortest(ms)", "Longest(ms)"))
	for _, stat := range stats {
		fmt.Printf("%s\n", stat)
	}

}
