package restproxy

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	ent "github.com/open-edge-platform/app-orch-catalog/internal/ent/generated"
	"github.com/open-edge-platform/app-orch-catalog/internal/ent/generated/enttest"
	"github.com/open-edge-platform/app-orch-catalog/internal/northbound"
	catalogv3 "github.com/open-edge-platform/app-orch-catalog/pkg/api/catalog/v3"
	restclient "github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	ginutils "github.com/open-edge-platform/orch-library/go/pkg/middleware/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createRestServerConnection(t *testing.T, dbClient *ent.Client) *restclient.ClientWithResponses {
	s, err := newTestService(dbClient)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	server := grpc.NewServer()
	s.Register(server)

	mux := runtime.NewServeMux(
		// convert header in response(going from gateway) from metadata received.
		runtime.WithMetadata(func(_ context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			uaHeader := request.Header.Get("User-Agent")
			projectIDHeader := request.Header.Get("ActiveProjectID")
			// send all the headers received from the client
			md := metadata.Pairs("authorization", authHeader, "user-agent", uaHeader, "ActiveProjectID", projectIDHeader)
			return md
		}),
		runtime.WithRoutingErrorHandler(ginutils.HandleRoutingError),
	)

	dialer := func(context.Context, string) (net.Conn, error) {
		listener := bufconn.Listen(1024 * 1024)

		go func() {
			if err := server.Serve(listener); err != nil {
				t.Error(err)
				t.Fail()
			}
		}()
		return listener.Dial()
	}

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	err = catalogv3.RegisterCatalogServiceHandler(ctx, mux, conn)
	assert.NoError(t, err)
	restProxy := httptest.NewServer(mux)

	catalogClient, err := restclient.NewClientWithResponses(restProxy.URL)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return catalogClient
}

func newTestService(dbClient *ent.Client) (northbound.Service, error) {
	s := northbound.Service{DatabaseClient: dbClient}
	return s, nil
}

type MockRestProxy interface {
	RestClient() *restclient.ClientWithResponses
	Close() error
}

type restProxy struct {
	restClient *restclient.ClientWithResponses
	dbClient   *ent.Client
}

func NewMockRestProxy(t *testing.T) MockRestProxy {
	// Set up database
	dbClient := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")

	// set up Rest client
	client := createRestServerConnection(t, dbClient)
	assert.NotNil(t, client)

	return restProxy{
		restClient: client,
		dbClient:   dbClient,
	}
}

func (p restProxy) RestClient() *restclient.ClientWithResponses { return p.restClient }
func (p restProxy) Close() error                                { return p.dbClient.Close() }
