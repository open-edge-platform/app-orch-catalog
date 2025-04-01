// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
)

type TestHTTPServer struct {
	K8SLoginReadHandler func(w http.ResponseWriter)
	SecretHandler       func(w http.ResponseWriter, r *http.Request)
	RevokeHandler       func(w http.ResponseWriter)
	Server              *httptest.Server
}

func (t *TestHTTPServer) WithK8SLoginReadHandler(K8SLoginReadHANDLER func(w http.ResponseWriter)) *TestHTTPServer {
	t.K8SLoginReadHandler = K8SLoginReadHANDLER
	return t
}

func (t *TestHTTPServer) WithSecretHandler(SecretHandler func(w http.ResponseWriter, r *http.Request)) *TestHTTPServer {
	t.SecretHandler = SecretHandler
	return t
}
func (t *TestHTTPServer) WithRevokeHandler(RevokeHandler func(w http.ResponseWriter)) *TestHTTPServer {
	t.RevokeHandler = RevokeHandler
	return t
}

func (s *NorthBoundTestSuite) NewTestHTTPServer() *TestHTTPServer {
	return &TestHTTPServer{
		K8SLoginReadHandler: s.handleK8SLogin,
		SecretHandler:       s.handleSecret,
		RevokeHandler:       s.handleRevoke,
	}
}

func (t *TestHTTPServer) Start() *TestHTTPServer {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case vaultK8SLoginURL:
			t.K8SLoginReadHandler(w)
		case vaultSecretBaseURL + `secret-path`:
			t.SecretHandler(w, r)
		case vaultRevokeSelfURL:
			t.RevokeHandler(w)
		}
	}))
	t.Server = server
	VaultServer = server.URL
	K8STokenFile = `testdata/k8stoken` // #nosec
	return t
}

func (t *TestHTTPServer) Stop() {
	t.Server.Close()
}

func (s *NorthBoundTestSuite) handleK8SLogin(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
		Errors []string `json:"errors"`
	}
	loginResp.Auth.ClientToken = "token"
	js, err := json.Marshal(loginResp)
	s.NoError(err)
	_, _ = w.Write(js)
}

func (s *NorthBoundTestSuite) handleK8SLoginBadJSON(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
		Errors []string `json:"errors"`
	}
	loginResp.Auth.ClientToken = "token"
	_, _ = w.Write([]byte("This is not the JSON you are looking for"))
}

func (s *NorthBoundTestSuite) handleRevoke(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *NorthBoundTestSuite) handleRevokeHTTPError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

var secrets = map[string]string{}

func (s *NorthBoundTestSuite) handleSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)
		secretData := map[string]interface{}{}
		rawData, err := io.ReadAll(r.Body)
		s.NoError(err)
		err = json.Unmarshal(rawData, &secretData)
		s.NoError(err)
		data := secretData["data"].(map[string]interface{})
		s.NotNil(data)
		secret := data["value"].(string)
		secretJSON := `{"data":{"data":{"value":"` + secret + `"}}}`
		secrets[r.URL.Path] = secretJSON
	} else if r.Method == http.MethodGet {
		secretJSON, ok := secrets[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(secretJSON))
		}
	} else if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusOK)
		delete(secrets, r.URL.Path)
	}
}

func (s *NorthBoundTestSuite) handleSecretBadHTTPStatus(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func NilContext() context.Context {
	return context.Context(nil)
}

func CancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func (s *NorthBoundTestSuite) TestVault() {
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	ss, err := SecretServiceFactory(s.ctx)
	s.NotNil(ss)
	s.NoError(err)

	err = ss.WriteSecret(s.ctx, "secret-path", "secret-data")
	s.NoError(err)
	secret, err := ss.ReadSecret(s.ctx, "secret-path")
	s.NoError(err)
	s.Equal("secret-data", secret)

	err = ss.DeleteSecret(s.ctx, "secret-path")
	s.NoError(err)
	_, err = ss.ReadSecret(s.ctx, "secret-path")
	s.Error(err)
}

func (s *NorthBoundTestSuite) TestLoadCAErrors() {
	saveReadAllFactory := readAllFactory
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	// Test io READ error
	readAllFactory = func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("readAll error")
	}
	c, err := getVaultHTTPClient()
	s.NoError(err)
	s.NotNil(c)
	readAllFactory = saveReadAllFactory
}

func (s *NorthBoundTestSuite) TestLoginErrors() {
	saveReadAllFactory := readAllFactory
	server := s.NewTestHTTPServer().WithK8SLoginReadHandler(s.handleK8SLoginBadJSON).Start()
	defer server.Stop()

	// login error - can't get token
	K8STokenFile = `testdata/k8stoken-no-such-file` // #nosec
	c, _ := getVaultHTTPClient()
	token, err := loginToVault(s.ctx, c)
	s.Equal("", token)
	s.Error(err)

	// login error - can't unmarshal JSON login response
	K8STokenFile = `testdata/k8stoken` // #nosec
	c, _ = getVaultHTTPClient()
	token, err = loginToVault(s.ctx, c)
	s.Equal("", token)
	s.Error(err)

	// Test io READ error
	c, _ = getVaultHTTPClient()
	readAllFactory = func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("readAll error")
	}
	token, err = loginToVault(s.ctx, c)
	s.Equal("", token)
	s.Error(err)
	readAllFactory = saveReadAllFactory

	// Test unable to create request
	c, _ = getVaultHTTPClient()
	emptyContext := context.Context(nil)
	token, err = loginToVault(emptyContext, c)
	s.Equal("", token)
	s.Error(err)

	// Test unable to read from server
	c, _ = getVaultHTTPClient()
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()
	token, err = loginToVault(ctx, c)
	s.Equal("", token)
	s.Error(err)
}

func (s *NorthBoundTestSuite) TestNewSecretServiceErrors() {
	saveReadAllFactory := readAllFactory

	// Test can't create client error
	server := s.NewTestHTTPServer().Start()
	ss, err := newSecretService(s.ctx)
	s.NotNil(ss)
	s.NoError(err)
	readAllFactory = saveReadAllFactory
	server.Stop()

	// login error - can't get token
	server = s.NewTestHTTPServer().Start()
	K8STokenFile = `testdata/k8stoken-no-such-file` // #nosec
	ss, err = newSecretService(s.ctx)
	s.Nil(ss)
	s.Error(err)
	server.Stop()
}

func (s *NorthBoundTestSuite) TestWriteSecretErrors() {
	server := s.NewTestHTTPServer().WithSecretHandler(s.handleSecretBadHTTPStatus).Start()
	defer server.Stop()

	// Test can't create client error
	ss, err := newSecretService(s.ctx)
	s.NotNil(ss)
	s.NoError(err)
	err = ss.WriteSecret(NilContext(), "path", "secret")
	s.Error(err)

	ss, err = newSecretService(s.ctx)
	s.NotNil(ss)
	s.NoError(err)

	// Test unable to create request
	err = ss.WriteSecret(CancelledContext(), "secret-path", "secret")
	s.Error(err)

	// Post will return an error
	err = ss.WriteSecret(s.ctx, "secret-path", "secret")
	s.Error(err)
}

func (s *NorthBoundTestSuite) TestReadSecretErrors() {
	saveReadAllFactory := readAllFactory
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	ss, err := newSecretService(s.ctx)
	s.NotNil(ss)
	s.NoError(err)
	_, err = ss.ReadSecret(NilContext(), "path")
	s.Error(err)

	ss, err = newSecretService(s.ctx)
	s.NotNil(ss)
	s.NoError(err)

	_, err = ss.ReadSecret(CancelledContext(), "secret-path")
	s.Error(err)

	// secret not found
	_, err = ss.ReadSecret(s.ctx, "secret-path")
	s.Error(err)

	// read error making request
	_ = ss.WriteSecret(s.ctx, "secret-path", "xyzzy")
	readAllFactory = func(_ io.Reader) ([]byte, error) {
		return nil, errors.New("readAll error")
	}
	_, err = ss.ReadSecret(s.ctx, "secret-path")
	s.Error(err)

	// Bad JSON
	readAllFactory = saveReadAllFactory

	secrets["/v1/secret/data/secret-path"] = `{Not JSON`
	_, err = ss.ReadSecret(s.ctx, "secret-path")
	s.Error(err)

	// Wrong JSON key
	secrets["/v1/secret/data/secret-path"] = `{"data":{"data":{"XXXvalueXXX":"xyzzy"}}}`
	_, err = ss.ReadSecret(s.ctx, "secret-path")
	s.Error(err)

	// Wrong JSON map
	secrets["/v1/secret/data/secret-path"] = `{"data":{"XXXdataXXX":{"value":"xyzzy"}}}`
	_, err = ss.ReadSecret(s.ctx, "secret-path")
	s.Error(err)
}

func (s *NorthBoundTestSuite) TestDeleteSecretErrors() {
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	ss, err := newSecretService(s.ctx)
	s.NotNil(ss)
	s.NoError(err)

	err = ss.DeleteSecret(NilContext(), "path")
	s.Error(err)

	err = ss.DeleteSecret(CancelledContext(), "path")
	s.Error(err)
}

// TestLogout tests the successful path for logging out from the secrets service
func (s *NorthBoundTestSuite) TestLogout() {
	server := s.NewTestHTTPServer().Start()
	defer server.Stop()

	ss, err := SecretServiceFactory(s.ctx)
	s.NoError(err)

	vs := ss.(*vaultServer)
	vs.vaultToken = "AAAAA"
	ss.Logout(s.ctx)
	s.Equal("", vs.vaultToken)
}

// TestLogoutHTTPError tests when the secret HTTP server returns an error on the revoke
func (s *NorthBoundTestSuite) TestLogoutHTTPError() {
	server := s.NewTestHTTPServer().WithRevokeHandler(s.handleRevokeHTTPError).Start()
	defer server.Stop()

	ss, err := SecretServiceFactory(s.ctx)
	s.NoError(err)

	vs := ss.(*vaultServer)
	vs.vaultToken = "AAAAA"
	ss.Logout(s.ctx)
	s.Equal("AAAAA", vs.vaultToken)
}

// TestLogoutRequestFailure tests failing to complete the HTTP request to the secrets server due to a cancelled context
func (s *NorthBoundTestSuite) TestLogoutRequestFailure() {
	server := s.NewTestHTTPServer().Start()
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()
	defer server.Stop()

	ss, err := SecretServiceFactory(s.ctx)
	s.NoError(err)

	vs := ss.(*vaultServer)
	vs.vaultToken = "AAAAA"
	ss.Logout(ctx)
	s.Equal("AAAAA", vs.vaultToken)
}

// TestLogoutRequestFailure tests failing to create an HTTP request for the revoke call
func (s *NorthBoundTestSuite) TestLogoutRequestCreationFailure() {
	server := s.NewTestHTTPServer().Start()
	var ctx context.Context
	defer server.Stop()

	ss, err := SecretServiceFactory(s.ctx)
	s.NoError(err)

	vs := ss.(*vaultServer)
	vs.vaultToken = "AAAAA"
	ss.Logout(ctx)
	s.Equal("AAAAA", vs.vaultToken)
}
