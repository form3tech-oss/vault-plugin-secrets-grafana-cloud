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
| `user` (optional) | The user ID that is used to authenticate with the grafana cloud prometheus endpoint. There is only one of these per organisation see the grafana cloud organisation dashboard for more details. | 

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

1. Setup role

```shell
vault write grafanacloud/role/examplerole \
    gc_role="Viewer" # The desired grafana cloud role for api keys generated for this role \
    ttl="300"        # Default lease for generated api keys \
    max_ttl="3600"   # Maximum time for role (see https://learn.hashicorp.com/tutorials/vault/tokens#ttl-and-max-ttl)
```

Valid values for `gc_role` are `Viewer`, `Admin`, `Editor`, `MetricsPublisher`, `PluginPublisher`

2. Retrieve a new grafana cloud API key from Vault

```shell
vault read grafanacloud/creds/examplerole 

Key                Value
---                -----
lease_id           hashicups/creds/examplerole/$LEASE_ID
lease_duration     5m
lease_renewable    true
token              $GRAFANA_CLOUD_TOKEN
user               $CONFIGURED_USER_ID
```

3. Use the token in the grafana cloud API

```shell
# List stacks
curl -H "Authorization: Bearer $GRAFANA_CLOUD_TOKEN" https://grafana.com/api/orgs/<org_slug>/instances
```