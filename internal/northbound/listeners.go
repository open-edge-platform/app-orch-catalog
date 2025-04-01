// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"sync"
)

// EventListeners tracks current listeners for different type of entity events.
type EventListeners struct {
	lock sync.RWMutex

	registryListeners          map[chan *catalogv3.WatchRegistriesResponse]*catalogv3.WatchRegistriesRequest
	artifactListeners          map[chan *catalogv3.WatchArtifactsResponse]*catalogv3.WatchArtifactsRequest
	applicationListeners       map[chan *catalogv3.WatchApplicationsResponse]*catalogv3.WatchApplicationsRequest
	deploymentPackageListeners map[chan *catalogv3.WatchDeploymentPackagesResponse]*catalogv3.WatchDeploymentPackagesRequest
}

func NewEventListeners() *EventListeners {
	return &EventListeners{
		registryListeners:          make(map[chan *catalogv3.WatchRegistriesResponse]*catalogv3.WatchRegistriesRequest),
		artifactListeners:          make(map[chan *catalogv3.WatchArtifactsResponse]*catalogv3.WatchArtifactsRequest),
		applicationListeners:       make(map[chan *catalogv3.WatchApplicationsResponse]*catalogv3.WatchApplicationsRequest),
		deploymentPackageListeners: make(map[chan *catalogv3.WatchDeploymentPackagesResponse]*catalogv3.WatchDeploymentPackagesRequest),
	}
}

func (el *EventListeners) addRegistryListener(ch chan *catalogv3.WatchRegistriesResponse, req *catalogv3.WatchRegistriesRequest) {
	el.lock.Lock()
	defer el.lock.Unlock()
	el.registryListeners[ch] = req
}

func (el *EventListeners) deleteRegistryListener(ch chan *catalogv3.WatchRegistriesResponse) {
	el.lock.Lock()
	defer el.lock.Unlock()
	delete(el.registryListeners, ch)
}

func (el *EventListeners) sendRegistryEvents(event *catalogv3.WatchRegistriesResponse) {
	el.lock.RLock()
	defer el.lock.RUnlock()
	for ch, req := range el.registryListeners {
		if req.ProjectId == "" || req.ProjectId == event.Event.ProjectId {
			ch <- event
		}
	}
}

func (el *EventListeners) addArtifactListener(ch chan *catalogv3.WatchArtifactsResponse, req *catalogv3.WatchArtifactsRequest) {
	el.lock.Lock()
	defer el.lock.Unlock()
	el.artifactListeners[ch] = req
}

func (el *EventListeners) deleteArtifactListener(ch chan *catalogv3.WatchArtifactsResponse) {
	el.lock.Lock()
	defer el.lock.Unlock()
	delete(el.artifactListeners, ch)
}

func (el *EventListeners) sendArtifactEvents(event *catalogv3.WatchArtifactsResponse) {
	el.lock.RLock()
	defer el.lock.RUnlock()
	for ch, req := range el.artifactListeners {
		if req.ProjectId == "" || req.ProjectId == event.Event.ProjectId {
			ch <- event
		}
	}
}

func (el *EventListeners) addApplicationListener(ch chan *catalogv3.WatchApplicationsResponse, req *catalogv3.WatchApplicationsRequest) {
	el.lock.Lock()
	defer el.lock.Unlock()
	el.applicationListeners[ch] = req
}

func (el *EventListeners) deleteApplicationListener(ch chan *catalogv3.WatchApplicationsResponse) {
	el.lock.Lock()
	defer el.lock.Unlock()
	delete(el.applicationListeners, ch)
}

func (el *EventListeners) sendApplicationEvents(event *catalogv3.WatchApplicationsResponse) {
	el.lock.RLock()
	defer el.lock.RUnlock()
	for ch, req := range el.applicationListeners {
		if req.ProjectId == "" || req.ProjectId == event.Event.ProjectId {
			if kindMatches(req.Kinds, event.Application.Kind) {
				ch <- event
			}
		}
	}
}

func (el *EventListeners) addDeploymentPackageListener(ch chan *catalogv3.WatchDeploymentPackagesResponse, req *catalogv3.WatchDeploymentPackagesRequest) {
	el.lock.Lock()
	defer el.lock.Unlock()
	el.deploymentPackageListeners[ch] = req
}

func (el *EventListeners) deleteDeploymentPackageListener(ch chan *catalogv3.WatchDeploymentPackagesResponse) {
	el.lock.Lock()
	defer el.lock.Unlock()
	delete(el.deploymentPackageListeners, ch)
}

func (el *EventListeners) sendDeploymentPackageEvents(event *catalogv3.WatchDeploymentPackagesResponse) {
	el.lock.RLock()
	defer el.lock.RUnlock()
	for ch, req := range el.deploymentPackageListeners {
		if req.ProjectId == "" || req.ProjectId == event.Event.ProjectId {
			if kindMatches(req.Kinds, event.DeploymentPackage.Kind) {
				ch <- event
			}
		}
	}
}

func kindMatches(kinds []catalogv3.Kind, kind catalogv3.Kind) bool {
	if len(kinds) == 0 {
		return true
	}
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}
