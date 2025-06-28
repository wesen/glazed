package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// FileInfoCommand demonstrates a GlazeCommand for file information
type FileInfoCommand struct {
	*cmds.CommandDescription
}

type FileInfoSettings struct {
	ShowHidden bool `glazed.parameter:"show-hidden"`
	Recursive  bool `glazed.parameter:"recursive"`
	SizeFilter int  `glazed.parameter:"size-filter"`
}

var _ cmds.GlazeCommand = &FileInfoCommand{}

func (c *FileInfoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &FileInfoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Get file paths from arguments
	filesParam, ok := parsedLayers.GetParameter(layers.DefaultSlug, "files")
	if !ok {
		return fmt.Errorf("files parameter not found")
	}

	paths, ok := filesParam.Value.([]string)
	if !ok {
		return fmt.Errorf("invalid file paths")
	}

	for _, path := range paths {
		if err := c.processPath(ctx, gp, path, settings); err != nil {
			// Log error but continue with other files
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", path, err)
			continue
		}
	}

	return nil
}

func (c *FileInfoCommand) processPath(ctx context.Context, gp middlewares.Processor, path string, settings *FileInfoSettings) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// If it's a directory and recursive is enabled, walk the directory
	if info.IsDir() && settings.Recursive {
		return filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip hidden files unless requested
			if !settings.ShowHidden && isHidden(fileInfo.Name()) {
				if fileInfo.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Apply size filter
			if settings.SizeFilter > 0 && fileInfo.Size() < int64(settings.SizeFilter) {
				return nil
			}

			return c.addFileInfo(ctx, gp, filePath, fileInfo)
		})
	}

	// Single file
	if !settings.ShowHidden && isHidden(info.Name()) {
		return nil
	}

	if settings.SizeFilter > 0 && info.Size() < int64(settings.SizeFilter) {
		return nil
	}

	return c.addFileInfo(ctx, gp, path, info)
}

func (c *FileInfoCommand) addFileInfo(ctx context.Context, gp middlewares.Processor, path string, info os.FileInfo) error {
	// Get file type
	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	} else if info.Mode()&os.ModeSymlink != 0 {
		fileType = "symlink"
	} else if info.Mode()&os.ModeDevice != 0 {
		fileType = "device"
	}

	// Format permissions
	permissions := info.Mode().Perm().String()

	// Calculate size in human-readable format
	sizeStr := formatSize(info.Size())

	row := types.NewRow(
		types.MRP("path", path),
		types.MRP("name", info.Name()),
		types.MRP("type", fileType),
		types.MRP("size_bytes", info.Size()),
		types.MRP("size_human", sizeStr),
		types.MRP("permissions", permissions),
		types.MRP("mode", info.Mode().String()),
		types.MRP("modified", info.ModTime().Format(time.RFC3339)),
		types.MRP("modified_unix", info.ModTime().Unix()),
		types.MRP("is_dir", info.IsDir()),
	)

	return gp.AddRow(ctx, row)
}

func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func NewFileInfoCommand() (*FileInfoCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"file-info",
		cmds.WithShort("Get information about files and directories"),
		cmds.WithLong(`
Get detailed information about files and directories including size,
permissions, modification time, and type.

Supports recursive directory processing and filtering options.

Examples:
  file-info README.md
  file-info --recursive --show-hidden /etc
  file-info --size-filter 1024 *.log --output json
  file-info dir1 dir2 file.txt --fields name,size_human,modified
		`),

		cmds.WithArguments(
			parameters.NewParameterDefinition(
				"files",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Files or directories to analyze"),
				parameters.WithRequired(true),
			),
		),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"show-hidden",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include hidden files and directories"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"recursive",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Process directories recursively"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"size-filter",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Only show files larger than this many bytes"),
				parameters.WithDefault(0),
			),
		),

		cmds.WithLayersList(glazedLayer),
	)

	return &FileInfoCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "file-info-glaze",
		Short: "File information tool using GlazeCommand",
	}

	fileInfoCmd, err := NewFileInfoCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(fileInfoCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
