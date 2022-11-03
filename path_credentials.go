package secretsengine

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// pathCredentials extends the Vault API with a `/creds`
// endpoint for a role. You can choose whether
// or not certain attributes should be displayed,
// required, and named.
func pathCredentials(b *grafanaCloudBackend) *framework.Path {
	return &framework.Path{
		Pattern: "creds/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeLowerCaseString,
				Description: "Name of the role",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation:   &framework.PathOperation{Callback: b.pathCredentialsRead},
			logical.UpdateOperation: &framework.PathOperation{Callback: b.pathCredentialsRead},
		},
		HelpSynopsis:    pathCredentialsHelpSyn,
		HelpDescription: pathCredentialsHelpDesc,
	}
}

//nolint:gosec // help string, not credential.
const pathCredentialsHelpSyn = `
Generate a Grafana Cloud API key from a specific Vault role.
`

//nolint:gosec // help string, not credential.
const pathCredentialsHelpDesc = `
This path generates a Grafana Cloud API key based on a particular role.
`

func (b *grafanaCloudBackend) createKey(ctx context.Context, s logical.Storage, roleName string, roleEntry *grafanaCloudRoleEntry) (*GrafanaCloudKey, error) {
	client, err := b.getClient(ctx, s)
	if err != nil {
		return nil, err
	}

	config, err := getConfig(ctx, s)
	if err != nil {
		return nil, NewInternalError("error reading secrets engine configuration", err)
	}

	var token *GrafanaCloudKey

	if roleEntry.APIType == CloudAPIType {
		token, err = createCloudKey(ctx, client, config.Organisation, roleName, config, roleEntry.GrafanaCloudRole)
	} else {
		token, err = createGrafanaKey(ctx, client, roleEntry.StackSlug, roleName, int64(roleEntry.TTL.Seconds()), config, roleEntry.GrafanaCloudRole)
	}

	if err != nil || token == nil {
		return nil, NewInternalError("error creating Grafana Cloud token", err)
	}

	return token, nil
}

func (b *grafanaCloudBackend) createUserCreds(ctx context.Context, req *logical.Request, roleName string,
	role *grafanaCloudRoleEntry,
) (*logical.Response, error) {
	key, err := b.createKey(ctx, req.Storage, roleName, role)
	if err != nil {
		return nil, err
	}

	responseData := map[string]interface{}{
		"type":  key.Type,
		"token": key.Token,
	}

	if key.User != "" {
		responseData["user"] = key.User
	}

	if key.PrometheusUser != "" {
		responseData["prometheus_user"] = key.PrometheusUser
	}

	if key.LokiUser != "" {
		responseData["loki_user"] = key.LokiUser
	}

	if key.TempoUser != "" {
		responseData["tempo_user"] = key.TempoUser
	}

	if key.AlertmanagerUser != "" {
		responseData["alertmanager_user"] = key.AlertmanagerUser
	}

	if key.GraphiteUser != "" {
		responseData["graphite_user"] = key.GraphiteUser
	}

	if key.PrometheusURL != "" {
		responseData["prometheus_url"] = key.PrometheusURL
	}

	if key.LokiURL != "" {
		responseData["loki_url"] = key.LokiURL
	}

	if key.TempoURL != "" {
		responseData["tempo_url"] = key.TempoURL
	}

	if key.AlertmanagerURL != "" {
		responseData["alertmanager_url"] = key.AlertmanagerURL
	}

	if key.GraphiteURL != "" {
		responseData["graphite_url"] = key.GraphiteURL
	}

	internalData := map[string]interface{}{
		"id":   key.ID,
		"name": key.Name,
		"type": key.Type,
	}

	if key.StackSlug != "" {
		internalData["stack_slug"] = key.StackSlug
	}

	resp := b.Secret(grafanaCloudKeyType).Response(responseData, internalData)

	if role.TTL > 0 {
		resp.Secret.TTL = role.TTL
	}

	if role.MaxTTL > 0 {
		resp.Secret.MaxTTL = role.MaxTTL
	}

	return resp, nil
}

func (b *grafanaCloudBackend) pathCredentialsRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	roleName := d.Get("name").(string)

	roleEntry, err := b.getRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, NewInternalError("error retrieving role", err)
	}

	if roleEntry == nil {
		return nil, NewInternalError("error retrieving role: role is nil", nil)
	}

	return b.createUserCreds(ctx, req, roleName, roleEntry)
}
