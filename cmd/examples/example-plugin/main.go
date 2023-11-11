package main

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/plugin"
	"github.com/spf13/cobra"
	"os"
)

var mockDescriptor = plugin.Descriptor{
	Name:        "database_plugin",
	Version:     "v1.0.0",
	Description: "A sample database plugin",
	Author:      "OpenAI",
	Interfaces: []plugin.Interface{
		{
			Name:        "logger",
			Version:     "v1.0",
			Description: "Provides logging capabilities.",
		},
	},
	Handshake: plugin.Handshake{
		ProtocolVersion:  1,
		MagicCookieKey:   "MagicKey",
		MagicCookieValue: "MagicValue",
	},
	Options: []plugin.Option{
		{
			Key:         "log_level",
			Default:     "info",
			Description: "Defines the log level.",
		},
	},
}

var getDescriptorCmd = &cobra.Command{
	Use:   "get-descriptor",
	Short: "Retrieve the plugin descriptor",
	Run: func(cmd *cobra.Command, args []string) {
		err := mockDescriptor.PrintAsYAML(os.Stdout)
		cobra.CheckErr(err)
		_ = mockDescriptor
	},
}

func main() {
	rootCmd := &cobra.Command{Use: "plugin"}
	rootCmd.AddCommand(getDescriptorCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
