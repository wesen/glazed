package middlewares

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// GatherFlagsFromYAMLConfig loads parameter values from a YAML configuration file
// and, when available, a username-specific override file in the format
// <name>.<username><ext>. Values from the override replace values from the base config.
//
// The YAML file must follow the structure:
//
//	layerSlug:
//	  parameter-name: value
//
// parseOptions are applied to each parsed parameter to annotate the parse steps.
func GatherFlagsFromYAMLConfig(configFile string, parseOptions ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			baseMap, err := loadYAMLConfigMap(configFile)
			if err != nil {
				if os.IsNotExist(err) {
					return errors.Wrapf(err, "config file %s does not exist", configFile)
				}
				return errors.Wrapf(err, "failed to load config file %s", configFile)
			}

			err = applyConfigMap(layers_, parsedLayers, baseMap, parseOptions, map[string]interface{}{
				"configFile": configFile,
				"kind":       "base",
			})
			if err != nil {
				return err
			}

			username := resolveCurrentUsername()
			if username == "" {
				return nil
			}

			overridePath := overrideConfigPath(configFile, username)
			overrideMap, err := loadYAMLConfigMap(overridePath)
			switch {
			case err == nil:
				return applyConfigMap(layers_, parsedLayers, overrideMap, parseOptions, map[string]interface{}{
					"configFile": overridePath,
					"kind":       "override",
					"username":   username,
				})
			case os.IsNotExist(err):
				return nil
			default:
				return errors.Wrapf(err, "failed to load override config file %s", overridePath)
			}
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

func overrideConfigPath(path string, username string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	filename := name + "." + username + ext
	if dir == "." {
		return filename
	}

	return filepath.Join(dir, filename)
}

func resolveCurrentUsername() string {
	if username := sanitizeUsername(os.Getenv("GLAZED_USERNAME")); username != "" {
		return username
	}

	if username := sanitizeUsername(os.Getenv("USER")); username != "" {
		return username
	}

	if username := sanitizeUsername(os.Getenv("USERNAME")); username != "" {
		return username
	}

	if u, err := user.Current(); err == nil {
		username := sanitizeUsername(u.Username)
		if username != "" {
			return username
		}
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
