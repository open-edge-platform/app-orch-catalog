// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net/http"
	"sync"
)

// EventHandler represents entity capable of relaying various Watch<X> events to a web-socket channel
type EventHandler struct {
	grpcEndpoint string
	grpcClient   catalogv3.CatalogServiceClient

	lock sync.RWMutex

	// map of session ID to session
	sessions map[string]*Session
}

const (
	bufferSize = 1024 * 1024

	publisherKind         = "Publisher"
	registryKind          = "Registry"
	artifactKind          = "Artifact"
	applicationKind       = "Application"
	deploymentPackageKind = "DeploymentPackage"

	subscribeOp    = "subscribe"
	subscribedOp   = "subscribed"
	unsubscribeOp  = "unsubscribe"
	unsubscribedOp = "unsubscribed"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  bufferSize,
	WriteBufferSize: bufferSize,
}

// NewEventHandler creates a new event handler relay.
func NewEventHandler(endpoint string, opts []grpc.DialOption) (*EventHandler, error) {
	log.Infow("Creating EventHandler", dazl.String("grpcEndpoint", endpoint))

	ctx := context.Background()
	conn, err := grpc.Dial(endpoint, opts...)

	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()
	client := catalogv3.NewCatalogServiceClient(conn)

	return &EventHandler{
		grpcEndpoint: endpoint,
		grpcClient:   client,
		sessions:     make(map[string]*Session, 8),
	}, nil
}

// Watch upgrades the incoming GET request into a web-socket connection on which it creates a session
// for accepting subscriptions and relaying of subsequent events.
func (h *EventHandler) Watch(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	go h.startSession(c, ws)
}

func (h *EventHandler) addSession(s *Session) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.sessions[s.ID()] = s
}

func (h *EventHandler) removeSession(s *Session) {
	h.lock.Lock()
	defer h.lock.Unlock()
	delete(h.sessions, s.ID())
	_ = s.Close()
}

func (h *EventHandler) startSession(c *gin.Context, ws *websocket.Conn) {
	session := NewSession(ws, h.grpcClient, c)
	h.addSession(session)
	session.Start(c)

	log.Infof("Started watch session for project %s", session.projectUUID)

	defer func() {
		if err := recover(); err != nil {
			log.Warnf("Watch session error: %v", err)
		}
		h.removeSession(session)
	}()

	for {
		select {
		case msg, ok := <-session.Listen():
			if !ok {
				return
			}
			switch msg.Op {
			case subscribeOp:
				session.addSubscription(c, msg.Kind, msg.Project)
			case unsubscribeOp:
				session.removeSubscription(msg.Kind)
			}

		case err := <-session.Error():
			log.Warnf("Session error: %v", err)
		case <-session.Done():
			break
		}
	}
}
