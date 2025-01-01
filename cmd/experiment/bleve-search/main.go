package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/go-go-golems/glazed/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// BleveHelpIndex wraps a bleve index with help system specific functionality
type BleveHelpIndex struct {
	index      bleve.Index
	helpSystem *help.HelpSystem
}

// SectionDocument represents a help section in the search index
type SectionDocument struct {
	Slug           string
	Title          string
	SubTitle       string
	Short          string
	Content        string
	Topics         []string
	Flags          []string
	Commands       []string
	SectionType    string
	IsTopLevel     bool
	IsTemplate     bool
	ShowPerDefault bool
	Order          int
}

// createIndexMapping creates a bleve index mapping for help sections
func createIndexMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	// Text fields with term vectors for better search
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = true
	textFieldMapping.IncludeTermVectors = true

	// Keyword fields for exact matches
	keywordFieldMapping := bleve.NewKeywordFieldMapping()
	keywordFieldMapping.Store = true

	// Boolean fields
	booleanFieldMapping := bleve.NewBooleanFieldMapping()
	booleanFieldMapping.Store = true

	// Numeric fields
	numericFieldMapping := bleve.NewNumericFieldMapping()
	numericFieldMapping.Store = true

	// Document mapping
	documentMapping := bleve.NewDocumentMapping()

	// Add field mappings
	documentMapping.AddFieldMappingsAt("Slug", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("Title", textFieldMapping)
	documentMapping.AddFieldMappingsAt("SubTitle", textFieldMapping)
	documentMapping.AddFieldMappingsAt("Short", textFieldMapping)
	documentMapping.AddFieldMappingsAt("Content", textFieldMapping)
	documentMapping.AddFieldMappingsAt("Topics", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("Flags", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("Commands", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("SectionType", keywordFieldMapping)
	documentMapping.AddFieldMappingsAt("IsTopLevel", booleanFieldMapping)
	documentMapping.AddFieldMappingsAt("IsTemplate", booleanFieldMapping)
	documentMapping.AddFieldMappingsAt("ShowPerDefault", booleanFieldMapping)
	documentMapping.AddFieldMappingsAt("Order", numericFieldMapping)

	indexMapping.AddDocumentMapping("section", documentMapping)

	return indexMapping
}

// NewBleveHelpIndex creates a new BleveHelpIndex with the given help system
func NewBleveHelpIndex(hs *help.HelpSystem, indexPath string) (*BleveHelpIndex, error) {
	var index bleve.Index
	var err error

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// Create new index if it doesn't exist
		mapping := createIndexMapping()
		index, err = bleve.New(indexPath, mapping)
	} else {
		// Open existing index
		index, err = bleve.Open(indexPath)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to create/open bleve index")
	}

	return &BleveHelpIndex{
		index:      index,
		helpSystem: hs,
	}, nil
}

// IndexSection adds a single help section to the index
func (bhi *BleveHelpIndex) IndexSection(section *help.Section) error {
	doc := SectionDocument{
		Slug:           section.Slug,
		Title:          section.Title,
		SubTitle:       section.SubTitle,
		Short:          section.Short,
		Content:        section.Content,
		Topics:         section.Topics,
		Flags:          section.Flags,
		Commands:       section.Commands,
		SectionType:    section.SectionType.String(),
		IsTopLevel:     section.IsTopLevel,
		IsTemplate:     section.IsTemplate,
		ShowPerDefault: section.ShowPerDefault,
		Order:          section.Order,
	}

	return bhi.index.Index(section.Slug, doc)
}

// IndexAllSections indexes all sections from the help system
func (bhi *BleveHelpIndex) IndexAllSections() error {
	batch := bhi.index.NewBatch()
	for _, section := range bhi.helpSystem.Sections {
		doc := SectionDocument{
			Slug:           section.Slug,
			Title:          section.Title,
			SubTitle:       section.SubTitle,
			Short:          section.Short,
			Content:        section.Content,
			Topics:         section.Topics,
			Flags:          section.Flags,
			Commands:       section.Commands,
			SectionType:    section.SectionType.String(),
			IsTopLevel:     section.IsTopLevel,
			IsTemplate:     section.IsTemplate,
			ShowPerDefault: section.ShowPerDefault,
			Order:          section.Order,
		}
		err := batch.Index(section.Slug, doc)
		if err != nil {
			return errors.Wrapf(err, "failed to index section %s", section.Slug)
		}
	}
	return bhi.index.Batch(batch)
}

var rootCmd = &cobra.Command{
	Use:   "bleve-search",
	Short: "Search help documentation using bleve",
	Long:  "A tool to index and search help documentation using the bleve search engine",
}

var indexPath string

func init() {
	rootCmd.PersistentFlags().StringVar(&indexPath, "index", "/tmp/bleve-help-index", "Path to the bleve index")

	rootCmd.AddCommand(createIndexCommand())
	rootCmd.AddCommand(createSearchCommand())
	rootCmd.AddCommand(createFieldSearchCommand())
}

func createIndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index all help documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create help system and load docs
			helpSystem := help.NewHelpSystem()
			if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
				return errors.Wrap(err, "failed to load help system")
			}

			// Ensure index directory exists
			if err := os.MkdirAll(filepath.Dir(indexPath), 0755); err != nil {
				return errors.Wrap(err, "failed to create index directory")
			}

			// Create and initialize index
			bleveIndex, err := NewBleveHelpIndex(helpSystem, indexPath)
			if err != nil {
				return errors.Wrap(err, "failed to create bleve index")
			}
			defer bleveIndex.index.Close()

			// Index all sections
			if err := bleveIndex.IndexAllSections(); err != nil {
				return errors.Wrap(err, "failed to index sections")
			}

			docCount, err := bleveIndex.index.DocCount()
			if err != nil {
				return errors.Wrap(err, "failed to get document count")
			}

			fmt.Printf("Successfully indexed %d help sections to %s\n", docCount, indexPath)
			return nil
		},
	}

	return cmd
}

func createSearchCommand() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search help documentation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open index
			index, err := bleve.Open(indexPath)
			if err != nil {
				return errors.Wrap(err, "failed to open index")
			}
			defer index.Close()

			// Create search query
			query := bleve.NewQueryStringQuery(args[0])
			searchRequest := bleve.NewSearchRequest(query)
			searchRequest.Size = limit
			searchRequest.Fields = []string{"Title", "Short", "Content", "SectionType"}

			// Execute search
			searchResult, err := index.Search(searchRequest)
			if err != nil {
				return errors.Wrap(err, "failed to execute search")
			}

			// Print results
			fmt.Printf("Found %d matches in %s\n\n", searchResult.Total, searchResult.Took)
			for _, hit := range searchResult.Hits {
				fmt.Printf("Score: %f\n", hit.Score)
				fmt.Printf("Type: %s\n", hit.Fields["SectionType"])
				fmt.Printf("Title: %s\n", hit.Fields["Title"])
				if short, ok := hit.Fields["Short"].(string); ok && short != "" {
					fmt.Printf("Short: %s\n", short)
				}
				fmt.Println(strings.Repeat("-", 80))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results to return")
	return cmd
}

func createFieldSearchCommand() *cobra.Command {
	var (
		limit       int
		sectionType string
		isTopLevel  bool
	)

	cmd := &cobra.Command{
		Use:   "field-search [query]",
		Short: "Search help documentation with field filters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open index
			index, err := bleve.Open(indexPath)
			if err != nil {
				return errors.Wrap(err, "failed to open index")
			}
			defer index.Close()

			// Create compound query
			var queries []query.Query
			queries = append(queries, bleve.NewQueryStringQuery(args[0]))

			if sectionType != "" {
				typeQuery := bleve.NewTermQuery(sectionType)
				typeQuery.SetField("SectionType")
				queries = append(queries, typeQuery)
			}

			if cmd.Flags().Changed("top-level") {
				topLevelQuery := bleve.NewBoolFieldQuery(isTopLevel)
				topLevelQuery.SetField("IsTopLevel")
				queries = append(queries, topLevelQuery)
			}

			// Combine all queries with AND
			query := bleve.NewConjunctionQuery(queries...)

			// Create and execute search request
			searchRequest := bleve.NewSearchRequest(query)
			searchRequest.Size = limit
			searchRequest.Fields = []string{"Title", "Short", "Content", "SectionType", "IsTopLevel"}

			searchResult, err := index.Search(searchRequest)
			if err != nil {
				return errors.Wrap(err, "failed to execute search")
			}

			// Print results
			fmt.Printf("Found %d matches in %s\n\n", searchResult.Total, searchResult.Took)
			for _, hit := range searchResult.Hits {
				fmt.Printf("Score: %f\n", hit.Score)
				fmt.Printf("Type: %s\n", hit.Fields["SectionType"])
				fmt.Printf("Title: %s\n", hit.Fields["Title"])
				if short, ok := hit.Fields["Short"].(string); ok && short != "" {
					fmt.Printf("Short: %s\n", short)
				}
				fmt.Printf("Top Level: %v\n", hit.Fields["IsTopLevel"])
				fmt.Println(strings.Repeat("-", 80))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results to return")
	cmd.Flags().StringVar(&sectionType, "type", "", "Filter by section type (GeneralTopic, Example, Application, Tutorial)")
	cmd.Flags().BoolVar(&isTopLevel, "top-level", false, "Filter by top level status")

	return cmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
