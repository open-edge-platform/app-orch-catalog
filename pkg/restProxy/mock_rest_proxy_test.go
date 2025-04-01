package restproxy

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/open-edge-platform/app-orch-catalog/pkg/restClient"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMockRestProxy(t *testing.T) {
	p := NewMockRestProxy(t)
	assert.NotNil(t, p)
	assert.NotNil(t, p.RestClient())
	c := p.RestClient().ClientInterface.(*restClient.Client)
	assert.NotNil(t, c)
	assert.NoError(t, p.Close())
}
