// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	vaultK8STokenFile  = `/var/run/secrets/kubernetes.io/serviceaccount/token` // #nosec
	vaultK8SLoginURL   = `/v1/auth/kubernetes/login`
	vaultSecretBaseURL = `/v1/secret/data/`           // #nosec
	vaultRevokeSelfURL = `/v1/auth/token/revoke-self` // #nosec
)

type SecretService interface {
	ReadSecret(ctx context.Context, path string) (string, error)
	WriteSecret(ctx context.Context, path string, secret string) error
	DeleteSecret(ctx context.Context, path string) error
	Logout(ctx context.Context)
}

type vaultServer struct {
	httpClient *http.Client
	vaultToken string
}

func newSecretService(ctx context.Context) (SecretService, error) {
	ss := &vaultServer{}
	err := ss.login(ctx)
	if err != nil {
		return nil, err
	}
	return ss, err
}

var SecretServiceFactory = newSecretService
var K8STokenFile = vaultK8STokenFile // #nosec
var VaultServerAddress = os.Getenv("VAULT_SERVER_ADDRESS")

func readAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

var readAllFactory = readAll

func httpsVaultURL(path string) string {
	return VaultServer + path
}

func getVaultHTTPClient() (*http.Client, error) {
	return &http.Client{
		Timeout: 10 * time.Second,
	}, nil
}

func loginToVault(ctx context.Context, httpClient *http.Client) (string, error) {
	tokenData, err := os.ReadFile(K8STokenFile)
	if err != nil {
		return "", err
	}
	loginReq := struct {
		JWT  string `json:"jwt"`
		Role string `json:"role"`
	}{
		JWT:  string(tokenData),
		Role: os.Getenv("SERVICE_ACCOUNT"),
	}
	body, _ := json.Marshal(loginReq)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		httpsVaultURL(vaultK8SLoginURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	log.Debugf("Logging in with URL %s", req.URL.String())
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var loginResp struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
		Errors []string `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		return "", err
	}
	return loginResp.Auth.ClientToken, nil

}

func (v *vaultServer) Logout(ctx context.Context) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		httpsVaultURL(vaultRevokeSelfURL),
		nil,
	)
	if err != nil {
		log.Infof("Error creating request revoking vault token: %v", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", v.vaultToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		log.Infof("Error invoking request revoking vault token: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		log.Infof("http error on revoke: %d", resp.StatusCode)
		return
	}
	v.vaultToken = ""
}

func (v *vaultServer) login(ctx context.Context) error {
	httpClient, err := getVaultHTTPClient()
	if err != nil {
		return err
	}
	// log in to Vault
	vaultToken, err := loginToVault(ctx, httpClient)
	if err != nil {
		return err
	}
	v.vaultToken = vaultToken
	v.httpClient = httpClient
	return nil
}

func (v *vaultServer) WriteSecret(ctx context.Context, path string, dataBlob string) error {
	data := `{"data":{"value":"` + dataBlob + `"}}`
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		httpsVaultURL(vaultSecretBaseURL+path),
		bytes.NewReader([]byte(data)),
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", v.vaultToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error on write: %d", resp.StatusCode)
	}
	return nil
}

func (v *vaultServer) ReadSecret(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		httpsVaultURL(vaultSecretBaseURL+path),
		nil,
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", v.vaultToken)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("secret not found")
	}
	getRawData, err := readAllFactory(resp.Body)
	if err != nil {
		return "", err
	}
	var secret vault.Secret
	err = json.Unmarshal(getRawData, &secret)
	if err != nil {
		return "", err
	}

	dataMap, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", errors.New("secret not found")
	}
	encodedSecretValue, ok := dataMap["value"].(string)
	if !ok {
		return "", errors.New("secret not found")
	}
	return encodedSecretValue, nil
}

func (v *vaultServer) DeleteSecret(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		httpsVaultURL(vaultSecretBaseURL+path),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", v.vaultToken)
	_, err = v.httpClient.Do(req)

	return err
}
