package middlewares

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherFlagsFromYAMLConfigLoadsBaseAndOverride(t *testing.T) {
	t.Setenv("USER", "test-user")

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	overrideFile := filepath.Join(tempDir, "config.test-user.yaml")

	baseConfig := `
test:
  base-only: "from-base"
  override-me: "base-value"
`
	overrideConfig := `
test:
  override-me: "override-value"
  override-only: 42
`

	require.NoError(t, os.WriteFile(configFile, []byte(baseConfig), 0644))
	require.NoError(t, os.WriteFile(overrideFile, []byte(overrideConfig), 0644))

	baseOnlyParam := &parameters.ParameterDefinition{
		Name: "base-only",
		Type: parameters.ParameterTypeString,
	}
	overrideParam := &parameters.ParameterDefinition{
		Name: "override-me",
		Type: parameters.ParameterTypeString,
	}
	overrideOnlyParam := &parameters.ParameterDefinition{
		Name: "override-only",
		Type: parameters.ParameterTypeInteger,
	}

	layer, err := layers.NewParameterLayer("test", "Test layer", layers.WithParameterDefinitions(
		baseOnlyParam,
		overrideParam,
		overrideOnlyParam,
	))
	require.NoError(t, err)

	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("test", layer)

	parsedLayers := layers.NewParsedLayers()

	middleware := GatherFlagsFromYAMLConfig(
		configFile,
		parameters.WithParseStepSource("yaml-config"),
	)

	handler := middleware(Identity)
	require.NoError(t, handler(parameterLayers, parsedLayers))

	parsedLayer, ok := parsedLayers.Get("test")
	require.True(t, ok)

	baseOnlyValue, ok := parsedLayer.Parameters.Get("base-only")
	require.True(t, ok)
	assert.Equal(t, "from-base", baseOnlyValue.Value)

	overrideValue, ok := parsedLayer.Parameters.Get("override-me")
	require.True(t, ok)
	assert.Equal(t, "override-value", overrideValue.Value)

	overrideOnlyValue, ok := parsedLayer.Parameters.Get("override-only")
	require.True(t, ok)
	assert.Equal(t, 42, overrideOnlyValue.Value)
}

func TestGatherFlagsFromYAMLConfigMissingOverride(t *testing.T) {
	t.Setenv("USER", "missing-user")

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	baseConfig := `
test:
  base-param: "value"
`
	require.NoError(t, os.WriteFile(configFile, []byte(baseConfig), 0644))

	baseParam := &parameters.ParameterDefinition{
		Name: "base-param",
		Type: parameters.ParameterTypeString,
	}

	layer, err := layers.NewParameterLayer("test", "Test layer", layers.WithParameterDefinitions(baseParam))
	require.NoError(t, err)

	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("test", layer)

	parsedLayers := layers.NewParsedLayers()

	middleware := GatherFlagsFromYAMLConfig(configFile)

	handler := middleware(Identity)
	require.NoError(t, handler(parameterLayers, parsedLayers))

	parsedLayer, ok := parsedLayers.Get("test")
	require.True(t, ok)

	baseValue, ok := parsedLayer.Parameters.Get("base-param")
	require.True(t, ok)
	assert.Equal(t, "value", baseValue.Value)
}

func TestGatherFlagsFromYAMLConfigMissingBaseReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	parameterLayers := layers.NewParameterLayers()
	layer, err := layers.NewParameterLayer("test", "Test layer")
	require.NoError(t, err)
	parameterLayers.Set("test", layer)

	parsedLayers := layers.NewParsedLayers()

	middleware := GatherFlagsFromYAMLConfig(configFile)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}
