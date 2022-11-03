package secretsengine

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	uuid "github.com/google/uuid"
	grafanaclient "github.com/grafana/grafana-api-golang-client"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	tempKeyPrefix = "vault-temp"
	tempKeyTTL    = time.Second * 30
)

type GrafanaCloudKey struct {
	ID               string
	Name             string
	Token            string
	Type             string
	StackSlug        string
	User             string
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
}

func (b *grafanaCloudBackend) grafanaCloudKey() *framework.Secret {
	return &framework.Secret{
		Type: grafanaCloudKeyType,
		Fields: map[string]*framework.FieldSchema{
			"user": {
				Type:        framework.TypeString,
				Description: "Grafana cloud api credentials username",
			},
			"token": {
				Type:        framework.TypeString,
				Description: "Grafana cloud api credentials Token",
			},
			"type": {
				Type:        framework.TypeString,
				Description: "Grafana cloud api type",
			},
		},
		Revoke: b.keyRevoke,
		Renew:  b.keyRenew,
	}
}

func deleteGrafanaAPIKey(client *grafanaclient.Client, stackSlug string, tokenID int64) error {
	instanceClient, cleanup, err := client.CreateTemporaryStackGrafanaClient(stackSlug, tempKeyPrefix, tempKeyTTL)
	if err != nil {
		return fmt.Errorf("error creating temporary Grafana API key: %w", err)
	}

	defer func() {
		err := cleanup()
		if err != nil {
			log.Printf("error deleting temporary Grafana Cloud HTTP API key: %s", err)
		}
	}()

	_, err = instanceClient.DeleteAPIKey(tokenID)
	if err != nil {
		return err
	}

	return nil
}

func (b *grafanaCloudBackend) keyRevoke(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	c, err := b.getClient(ctx, req.Storage)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	org := config.Organisation
	tokenName := req.Secret.InternalData["name"].(string)
	apiType, found := req.Secret.InternalData["type"].(string)

	if !found || apiType == "Cloud" {
		err = c.DeleteCloudAPIKey(org, tokenName)
	} else {
		tokenID := req.Secret.InternalData["id"].(string)
		stackSlug := req.Secret.InternalData["stack_slug"].(string)
		id, _ := strconv.ParseInt(tokenID, 10, 64)
		err = deleteGrafanaAPIKey(c, stackSlug, id)
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *grafanaCloudBackend) keyRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleRaw, ok := req.Secret.InternalData["role"]
	if !ok {
		return nil, NewInternalError("secret is missing role internal data", nil)
	}

	role := roleRaw.(string)
	roleEntry, err := b.getRole(ctx, req.Storage, role)
	if err != nil {
		return nil, NewInternalError("error retrieving role", err)
	}

	if roleEntry == nil {
		return nil, NewInternalError("error retrieving role: role is nil", nil)
	}

	resp := &logical.Response{Secret: req.Secret}

	if roleEntry.TTL > 0 {
		resp.Secret.TTL = roleEntry.TTL
	}
	if roleEntry.MaxTTL > 0 {
		resp.Secret.MaxTTL = roleEntry.MaxTTL
	}

	return resp, nil
}

func createCloudKey(
	_ context.Context,
	c *grafanaclient.Client,
	organisation string,
	roleName string,
	config *grafanaCloudConfig,
	grafanaCloudRole string,
) (*GrafanaCloudKey, error) {
	suffix := uuid.New().String()
	tokenName := fmt.Sprintf("%s_%s", roleName, suffix)

	key, err := c.CreateCloudAPIKey(
		organisation,
		&grafanaclient.CreateCloudAPIKeyInput{
			Name: tokenName,
			Role: grafanaCloudRole,
		})
	if err != nil {
		return nil, fmt.Errorf("error creating Grafana Cloud key: %w", err)
	}

	return &GrafanaCloudKey{
		Name:             key.Name,
		Token:            key.Token,
		Type:             CloudAPIType,
		User:             config.User,
		PrometheusUser:   config.PrometheusUser,
		PrometheusURL:    config.PrometheusURL,
		LokiUser:         config.LokiUser,
		LokiURL:          config.LokiURL,
		TempoUser:        config.TempoUser,
		TempoURL:         config.TempoURL,
		AlertmanagerUser: config.AlertmanagerUser,
		AlertmanagerURL:  config.AlertmanagerURL,
		GraphiteUser:     config.GraphiteUser,
		GraphiteURL:      config.GraphiteURL,
	}, nil
}

func createGrafanaKey(
	_ context.Context,
	c *grafanaclient.Client,
	stackSlug string,
	roleName string,
	secondsToLive int64,
	config *grafanaCloudConfig,
	grafanaCloudRole string,
) (*GrafanaCloudKey, error) {
	suffix := uuid.New().String()
	tokenName := fmt.Sprintf("%s_%s", roleName, suffix)
	instanceClient, cleanup, err := c.CreateTemporaryStackGrafanaClient(stackSlug, tempKeyPrefix, tempKeyTTL)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary Grafana API key: %w", err)
	}

	defer func() {
		err := cleanup()
		if err != nil {
			log.Printf("error deleting temporary Grafana API key: %s", err)
		}
	}()

	key, err := instanceClient.CreateAPIKey(grafanaclient.CreateAPIKeyRequest{
		Name:          tokenName,
		Role:          grafanaCloudRole,
		SecondsToLive: secondsToLive,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Grafana API key: %w", err)
	}

	return &GrafanaCloudKey{
		ID:               strconv.FormatInt(key.ID, 10),
		Name:             key.Name,
		StackSlug:        stackSlug,
		Type:             GrafanaAPIType,
		Token:            key.Key,
		User:             config.User,
		PrometheusUser:   config.PrometheusUser,
		PrometheusURL:    config.PrometheusURL,
		LokiUser:         config.LokiUser,
		LokiURL:          config.LokiURL,
		TempoUser:        config.TempoUser,
		TempoURL:         config.TempoURL,
		AlertmanagerUser: config.AlertmanagerUser,
		AlertmanagerURL:  config.AlertmanagerURL,
		GraphiteUser:     config.GraphiteUser,
		GraphiteURL:      config.GraphiteURL,
	}, nil
}
