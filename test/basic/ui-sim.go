// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"context"
	"fmt"
	restapi "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"math/rand"
	"sync"
	"time"
)

type WorkloadSimulator interface {
	Start()
	Stop(*sync.WaitGroup)
	Run()
	SimulateWork()
}

type Simulator struct {
	WorkloadSimulator
	id        string
	client    *restapi.ClientWithResponses
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan bool
	latencies map[string]*Latency
	wg        *sync.WaitGroup
}

func addStat(stats map[string]*Latency, name string, duration time.Duration) {
	stat, ok := stats[name]
	if !ok {
		stat = NewLatency(name)
		stats[name] = stat
	}
	stat.Add(duration)
}

func (s *Simulator) measure(name string, f func() error) error {
	start := time.Now()
	if err := f(); err != nil {
		return err
	}
	end := time.Now()
	duration := end.Sub(start)
	addStat(s.latencies, name, duration)
	return nil
}

// Start starts the simulator asynchronously
func (s *Simulator) Start() {
	fmt.Printf("%s: starting...\n", s.id)
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 15*time.Minute)
	s.latencies = make(map[string]*Latency, 8)
	s.done = make(chan bool, 8)
	go s.Run()
}

// Stop stops the simulator
func (s *Simulator) Stop(wg *sync.WaitGroup) {
	fmt.Printf("%s: stopping...\n", s.id)
	s.wg = wg
	s.done <- true
}

// Run runs the simulator via repeated invocations of SimulateWork
func (s *Simulator) Run() {
	fmt.Printf("%s: started\n", s.id)
	for {
		select {
		case <-s.done:
			fmt.Printf("%s: stopped\n", s.id)
			s.wg.Done()
			return
		default:
			s.SimulateWork()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

type UISimulator struct {
	Simulator
	registries   []restapi.Registry
	artifacts    []restapi.Artifact
	applications []restapi.Application
	packages     []restapi.DeploymentPackage
}

func NewUISimulator(id string, client *restapi.ClientWithResponses) *UISimulator {
	sim := &UISimulator{
		Simulator{id: id, client: client}, nil, nil, nil, nil,
	}
	sim.WorkloadSimulator = sim
	return sim
}

// Define various UI activities
type Activity int

const (
	AwayFromKeyboard  Activity = 0
	PageRegistries             = 1
	PageArtifacts              = 2
	PageApplications           = 3
	PagePackages               = 4
	CreateApplication          = 5
	DeleteApplication          = 6
	CreatePackage              = 7
	DeletePackage              = 8
	LastActivity               = 9
)

// Probability distribution function for UI activities
var activityPDF = map[Activity]float32{
	AwayFromKeyboard:  0.10,
	PageRegistries:    0.05,
	PageArtifacts:     0.10,
	PageApplications:  0.30,
	PagePackages:      0.25,
	CreateApplication: 0.05,
	DeleteApplication: 0.05,
	CreatePackage:     0.05,
	DeletePackage:     0.05,
}

// Cummulative distribution function for UI activities; computed from PDF
var activityCDF = make([]float32, LastActivity)

func init() {
	activityCDF[0] = activityPDF[AwayFromKeyboard]
	for i := 1; i < LastActivity; i++ {
		activityCDF[i] = activityCDF[i-1] + activityPDF[Activity(i)]
	}
}

func (s *UISimulator) selectNextActivity() Activity {
	i := 0
	r := rand.Float32()
	for r > activityCDF[i] {
		i++
	}
	return Activity(i)
}

func (s *UISimulator) SimulateWork() {
	action := s.selectNextActivity()

	var err error
	switch action {
	case PageRegistries:
		err = s.paginateRegistries(0, 3, 25)
	case PageArtifacts:
		err = s.paginateArtifacts(0, 3, 25)
	case PageApplications:
		err = s.paginateApplications(0, 3, 25)
	case PagePackages:
		err = s.paginatePackages(0, 3, 25)
	case CreateApplication:
		if len(s.registries) > 0 {
			err = s.createApplication()
		}
	case DeleteApplication:
		if len(s.applications) > 0 {
			err = s.deleteApplication()
		}
	case CreatePackage:
		if len(s.applications) > 0 {
			err = s.createPackage()
		}
	case DeletePackage:
		if len(s.packages) > 0 {
			err = s.deletePackage()
		}
	case AwayFromKeyboard:
		s.awayFromKeyboard()
	default:
		fmt.Printf("not implemented yet\n")
	}
	if err != nil {
		fmt.Printf("%s: %v\n", s.id, err)
	}
}

const (
	minPause = 7
	maxPause = 15
)

func (s *UISimulator) shortPause() time.Duration {
	return time.Duration(500+rand.Intn(2000)) * time.Millisecond
}

func (s *UISimulator) normalPause() time.Duration {
	return time.Duration(minPause+rand.Intn(maxPause)) * time.Second
}

func (s *UISimulator) longPause() time.Duration {
	return time.Duration(2*minPause+rand.Intn(3*maxPause)) * time.Second
}

func (s *UISimulator) awayFromKeyboard() {
	time.Sleep(s.longPause())
}

func (s *UISimulator) createApplication() error {
	registry := s.registries[rand.Intn(len(s.registries))]
	name := fmt.Sprintf("app%04d", rand.Intn(1000))
	fmt.Printf("%s: Create application %s", s.id, name)
	return s.measure("Create Application", func() error {
		p1 := "p1"
		_, err := s.client.CatalogServiceCreateApplicationWithResponse(s.ctx, restapi.CatalogServiceCreateApplicationJSONRequestBody{
			Name: name, Version: "0.1",
			ChartName: fmt.Sprintf("%s-chart", name), ChartVersion: "0.1", HelmRegistryName: registry.Name,
			Profiles: &[]restapi.Profile{
				profileREST("p1", "Profile One", "First profile", "some odd yaml goes here"),
				profileREST("p2", "Profile Two", "Second profile", "some other yaml goes here"),
				profileREST("p3", "Profile Three", "Third profile", "some weird yaml here"),
			},
			DefaultProfileName: &p1,
		}, addHeaders)
		return err
	})
}

func (s *UISimulator) deleteApplication() error {
	application := s.applications[rand.Intn(len(s.applications))]
	fmt.Printf("%s: Delete application %s", s.id, application.Name)
	return s.measure("Delete Application", func() error {
		_, err := s.client.CatalogServiceDeleteApplicationWithResponse(s.ctx, application.Name, application.Version, addHeaders)
		return err
	})
}

func (s *UISimulator) createPackage() error {
	application := s.applications[rand.Intn(len(s.applications))]
	name := fmt.Sprintf("pkg%04d", rand.Intn(1000))
	fmt.Printf("%s: Create package %s", s.id, name)
	return s.measure("Create Package", func() error {
		p1 := "p1"
		_, err := s.client.CatalogServiceCreateDeploymentPackageWithResponse(s.ctx, restapi.CatalogServiceCreateDeploymentPackageJSONRequestBody{
			Name: name, Version: "0.2",
			ApplicationReferences: []restapi.ApplicationReference{{Name: application.Name, Version: application.Version}},
			Profiles: &[]restapi.DeploymentProfile{
				packageRESTProfile("p1", "Profile One", "First profile", map[string]string{application.Name: "p1"}),
			},
			DefaultProfileName: &p1,
		}, addHeaders)
		return err
	})
}

func (s *UISimulator) deletePackage() error {
	pkg := s.packages[rand.Intn(len(s.packages))]
	fmt.Printf("%s: Delete package %s", s.id, pkg.Name)
	return s.measure("Delete Package", func() error {
		_, err := s.client.CatalogServiceDeleteDeploymentPackageWithResponse(s.ctx, pkg.Name, pkg.Version, addHeaders)
		return err
	})
}

func (s *UISimulator) paginate(name string, initialOffset int32, pages int32, pageSize int32, f func(int32, int32) error) error {
	fmt.Printf("%s: %s from %d, %d times by %d items\n", s.id, name, initialOffset, pages, pageSize)
	if err := s.measure(name, func() error { return f(initialOffset, pageSize) }); err != nil {
		return err
	}
	for i := int32(1); i <= pages; i++ {
		s.shortPause()
		offset := initialOffset + i*pageSize
		if err := s.measure(name, func() error { return f(offset, pageSize) }); err != nil {
			return err
		}
	}
	return nil
}

func (s *UISimulator) paginateRegistries(initialOffset int32, pages int32, pageSize int32) error {
	return s.paginate("Page Registries", initialOffset, pages, pageSize,
		func(offset int32, size int32) error {
			resp, err := s.client.CatalogServiceListRegistriesWithResponse(s.ctx,
				&restapi.CatalogServiceListRegistriesParams{Offset: &offset, PageSize: &size}, addHeaders)
			if err == nil && resp.HTTPResponse.StatusCode == 200 {
				s.registries = resp.JSON200.Registries
			}
			return err
		})
}

func (s *UISimulator) paginateArtifacts(initialOffset int32, pages int32, pageSize int32) error {
	return s.paginate("Page Artifacts", initialOffset, pages, pageSize,
		func(offset int32, size int32) error {
			resp, err := s.client.CatalogServiceListArtifactsWithResponse(s.ctx,
				&restapi.CatalogServiceListArtifactsParams{Offset: &offset, PageSize: &size}, addHeaders)
			if err == nil && resp.HTTPResponse.StatusCode == 200 {
				s.artifacts = resp.JSON200.Artifacts
			}
			return err
		})
}

func (s *UISimulator) paginateApplications(initialOffset int32, pages int32, pageSize int32) error {
	return s.paginate("Page Applications", initialOffset, pages, pageSize,
		func(offset int32, size int32) error {
			resp, err := s.client.CatalogServiceListApplicationsWithResponse(s.ctx,
				&restapi.CatalogServiceListApplicationsParams{Offset: &offset, PageSize: &size}, addHeaders)
			if err == nil && resp.HTTPResponse.StatusCode == 200 {
				s.applications = resp.JSON200.Applications
			}
			return err
		})
}

func (s *UISimulator) paginatePackages(initialOffset int32, pages int32, pageSize int32) error {
	return s.paginate("Page Packages", initialOffset, pages, pageSize,
		func(offset int32, size int32) error {
			resp, err := s.client.CatalogServiceListDeploymentPackagesWithResponse(s.ctx,
				&restapi.CatalogServiceListDeploymentPackagesParams{Offset: &offset, PageSize: &size}, addHeaders)
			if err == nil && resp.HTTPResponse.StatusCode == 200 {
				s.packages = resp.JSON200.DeploymentPackages
			}
			return err
		})
}
