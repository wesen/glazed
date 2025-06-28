package main

import (
	"os"
	"testing"

	"tutorial-04/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductionProcessorCommand_Integration(t *testing.T) {
	// Create a temporary log file
	logContent := `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
192.168.1.1 - - [10/Oct/2000:13:55:37 -0700] "POST /form HTTP/1.0" 404 1234
10.0.0.1 - - [10/Oct/2000:13:55:38 -0700] "GET /index.html HTTP/1.0" 200 5432
`

	tmpFile, err := os.CreateTemp("", "test-access-*.log")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(logContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test the command
	cmd, err := NewProductionProcessorCommand()
	require.NoError(t, err)

	// This would require more setup to test the full command
	// For now, just test that it creates successfully
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.CommandDescription)
}

func TestConfigLoading(t *testing.T) {
	// Create a temporary config file
	configContent := `
app:
  name: test-processor
  version: 1.0.0
  environment: test

logging:
  level: debug
  format: console

server:
  host: localhost
  port: 8080
  metrics_enabled: true
  metrics_port: 9090

storage:
  type: local
  path: ./test-data
  retention_days: 7
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load the config
	cfg, err := config.LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "test-processor", cfg.App.Name)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, 9090, cfg.Server.MetricsPort)
	assert.Equal(t, 7, cfg.Storage.Retention)
}
