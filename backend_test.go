package secretsengine

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
)

const (
	envVarRunAccTests              = "VAULT_ACC"
	envVarGrafanaCloudOrganisation = "TEST_GRAFANA_CLOUD_ORGANISATION"
	//nolint:gosec // test key.
	envVarGrafanaCloudAPIKey = "TEST_GRAFANA_CLOUD_API_KEY"
	envVarGrafanaCloudURL    = "TEST_GRAFANA_CLOUD_URL"
	envVarGrafanaCloudStack  = "TEST_GRAFANA_CLOUD_STACK"
	envVarCATarPath          = "TEST_GRAFANA_CLOUD_CA_TAR_PATH"
)

// getTestBackend will help you construct a test backend object.
// Update this function with your target backend.
func getTestBackend(tb testing.TB) (*grafanaCloudBackend, logical.Storage) {
	tb.Helper()

	config := logical.TestBackendConfig()
	config.StorageView = new(logical.InmemStorage)
	config.Logger = hclog.NewNullLogger()
	config.System = logical.TestSystemView()

	b, err := Factory(context.Background(), config)
	if err != nil {
		tb.Fatal(err)
	}

	return b.(*grafanaCloudBackend), config.StorageView
}

// runAcceptanceTests will separate unit tests from
// acceptance tests, which will make active requests
// to your target API.
var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

type issuedToken struct {
	ID   string
	Name string
	Type string
}

// testEnv creates an object to store and track testing environment
// resources.
type testEnv struct {
	// Password string.
	Organisation     string
	Key              string
	URL              string
	StackSlug        string
	PrometheusUser   string
	PrometheusURL    string
	LokiUser         string
	LokiURL          string
	TempoUser        string
	TempoURL         string
	AlertmanagerUser string
	AlertmanagerURL  string
	GraphiteUser     string
	GraphiteURL      string

	Backend logical.Backend
	Context context.Context
	Storage logical.Storage

	// SecretToken tracks the API token, for checking rotations.
	SecretToken string

	// IssuedTokens tracks the generated tokens, to make sure we clean up.
	IssuedTokens []issuedToken
}

// AddConfig adds the configuration to the test backend.
// Make sure data includes all of the configuration
// attributes you need and the `config` path!
func (e *testEnv) AddConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"organisation":      e.Organisation,
			"key":               e.Key,
			"url":               e.URL,
			"stack_slug":        e.StackSlug,
			"prometheus_user":   e.PrometheusUser,
			"prometheus_url":    e.PrometheusURL,
			"loki_user":         e.LokiUser,
			"loki_url":          e.LokiURL,
			"tempo_user":        e.TempoUser,
			"tempo_url":         e.TempoURL,
			"alertmanager_user": e.AlertmanagerUser,
			"alertmanager_url":  e.AlertmanagerURL,
			"graphite_user":     e.GraphiteUser,
			"graphite_url":      e.GraphiteURL,
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

// AddCloudUserTokenRole adds a role for the Grafana Cloud
// API token.
func (e *testEnv) AddCloudUserTokenRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-cloud-user-token",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"gc_role": "Viewer",
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, resp)
	require.Nil(t, err)
}

// ReadCloudUserToken retrieves the user token
// based on a Vault role.
func (e *testEnv) ReadCloudUserToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-cloud-user-token",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Secret.InternalData["name"])
	require.NotNil(t, resp.Secret)
	require.Equal(t, testPrometheusUser, resp.Data["user"])
	require.Equal(t, testPrometheusUser, resp.Data["prometheus_user"])
	require.Equal(t, testLokiUser, resp.Data["loki_user"])
	require.Equal(t, testTempoUser, resp.Data["tempo_user"])
	require.Equal(t, testAlertmanagerUser, resp.Data["alertmanager_user"])
	require.Equal(t, testGraphiteUser, resp.Data["graphite_user"])

	require.Equal(t, testPrometheusURL, resp.Data["prometheus_url"])
	require.Equal(t, testLokiURL, resp.Data["loki_url"])
	require.Equal(t, testTempoURL, resp.Data["tempo_url"])
	require.Equal(t, testAlertmanagerURL, resp.Data["alertmanager_url"])
	require.Equal(t, testGraphiteURL, resp.Data["graphite_url"])

	if e.SecretToken != "" {
		require.NotEqual(t, e.SecretToken, resp.Data["token"])
	}

	e.SecretToken = resp.Data["token"].(string)

	if t, ok := resp.Secret.InternalData["name"]; ok {
		e.IssuedTokens = append(e.IssuedTokens, issuedToken{
			Name: t.(string),
			Type: CloudAPIType,
		})
	}
}

// AddGrafanaUserTokenRole adds a role for the Grafana Cloud
// API token.
func (e *testEnv) AddGrafanaUserTokenRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test-grafana-user-token",
		Storage:   e.Storage,
		Data: map[string]interface{}{
			"api_type":   GrafanaAPIType,
			"stack_slug": e.StackSlug,
			"gc_role":    "Viewer",
			"ttl":        "120s",
			"max_ttl":    "300s",
		},
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.Nil(t, resp)
}

// ReadGrafanaUserToken retrieves the user token
// based on a Vault role.
func (e *testEnv) ReadGrafanaUserToken(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "creds/test-grafana-user-token",
		Storage:   e.Storage,
	}
	resp, err := e.Backend.HandleRequest(e.Context, req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Secret)

	require.NotEmpty(t, resp.Secret.InternalData["id"])
	require.NotEmpty(t, resp.Secret.InternalData["name"])
	require.NotEmpty(t, resp.Secret.InternalData["type"])

	require.Equal(t, GrafanaAPIType, resp.Secret.InternalData["type"])
	require.Equal(t, e.StackSlug, resp.Secret.InternalData["stack_slug"])

	if e.SecretToken != "" {
		require.NotEqual(t, e.SecretToken, resp.Data["token"])
	}

	e.SecretToken = resp.Data["token"].(string)

	if t, ok := resp.Secret.InternalData["name"]; ok {
		e.IssuedTokens = append(e.IssuedTokens, issuedToken{
			ID:   resp.Secret.InternalData["id"].(string),
			Name: t.(string),
			Type: GrafanaAPIType,
		})
	}
}

// CleanupUserTokens removes the tokens
// when the test completes.
func (e *testEnv) CleanupUserTokens(t *testing.T) {
	if len(e.IssuedTokens) == 0 {
		t.Fatalf("expected issued tokens, got: %d", len(e.IssuedTokens))
	}

	for _, token := range e.IssuedTokens {
		b := e.Backend.(*grafanaCloudBackend)
		client, err := b.getClient(e.Context, e.Storage)
		if err != nil {
			t.Fatal("fatal getting client")
		}

		if token.Type == CloudAPIType {
			err = client.DeleteCloudAPIKey(e.Organisation, token.Name)
		} else {
			tokenID, _ := strconv.ParseInt(token.ID, 10, 64)
			err = deleteGrafanaAPIKey(client, e.StackSlug, tokenID)
		}

		if err != nil {
			t.Fatalf("unexpected error deleting Cloud API key: %s", err)
		}
	}
}
