package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/go-go-golems/glazed/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func (sd SectionDocument) Type() string {
	return "section"
}

func init() {
	// Initialize zerolog with console writer and debug level
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(consoleWriter).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Caller().
		Logger()
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

// createNewIndex creates a new bleve index with proper logging
func createNewIndex(indexPath string, mapping mapping.IndexMapping) (bleve.Index, error) {
	startTime := time.Now()
	log.Debug().Str("path", indexPath).Msg("Creating new bleve index")

	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		log.Error().Err(err).Str("path", indexPath).Msg("Failed to create bleve index")
		return nil, errors.Wrap(err, "failed to create bleve index")
	}

	elapsed := time.Since(startTime)
	log.Debug().
		Str("path", indexPath).
		Dur("duration", elapsed).
		Msg("Successfully created new bleve index")

	return index, nil
}

// openExistingIndex opens an existing bleve index with proper logging
func openExistingIndex(indexPath string) (bleve.Index, error) {
	startTime := time.Now()
	log.Debug().Str("path", indexPath).Msg("Opening existing bleve index")

	index, err := bleve.Open(indexPath)
	if err != nil {
		log.Error().Err(err).Str("path", indexPath).Msg("Failed to open bleve index")
		return nil, errors.Wrap(err, "failed to open bleve index")
	}

	elapsed := time.Since(startTime)
	log.Debug().
		Str("path", indexPath).
		Dur("duration", elapsed).
		Msg("Successfully opened existing bleve index")

	return index, nil
}

// NewBleveHelpIndex creates a new BleveHelpIndex with the given help system
func NewBleveHelpIndex(hs *help.HelpSystem, indexPath string) (*BleveHelpIndex, error) {
	var index bleve.Index
	var err error

	mapping := createIndexMapping()

	if memoryIndex {
		log.Debug().Msg("Creating in-memory bleve index")
		index, err = bleve.NewMemOnly(mapping)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create in-memory bleve index")
			return nil, errors.Wrap(err, "failed to create in-memory bleve index")
		}
		log.Debug().Msg("Successfully created in-memory bleve index")
	} else {
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			// Create new index if it doesn't exist
			index, err = createNewIndex(indexPath, mapping)
			if err != nil {
				return nil, err
			}
		} else {
			// Open existing index
			index, err = openExistingIndex(indexPath)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
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

var (
	indexPath   string
	memoryIndex bool
)

func init() {
	rootCmd.PersistentFlags().StringVar(&indexPath, "index", "/tmp/bleve-help-index", "Path to the bleve index")
	rootCmd.PersistentFlags().BoolVar(&memoryIndex, "memory", true, "Create index in memory instead of on disk")

	rootCmd.AddCommand(createIndexCommand())
	rootCmd.AddCommand(createSearchCommand())
	rootCmd.AddCommand(createFieldSearchCommand())
	rootCmd.AddCommand(createDebugCommand())
}

func createIndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index all help documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			startTime := time.Now()

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

			indexingDuration := time.Since(startTime)
			fmt.Printf("Successfully indexed %d help sections to %s in %v\n", docCount, indexPath, indexingDuration)
			return nil
		},
	}

	return cmd
}

// initializeHelpSystem creates and initializes a help system with proper logging
func initializeHelpSystem() (*help.HelpSystem, error) {
	log.Debug().Msg("Initializing help system")
	helpSystem := help.NewHelpSystem()
	if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
		log.Error().Err(err).Msg("Failed to load help system")
		return nil, errors.Wrap(err, "failed to load help system")
	}
	log.Debug().Msg("Successfully initialized help system")
	return helpSystem, nil
}

// openOrCreateIndex opens an existing index or creates a new one with all sections indexed
func openOrCreateIndex() (bleve.Index, error) {
	var index bleve.Index
	var err error

	// Initialize help system first, outside of timing measurement
	helpSystem, err := initializeHelpSystem()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	if memoryIndex {
		log.Debug().Msg("Creating in-memory index")
		bleveIndex, err := NewBleveHelpIndex(helpSystem, indexPath)
		if err != nil {
			return nil, err
		}
		index = bleveIndex.index
	} else {
		bleveIndex, err := NewBleveHelpIndex(helpSystem, indexPath)
		if err != nil {
			return nil, err
		}
		index = bleveIndex.index
	}

	// Index all sections
	if err := indexAllSections(index, helpSystem); err != nil {
		return nil, err
	}

	docCount, err := index.DocCount()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get document count")
		return nil, errors.Wrap(err, "failed to get document count")
	}

	indexingDuration := time.Since(startTime)
	storageType := "disk"
	if memoryIndex {
		storageType = "memory"
	}

	log.Info().
		Str("type", storageType).
		Str("path", indexPath).
		Uint64("documents", docCount).
		Dur("duration", indexingDuration).
		Msg("Successfully initialized index")

	return index, nil
}

// indexAllSections indexes all sections from the help system with proper logging
func indexAllSections(index bleve.Index, helpSystem *help.HelpSystem) error {
	log.Debug().Msg("Starting to index all sections")
	startTime := time.Now()

	batch := index.NewBatch()
	for _, section := range helpSystem.Sections {
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
		if err := batch.Index(section.Slug, doc); err != nil {
			log.Error().Err(err).Str("slug", section.Slug).Msg("Failed to index section")
			return errors.Wrapf(err, "failed to index section %s", section.Slug)
		}
	}

	if err := index.Batch(batch); err != nil {
		log.Error().Err(err).Msg("Failed to execute batch indexing")
		return errors.Wrap(err, "failed to execute batch indexing")
	}

	elapsed := time.Since(startTime)
	log.Debug().
		Int("sections", len(helpSystem.Sections)).
		Dur("duration", elapsed).
		Msg("Successfully indexed all sections")

	return nil
}

func createSearchCommand() *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search help documentation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open or create index
			index, err := openOrCreateIndex()
			if err != nil {
				return err
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

	validTypes := map[string]bool{
		"GeneralTopic": true,
		"Example":      true,
		"Application":  true,
		"Tutorial":     true,
	}

	cmd := &cobra.Command{
		Use:   "field-search [query]",
		Short: "Search help documentation with field filters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open or create index
			index, err := openOrCreateIndex()
			if err != nil {
				return err
			}
			defer index.Close()

			// Validate section type
			// if sectionType != "" && !validTypes[sectionType] {
			// 	return fmt.Errorf("invalid section type %q. Valid types are: GeneralTopic, Example, Application, Tutorial", sectionType)
			// }
			_ = validTypes

			// Create compound query
			var queries []query.Query

			// If no search term is provided, match all documents
			if len(args) > 0 {
				queries = append(queries, bleve.NewQueryStringQuery(args[0]))
			} else {
				queries = append(queries, bleve.NewMatchAllQuery())
			}

			if sectionType != "" {
				// Create a match query instead of term query for case-insensitive matching
				typeQuery := bleve.NewTermQuery(sectionType)
				typeQuery.SetField("SectionType")
				fmt.Printf("Searching for section type: %q\n", sectionType)
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

func createDebugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Show debug information about the index",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Open or create index
			index, err := openOrCreateIndex()
			if err != nil {
				return err
			}
			defer index.Close()

			// Create a match all query to get all documents
			query := bleve.NewMatchAllQuery()
			searchRequest := bleve.NewSearchRequest(query)
			searchRequest.Size = 1000 // Get all documents
			searchRequest.Fields = []string{"SectionType"}

			// Execute search
			searchResult, err := index.Search(searchRequest)
			if err != nil {
				return errors.Wrap(err, "failed to execute search")
			}

			// Print unique section types
			sectionTypes := make(map[string]int)
			for _, hit := range searchResult.Hits {
				if sType, ok := hit.Fields["SectionType"].(string); ok {
					sectionTypes[sType]++
				}
			}

			fmt.Println("Section types in index:")
			for sType, count := range sectionTypes {
				fmt.Printf("  %s: %d documents\n", sType, count)
			}

			return nil
		},
	}

	return cmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
