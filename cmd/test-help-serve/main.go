package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed doc/*
var docFS embed.FS

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create a help system
	hs := help.NewHelpSystem()
	
	// Load help sections from embedded filesystem
	err := hs.LoadSectionsFromFS(docFS, "doc")
	if err != nil {
		log.Error().Err(err).Msg("Failed to load help sections")
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "test-help",
		Short: "Test application for help system web UI",
		Long:  "A test application to demonstrate the glazed help system web UI functionality",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to the test help application!")
			fmt.Println("Try running: test-help help serve")
		},
	}

	// Add some subcommands for testing
	rootCmd.AddCommand(&cobra.Command{
		Use:   "process",
		Short: "Process data files",
		Long:  "Process various data files with different options",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Processing data...")
		},
	})

	processCmd := rootCmd.Commands()[0]
	processCmd.Flags().String("input", "", "Input file path")
	processCmd.Flags().String("output", "", "Output file path") 
	processCmd.Flags().Bool("verbose", false, "Verbose output")

	// Setup help system with the root command
	hs.SetupCobraRootCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}
