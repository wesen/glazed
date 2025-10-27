package middlewares

import (
	"bytes"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// yamlConfigOverrideTemplateData contains information passed to the override filename template.
type yamlConfigOverrideTemplateData struct {
	// Path is the absolute or relative path that was provided to the middleware.
	Path string
	// Dir is the directory portion of Path, without a trailing separator. May be empty.
	Dir string
	// Base is the last element of Path (including the extension).
	Base string
	// Name is Base without the extension.
	Name string
	// Ext is the file extension for Path (including the leading dot).
	Ext string
	// Username is the resolved username that should be injected into the template.
	Username string
}

// YAMLConfigOption customises how GatherFlagsFromYAMLConfig behaves.
type YAMLConfigOption func(*yamlConfigConfig)

type yamlConfigConfig struct {
	overrideTemplate string
	username         string
	parseOptions     []parameters.ParseStepOption
	required         bool
}

const defaultOverrideTemplate = `{{ if .Dir }}{{ .Dir }}/{{ end }}{{ .Name }}.{{ .Username }}{{ .Ext }}`

// WithYAMLConfigOverrideTemplate allows providing a text/template that is used to build
// the override filename. The template receives yamlConfigOverrideTemplateData.
func WithYAMLConfigOverrideTemplate(tmpl string) YAMLConfigOption {
	return func(c *yamlConfigConfig) {
		c.overrideTemplate = tmpl
	}
}

// WithYAMLConfigUsername sets the username that should be injected into the override template.
// Useful for tests or when the operating system user must be overridden.
func WithYAMLConfigUsername(username string) YAMLConfigOption {
	return func(c *yamlConfigConfig) {
		c.username = username
	}
}

// WithYAMLConfigParseOptions appends parse step options that will be applied when parameters
// are populated from the YAML files.
func WithYAMLConfigParseOptions(options ...parameters.ParseStepOption) YAMLConfigOption {
	return func(c *yamlConfigConfig) {
		c.parseOptions = append(c.parseOptions, options...)
	}
}

// WithYAMLConfigRequired ensures the base YAML config must exist; missing files trigger an error.
func WithYAMLConfigRequired(required bool) YAMLConfigOption {
	return func(c *yamlConfigConfig) {
		c.required = required
	}
}

// GatherFlagsFromYAMLConfig loads parameter values from a YAML configuration file and, when present,
// from a username-specific override file. The override filename is built from a configurable
// template that receives the base file information and the resolved username.
//
// The YAML file must follow the structure:
//
//	layerSlug:
//	  parameter-name: value
//
// Values are merged into the parsed layers using updateFromMap semantics, meaning overrides replace
// previously populated values.
func GatherFlagsFromYAMLConfig(configFile string, options ...YAMLConfigOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			cfg := &yamlConfigConfig{
				overrideTemplate: defaultOverrideTemplate,
			}
			for _, opt := range options {
				opt(cfg)
			}

			if cfg.username == "" {
				cfg.username = resolveCurrentUsername()
			}

			baseMap, err := loadYAMLConfigMap(configFile)
			switch {
			case err == nil:
				err = applyConfigMap(layers_, parsedLayers, baseMap, cfg.parseOptions, map[string]interface{}{
					"configFile": configFile,
					"kind":       "base",
				})
				if err != nil {
					return err
				}
			case os.IsNotExist(err):
				if cfg.required {
					return errors.Wrapf(err, "config file %s does not exist", configFile)
				}
			case err != nil:
				return errors.Wrapf(err, "failed to load config file %s", configFile)
			}

			if cfg.overrideTemplate == "" || cfg.username == "" {
				return nil
			}

			overridePath, err := buildOverridePath(configFile, cfg.overrideTemplate, cfg.username)
			if err != nil {
				return errors.Wrap(err, "failed to build override config path")
			}

			overrideMap, err := loadYAMLConfigMap(overridePath)
			switch {
			case err == nil:
				err = applyConfigMap(layers_, parsedLayers, overrideMap, cfg.parseOptions, map[string]interface{}{
					"configFile": overridePath,
					"kind":       "override",
					"username":   cfg.username,
				})
				if err != nil {
					return err
				}
			case os.IsNotExist(err):
				// Missing overrides are fine.
				return nil
			case err != nil:
				return errors.Wrapf(err, "failed to load override config file %s", overridePath)
			}

			return nil
		}
	}
}

func applyConfigMap(
	layers_ *layers.ParameterLayers,
	parsedLayers *layers.ParsedLayers,
	config map[string]map[string]interface{},
	parseOptions []parameters.ParseStepOption,
	metadata map[string]interface{},
) error {
	if len(config) == 0 {
		return nil
	}

	options := append([]parameters.ParseStepOption{}, parseOptions...)
	if metadata != nil {
		options = append(options, parameters.WithParseStepMetadata(metadata))
	}

	return updateFromMap(layers_, parsedLayers, config, options...)
}

func loadYAMLConfigMap(path string) (map[string]map[string]interface{}, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(bytes) == 0 {
		return map[string]map[string]interface{}{}, nil
	}

	result := map[string]map[string]interface{}{}
	if err := yaml.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	if result == nil {
		return map[string]map[string]interface{}{}, nil
	}

	return result, nil
}

func buildOverridePath(path string, tmpl string, username string) (string, error) {
	dir := filepath.Dir(path)
	if dir == "." {
		dir = ""
	}

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	t, err := template.New("yaml-config-override").Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := yamlConfigOverrideTemplateData{
		Path:     path,
		Dir:      dir,
		Base:     base,
		Name:     name,
		Ext:      ext,
		Username: username,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func resolveCurrentUsername() string {
	if u, err := user.Current(); err == nil {
		username := sanitizeUsername(u.Username)
		if username != "" {
			return username
		}
	}

	if username := sanitizeUsername(os.Getenv("USER")); username != "" {
		return username
	}

	if username := sanitizeUsername(os.Getenv("USERNAME")); username != "" {
		return username
	}

	return ""
}

func sanitizeUsername(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}

	if idx := strings.LastIndexAny(username, `/\`); idx >= 0 && idx < len(username)-1 {
		username = username[idx+1:]
	}

	return username
}
