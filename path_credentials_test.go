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
	testPrometheusUrl    = "http://prometheus"
	testLokiUrl          = "http://loki"
	testTempoUrl         = "http://tempo"
	testAlertmanagerUrl  = "http://alertmanager"
	testGraphiteUrl      = "http://graphite"
)

// newAcceptanceTestEnv creates a test environment for credentials
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
		PrometheusUser:   testPrometheusUser,
		PrometheusURL:    testPrometheusUrl,
		LokiUser:         testLokiUser,
		LokiURL:          testLokiUrl,
		TempoUser:        testTempoUser,
		TempoURL:         testTempoUrl,
		AlertmanagerUser: testAlertmanagerUser,
		AlertmanagerURL:  testAlertmanagerUrl,
		GraphiteUser:     testGraphiteUser,
		GraphiteURL:      testGraphiteUrl,

		Backend: b,
		Context: ctx,
		Storage: &logical.InmemStorage{},
	}, nil
}

// TestAcceptanceUserToken tests a series of steps to make
// sure the role and token creation work correctly.
func TestAcceptanceUserToken(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	acceptanceTestEnv, err := newAcceptanceTestEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("add config", acceptanceTestEnv.AddConfig)
	t.Run("add user token role", acceptanceTestEnv.AddUserTokenRole)
	t.Run("read user token cred", acceptanceTestEnv.ReadUserToken)
	t.Run("read user token cred", acceptanceTestEnv.ReadUserToken)
	t.Run("cleanup user tokens", acceptanceTestEnv.CleanupUserTokens)
}
