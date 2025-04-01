// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/jsonrenderer"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/dazl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"regexp"
)

type FileHandler struct {
	grpcEndpoint string
	grpcClient   catalogv3.CatalogServiceClient
}

func (h *FileHandler) Upload(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	uaHeader := c.Request.Header.Get("User-Agent")
	projectHeader := c.Request.Header.Get(ActiveProjectID)

	mdCtx := metadata.NewOutgoingContext(context.TODO(),
		metadata.Pairs("authorization", authHeader, "user-agent", uaHeader, "activeprojectid", projectHeader))

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	returnStatus := http.StatusOK

	// Regular expression matcher for extracting full filename from Content-Disposition header
	ffnregex, _ := regexp.Compile("filename=\\\"(.*)\\\"")

	files := form.File["files"]
	filesCount := len(files)
	// the catalog returns a sessionID after the first call
	// keep track of that for subsequent calls
	sessionID := ""
	responses := &catalogv3.UploadMultipleCatalogEntitiesResponse{
		Responses: []*catalogv3.UploadCatalogEntitiesResponse{},
	}
	for index, file := range files {
		path := file.Filename // use the base filename as fallback path
		// Otherwise, extract the full filename path from the content disposition header
		match := ffnregex.FindStringSubmatch(file.Header.Get("Content-Disposition"))
		if len(match) > 1 {
			path = match[1]
		}

		log.Infof("processing-file: %s", path)
		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		content, err := io.ReadAll(openedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		res, err := h.grpcClient.UploadCatalogEntities(mdCtx, &catalogv3.UploadCatalogEntitiesRequest{
			SessionId:  sessionID,
			LastUpload: (index + 1) == filesCount,
			Upload: &catalogv3.Upload{
				FileName: path,
				Artifact: content,
			},
		})

		if err != nil {
			responses.Responses = append(responses.Responses,
				&catalogv3.UploadCatalogEntitiesResponse{
					SessionId:     sessionID,
					ErrorMessages: []string{err.Error()},
				})
			returnStatus = http.StatusBadRequest
			log.Errorw("error processing file", dazl.String("file", file.Filename), dazl.Error(err))
		} else {
			// if there is a response, we update the sessionID and the response map
			sessionID = res.SessionId
			responses.Responses = append(responses.Responses, res)
		}
	}

	// use protojson to marshal the data into JSON as that's the same that gRPC-gateway uses
	// otherwise the casing in the JSON are messed up (we'll get snake_case instead of the expected camelCase)
	renderer := jsonrenderer.JSONFromProto{Data: responses}
	c.Render(returnStatus, renderer)
}

func NewFileHandler(endpoint string, opts []grpc.DialOption) (*FileHandler, error) {
	log.Infow("Creating FileHandler", dazl.String("grpcEndpoint", endpoint))

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

	return &FileHandler{
		grpcEndpoint: endpoint,
		grpcClient:   client,
	}, nil
}
