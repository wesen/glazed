package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/templates/html"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "help-system",
	Short: "Showcase HTML help templates",
	Long:  "A demonstration of the HTML help templates in the glazed help system",
}

func main() {
	helpSystem := help.NewHelpSystem()
	err := doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	// Add commands
	rootCmd.AddCommand(
		createTopicsCommand(helpSystem),
		createSearchCommand(helpSystem),
		createViewCommand(helpSystem),
		createExamplesCommand(helpSystem),
		createCommandCommand(helpSystem),
		createServeCommand(helpSystem),
	)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}

func createTopicsCommand(hs *help.HelpSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "topics",
		Short: "List all topics in HTML format",
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := cmd.Flags().GetString("output")
			withStyle, _ := cmd.Flags().GetBool("style")

			// Create a query that returns all topics
			query := help.NewSectionQuery().ReturnTopics()
			data, _ := hs.ComputeRenderData(query)
			helpData := data["Help"].(*help.HelpPage)

			// Convert to HTML types
			topics := make([]html.Topic, 0)
			for _, s := range helpData.AllGeneralTopics {
				topics = append(topics, html.Topic{
					Title:   s.Title,
					Short:   s.Short,
					Content: s.Content,
					Slug:    s.Slug,
				})
			}

			htmlHelp := html.Help{
				DefaultGeneralTopics: topics,
			}

			renderToOutput(html.HelpShortSectionList(htmlHelp, "", ""), output, withStyle)
		},
	}

	addCommonFlags(cmd)
	return cmd
}

func createSearchCommand(hs *help.HelpSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search help content and show results in HTML",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := cmd.Flags().GetString("output")
			withStyle, _ := cmd.Flags().GetBool("style")
			typeFilter, _ := cmd.Flags().GetString("type")

			// Build query based on search term and type filter
			query := help.NewSectionQuery()
			if typeFilter != "" {
				switch strings.ToLower(typeFilter) {
				case "topics":
					query = query.ReturnTopics()
				case "examples":
					query = query.ReturnExamples()
				case "applications":
					query = query.ReturnApplications()
				case "tutorials":
					query = query.ReturnTutorials()
				}
			}

			data, _ := hs.ComputeRenderData(query)
			helpData := data["Help"].(*help.HelpPage)

			// Convert to HTML types
			topics := make([]html.Topic, 0)
			for _, s := range helpData.AllGeneralTopics {
				topics = append(topics, html.Topic{
					Title:   s.Title,
					Short:   s.Short,
					Content: s.Content,
					Slug:    s.Slug,
				})
			}

			htmlHelp := html.Help{
				DefaultGeneralTopics: topics,
			}

			renderToOutput(html.HelpShortSectionList(htmlHelp, "", ""), output, withStyle)
		},
	}

	cmd.Flags().String("type", "", "Filter by type (topics, examples, applications, tutorials)")
	addCommonFlags(cmd)
	return cmd
}

func createViewCommand(hs *help.HelpSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [topic]",
		Short: "View a specific topic in HTML format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := cmd.Flags().GetString("output")
			withStyle, _ := cmd.Flags().GetBool("style")

			section, err := hs.GetSectionWithSlug(args[0])
			if err != nil {
				fmt.Printf("Topic not found: %s\n", args[0])
				os.Exit(1)
			}

			topic := html.Topic{
				Title:   section.Title,
				Short:   section.Short,
				Content: section.Content,
				Slug:    section.Slug,
			}

			shortTopicData := html.ShortTopicData{
				Topic:          topic,
				NoResultsFound: false,
				RequestedTypes: "",
				QueryString:    "",
			}

			renderToOutput(html.HelpShortTopic(shortTopicData), output, withStyle)
		},
	}

	addCommonFlags(cmd)
	return cmd
}

func createExamplesCommand(hs *help.HelpSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "examples",
		Short: "Show all examples in HTML format",
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := cmd.Flags().GetString("output")
			withStyle, _ := cmd.Flags().GetBool("style")

			// Create a query that returns all examples
			query := help.NewSectionQuery().ReturnExamples()
			data, _ := hs.ComputeRenderData(query)
			helpData := data["Help"].(*help.HelpPage)

			// Convert to HTML types
			examples := make([]html.Topic, 0)
			for _, s := range helpData.AllExamples {
				examples = append(examples, html.Topic{
					Title:   s.Title,
					Short:   s.Short,
					Content: s.Content,
					Slug:    s.Slug,
				})
			}

			htmlHelp := html.Help{
				DefaultExamples: examples,
			}

			renderToOutput(html.HelpShortSectionList(htmlHelp, "", ""), output, withStyle)
		},
	}

	addCommonFlags(cmd)
	return cmd
}

func createCommandCommand(hs *help.HelpSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command [command-name]",
		Short: "Show command help in HTML format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			output, _ := cmd.Flags().GetString("output")
			withStyle, _ := cmd.Flags().GetBool("style")

			// Create a query for the specific command
			query := help.NewSectionQuery().ReturnOnlyCommands(args[0])
			data, _ := hs.ComputeRenderData(query)
			helpData := data["Help"].(*help.HelpPage)

			// Convert to HTML types
			topics := make([]html.Topic, 0)
			for _, s := range helpData.AllGeneralTopics {
				topics = append(topics, html.Topic{
					Title:   s.Title,
					Short:   s.Short,
					Content: s.Content,
					Slug:    s.Slug,
				})
			}

			htmlHelp := html.Help{
				DefaultGeneralTopics: topics,
			}

			renderToOutput(html.HelpShortSectionList(htmlHelp, "", ""), output, withStyle)
		},
	}

	addCommonFlags(cmd)
	return cmd
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().String("output", "", "Output file (default stdout)")
	cmd.Flags().Bool("style", false, "Include CSS styling")
}

func renderToOutput(component interface {
	Render(ctx context.Context, w io.Writer) error
}, output string, withStyle bool) {
	var w io.Writer = os.Stdout
	if output != "" {
		f, err := os.Create(output)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	if withStyle {
		fmt.Fprintln(w, `<style>
.help-page { max-width: 800px; margin: 0 auto; font-family: system-ui; }
.topic { margin: 1em 0; padding: 1em; border: 1px solid #ddd; border-radius: 4px; }
.command-help { margin: 1em 0; }
.flag-group { margin: 1em 0; }
.flag-entry { margin: 0.5em 0; }
code { background: #f5f5f5; padding: 0.2em 0.4em; border-radius: 3px; }
pre { background: #f5f5f5; padding: 1em; border-radius: 4px; overflow-x: auto; }
.error { color: #e00; }
</style>`)
	}

	err := component.Render(context.Background(), w)
	if err != nil {
		fmt.Printf("Error rendering template: %v\n", err)
		os.Exit(1)
	}
}
