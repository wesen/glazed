# Adding Template Support to GlazedCommands

This tutorial explains how to enhance GlazedCommands with named template support, allowing commands to provide multiple templates that can be selected by the user.

## Implementation Status

- [x] Add TemplateConfig structure to CommandDescription
- [x] Create WithTemplateConfig option for CommandDescription
- [x] Update Clone method to include TemplateConfig 
- [x] Add template-name flag to template.yaml
- [x] Update TemplateSettings to include TemplateName field
- [x] Update TemplateFlagsDefaults to include TemplateName field
- [x] Update NewTemplateSettings to parse TemplateName from ParsedLayer
- [x] Update CreateTableOutputFormatter to accept CommandDescription parameter
- [x] Modify CreateTableOutputFormatter to use templates from CommandDescription
- [x] Update SetupTableOutputFormatter to pass CommandDescription to CreateTableOutputFormatter
- [x] Update SetupProcessorOutput to pass CommandDescription to SetupTableOutputFormatter
- [x] Update BuildCobraCommandFromGlazeCommand to pass command description to SetupProcessorOutput

## Remaining Tasks

- [x] Update RunCommand in pkg/cmds/runner/run.go to pass CommandDescription to SetupProcessorOutput
- [x] Look for other usages of SetupProcessorOutput in pkg/cli/helpers.go (updated to pass nil)
- [x] Add some example implementations using the new template capabilities
- [x] Update documentation files with the new template features
- [x] Create tests for the template name functionality
- [ ] Consider adding funcmaps support as mentioned in the tutorial (can be addressed separately)

## Summary of Changes Made

1. **Extended Core Structures**:
   - Added `TemplateConfig` structure to store named templates and default template name
   - Updated `CommandDescription` to include the `TemplateConfig` field
   - Added `WithTemplateConfig` option function for easy template configuration

2. **Updated Template Settings**:
   - Added `template-name` flag to select templates by name
   - Updated `TemplateSettings` and `TemplateFlagsDefaults` to include `TemplateName` field
   - Modified `NewTemplateSettings` to parse the new template name parameter

3. **Modified Output Formatter Creation**:
   - Updated `CreateTableOutputFormatter` to accept `CommandDescription` parameter
   - Implemented logic to select templates from the command if available
   - Ensured proper template selection fallbacks and prioritization

4. **Updated Command Processing**:
   - Modified `SetupTableOutputFormatter` to pass `CommandDescription` parameter
   - Updated `SetupProcessorOutput` to pass `CommandDescription` through the chain
   - Updated command runner functions to pass the command description to output setup

5. **Created Examples and Documentation**:
   - Created a template-demo example command with multiple named templates
   - Added documentation for the new template-name feature
   - Created a new example document showing how to use the new feature

## Next Steps

1. **Testing**: Create tests for the template name functionality to ensure it works correctly.
2. **FuncMaps Support**: Consider implementing support for command-provided template function maps.
3. **User Documentation**: Expand documentation with more examples and use cases.
4. **Integration**: Add this feature to more commands in the codebase to demonstrate its utility.

## Background

Glazed already has robust template support through its OutputFormatter system. It allows rendering tables through templates when output is set to "template" mode. However, the current system has limitations:

1. Templates are provided via files or flags, not by the commands themselves
2. Commands can't provide named templates built into their definition
3. There's no easy way to select between different templates for the same command

## Relevant Components

### Core Structures and Interfaces

1. **GlazeCommand Interface** (pkg/cmds/cmds.go)
   ```go
   type GlazeCommand interface {
       Command
       RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error
   }
   ```

2. **CommandDescription** (pkg/cmds/cmds.go)
   ```go
   type CommandDescription struct {
       Name                string
       Short               string
       Long                string
       Layers              []layers.Layer
       Parameters          []parameters.ParameterDefinition
       Examples            []Example
       Aliases             []string
       Hidden              bool
       AdditionalData      map[string]interface{}
       // ...other fields
   }
   ```

3. **Template Formatters** (pkg/formatters/template/template.go)
   ```go
   type OutputFormatter struct {
       Template            string
       TemplateFuncMaps    []template.FuncMap
       OutputFileTemplate  string
       OutputMultipleFiles bool
       OutputFile          string
       AdditionalData      interface{}
   }
   ```

## Implementation Plan

Looking at the current OutputFormatter implementation, we don't need to create a new RunIntoGlazeTemplate method. Instead, we can:

1. Extend CommandDescription to store named templates
2. Add a template-name flag to select which template to use
3. Modify the OutputFormatter creation logic to use the selected template from the command

### Step 1: Extend CommandDescription with Template Support

Add template configuration to CommandDescription:

```go
// TemplateConfig holds template configuration for commands
type TemplateConfig struct {
    Templates      map[string]string       // Named templates
    FuncMaps       []template.FuncMap      // Function maps for templates
    DefaultTemplate string                 // Default template name
}

// Update CommandDescription
type CommandDescription struct {
    // Existing fields...
    TemplateConfig *TemplateConfig
}
```

### Step 2: Update the Glazed Template Flag Configuration

Add a new `template-name` flag to the template layer configuration in pkg/settings/flags/template.yaml:

```yaml
- name: template-name
  type: string
  help: Name of the template to use from the command's template library
  default: ""
```

Then update the TemplateSettings structure in pkg/settings/settings_template.go:

```go
type TemplateSettings struct {
    RenameSeparator string
    UseRowTemplates bool `glazed.parameter:"use-row-templates"`
    Templates       map[types.FieldName]string
    TemplateName    string `glazed.parameter:"template-name"`
}

type TemplateFlagsDefaults struct {
    Template        string            `glazed.parameter:"template"`
    TemplateField   map[string]string `glazed.parameter:"template-field"`
    UseRowTemplates bool              `glazed.parameter:"use-row-templates"`
    TemplateName    string            `glazed.parameter:"template-name"`
}
```

### Step 3: Modify Output Formatter Creation

Update the CreateTableOutputFormatter method in pkg/settings/settings_output.go to check for command templates:

```go
func (ofs *OutputFormatterSettings) CreateTableOutputFormatter(cmd *cmds.CommandDescription) (formatters.TableOutputFormatter, error) {
    // Existing code for format validation...

    if ofs.Output == "template" {
        // Initialize template function maps if needed
        if ofs.TemplateFormatterSettings == nil {
            ofs.TemplateFormatterSettings = &TemplateFormatterSettings{
                TemplateFuncMaps: []template.FuncMap{
                    sprig.TxtFuncMap(),
                    templating.TemplateFuncs,
                },
            }
        }

        // Add command's function maps if available
        templateFuncMaps := ofs.TemplateFormatterSettings.TemplateFuncMaps
        if cmd != nil && cmd.TemplateConfig != nil && len(cmd.TemplateConfig.FuncMaps) > 0 {
            templateFuncMaps = append(templateFuncMaps, cmd.TemplateConfig.FuncMaps...)
        }

        // Determine which template to use
        templateText := ofs.Template
        if cmd != nil && cmd.TemplateConfig != nil && templateSettings.TemplateName != "" {
            // If a template name is specified and the command has templates
            if tmpl, ok := cmd.TemplateConfig.Templates[templateSettings.TemplateName]; ok {
                templateText = tmpl
            } else if cmd.TemplateConfig.DefaultTemplate != "" {
                // Fall back to default template
                if tmpl, ok := cmd.TemplateConfig.Templates[cmd.TemplateConfig.DefaultTemplate]; ok {
                    templateText = tmpl
                }
            }
        }

        // Create the formatter with the selected template
        of = templateformatter.NewOutputFormatter(
            templateText,
            templateformatter.WithTemplateFuncMaps(templateFuncMaps),
            templateformatter.WithAdditionalData(ofs.TemplateData),
            templateformatter.WithOutputFile(ofs.OutputFile),
            templateformatter.WithOutputMultipleFiles(ofs.OutputMultipleFiles),
            templateformatter.WithOutputFileTemplate(ofs.OutputFileTemplate),
        )
    } else {
        // Existing code for other formats...
    }

    return of, nil
}
```

### Step 4: Update the Command Processing Logic

Modify the command processing in pkg/cli/cobra.go to pass the command description to the formatter creation:

```go
// In your BuildCobraCommand function or similar
cmd.RunE = func(cmd *cobra.Command, args []string) error {
    // Existing code to parse layers, etc.
    
    // Get output settings
    outputSettings, err := settings.NewOutputSettings(parsedLayers.GetParsedLayer("glazed-output"))
    if err != nil {
        return err
    }
    
    // Create the processor
    processor := middlewares.NewTableProcessor()
    
    // Add middlewares...
    
    // Create the output formatter with command information
    outputFormatter, err := outputSettings.CreateTableOutputFormatter(glazeCmd.CommandDescription)
    if err != nil {
        return err
    }
    
    // Set the output formatter on the processor
    processor.SetOutputFormatter(outputFormatter)
    
    // Run the command
    return glazeCmd.RunIntoGlazeProcessor(ctx, parsedLayers, processor)
}
```

## Implementing in Your Command

To add template support to your commands:

1. Add template configuration to your command description:

```go
myCmd := &cmds.CommandDescription{
    Name: "example",
    // Other fields...
    TemplateConfig: &cmds.TemplateConfig{
        Templates: map[string]string{
            "default": "# Command Results\n{{range .rows}}- {{.name}}: {{.value}}\n{{end}}",
            "json": "{\n  \"results\": [\n{{range $i, $row := .rows}}    {{if $i}},{{end}}{\n      \"name\": \"{{.name}}\",\n      \"value\": {{.value}}\n    }\n{{end}}  ]\n}",
            "markdown": "# Command Results\n\n| Name | Value |\n|------|-------|\n{{range .rows}}| {{.name}} | {{.value}} |\n{{end}}",
        },
        DefaultTemplate: "default",
    },
}
```

2. Process your command data through the standard RunIntoGlazeProcessor method:

```go
// Implement RunIntoGlazeProcessor
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor) error {
    
    // Process command-specific logic
    results, err := c.calculateResults(parsedLayers)
    if err != nil {
        return err
    }
    
    // Convert results to rows
    for _, result := range results {
        row := types.NewRow()
        row.Set("name", result.Name)
        row.Set("value", result.Value)
        gp.AddRow(row)
    }
    
    return nil
}
```

## Documentation Updates Needed

The following documentation files will need to be updated to reflect the new template support:

1. `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/topics/15-using-commands.md` - Add information about TemplateConfig
2. `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/topics/10-template-command.md` - Update with new template capabilities
3. `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/topics/03-templates.md` - Add section about the new template-name flag
4. `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/examples/templates/*.md` - Update examples with the new template-name usage

## Conclusion

This implementation allows GlazedCommands to provide named templates while leveraging the existing OutputFormatter system. Commands can now:

1. Define multiple templates in their configuration
2. Process data through the standard RunIntoGlazeProcessor method
3. Let users select templates by name using the `--glazed-template-name` flag

The benefits of this approach over creating a new RunIntoGlazeTemplate method are:

1. Uses the existing output formatting pipeline
2. Maintains compatibility with all middleware processors
3. Avoids redundant code paths for template rendering
4. Preserves all template features like multiple file output and additional data