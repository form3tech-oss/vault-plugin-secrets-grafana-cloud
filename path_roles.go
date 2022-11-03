package secretsengine

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	pathRoleHelpSynopsis    = `Manages the Vault role for generating Grafana Cloud API tokens.`
	pathRoleHelpDescription = `
This path allows you to read and write roles used to generate Grafana Cloud tokens.
`

	pathRoleListHelpSynopsis           = `List the existing roles in Grafana Cloud backend`
	pathRoleListHelpDescription        = `Roles will be listed by the role name.`
	CloudAPIType                string = "Cloud"
	GrafanaAPIType              string = "Grafana"
)

// grafanaCloudRoleEntry defines the data required
// for a Vault role to access and call the Grafana Cloud
// token endpoints.
type grafanaCloudRoleEntry struct {
	APIType          string        `json:"api_type"`
	StackSlug        string        `json:"stack_slug"`
	GrafanaCloudRole string        `json:"gc_role"`
	TTL              time.Duration `json:"ttl"`
	MaxTTL           time.Duration `json:"max_ttl"`
}

// grafanaCloudValidRoles valid roles in Grafana Cloud
// to whom keys may be generated.
//
//nolint:gochecknoglobals // required for verifying role is valid.
var grafanaCloudValidRoles = map[string]bool{
	"Viewer":           true,
	"Admin":            true,
	"Editor":           true,
	"MetricsPublisher": true,
	"PluginPublisher":  true,
}

// grafanaCloudValidAPITypes API key types supported by Grafana Cloud.
//
//nolint:gochecknoglobals // required for verifying role is valid.
var grafanaCloudValidAPITypes = map[string]bool{
	CloudAPIType:   true,
	GrafanaAPIType: true,
}

// toResponseData returns response data for a role.
func (r *grafanaCloudRoleEntry) toResponseData() map[string]interface{} {
	respData := map[string]interface{}{
		"api_type":   r.APIType,
		"stack_slug": r.StackSlug,
		"gc_role":    r.GrafanaCloudRole,
		"ttl":        r.TTL.Seconds(),
		"max_ttl":    r.MaxTTL.Seconds(),
	}
	return respData
}

// pathRole extends the Vault API with a `/role`
// endpoint for the backend. You can choose whether
// or not certain attributes should be displayed,
// required, and named. You can also define different
// path patterns to list all roles.
func pathRole(b *grafanaCloudBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: "roles/" + framework.GenericNameRegex("name"),
			Fields: map[string]*framework.FieldSchema{
				"name": {
					Type:        framework.TypeString,
					Description: "The actual Role name",
					Required:    true,
				},
				"gc_role": {
					Type:        framework.TypeString,
					Description: "The Grafana Cloud role, i.e. the key authorization level",
					Required:    true,
				},
				"api_type": {
					Type:        framework.TypeString,
					Description: "The Grafana Cloud API type (Cloud or Grafana)",
					Default:     CloudAPIType,
					Required:    true,
				},
				"stack_slug": {
					Type:        framework.TypeString,
					Description: "The stack slug to create Grafana API keys for",
				},
				"ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Default lease for generated credentials. If not set or set to 0, will use system default.",
				},
				"max_ttl": {
					Type:        framework.TypeDurationSecond,
					Description: "Maximum time for role. If not set or set to 0, will use system default.",
				},
			},
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.pathRolesRead,
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.pathRolesWrite,
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.pathRolesWrite,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.pathRolesDelete,
				},
			},
			HelpSynopsis:    pathRoleHelpSynopsis,
			HelpDescription: pathRoleHelpDescription,
		},
		{
			Pattern: "roles/?$",
			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ListOperation: &framework.PathOperation{
					Callback: b.pathRolesList,
				},
			},
			HelpSynopsis:    pathRoleListHelpSynopsis,
			HelpDescription: pathRoleListHelpDescription,
		},
	}
}

func (b *grafanaCloudBackend) getRole(ctx context.Context, s logical.Storage, name string) (*grafanaCloudRoleEntry, error) {
	if name == "" {
		return nil, NewInvalidConfigurationError("missing role name", nil)
	}

	entry, err := s.Get(ctx, "roles/"+name)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var role grafanaCloudRoleEntry

	if err := entry.DecodeJSON(&role); err != nil {
		return nil, err
	}
	return &role, nil
}

func (b *grafanaCloudBackend) pathRolesRead(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entry, err := b.getRole(ctx, req.Storage, d.Get("name").(string))
	if err != nil {
		return nil, err
	}

	if entry == nil {
		//nolint:godox
		// TODO: shouldn't we return a Response or an error here?
		return nil, nil
	}

	return &logical.Response{
		Data: entry.toResponseData(),
	}, nil
}

func setRole(ctx context.Context, s logical.Storage, name string, roleEntry *grafanaCloudRoleEntry) error {
	entry, err := logical.StorageEntryJSON("roles/"+name, roleEntry)
	if err != nil {
		return err
	}

	if entry == nil {
		return NewInternalError("failed to create storage entry for role", nil)
	}

	if err := s.Put(ctx, entry); err != nil {
		return err
	}

	return nil
}

func (b *grafanaCloudBackend) pathRolesWrite(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	name, ok := d.GetOk("name")
	if !ok {
		return logical.ErrorResponse("missing role name"), nil
	}

	roleEntry, err := b.getRole(ctx, req.Storage, name.(string))
	if err != nil {
		return nil, err
	}

	if roleEntry == nil {
		roleEntry = &grafanaCloudRoleEntry{}
	}

	createOperation := req.Operation == logical.CreateOperation

	if gcAPIType, ok := d.GetOk("api_type"); ok {
		roleEntry.APIType = gcAPIType.(string)
		if _, ok := grafanaCloudValidAPITypes[roleEntry.APIType]; !ok {
			return logical.ErrorResponse(
				fmt.Sprintf("provided api_type %s is not valid. Valid values are '%s' and '%s'", roleEntry.APIType, CloudAPIType, GrafanaAPIType)), nil
		}
	} else {
		roleEntry.APIType = CloudAPIType
	}

	if gcRole, ok := d.GetOk("gc_role"); ok {
		roleEntry.GrafanaCloudRole = gcRole.(string)
		if _, ok := grafanaCloudValidRoles[roleEntry.GrafanaCloudRole]; !ok {
			return logical.ErrorResponse(fmt.Sprintf("provided gc_role %s is not valid", roleEntry.GrafanaCloudRole)), nil
		}
	}

	if stackSlug, ok := d.GetOk("stack_slug"); ok {
		roleEntry.StackSlug = stackSlug.(string)
		if roleEntry.APIType == GrafanaAPIType && roleEntry.StackSlug == "" {
			return logical.ErrorResponse("need to specify a stack_slug for Grafana API keys."), nil
		}
	} else if roleEntry.APIType == GrafanaAPIType {
		return logical.ErrorResponse("need to specify a stack_slug for Grafana API keys."), nil
	}

	if !ok && createOperation {
		return nil, NewInvalidConfigurationError("missing gc_role value", nil)
	}

	if ttlRaw, ok := d.GetOk("ttl"); ok {
		roleEntry.TTL = time.Duration(ttlRaw.(int)) * time.Second
	} else if createOperation {
		// Use default value
		roleEntry.TTL = time.Duration(d.Get("ttl").(int)) * time.Second
	}

	if maxTTLRaw, ok := d.GetOk("max_ttl"); ok {
		roleEntry.MaxTTL = time.Duration(maxTTLRaw.(int)) * time.Second
	} else if createOperation {
		// Use default value
		roleEntry.MaxTTL = time.Duration(d.Get("max_ttl").(int)) * time.Second
	}

	if roleEntry.MaxTTL != 0 && roleEntry.TTL > roleEntry.MaxTTL {
		return logical.ErrorResponse("ttl cannot be greater than max_ttl"), nil
	}

	if err := setRole(ctx, req.Storage, name.(string), roleEntry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *grafanaCloudBackend) pathRolesDelete(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, "roles/"+d.Get("name").(string))
	if err != nil {
		return nil, fmt.Errorf("error deleting grafanaCloud role: %w", err)
	}

	return nil, nil
}

func (b *grafanaCloudBackend) pathRolesList(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	entries, err := req.Storage.List(ctx, "roles/")
	if err != nil {
		return nil, err
	}

	return logical.ListResponse(entries), nil
}
