package settings

import (
	"bytes"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateNameSelection(t *testing.T) {
	// Create a mock command with template config
	cmd := &cmds.CommandDescription{
		Name: "test-command",
		TemplateConfig: &cmds.TemplateConfig{
			Templates: map[string]string{
				"default":  "Default: {{range .rows}}{{.name}}={{.value}}\n{{end}}",
				"json":     "JSON: {{range .rows}}{{.name}}:{{.value}}\n{{end}}",
				"markdown": "MD: {{range .rows}}* {{.name}}: {{.value}}\n{{end}}",
			},
			DefaultTemplate: "default",
		},
	}

	// Create test cases
	tests := []struct {
		name         string
		templateName string
		template     string
		expected     string
	}{
		{
			name:         "Default template when no name specified",
			templateName: "",
			template:     "",
			expected:     "Default: name1=100\nname2=200\n",
		},
		{
			name:         "Named template selection",
			templateName: "markdown",
			template:     "",
			expected:     "MD: * name1: 100\n* name2: 200\n",
		},
		{
			name:         "Custom template overrides named templates",
			templateName: "markdown",
			template:     "Custom: {{range .rows}}{{.name}}->{{.value}}\n{{end}}",
			expected:     "Custom: name1->100\nname2->200\n",
		},
		{
			name:         "Invalid template name falls back to default",
			templateName: "non-existent",
			template:     "",
			expected:     "Default: name1=100\nname2=200\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a parsed layer with the necessary template settings
			parsedLayer := &layers.ParsedLayer{
				Layer:      &TemplateParameterLayer{},
				Parameters: parameters.NewParsedParameters(),
			}

			// Set template name if specified
			if tt.templateName != "" {
				err := parsedLayer.Parameters.Set("template-name", tt.templateName)
				require.NoError(t, err)
			}

			// Set template if specified
			if tt.template != "" {
				err := parsedLayer.Parameters.Set("template", tt.template)
				require.NoError(t, err)
			}

			// Set output to template
			err := parsedLayer.Parameters.Set("output", "template")
			require.NoError(t, err)

			// Create output formatter settings
			outputSettings, err := NewOutputFormatterSettings(parsedLayer)
			require.NoError(t, err)

			// Create output formatter
			of, err := outputSettings.CreateTableOutputFormatter(cmd)
			require.NoError(t, err)

			// Set up processor and add some test data
			processor := middlewares.NewTableProcessor()
			processor.SetOutputFormatter(of)

			row1 := types.NewRow()
			row1.Set("name", "name1")
			row1.Set("value", 100)
			processor.AddRow(row1)

			row2 := types.NewRow()
			row2.Set("name", "name2")
			row2.Set("value", 200)
			processor.AddRow(row2)

			// Capture output
			buf := &bytes.Buffer{}
			_, err = of.WriteTable(buf, processor.GetTable())
			require.NoError(t, err)

			// Check output
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}