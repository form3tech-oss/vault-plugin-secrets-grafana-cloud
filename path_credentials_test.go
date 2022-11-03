package secretsengine

import (
	"context"
	"os"
	"testing"
	"time"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/helper/logging"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	testPrometheusUser   = "1"
	testLokiUser         = "2"
	testTempoUser        = "3"
	testAlertmanagerUser = "4"
	testGraphiteUser     = "5"
	testPrometheusURL    = "http://prometheus"
	testLokiURL          = "http://loki"
	testTempoURL         = "http://tempo"
	testAlertmanagerURL  = "http://alertmanager"
	testGraphiteURL      = "http://graphite"
)

// newAcceptanceTestEnv creates a test environment for credentials.
func newAcceptanceTestEnv() (*testEnv, error) {
	ctx := context.Background()

	maxLease, _ := time.ParseDuration("60s")
	defaultLease, _ := time.ParseDuration("30s")
	conf := &logical.BackendConfig{
		System: &logical.StaticSystemView{
			DefaultLeaseTTLVal: defaultLease,
			MaxLeaseTTLVal:     maxLease,
		},
		Logger: logging.NewVaultLogger(log.Debug),
	}
	b, err := Factory(ctx, conf)
	if err != nil {
		return nil, err
	}
	return &testEnv{
		Organisation:     os.Getenv(envVarGrafanaCloudOrganisation),
		Key:              os.Getenv(envVarGrafanaCloudAPIKey),
		URL:              os.Getenv(envVarGrafanaCloudURL),
		StackSlug:        os.Getenv(envVarGrafanaCloudStack),
		PrometheusUser:   testPrometheusUser,
		PrometheusURL:    testPrometheusURL,
		LokiUser:         testLokiUser,
		LokiURL:          testLokiURL,
		TempoUser:        testTempoUser,
		TempoURL:         testTempoURL,
		AlertmanagerUser: testAlertmanagerUser,
		AlertmanagerURL:  testAlertmanagerURL,
		GraphiteUser:     testGraphiteUser,
		GraphiteURL:      testGraphiteURL,

		Backend: b,
		Context: ctx,
		Storage: &logical.InmemStorage{},
	}, nil
}

// TestAcceptanceCloudAPIKey tests a series of steps to make
// sure the role and token creation work correctly.
func TestAcceptanceCloudAPIKey(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add cloud user token role", acceptanceTestEnv.AddCloudUserTokenRole)
	t.Run("read cloud user token cred", acceptanceTestEnv.ReadCloudUserToken)
	t.Run("read cloud user token cred", acceptanceTestEnv.ReadCloudUserToken)
	t.Run("cleanup user tokens", acceptanceTestEnv.CleanupUserTokens)
}

func TestAcceptanceGrafanaAPIKey(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add grafana user token role", acceptanceTestEnv.AddGrafanaUserTokenRole)
	t.Run("read grafana user token cred", acceptanceTestEnv.ReadGrafanaUserToken)
	t.Run("read grafana user token cred", acceptanceTestEnv.ReadGrafanaUserToken)
	t.Run("cleanup user tokens", acceptanceTestEnv.CleanupUserTokens)
}
