// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package restproxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/open-edge-platform/app-orch-catalog/internal/shared/version"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	"github.com/open-edge-platform/orch-library/go/dazl"
	ginlogger "github.com/open-edge-platform/orch-library/go/pkg/logging/gin"
	ginutils "github.com/open-edge-platform/orch-library/go/pkg/middleware/gin"
	"github.com/open-edge-platform/orch-library/go/pkg/middleware/projectcontext"
	openapiutils "github.com/open-edge-platform/orch-library/go/pkg/openapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var log = dazl.GetPackageLogger()

// Config defines configurable parameters for setting up the REST proxy.
type Config struct {
	Port               int
	OIDCExternal       string
	GRPCEndpoint       string
	BasePath           string
	SpecFilePath       string
	AllowedCorsOrigins string
	// ProjectServiceURL is the URL of the project lookup service used to resolve
	// a project name (from the URL path) to its UUID for OPA authorization.
	// Currently points to the Nexus API GW as an interim solution;
	// will be replaced by the Tenant Manager REST API.
	ProjectServiceURL string
}

// RESTProxy represents the REST proxy state
type RESTProxy struct {
	cfg    *Config
	engine *gin.Engine
}

var allowedHeaders = map[string]struct{}{
	"x-request-id": {},
}

const ActiveProjectID = "ActiveProjectID"

func isHeaderAllowed(s string) (string, bool) {
	// check if allowedHeaders contain the header
	if _, isAllowed := allowedHeaders[s]; isAllowed {
		// send uppercase header
		return strings.ToUpper(s), true
	}
	// if not in the allowed header, don't send the header
	return s, false
}

func NewRESTProxy(cfg *Config) (*RESTProxy, error) {
	log.Infof("Creating REST proxy on port %d", cfg.Port)
	version.LogVersion("  ")

	gin.DefaultWriter = ginlogger.NewWriter(log)

	// creating mux for gRPC gateway. This will multiplex or route request different gRPC service
	mux := runtime.NewServeMux(
		// convert header in response(going from gateway) from metadata received.
		runtime.WithOutgoingHeaderMatcher(isHeaderAllowed),
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			uaHeader := request.Header.Get("User-Agent")
			projectIDHeader := request.Header.Get(ActiveProjectID)
			// Resolve project UUID from the URL path project name and set ActiveProjectID
			// for OPA authorization. For /v3/projects/{name}/catalog/... paths the name
			// is resolved to a UUID via the project service (Nexus API GW today, Tenant
			// Manager REST API in future. Falls back to JWT extraction
			// for legacy /catalog.orchestrator.apis/v3/... paths.
			projectUUID, err := projectcontext.ResolveAndValidateProjectID(
				ctx,
				request.URL.Path,
				authHeader,
				projectIDHeader,
				projectcontext.ProjectResolverConfig{
					ProjectServiceURL:     cfg.ProjectServiceURL,
					ErrorOnMissingProject: false,
				},
			)
			if err != nil {
				log.Warnf("Failed to resolve project ID: %v", err)
			} else if projectUUID != "" {
				projectIDHeader = projectUUID
			}
			return metadata.Pairs("authorization", authHeader, "user-agent", uaHeader, "activeprojectid", projectIDHeader)
		}),
		runtime.WithRoutingErrorHandler(ginutils.HandleRoutingError),
	)

	// setting up a dial-up for gRPC service by specifying endpoint/target url
	err := catalogv3.RegisterCatalogServiceHandlerFromEndpoint(context.Background(), mux, cfg.GRPCEndpoint,
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		return nil, err
	}

	engine := gin.New()

	rp := &RESTProxy{
		cfg:    cfg,
		engine: engine,
	}

	// check if another method is allowed for the current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed' and HTTP status code 405
	// otherwise will return 'Page Not Found' and HTTP status code 404.
	engine.HandleMethodNotAllowed = true
	engine.Handle("GET", "/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})
	engine.Handle("GET", "/openidc-issuer", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"externalURL": cfg.OIDCExternal})
	})

	// Instantiate the EventHandler module
	eventHandler, err := NewEventHandler(cfg.GRPCEndpoint, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		return nil, err
	}
	engine.Handle("GET", fmt.Sprintf("%scatalog.orchestrator.apis/events", cfg.BasePath), func(c *gin.Context) {
		eventHandler.Watch(c)
	})

	// Instantiate the ChartsHandler module
	chartsHandler, err := NewChartsHandler(cfg.GRPCEndpoint, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		return nil, err
	}
	engine.Handle("GET", fmt.Sprintf("%scatalog.orchestrator.apis/charts", cfg.BasePath), func(c *gin.Context) {
		chartsHandler.FetchChartsList(c)
	})

	// Instantiate the FileHandler module
	fileHandler, err := NewFileHandler(cfg.GRPCEndpoint, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		return nil, err
	}

	// Add download to the mux so it is part of the /v3/ API
	err = mux.HandlePath("GET", "/catalog.orchestrator.apis/v3/deployment_packages/{deployment_package_name}/versions/{version}/download", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		fileHandler.Download(mux, w, r, pathParams)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register download handler: %w", err)
	}

	// FIXME understand how to make this endpoint part of the openapi specs
	// Example:
	// curl -X POST http://localhost:8080/upload \
	//  -F "files=@path-to-file/file1.zip" \
	//  -F "files=@path-to-file/file2.zip" \
	//  -H "Content-Type: multipart/form-data"
	engine.Handle("POST", fmt.Sprintf("%scatalog.orchestrator.apis/upload", cfg.BasePath), func(c *gin.Context) {
		fileHandler.Upload(c)
	})
	// New multi-tenant upload path: /v3/projects/{projectName}/catalog/upload
	// Registered on the grpc-gateway mux (not gin) because the gin wildcard route
	// v3/projects/:projectName/catalog/* would conflict with a static gin route.
	// The WithMetadata handler above already resolves ActiveProjectID from the URL
	// path for all /v3/projects/{name}/catalog/... requests, so no extra UUID
	// resolution is needed here.
	err = mux.HandlePath("POST", "/v3/projects/{projectName}/catalog/upload", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		fileHandler.UploadHTTP(w, r)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register multi-tenant upload handler: %w", err)
	}
	spec, err := openapiutils.LoadOpenAPISpec(cfg.SpecFilePath)
	if err != nil {
		return nil, err
	}
	// Restrict GET verb for different endpoints of the API
	allPaths := openapiutils.ExtractAllPaths(spec)

	var allowedMethods []string
	for verb := range allPaths {
		allowedMethods = append(allowedMethods, verb)
	}

	corsOrigins := strings.Split(cfg.AllowedCorsOrigins, ",")
	if len(corsOrigins) > 1 {
		config := cors.DefaultConfig()
		config.AllowOrigins = corsOrigins
		engine.Use(cors.New(config))
	}

	engine.Use(ginlogger.NewGinLogger(log))
	engine.Use(secure.New(secure.Config{ContentTypeNosniff: true}))
	engine.Use(ginutils.UnicodePrintableCharsChecker())
	engine.Use(ginutils.PathParamUnicodeCheckerMiddleware())
	// Add query parameter validation middleware for registry endpoints
	engine.Use(QueryParameterValidationMiddleware())
	engine.StaticFile(fmt.Sprintf("%scatalog.orchestrator.apis/api/v3", cfg.BasePath), cfg.SpecFilePath)
	engine.Group(fmt.Sprintf("%scatalog.orchestrator.apis/v3/*{grpc_gateway}", cfg.BasePath)).Match(allowedMethods, "", gin.WrapH(mux))
	// Route /v3/projects/{name}/catalog/... paths to the grpc-gateway mux.
	// The project name is resolved to a UUID by the WithMetadata handler above.
	// Use a specific pattern to avoid routing conflicts with future non-catalog /v3/projects/... endpoints.
	engine.Group(fmt.Sprintf("%sv3/projects/:projectName/catalog/*{grpc_gateway}", cfg.BasePath)).Match(allowedMethods, "", gin.WrapH(mux))

	engine.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Ok")
	})

	return rp, nil
}

func (p *RESTProxy) Run() error {
	log.Infof("Starting rest-proxy on port %d", p.cfg.Port)
	return p.engine.Run(fmt.Sprintf(":%d", p.cfg.Port))
}
