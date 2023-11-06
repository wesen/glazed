package pluginpackage

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"os"
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

//nolint:deacode,unused
var mockDescriptor = Descriptor{
	// ... [Same mockDescriptor content as before]
}

//nolint:deadcode,unused
var getDescriptorCmd = &cobra.Command{
	Use:   "get-descriptor",
	Short: "Retrieve the plugin descriptor",
	Run: func(cmd *cobra.Command, args []string) {
		err := mockDescriptor.PrintAsYAML(os.Stdout)
		cobra.CheckErr(err)
		_ = mockDescriptor
	},
}

//nolint:deadcode,unused
func main() {
	rootCmd := &cobra.Command{Use: "plugin"}
	rootCmd.AddCommand(getDescriptorCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
