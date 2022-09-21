package secretsengine

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	grafanclient "github.com/grafana/grafana-api-golang-client"
	uuid "github.com/google/uuid"
	"github.com/hashicorp/vault/sdk/logical"
)

type GrafanaCloudKey struct {
	Name             string
	Token            string
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
		},
		Revoke: b.keyRevoke,
		Renew:  b.keyRenew,
	}
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
	tokenID := req.Secret.InternalData["name"].(string)
	err = c.DeleteCloudAPIKey(org, tokenID)
	if err != nil {
		return nil, err
	}

	return &logical.Response{}, nil
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

func createKey(_ context.Context, c *grafanclient.Client, organisation, roleName string, config *grafanaCloudConfig, grafanaCloudRole string) (*GrafanaCloudKey, error) {
	suffix := uuid.New().String()
	tokenName := fmt.Sprintf("%s_%s", roleName, suffix)

	key, err := c.CreateCloudAPIKey(
		organisation,
		&grafanclient.CreateCloudAPIKeyInput{
			Name: tokenName,
			Role: grafanaCloudRole,
		})

	if err != nil {
		return nil, fmt.Errorf("error creating Grafana Cloud key: %w", err)
	}

	return &GrafanaCloudKey{
		Name:             key.Name,
		Token:            key.Token,
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
