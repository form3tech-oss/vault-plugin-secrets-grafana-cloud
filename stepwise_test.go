package secretsengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	stepwise "github.com/hashicorp/vault-testing-stepwise"
	dockerEnvironment "github.com/hashicorp/vault-testing-stepwise/environments/docker"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

// TestAccUserToken runs a series of acceptance tests to check the
// end-to-end workflow of the backend. It creates a Vault Docker container
// and loads a temporary plugin.
func TestAccUserToken(t *testing.T) {
	t.Parallel()
	if !runAcceptanceTests {
		t.SkipNow()
	}
	envOptions := &stepwise.MountOptions{
		RegistryName:    "grafana",
		PluginType:      stepwise.PluginTypeSecrets,
		PluginName:      "vault-plugin-secrets-grafanacloud",
		MountPathPrefix: "grafana",
	}

	roleName := "vault-stepwise-user-role"

	cred := new(string)
	stepwise.Run(t, stepwise.Case{
		Precheck:    func() { testAccPreCheck(t) },
		Environment: dockerEnvironment.NewEnvironment("grafana-cloud", envOptions),
		Steps: []stepwise.Step{
			testAddCA(t),
			testAccConfig(t),
			testAccUserRole(t, roleName),
			testAccUserRoleRead(t, roleName),
			testAccUserCredRead(t, roleName, cred),
		},
	})
}

var initSetup sync.Once

func testAccPreCheck(t *testing.T) {
	initSetup.Do(func() {
		// Ensure test variables are set
		if v := os.Getenv(envVarGrafanaCloudAPIKey); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudAPIKey))
		}
		if v := os.Getenv(envVarGrafanaCloudURL); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudURL))
		}
		if v := os.Getenv(envVarGrafanaCloudOrganisation); v == "" {
			t.Skip(fmt.Printf("%s not set", envVarGrafanaCloudOrganisation))
		}
	})
}

// testAddCA will add (if given) the CA tar to the vault container.
func testAddCA(t *testing.T) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.HelpOperation,
		Assert: func(_ *api.Secret, _ error) error {
			caPath := os.Getenv(envVarCATarPath)
			if caPath == "" {
				return nil
			}

			cli, err := client.NewClientWithOpts()
			if err != nil {
				return fmt.Errorf("failed to create new Docker CLI client: %s", err.Error())
			}

			container, err := fetchVaultContainer(cli)
			if err != nil {
				return err
			}

			if err := copyCATarToContainer(cli, caPath, container); err != nil {
				return err
			}

			return updateContainerCACertificates(cli, container)
		},
	}
}

// fetchVaultContainer will attempt to find the vault container started by stepwise.
func fetchVaultContainer(cli client.APIClient) (types.Container, error) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "name",
				Value: "test-grafana-cloud-*",
			},
		),
	})
	if err != nil {
		return types.Container{}, fmt.Errorf("failed to list containers: %s", err.Error())
	}

	for _, container := range containers {
		if strings.HasSuffix(container.Names[0], "vault-0") {
			return container, nil
		}
	}

	return types.Container{}, errors.New("could not find container 'test-grafana-cloud-*-vault-0")
}

// copyCATarToContainer will attempt to read the given file and copy it to the container.
func copyCATarToContainer(cli client.APIClient, caPath string, container types.Container) error {
	caBytes, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA tar file: %s", err.Error())
	}

	if err := cli.CopyToContainer(
		context.Background(),
		container.ID,
		"/usr/local/share/ca-certificates/",
		bytes.NewReader(caBytes),
		types.CopyToContainerOptions{}); err != nil {

		return fmt.Errorf("failed to copy file to container: %s", err.Error())
	}

	return nil
}

// updateContainerCACertificates will run the update-ca-certificates command within the container.
func updateContainerCACertificates(cli client.APIClient, container types.Container) error {
	exec, err := cli.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{
		Cmd: []string{
			"update-ca-certificates",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create container exec command: %s", err.Error())
	}

	if err := cli.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{}); err != nil {
		return fmt.Errorf("failed to run container exec command: %s", err.Error())
	}

	return nil
}

func testAccConfig(t *testing.T) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "config",
		Data: map[string]interface{}{
			"organisation": os.Getenv(envVarGrafanaCloudOrganisation),
			"key":          os.Getenv(envVarGrafanaCloudAPIKey),
			"url":          os.Getenv(envVarGrafanaCloudURL),
		},
	}
}

func testAccUserRole(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.UpdateOperation,
		Path:      "roles/" + roleName,
		Data: map[string]interface{}{
			"gc_role": "Viewer",
			"ttl":     "1m",
			"max_ttl": "5m",
		},
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.Nil(t, resp)
			return nil
		},
	}
}

func testAccUserRoleRead(t *testing.T, roleName string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "roles/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.Equal(t, "Viewer", resp.Data["gc_role"])
			return nil
		},
	}
}

func testAccUserCredRead(t *testing.T, roleName string, apiKey *string) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ReadOperation,
		Path:      "creds/" + roleName,
		Assert: func(resp *api.Secret, err error) error {
			require.Nil(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.Data["token"])
			*apiKey = resp.Data["token"].(string)
			return nil
		},
	}
}
