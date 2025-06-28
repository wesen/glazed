package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// UserListCommand demonstrates a GlazeCommand that outputs structured data
type UserListCommand struct {
	*cmds.CommandDescription
}

// Settings struct for clean parameter access using glazed.parameter tags
type UserListSettings struct {
	Count   int    `glazed.parameter:"count"`
	Search  string `glazed.parameter:"search"`
	Verbose bool   `glazed.parameter:"verbose"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &UserListCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *UserListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers using struct tags
	s := &UserListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	if s.Verbose {
		fmt.Printf("Fetching users with count=%d, search='%s'\n", s.Count, s.Search)
	}

	// Sample data (in a real app, this would come from a database)
	users := []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Active   bool   `json:"active"`
	}{
		{1, "alice", "alice@example.com", "admin", true},
		{2, "bob", "bob@example.com", "user", true},
		{3, "carol", "carol@example.com", "editor", false},
		{4, "dave", "dave@example.com", "user", true},
		{5, "eve", "eve@example.com", "moderator", true},
	}

	// Apply search filter if provided
	var filteredUsers []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
		Active   bool   `json:"active"`
	}

	for _, user := range users {
		if s.Search == "" ||
			user.Username == s.Search ||
			user.Role == s.Search ||
			user.Email == s.Search {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Apply count limit
	if s.Count > 0 && s.Count < len(filteredUsers) {
		filteredUsers = filteredUsers[:s.Count]
	}

	// Output as structured rows - glazed handles all the formatting
	for _, user := range filteredUsers {
		row := types.NewRowFromStruct(&user, true)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewUserListCommand creates a new UserListCommand with all its parameters and layers
func NewUserListCommand() (*UserListCommand, error) {
	// Create the standard Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description with parameters
	cmdDesc := cmds.NewCommandDescription(
		"list-users",
		cmds.WithShort("List users with optional search filtering"),
		cmds.WithLong(`
List all users in the system with optional search filtering by username, email, or role.
Supports various output formats through standard Glazed flags.

Examples:
  list-users --count=3
  list-users --search=admin --output=json
  list-users --search=alice --fields=username,email --output=csv
  list-users --verbose --output=yaml
		`),
		// Define command flags (optional parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"count",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of users to return (0 = all)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"search",
				parameters.ParameterTypeString,
				parameters.WithHelp("Search users by username, email, or role"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose output"),
				parameters.WithDefault(false),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer, // Provides --output, --fields, --filter, etc.
		),
	)

	return &UserListCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "simple-command",
		Short: "Example glazed application",
		Long: `
This is a simple example of a glazed command-line application.
It demonstrates how to create commands with structured data output,
parameter layers, and rich help integration.
		`,
	}

	// Create and add the user list command
	userCmd, err := NewUserListCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating user command: %v\n", err)
		os.Exit(1)
	}

	// Convert glazed command to Cobra command
	userCobraCmd, err := cli.BuildCobraCommandFromCommand(userCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building Cobra command: %v\n", err)
		os.Exit(1)
	}

	// Add to root command
	rootCmd.AddCommand(userCobraCmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
