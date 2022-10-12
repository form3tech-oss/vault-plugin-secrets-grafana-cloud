# Grafana Cloud Vault Secrets Engine

A Vault secrets engine for creating dynamic API keys in grafana cloud.

## Why?
Managing API tokens securely can be a challange, keeping keys secure and rotating keys is hard when tokens are distributed over many deployments and users.

If you already use Vault to manage secrets then this plugin will enable vault to issue short lived Grafana Cloud API keys, putting the responsibility for the security and lifetime of those keys within Vault itself.

## Installation

1. Place the plugin in the Vault plugin directory 

    - **a - From Binary -**
    Download the latest release and copy the binary into the vault plugin directory

    - **b - From Docker -**
    The plugin is also published as a [docker container](https://hub.docker.com/r/form3tech/vault-plugin-secrets-grafanacloud) that can be mounted as a shared volume with the vault plugin directory.


2. Register the plugin with Vault, the SHA256 is published along with each release:

```shell
vault plugin register -sha256=$SHA256 secret vault-plugin-secrets-grafanacloud
```

3. Once the plugin is installed and registered it must be enabled:

```shell
vault secrets enable -path=grafanacloud vault-plugin-secrets-grafanacloud
```

## Setup

These setup steps can also be performed using the [terraform provider](https://github.com/form3tech-oss/terraform-provider-vault-grafanacloud).

To configure the plugin you will need the following details of the grafana cloud organisation that the plugin will create api keys in:

| Item              | Description                                                                                                                                                                                     | 
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `organisation`    | The organisation name in grafana cloud (e.g. https://grafana.com/orgs/<organisation>)                                                                                                           |
| `key`             | An admin API key that is used by the plugin authenticate with the grafana cloud api                                                                                                             | 
| `url`             | The url or the grafana cloud api (usually `https://grafana.com/api/`)                                                                                                                           | 
| `user` (optional) | (Deprecated) The user ID that is used to authenticate with the grafana cloud prometheus endpoint. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `prometheus_user` (optional) | The user ID that is used to authenticate with the grafana cloud prometheus endpoint. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `prometheus_url` (optional) | The URL at which Prometheus can be accessed. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `loki_user` (optional) | The user ID that is used to authenticate with the grafana cloud loki endpoint. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `loki_url` (optional) | The URL at which Loki can be accessed. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `tempo_user` (optional) | The user ID that is used to authenticate with the grafana cloud tempo endpoint. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `tempo_url` (optional) | The URL at which Tempo can be accessed. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `alertmanager_user` (optional) | The user ID that is used to authenticate with the grafana cloud alertmanager endpoint. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `alertmanager_url` (optional) | The URL at which Alertmanager can be accessed. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `graphite_user` (optional) | The user ID that is used to authenticate with the grafana cloud graphite endpoint. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 
| `graphite_url` (optional) | The URL at which Graphite can be accessed. There is only one of these per stack see the grafana cloud stack dashboard for more details. | 

Configure the plugin with the details of the grafana cloud organisation:

```shell
# Write the configuration
vault write grafanacloud/config \
     organisation="$ORGANISATION" \
     key="$KEY" \
     url="$URL" \
     user="$USER"
```

## Usage

After the secrets engine is configured Vault can be used to generate grafana cloud api tokens for a given role. These steps can also be performed using the [terraform provider](https://github.com/form3tech-oss/terraform-provider-vault-grafanacloud).

### Cloud API keys

1. Setup role

```shell
vault write grafanacloud/roles/cloudrole \
    api_type="Cloud" # The type of api key to generate
    gc_role="Viewer" # The desired grafana cloud role for api keys generated for this role \
    ttl="300"        # Default lease for generated api keys \
    max_ttl="3600"   # Maximum time for role (see https://learn.hashicorp.com/tutorials/vault/tokens#ttl-and-max-ttl)
```

Valid values for `gc_role` are `Viewer`, `Admin`, `Editor`, `MetricsPublisher`, `PluginPublisher`

2. Retrieve a new grafana cloud API key from Vault

Any user/url configuration provided to the backend will be populated on the credential.

```shell
vault read grafanacloud/creds/cloudrole 

Key                Value
---                -----
lease_id           hashicups/creds/cloudrole/$LEASE_ID
lease_duration     5m
lease_renewable    true
token              $GRAFANA_CLOUD_TOKEN
type               Cloud
user               $CONFIGURED_USER_ID
```

3. Use the token in the grafana cloud API

```shell
# List stacks
curl -H "Authorization: Bearer $GRAFANA_CLOUD_TOKEN" https://grafana.com/api/orgs/<org_slug>/instances
```

### HTTP API keys

1. Setup role

```shell
vault write grafanacloud/roles/httprole \
    api_type="HTTP"
    stack_slug=<your-stack-slug>
    gc_role="Viewer"
    ttl="300"
    max_ttl="3600"
```

Notice how HTTP API keys are scoped to a specific stack.

Valid values for `gc_role` are `Viewer`, `Admin` and `Editor`.

2. Retrieve a new grafana cloud HTTP API key from Vault

```shell
vault read grafanacloud/creds/httprole 

Key                Value
---                -----
lease_id           hashicups/creds/httprole/$LEASE_ID
lease_duration     5m
lease_renewable    true
token              $GRAFANA_CLOUD_HTTP_TOKEN
type               HTTP
```

3. Use the token in the grafana cloud HTTP API

```shell
# Search folders and dashboards
curl -H "Authorization: Bearer $GRAFANA_CLOUD_HTTP_TOKEN" https://<your-stack-slug>.grafana.net/api/search/
```

## Testing

Tests can be run using `make test`.

To run the integration tests, you need to set some environment variables:

```
VAULT_ACC=1
TEST_GRAFANA_CLOUD_ORGANISATION=my-org
TEST_GRAFANA_CLOUD_API_KEY=<token>
TEST_GRAFANA_CLOUD_STACK=<stack>
TEST_GRAFANA_CLOUD_URL=https://grafana.com/api
TEST_GRAFANA_CLOUD_CA_TAR_PATH=<optional path to tar archive containing CA file if required for making HTTP requests from a docker container>
```