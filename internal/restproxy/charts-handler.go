// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"strings"
)

// ChartsHandler provides reverse proxy for access to a list of Help charts provided by a named repository.
type ChartsHandler struct {
	grpcEndpoint string
	grpcClient   catalogv3.CatalogServiceClient
}

func readAdminSecret() (string, string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", "", err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}

	value, err := clientSet.CoreV1().Secrets("harbor").Get(context.Background(), "harbor-admin-credential", metav1.GetOptions{})
	if err != nil {
		log.Errorf("Can't read secret %v", err)
		return "", "", nil
	}

	credsStr := string(value.Data["credential"])

	creds := strings.Split(credsStr, ":")
	if len(creds) != 2 {
		return "", "", fmt.Errorf("unable to parse harbor admin credentials")
	}

	return creds[0], creds[1], nil
}

// NewChartsHandler creates a new charts handler.
func NewChartsHandler(endpoint string, opts []grpc.DialOption) (*ChartsHandler, error) {
	log.Infow("Creating ChartsHandler", dazl.String("grpcEndpoint", endpoint))

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

	return &ChartsHandler{
		grpcEndpoint: endpoint,
		grpcClient:   client,
	}, nil
}

// FetchChartsList relays the Help charts list request to the back-end registry with appropriate authentication.
func (h *ChartsHandler) FetchChartsList(c *gin.Context) {
	// Grab the publisher and registry parameters
	query := c.Request.URL.Query()
	registryName := query.Get("registry")
	chartName := query.Get("chart")

	mdCtx := metadata.NewOutgoingContext(c, metadata.Pairs(
		"authorization", c.Request.Header.Get("Authorization"),
		"user-agent", c.Request.Header.Get("User-Agent"),
		ActiveProjectID, c.Request.Header.Get(ActiveProjectID),
	))

	log.Infof("Getting charts for registry %s (%s)[%s]...", registryName, c.Request.Header.Get(ActiveProjectID), chartName)

	resp, err := h.grpcClient.GetRegistry(mdCtx, &catalogv3.GetRegistryRequest{RegistryName: registryName, ShowSensitiveInfo: true})
	if err != nil {
		log.Errorf("Unable to get registry: %+v", err)
		err = c.AbortWithError(http.StatusInternalServerError, err)
		if (err != nil) {
			// Can't do anything about it, but log it
			log.Errorf("Unable to abort with status: %+v", err)
		}
		return
	}

	registry := resp.Registry
	if registry.InventoryUrl == "" {
		log.Debugf("Registry %s does not support inventory retrieval", registry.Name)
		c.AbortWithStatus(http.StatusNoContent)
	}

	if strings.Contains(registry.InventoryUrl, "/api/v2.0/projects/") {
		fetchOCIChartsList(c, registry, chartName)
	} else {
		log.Warnf("Registrry %s Not supported non-OCI registry inventory url %s: %v", registry.Name, registry.InventoryUrl)
		c.AbortWithStatusJSON(http.StatusNotImplemented, gin.H{"message": "Not supported non-OCI registry inventory url"})
	}
}
