// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"sync"
	"time"
)

const (
	maxMessageSize = 8 * 1024
	maxWriteWait   = 5 * time.Second
)

var (
	pingPeriod  = 5 * time.Second
	maxPongWait = 15 * time.Second
)

// Message represents subscription control message
type Message struct {
	Op      string `json:"op"`
	Kind    string `json:"kind"`
	Project string `json:"project"`
	Payload []byte `json:"payload"`
}

// Session represents a watch-event session.
type Session struct {
	id         string
	ws         *websocket.Conn
	grpcClient catalogv3.CatalogServiceClient

	messages chan *Message
	err      chan error
	done     chan interface{}
	m        sync.Mutex
	once     sync.Once

	cancellations map[string]context.CancelFunc
	authHeader    string
	uaHeader      string
	projectUUID   string
}

// NewSession creates a new session backed by the specified socket connection.
func NewSession(ws *websocket.Conn, grpcClient catalogv3.CatalogServiceClient, c *gin.Context) *Session {
	return &Session{
		id:            uuid.NewString(),
		ws:            ws,
		grpcClient:    grpcClient,
		messages:      make(chan *Message),
		err:           make(chan error),
		done:          make(chan interface{}),
		cancellations: make(map[string]context.CancelFunc),
		authHeader:    c.Request.Header.Get("Authorization"),
		uaHeader:      c.Request.Header.Get("User-Agent"),
		projectUUID:   c.Request.Header.Get(ActiveProjectID),
	}
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.id
}

// Start kicks off the session using the supplied context
func (s *Session) Start(ctx context.Context) {
	s.ws.SetReadLimit(maxMessageSize)
	_ = s.ws.SetReadDeadline(time.Now().Add(maxPongWait))
	s.ws.SetPongHandler(func(string) error {
		_ = s.ws.SetReadDeadline(time.Now().Add(maxPongWait))
		return nil
	})
	s.once.Do(func() { go s.start(ctx) })
}

func (s *Session) start(ctx context.Context) {
	var wg sync.WaitGroup

	cancelCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		s.send(websocket.CloseMessage)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.receive(cancelCtx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.ping(cancelCtx)
	}()

	wg.Wait()
	s.done <- struct{}{}
}

func (s *Session) receive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg := &Message{}
			if err := s.ws.ReadJSON(msg); err != nil {
				s.handleError(err)
				return
			}
			s.messages <- msg
		}
	}
}

func (s *Session) ping(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.send(websocket.PingMessage)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Session) send(msgType int) {
	s.m.Lock()
	defer s.m.Unlock()

	_ = s.ws.SetWriteDeadline(time.Now().Add(maxWriteWait))
	if err := s.ws.WriteMessage(msgType, nil); err != nil {
		s.handleError(err)
	}
}

func (s *Session) handleError(err error) {
	if _, ok := err.(*websocket.CloseError); !ok {
		return
	}
	if errors.Is(err, websocket.ErrCloseSent) {
		return
	}
	s.err <- err
}

// Close closes the session
func (s *Session) Close() error {
	close(s.messages)
	return s.ws.Close()
}

// Send sends the specified message to the client
func (s *Session) Send(msg Message) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err := s.ws.SetWriteDeadline(time.Now().Add(maxWriteWait)); err != nil {
		return err
	}
	return s.ws.WriteJSON(msg)
}

// Listen returns the channel for reading messages from the client
func (s *Session) Listen() <-chan *Message {
	return s.messages
}

// Done returns the channel for reading the signal that the session is done
func (s *Session) Done() <-chan interface{} {
	return s.done
}

// Error returns the channel for reading errors encountered during interaction with the backing web socket
func (s *Session) Error() <-chan error {
	return s.err
}

func (s *Session) addSubscription(c *gin.Context, kind string, projectUUID string) {
	log.Infof("Add %s subscription %s for project %s", kind, s.ID(), projectUUID)
	// If this is the first subscriber added for this kind, let's start watching it against the catalog service
	if _, ok := s.cancellations[kind]; !ok {
		ctx, cancellation := context.WithCancel(c)
		s.cancellations[kind] = cancellation
		s.startWatching(ctx, kind, projectUUID)
	}
	_ = s.Send(Message{Op: subscribedOp, Kind: kind, Project: projectUUID})
}

func (s *Session) removeSubscription(kind string) {
	log.Infof("Remove %s subscription %s", kind, s.ID())
	if cf, ok := s.cancellations[kind]; ok {
		delete(s.cancellations, kind)
		cf()
	}
	_ = s.Send(Message{Op: unsubscribedOp, Kind: kind})
}

func (s *Session) startWatching(ctx context.Context, kind string, projectUUID string) {
	mdCtx := metadata.NewOutgoingContext(ctx, s.metadata())
	switch kind {
	case registryKind:
		go s.watchRegistries(mdCtx, projectUUID)
	case artifactKind:
		go s.watchArtifacts(mdCtx, projectUUID)
	case applicationKind:
		go s.watchApplications(mdCtx, projectUUID)
	case deploymentPackageKind:
		go s.watchDeploymentPackages(mdCtx, projectUUID)
	}
}

func (s *Session) metadata() metadata.MD {
	return metadata.Pairs("authorization", s.authHeader, "user-agent", s.uaHeader, "activeprojectid", s.projectUUID)
}

func (s *Session) watchRegistries(ctx context.Context, projectUUID string) {
	stream, err := s.grpcClient.WatchRegistries(ctx, &catalogv3.WatchRegistriesRequest{NoReplay: true, ProjectId: projectUUID})
	if err != nil {
		log.Warnf("Unable to subscribe for registry events: %v", err)
		return
	}

	log.Infof("Started watching registry events")
	for {
		event, err := stream.Recv()
		if !s.processError(err) {
			break
		}
		buf, _ := protojson.Marshal(event.Registry)
		s.processEvent(Message{Op: event.Event.Type, Kind: registryKind, Project: event.Event.ProjectId, Payload: buf})
	}
	log.Infof("Stopped watching registry events")
}

func (s *Session) watchArtifacts(ctx context.Context, projectUUID string) {
	stream, err := s.grpcClient.WatchArtifacts(ctx, &catalogv3.WatchArtifactsRequest{NoReplay: true, ProjectId: projectUUID})
	if err != nil {
		log.Warnf("Unable to subscribe for artifact events: %v", err)
		return
	}

	log.Infof("Started watching artifact events")
	for {
		event, err := stream.Recv()
		if !s.processError(err) {
			break
		}
		buf, _ := protojson.Marshal(event.Artifact)
		s.processEvent(Message{Op: event.Event.Type, Kind: artifactKind, Project: event.Event.ProjectId, Payload: buf})
	}
	log.Infof("Stopped watching artifact events")
}

func (s *Session) watchApplications(ctx context.Context, projectUUID string) {
	stream, err := s.grpcClient.WatchApplications(ctx, &catalogv3.WatchApplicationsRequest{NoReplay: true, ProjectId: projectUUID})
	if err != nil {
		log.Warnf("Unable to subscribe for application events: %v", err)
		return
	}

	log.Infof("Started watching application events")
	for {
		event, err := stream.Recv()
		if !s.processError(err) {
			break
		}
		buf, _ := protojson.Marshal(event.Application)
		s.processEvent(Message{Op: event.Event.Type, Kind: applicationKind, Project: event.Event.ProjectId, Payload: buf})
	}
	log.Infof("Stopped watching application events")
}

func (s *Session) watchDeploymentPackages(ctx context.Context, projectUUID string) {
	stream, err := s.grpcClient.WatchDeploymentPackages(ctx, &catalogv3.WatchDeploymentPackagesRequest{NoReplay: true, ProjectId: projectUUID})
	if err != nil {
		log.Warnf("Unable to subscribe for deployment package events: %v", err)
		return
	}

	log.Infof("Started watching deployment package events")
	for {
		event, err := stream.Recv()
		if !s.processError(err) {
			break
		}
		buf, _ := protojson.Marshal(event.DeploymentPackage)
		s.processEvent(Message{Op: event.Event.Type, Kind: deploymentPackageKind, Project: event.Event.ProjectId, Payload: buf})
	}
	log.Infof("Stopped watching deployment package events")
}

func (s *Session) processError(err error) bool {
	if err != nil && err != io.EOF {
		log.Warnf("Unable to read message: %v", err)
	}
	return err == nil
}

func (s *Session) processEvent(msg Message) {
	log.Debugf("Sending %v to %s", msg, s.ID())
	if err := s.Send(msg); err != nil {
		log.Warnf("Unable to send message: %v", err)
	}
}
