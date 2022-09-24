package secretsengine

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configStoragePath = "config"
)

// grafanaConfig includes the minimum configuration required to instantiate a new GrafanaCloud client.
type grafanaCloudConfig struct {
	Organisation     string `json:"organisation"`
	Key              string `json:"key"`
	URL              string `json:"url"`
	User             string `json:"user"`
	PrometheusUser   string `json:"prometheus_user"`
	PrometheusURL    string `json:"prometheus_url"`
	LokiUser         string `json:"loki_user"`
	LokiURL          string `json:"loki_url"`
	TempoUser        string `json:"tempo_user"`
	TempoURL         string `json:"tempo_url"`
	AlertmanagerUser string `json:"alertmanager_user"`
	AlertmanagerURL  string `json:"alertmanager_url"`
	GraphiteUser     string `json:"graphite_user"`
	GraphiteURL      string `json:"graphite_url"`
}

// pathConfig extends the Vault API with a `/config`
// endpoint for the backend. You can choose whether
// or not certain attributes should be displayed,
// required, and named. For example, password
// is marked as sensitive and will not be output
// when you read the configuration.
func pathConfig(b *grafanaCloudBackend) *framework.Path {
	return &framework.Path{
		Pattern: "config",
		Fields: map[string]*framework.FieldSchema{
			"key": {
				Type:        framework.TypeString,
				Description: "API key with Admin role to create user keys",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Admin Key",
					Sensitive: true,
				},
			},
			"organisation": {
				Type:        framework.TypeString,
				Description: "The Organisation slug for the Grafana Cloud API",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Organisation",
					Sensitive: false,
				},
			},
			"url": {
				Type:        framework.TypeString,
				Description: "The URL for the Grafana Cloud API",
				Required:    true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "URL",
					Sensitive: false,
				},
			},
			"user": {
				Type:        framework.TypeString,
				Description: "(Deprecated) The User that is needed to interact with prometheus, if set this is returned alongside every issued credential",
				Required:    false,
				Deprecated:  true,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "User",
					Sensitive: true,
				},
			},
			"prometheus_user": {
				Type:        framework.TypeString,
				Description: "The User that is needed to interact with prometheus, if set this is returned alongside every issued credential. This will also set 'user' for backwards compatibility",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Prometheus User",
					Sensitive: true,
				},
			},
			"prometheus_url": {
				Type:        framework.TypeString,
				Description: "The URL at which Prometheus can be accessed, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Prometheus URL",
					Sensitive: true,
				},
			},
			"loki_user": {
				Type:        framework.TypeString,
				Description: "The User that is needed to interact with loki, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Loki User",
					Sensitive: true,
				},
			},
			"loki_url": {
				Type:        framework.TypeString,
				Description: "The URL at which Loki can be accessed, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Loki URL",
					Sensitive: true,
				},
			},
			"tempo_user": {
				Type:        framework.TypeString,
				Description: "The User that is needed to interact with tempo, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Tempo User",
					Sensitive: true,
				},
			},
			"tempo_url": {
				Type:        framework.TypeString,
				Description: "The URL at which Tempo can be accessed, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Tempo URL",
					Sensitive: true,
				},
			},
			"alertmanager_user": {
				Type:        framework.TypeString,
				Description: "The User that is needed to interact with alertmanager, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Alertmanager User",
					Sensitive: true,
				},
			},
			"alertmanager_url": {
				Type:        framework.TypeString,
				Description: "The URL at which Alertmanager can be accessed, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Alertmanager URL",
					Sensitive: true,
				},
			},
			"graphite_user": {
				Type:        framework.TypeString,
				Description: "The User that is needed to interact with graphite, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Graphite User",
					Sensitive: true,
				},
			},
			"graphite_url": {
				Type:        framework.TypeString,
				Description: "The URL at which Graphite can be accessed, if set this is returned alongside every issued credential",
				Required:    false,
				DisplayAttrs: &framework.DisplayAttributes{
					Name:      "Graphite URL",
					Sensitive: true,
				},
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},
		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    pathConfigHelpSynopsis,
		HelpDescription: pathConfigHelpDescription,
	}
}

// pathConfigExistenceCheck verifies if the configuration exists.
func (b *grafanaCloudBackend) pathConfigExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

func getConfig(ctx context.Context, s logical.Storage) (*grafanaCloudConfig, error) {
	entry, err := s.Get(ctx, configStoragePath)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	config := new(grafanaCloudConfig)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("error reading root configuration: %w", err)
	}

	// return the config, we are done
	return config, nil
}

func (b *grafanaCloudBackend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"organisation":      config.Organisation,
			"key":               config.Key,
			"url":               config.URL,
			"user":              config.User,
			"prometheus_user":   config.PrometheusUser,
			"prometheus_url":    config.PrometheusURL,
			"loki_user":         config.LokiUser,
			"loki_url":          config.LokiURL,
			"tempo_user":        config.TempoUser,
			"tempo_url":         config.TempoURL,
			"alertmanager_user": config.AlertmanagerUser,
			"alertmanager_url":  config.AlertmanagerURL,
			"graphite_user":     config.GraphiteUser,
			"graphite_url":      config.GraphiteURL,
		},
	}, nil
}

func (b *grafanaCloudBackend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := getConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	createOperation := req.Operation == logical.CreateOperation

	if config == nil {
		if !createOperation {
			return nil, errors.New("config not found during update operation")
		}
		config = new(grafanaCloudConfig)
	}

	if organisation, ok := data.GetOk("organisation"); ok {
		config.Organisation = organisation.(string)
	}

	if config.Organisation == "" && createOperation {
		return nil, fmt.Errorf("missing organisation in configuration")
	}

	if key, ok := data.GetOk("key"); ok {
		config.Key = key.(string)
	}

	if config.Key == "" && createOperation {
		return nil, fmt.Errorf("missing key in configuration")
	}

	if configuredUrl, ok := data.GetOk("url"); ok {
		config.URL = configuredUrl.(string)
		if u, err := url.ParseRequestURI(config.URL); err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("invalid url in configuration")
		}
	} else if !ok && createOperation {
		return nil, fmt.Errorf("missing url in configuration")
	}

	if user, ok := data.GetOk("user"); ok {
		config.User = user.(string)
	}

	if user, ok := data.GetOk("prometheus_user"); ok {
		config.PrometheusUser = user.(string)
		// Set user for backwards compatibility.
		config.User = user.(string)
	}

	if user, ok := data.GetOk("loki_user"); ok {
		config.LokiUser = user.(string)
	}

	if user, ok := data.GetOk("tempo_user"); ok {
		config.TempoUser = user.(string)
	}

	if user, ok := data.GetOk("alertmanager_user"); ok {
		config.AlertmanagerUser = user.(string)
	}

	if user, ok := data.GetOk("graphite_user"); ok {
		config.GraphiteUser = user.(string)
	}

	if prometheusUrl, ok := data.GetOk("prometheus_url"); ok {
		config.PrometheusURL = prometheusUrl.(string)
		if u, err := url.ParseRequestURI(config.PrometheusURL); err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("invalid prometheus_url in configuration")
		}
	}

	if lokiUrl, ok := data.GetOk("loki_url"); ok {
		config.LokiURL = lokiUrl.(string)
		if u, err := url.ParseRequestURI(config.LokiURL); err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("invalid loki_url in configuration")
		}
	}

	if tempoUrl, ok := data.GetOk("tempo_url"); ok {
		config.TempoURL = tempoUrl.(string)
		if u, err := url.ParseRequestURI(config.TempoURL); err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("invalid tempo_url in configuration")
		}
	}

	if AlertmanagerUrl, ok := data.GetOk("alertmanager_url"); ok {
		config.AlertmanagerURL = AlertmanagerUrl.(string)
		if u, err := url.ParseRequestURI(config.AlertmanagerURL); err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("invalid alertmanager_url in configuration")
		}
	}

	if graphiteURL, ok := data.GetOk("graphite_url"); ok {
		config.GraphiteURL = graphiteURL.(string)
		if u, err := url.ParseRequestURI(config.GraphiteURL); err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("invalid graphite_url in configuration")
		}
	}

	entry, err := logical.StorageEntryJSON(configStoragePath, config)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	b.reset()

	return nil, nil
}

func (b *grafanaCloudBackend) pathConfigDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	err := req.Storage.Delete(ctx, configStoragePath)

	if err == nil {
		b.reset()
	}

	return nil, err
}

// pathConfigHelpSynopsis summarizes the help text for the configuration
const pathConfigHelpSynopsis = `Configure the Grafana Cloud backend.`

// pathConfigHelpDescription describes the help text for the configuration
const pathConfigHelpDescription = `
The Grafana Cloud secret backend requires credentials for managing
API keys that it issues.
`
