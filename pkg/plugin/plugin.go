package plugin

import (
	"gopkg.in/yaml.v3"
	"io"
)

type Descriptor struct {
	Name        string      `yaml:"name"`
	Version     string      `yaml:"version"`
	Description string      `yaml:"description"`
	Author      string      `yaml:"author"`
	Interfaces  []Interface `yaml:"interfaces"`
	Handshake   Handshake   `yaml:"handshake"`
	Options     []Option    `yaml:"options"`
}

type Interface struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type Handshake struct {
	ProtocolVersion  int    `yaml:"protocol_version"`
	MagicCookieKey   string `yaml:"magic_cookie_key"`
	MagicCookieValue string `yaml:"magic_cookie_value"`
}

type Option struct {
	Key         string `yaml:"key"`
	Default     string `yaml:"default"`
	Description string `yaml:"description"`
}

// PrintAsYAML prints the descriptor as YAML format to the given writer.
func (d *Descriptor) PrintAsYAML(w io.Writer) error {
	data, err := yaml.Marshal(d)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
