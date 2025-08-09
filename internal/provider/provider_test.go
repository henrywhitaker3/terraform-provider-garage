// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func providerConfig(ip string, port int, token string) string {
	return fmt.Sprintf(`
provider "garage" {
	host = "%s"
	scheme = "http"
	token = "%s"
}`, fmt.Sprintf("%s:%d", ip, port), token)
}

const (
	garageConfig = `metadata_dir = "/tmp/meta"
data_dir = "/tmp/data"
db_engine = "sqlite"

replication_factor = 1

rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "3e844293abd869741d142cad93e269b1c9ff3c41240566508554c250d7668ec0"

[s3_api]
s3_region = "garage"
api_bind_addr = "[::]:3900"
root_domain = ".s3.garage.localhost"

[s3_web]
bind_addr = "[::]:3902"
root_domain = ".web.garage.localhost"
index = "index.html"

[k2v_api]
api_bind_addr = "[::]:3904"

[admin]
api_bind_addr = "[::]:3903"
admin_token = "EVCNqzJY4StaQ7RGZ+triyhK6GCzgLNrhlqSvTMVyrI="
metrics_token = "neFhPdBSRjcrfTW4LKcnTMqJQAY6vOII+qdVQZK/Dtw="`
)

func garage(t *testing.T) (string, context.CancelFunc) {
	configPath := filepath.Join(t.TempDir(), "garage.toml")
	require.Nil(
		t,
		os.WriteFile(configPath, []byte(garageConfig), 0644),
	)
	req := testcontainers.ContainerRequest{
		Image:        "dxflrs/garage:v2.0.0",
		ExposedPorts: []string{"3900/tcp", "3901/tcp", "3902/tcp", "3903/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      configPath,
				ContainerFilePath: "/etc/garage.toml",
				FileMode:          0444,
			},
		},
		// WaitingFor: &wait.HostPortStrategy{
		// 	Port: "3903/tcp",
		// },
		WaitingFor: &wait.HTTPStrategy{
			Port: "3903",
			Path: "/",
			StatusCodeMatcher: func(status int) bool {
				return status == 400
			},
		},
	}

	ctx := context.Background()

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	require.Nil(t, err)

	rc, output, err := container.Exec(ctx, []string{"/garage", "status"})
	require.Nil(t, err)
	body, err := io.ReadAll(output)
	require.Nil(t, err)
	t.Log(string(body))
	require.Equal(t, 0, rc)

	nodeID := strings.Split(strings.Split(string(body), "\n")[4], " ")[0]
	t.Log("got node id", nodeID)

	rc, output, err = container.Exec(
		ctx,
		[]string{"/garage", "layout", "assign", "-z", "test", "-c", "1G", nodeID},
	)
	require.Nil(t, err)
	body, err = io.ReadAll(output)
	require.Nil(t, err)
	t.Log(string(body))
	require.Equal(t, 0, rc)

	rc, output, err = container.Exec(
		ctx,
		[]string{"/garage", "layout", "apply", "--version", "1"},
	)
	require.Nil(t, err)
	body, err = io.ReadAll(output)
	require.Nil(t, err)
	t.Log(string(body))
	require.Equal(t, 0, rc)

	port, err := container.MappedPort(ctx, "3903")
	require.Nil(t, err)

	return providerConfig(
			"127.0.0.1",
			port.Int(),
			"EVCNqzJY4StaQ7RGZ+triyhK6GCzgLNrhlqSvTMVyrI=",
		), func() {
			_ = container.Terminate(ctx)
		}
}

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"garage": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the scaffolding provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"garage": providerserver.NewProtocol6WithError(New("test")()),
	"echo":   echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}
