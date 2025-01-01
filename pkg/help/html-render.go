package help

import (
	"context"
	"io"
	"strings"

	"github.com/a-h/templ"
	glazed_cobra "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/help/templates/html"
	"github.com/spf13/cobra"
)

// RenderToHTML renders the help content to HTML using templ templates
func RenderToHTML(t templ.Component, output io.Writer) error {
	return t.Render(context.Background(), output)
}

// RenderTopicHelpToHTML renders a topic's help content to HTML
func (hs *HelpSystem) RenderTopicHelpToHTML(
	topicSection *Section,
	options *RenderOptions,
	output io.Writer) error {

	data, noResultsFound := hs.ComputeRenderData(options.Query)

	helpData := data["Help"].(*HelpPage)
	helpCommand := options.HelpCommand

	// Convert HelpPage to the HTML template's Help type
	htmlHelp := html.Help{
		DefaultGeneralTopics: convertTopics(helpData.DefaultGeneralTopics),
		OtherGeneralTopics:   convertTopics(helpData.OtherGeneralTopics),
		DefaultExamples:      convertTopics(helpData.DefaultExamples),
		OtherExamples:        convertTopics(helpData.OtherExamples),
		DefaultApplications:  convertTopics(helpData.DefaultApplications),
		OtherApplications:    convertTopics(helpData.OtherApplications),
		DefaultTutorials:     convertTopics(helpData.DefaultTutorials),
		OtherTutorials:       convertTopics(helpData.OtherTutorials),
	}

	// Convert Section to the HTML template's Topic type
	htmlTopic := html.Topic{
		Title:   topicSection.Title,
		Short:   topicSection.Short,
		Content: topicSection.Content,
		Slug:    topicSection.Slug,
	}

	if options.ListSections || noResultsFound {
		shortTopicData := html.ShortTopicData{
			Topic:          htmlTopic,
			NoResultsFound: noResultsFound,
			RequestedTypes: options.Query.GetRequestedTypesAsString(),
			QueryString:    options.Query.GetOnlyQueryAsString(),
		}
		return RenderToHTML(html.HelpShortTopic(shortTopicData), output)
	}

	if options.ShowShortTopic {
		return RenderToHTML(html.HelpTopic(htmlTopic), output)
	}

	if options.ShowAllSections {
		helpList := html.HelpList{
			AllGeneralTopics: convertTopics(helpData.AllGeneralTopics),
			AllExamples:      convertTopics(helpData.AllExamples),
			AllApplications:  convertTopics(helpData.AllApplications),
			AllTutorials:     convertTopics(helpData.AllTutorials),
		}
		return RenderToHTML(html.HelpLongSectionList(helpList, helpCommand), output)
	}

	return RenderToHTML(html.HelpShortSectionList(htmlHelp, helpCommand, topicSection.Slug), output)
}

// RenderCommandHelpPageToHTML renders a command's help page to HTML
func RenderCommandHelpPageToHTML(c *cobra.Command, options *RenderOptions, hs *HelpSystem, output io.Writer) error {
	isTopLevel := c.Parent() == nil
	userQuery := options.Query
	userQuery.OnlyTopLevel = options.OnlyTopLevel

	if !isTopLevel {
		userQuery = userQuery.SearchForCommand(c.Name())
	}

	data, noResultsFound := hs.ComputeRenderData(userQuery)
	helpData := data["Help"].(*HelpPage)

	// Convert cobra.Command to HTML template's Command type
	htmlCmd := html.Command{
		Name:  c.Name(),
		Short: c.Short,
		Long:  c.Long,
	}

	if options.ListSections || noResultsFound {
		shortHelpData := html.CobraShortHelpData{
			Command:        htmlCmd,
			NoResultsFound: noResultsFound,
			RequestedTypes: userQuery.GetRequestedTypesAsString(),
			QueryString:    userQuery.GetOnlyQueryAsString(),
			HelpCommand:    options.HelpCommand,
		}
		return RenderToHTML(html.CobraShortHelp(shortHelpData), output)
	}

	if options.ShowShortTopic {
		return RenderToHTML(html.CobraHelp(htmlCmd, options.HelpCommand), output)
	}

	// For full usage template
	flagGroupUsage := glazed_cobra.ComputeCommandFlagGroupUsage(c)

	// Handle short help layers
	if !options.LongHelp {
		shortHelpLayers_, ok := c.Annotations["shortHelpLayers"]
		if ok {
			shortHelpLayers := map[string]interface{}{}
			for _, v := range strings.Split(shortHelpLayers_, ",") {
				shortHelpLayers[v] = true
			}

			localGroupUsages := []*glazed_cobra.FlagGroupUsage{}
			inheritedGroupUsages := []*glazed_cobra.FlagGroupUsage{}
			for _, f := range flagGroupUsage.LocalGroupUsages {
				if _, ok = shortHelpLayers[f.Slug]; ok {
					localGroupUsages = append(localGroupUsages, f)
				}
			}
			for _, f := range flagGroupUsage.InheritedGroupUsages {
				if _, ok = shortHelpLayers[f.Slug]; ok {
					inheritedGroupUsages = append(inheritedGroupUsages, f)
				}
			}

			flagGroupUsage.LocalGroupUsages = localGroupUsages
			flagGroupUsage.InheritedGroupUsages = inheritedGroupUsages
		}
	}

	// Calculate max flag string length
	maxLength := 0
	if flagGroupUsage != nil {
		for _, group := range flagGroupUsage.LocalGroupUsages {
			for _, usage := range group.FlagUsages {
				if len(usage.FlagString) > maxLength {
					maxLength = len(usage.FlagString)
				}
			}
		}
	}

	// Convert to HTML template's CobraCommand type
	htmlCobraCmd := html.CobraCommand{
		Runnable:                   c.Runnable(),
		UseLine:                    c.UseLine(),
		HasAvailableSubCommands:    c.HasAvailableSubCommands(),
		CommandPath:                c.CommandPath(),
		Aliases:                    c.Aliases,
		NameAndAliases:             c.NameAndAliases(),
		HasExample:                 c.HasExample(),
		Example:                    c.Example,
		Commands:                   convertCobraCommands(c.Commands()),
		Groups:                     convertCobraGroups(c.Groups()),
		Sections:                   convertTopics(helpData.AllGeneralTopics),
		HasAvailableLocalFlags:     c.HasAvailableLocalFlags(),
		LocalFlags:                 c.LocalFlags().FlagUsages(),
		HasAvailableInheritedFlags: c.HasAvailableInheritedFlags(),
		InheritedFlags:             c.InheritedFlags().FlagUsages(),
		HasHelpSubCommands:         c.HasHelpSubCommands(),
		LongHelp:                   options.LongHelp,
	}

	if flagGroupUsage != nil {
		htmlCobraCmd.FlagGroupUsage = &html.FlagGroupUsage{
			LocalGroupUsages:     convertFlagGroups(flagGroupUsage.LocalGroupUsages),
			InheritedGroupUsages: convertFlagGroups(flagGroupUsage.InheritedGroupUsages),
			FlagUsageMaxLength:   maxLength,
		}
	}

	return RenderToHTML(html.CobraUsage(htmlCobraCmd), output)
}

// Helper functions to convert between types

func convertTopics(sections []*Section) []html.Topic {
	var topics []html.Topic
	for _, s := range sections {
		topics = append(topics, html.Topic{
			Title:   s.Title,
			Short:   s.Short,
			Content: s.Content,
			Slug:    s.Slug,
		})
	}
	return topics
}

func convertCobraCommands(cmds []*cobra.Command) []html.CobraSubCommand {
	var htmlCmds []html.CobraSubCommand
	for _, cmd := range cmds {
		htmlCmds = append(htmlCmds, html.CobraSubCommand{
			Name:               cmd.Name(),
			Short:              cmd.Short,
			IsAvailableCommand: cmd.IsAvailableCommand(),
			GroupID:            cmd.GroupID,
			NamePadding:        cmd.NamePadding(),
		})
	}
	return htmlCmds
}

func convertCobraGroups(groups []*cobra.Group) []html.CobraGroup {
	var htmlGroups []html.CobraGroup
	for _, g := range groups {
		htmlGroups = append(htmlGroups, html.CobraGroup{
			ID:    g.ID,
			Title: g.Title,
		})
	}
	return htmlGroups
}

func convertFlagGroups(groups []*glazed_cobra.FlagGroupUsage) []html.FlagGroup {
	var htmlGroups []html.FlagGroup
	for _, g := range groups {
		htmlGroups = append(htmlGroups, html.FlagGroup{
			Name:       g.Name,
			FlagUsages: convertFlagUsages(g.FlagUsages),
		})
	}
	return htmlGroups
}

func convertFlagUsages(usages []*glazed_cobra.FlagUsage) []html.FlagUsage {
	var htmlUsages []html.FlagUsage
	for _, u := range usages {
		htmlUsages = append(htmlUsages, html.FlagUsage{
			FlagString: u.FlagString,
			Help:       u.Help,
			Default:    u.Default,
		})
	}
	return htmlUsages
}
