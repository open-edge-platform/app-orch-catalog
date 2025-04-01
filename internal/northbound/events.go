// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
)

type EventType string

const (
	CreatedEvent  EventType = "created"
	UpdatedEvent  EventType = "updated"
	DeletedEvent  EventType = "deleted"
	ReplayedEvent EventType = "replayed"
)

// Creates an API event type
func event(eventType EventType, projectUUID string) *catalogv3.Event {
	return &catalogv3.Event{Type: string(eventType), ProjectId: projectUUID}
}

// RegistryEvents is a queue of registry events.
type RegistryEvents struct {
	queue []*catalogv3.WatchRegistriesResponse
}

func (re *RegistryEvents) append(eventType EventType, projectUUID string, r *catalogv3.Registry) {
	re.queue = append(re.queue, &catalogv3.WatchRegistriesResponse{Event: event(eventType, projectUUID), Registry: r})
}

func (re *RegistryEvents) sendToAll(listeners *EventListeners) {
	for _, e := range re.queue {
		listeners.sendRegistryEvents(e)
	}
}

// ArtifactEvents is a queue of artifact events.
type ArtifactEvents struct {
	queue []*catalogv3.WatchArtifactsResponse
}

func (are *ArtifactEvents) append(eventType EventType, projectUUID string, ar *catalogv3.Artifact) {
	are.queue = append(are.queue, &catalogv3.WatchArtifactsResponse{Event: event(eventType, projectUUID), Artifact: ar})
}

func (are *ArtifactEvents) sendToAll(listeners *EventListeners) {
	for _, e := range are.queue {
		listeners.sendArtifactEvents(e)
	}
}

// ApplicationEvents is a queue of application events.
type ApplicationEvents struct {
	queue []*catalogv3.WatchApplicationsResponse
}

func (ape *ApplicationEvents) append(eventType EventType, projectUUID string, app *catalogv3.Application) {
	ape.queue = append(ape.queue, &catalogv3.WatchApplicationsResponse{Event: event(eventType, projectUUID), Application: app})
}

func (ape *ApplicationEvents) sendToAll(listeners *EventListeners) {
	for _, e := range ape.queue {
		listeners.sendApplicationEvents(e)
	}
}

// DeploymentPackageEvents is a queue of deployment package events.
type DeploymentPackageEvents struct {
	queue []*catalogv3.WatchDeploymentPackagesResponse
}

func (dpe *DeploymentPackageEvents) append(eventType EventType, projectUUID string, p *catalogv3.DeploymentPackage) {
	dpe.queue = append(dpe.queue, &catalogv3.WatchDeploymentPackagesResponse{Event: event(eventType, projectUUID), DeploymentPackage: p})
}

func (dpe *DeploymentPackageEvents) sendToAll(listeners *EventListeners) {
	for _, e := range dpe.queue {
		listeners.sendDeploymentPackageEvents(e)
	}
}
