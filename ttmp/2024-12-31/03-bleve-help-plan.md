# Bleve Help System Implementation Plan

## Overview

This document outlines the plan to integrate the Bleve search engine into the glazed help system. 
The goal is to replace the current query system with a more powerful full-text search capability while maintaining backward compatibility.

## Current System Analysis

The current help system uses a custom query builder pattern (`SectionQuery`) with boolean filters for:
- Section types (GeneralTopic, Example, Application, Tutorial)
- Visibility filters (ShowPerDefault, TopLevel)
- Metadata matching (Topics, Flags, Commands, Slugs)

## New System Design

### 1. Core Components

#### BleveHelpIndex
```go
type BleveHelpIndex struct {
    index bleve.Index
    helpSystem *HelpSystem
}

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
```

#### Query Builder
```go
type BleveSectionQuery struct {
    must        []bleve.Query
    should      []bleve.Query
    mustNot     []bleve.Query
    sortBy      []string
    from        int
    size        int
}
```

### 2. Implementation Phases

#### Phase 1: Core Infrastructure
1. Create BleveHelpIndex structure
2. Implement document mapping
3. Setup in-memory index creation
4. Add section indexing functionality

#### Phase 2: Query System
1. Implement BleveSectionQuery builder
2. Create search execution methods
3. Add result conversion utilities
4. Implement sorting and pagination

#### Phase 3: Integration
1. Modify HelpSystem to include search index
2. Update section addition/removal to maintain index
3. Create compatibility layer for existing query methods
4. Add new search capabilities

#### Phase 4: Migration
1. Update existing help commands to use new search
2. Add new search-specific commands
3. Document new search capabilities

### 3. Detailed Technical Specifications

#### Document Mapping
```go
func createIndexMapping() *mapping.IndexMapping {
    indexMapping := bleve.NewIndexMapping()
    
    // Text fields
    textFieldMapping := bleve.NewTextFieldMapping()
    textFieldMapping.Store = true
    textFieldMapping.IncludeTermVectors = true
    
    // Keyword fields
    keywordFieldMapping := bleve.NewKeywordFieldMapping()
    keywordFieldMapping.Store = true
    
    // Boolean fields
    booleanFieldMapping := bleve.NewBooleanFieldMapping()
    booleanFieldMapping.Store = true
    
    // Document mapping
    documentMapping := bleve.NewDocumentMapping()
    documentMapping.AddFieldMappingsAt("Title", textFieldMapping)
    documentMapping.AddFieldMappingsAt("Content", textFieldMapping)
    documentMapping.AddFieldMappingsAt("Topics", keywordFieldMapping)
    documentMapping.AddFieldMappingsAt("Commands", keywordFieldMapping)
    documentMapping.AddFieldMappingsAt("IsTopLevel", booleanFieldMapping)
    
    indexMapping.AddDocumentMapping("section", documentMapping)
    
    return indexMapping
}
```

#### Search Implementation
```go
func (bhi *BleveHelpIndex) Search(query *BleveSectionQuery) ([]*Section, error) {
    searchRequest := bleve.NewSearchRequest(query.BuildQuery())
    searchRequest.SortBy(query.sortBy)
    searchRequest.From = query.from
    searchRequest.Size = query.size
    
    searchRequest.Fields = []string{
        "Slug", "Title", "SubTitle", "Short",
        "Topics", "Flags", "Commands",
        "SectionType", "IsTopLevel", "ShowPerDefault",
    }
    
    results, err := bhi.index.Search(searchRequest)
    if err != nil {
        return nil, err
    }
    
    return bhi.convertResultsToSections(results)
}
```

### 4. Query Patterns and User Behaviors

This section details the various ways users interact with the help system and how these translate to Bleve queries.

#### Command Reference Queries

**User Intent:**
- Quick lookup of command documentation
- Finding command examples
- Viewing command flags and options

**Implementation:**
```go
// Command quick reference query
func (bhi *BleveHelpIndex) CommandQuickReference(command string) *bleve.Query {
    // Exact command match with high boost
    commandQuery := bleve.NewTermQuery(command).SetField("Commands").SetBoost(2.0)
    
    // Title/short description match
    textQuery := bleve.NewDisjunctionQuery(
        bleve.NewMatchQuery(command).SetField("Title"),
        bleve.NewMatchQuery(command).SetField("Short"),
    )
    
    // Combine with preference for default sections
    query := bleve.NewBooleanQuery()
    query.AddMust(bleve.NewDisjunctionQuery(commandQuery, textQuery))
    query.AddShould(bleve.NewTermQuery("true").SetField("ShowPerDefault"))
    
    return query
}
```

#### Topic Exploration

**User Intent:**
- Understanding concepts
- Learning about features
- Finding related functionality

**Implementation:**
```go
// Topic exploration query
func (bhi *BleveHelpIndex) TopicExploration(topic string) *bleve.Query {
    // Full text search across all content
    contentQuery := bleve.NewDisjunctionQuery(
        bleve.NewMatchQuery(topic).SetField("Title").SetBoost(2.0),
        bleve.NewMatchQuery(topic).SetField("Content"),
        bleve.NewMatchQuery(topic).SetField("Topics"),
    )
    
    // Prefer general topics and tutorials
    typeQuery := bleve.NewDisjunctionQuery(
        bleve.NewTermQuery("GeneralTopic").SetField("SectionType"),
        bleve.NewTermQuery("Tutorial").SetField("SectionType"),
    )
    
    query := bleve.NewBooleanQuery()
    query.AddMust(contentQuery)
    query.AddShould(typeQuery)
    
    return query
}
```

#### Example Search

**User Intent:**
- Finding practical examples
- Looking for specific use cases
- Learning by example

**Implementation:**
```go
// Example search query
func (bhi *BleveHelpIndex) ExampleSearch(searchText string) *bleve.Query {
    // Match example sections
    typeQuery := bleve.NewTermQuery("Example").SetField("SectionType")
    
    // Match content
    contentQuery := bleve.NewDisjunctionQuery(
        bleve.NewMatchQuery(searchText).SetField("Short").SetBoost(2.0),
        bleve.NewMatchQuery(searchText).SetField("Content"),
        bleve.NewMatchQuery(searchText).SetField("Commands"),
    )
    
    query := bleve.NewBooleanQuery()
    query.AddMust(typeQuery, contentQuery)
    
    return query
}
```

#### Feature Discovery

**User Intent:**
- Exploring available capabilities
- Finding new features
- Understanding tool capabilities

**Implementation:**
```go
// Feature discovery query
func (bhi *BleveHelpIndex) FeatureDiscovery(searchText string) *bleve.Query {
    // Broad match across all fields
    query := bleve.NewDisjunctionQuery(
        bleve.NewMatchQuery(searchText).SetField("Title").SetBoost(2.0),
        bleve.NewMatchQuery(searchText).SetField("Short"),
        bleve.NewMatchQuery(searchText).SetField("Content"),
        bleve.NewMatchQuery(searchText).SetField("Topics"),
        bleve.NewMatchQuery(searchText).SetField("Commands"),
    )
    
    // Boost top-level sections
    topLevelBoost := bleve.NewTermQuery("true").SetField("IsTopLevel").SetBoost(1.5)
    
    booleanQuery := bleve.NewBooleanQuery()
    booleanQuery.AddMust(query)
    booleanQuery.AddShould(topLevelBoost)
    
    return booleanQuery
}
```

#### Troubleshooting

**User Intent:**
- Finding solutions to problems
- Understanding errors
- Debugging issues

**Implementation:**
```go
// Troubleshooting query
func (bhi *BleveHelpIndex) TroubleshootingSearch(searchText string) *bleve.Query {
    // Content-focused search
    contentQuery := bleve.NewDisjunctionQuery(
        bleve.NewMatchQuery(searchText).SetField("Content").SetBoost(2.0),
        bleve.NewMatchQuery(searchText).SetField("Short"),
        bleve.NewMatchQuery(searchText).SetField("Title"),
    )
    
    // Prefer examples and tutorials
    typeQuery := bleve.NewDisjunctionQuery(
        bleve.NewTermQuery("Example").SetField("SectionType"),
        bleve.NewTermQuery("Tutorial").SetField("SectionType"),
    )
    
    query := bleve.NewBooleanQuery()
    query.AddMust(contentQuery)
    query.AddShould(typeQuery)
    
    return query
}
```

#### Related Content Discovery

**User Intent:**
- Finding related topics
- Discovering connected features
- Understanding feature relationships

**Implementation:**
```go
// Related content query
func (bhi *BleveHelpIndex) RelatedContent(section *Section) *bleve.Query {
    // Match topics
    topicsQuery := bleve.NewDisjunctionQuery()
    for _, topic := range section.Topics {
        topicsQuery.AddQuery(bleve.NewTermQuery(topic).SetField("Topics"))
    }
    
    // Match commands
    commandsQuery := bleve.NewDisjunctionQuery()
    for _, cmd := range section.Commands {
        commandsQuery.AddQuery(bleve.NewTermQuery(cmd).SetField("Commands"))
    }
    
    // Combine queries
    query := bleve.NewBooleanQuery()
    query.AddShould(topicsQuery, commandsQuery)
    
    // Exclude current section
    query.AddMustNot(bleve.NewTermQuery(section.Slug).SetField("Slug"))
    
    return query
}
```

These query patterns cover the main ways users interact with the help system while providing relevant and contextual results. The implementation uses field boosting, type filtering, and relevance scoring to ensure the most appropriate results are returned first.

### 5. Existing Related Content Queries

The current help system implements several patterns for finding related content that need to be preserved and enhanced in the Bleve implementation.

#### Default General Topics

**Current Implementation:**
```go
// Returns topics that:
// - Are of type GeneralTopic
// - Match the section's topics
// - Are marked as ShowPerDefault
func (s *Section) DefaultGeneralTopic() []*Section {
    return NewSectionQuery().
        ReturnTopics().
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}
```

**Bleve Implementation:**
```go
func (bhi *BleveHelpIndex) DefaultGeneralTopics(section *Section) *bleve.Query {
    query := bleve.NewBooleanQuery()
    
    // Must be a general topic
    query.AddMust(bleve.NewTermQuery("GeneralTopic").SetField("SectionType"))
    
    // Must match one of the section's topics
    topicsQuery := bleve.NewDisjunctionQuery()
    for _, topic := range section.Topics {
        topicsQuery.AddQuery(bleve.NewTermQuery(topic).SetField("Topics"))
    }
    query.AddMust(topicsQuery)
    
    // Must be shown by default
    query.AddMust(bleve.NewTermQuery("true").SetField("ShowPerDefault"))
    
    // Exclude the current section
    query.AddMustNot(bleve.NewTermQuery(section.Slug).SetField("Slug"))
    
    return query
}
```

#### Default Examples

**Current Implementation:**
```go
// Returns examples that:
// - Are of type Example
// - Match the section's topics
// - Are marked as ShowPerDefault
func (s *Section) DefaultExamples() []*Section {
    return NewSectionQuery().
        ReturnExamples().
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}
```

**Bleve Implementation:**
```go
func (bhi *BleveHelpIndex) DefaultExamples(section *Section) *bleve.Query {
    query := bleve.NewBooleanQuery()
    
    // Must be an example
    query.AddMust(bleve.NewTermQuery("Example").SetField("SectionType"))
    
    // Must match one of the section's topics
    topicsQuery := bleve.NewDisjunctionQuery()
    for _, topic := range section.Topics {
        topicsQuery.AddQuery(bleve.NewTermQuery(topic).SetField("Topics"))
    }
    query.AddMust(topicsQuery)
    
    // Must be shown by default
    query.AddMust(bleve.NewTermQuery("true").SetField("ShowPerDefault"))
    
    // Exclude the current section
    query.AddMustNot(bleve.NewTermQuery(section.Slug).SetField("Slug"))
    
    return query
}
```

#### Other Examples

**Current Implementation:**
```go
// Returns examples that:
// - Are of type Example
// - Match the section's topics
// - Are NOT marked as ShowPerDefault
func (s *Section) OtherExamples() []*Section {
    return NewSectionQuery().
        ReturnExamples().
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyNotShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}
```

**Bleve Implementation:**
```go
func (bhi *BleveHelpIndex) OtherExamples(section *Section) *bleve.Query {
    query := bleve.NewBooleanQuery()
    
    // Must be an example
    query.AddMust(bleve.NewTermQuery("Example").SetField("SectionType"))
    
    // Must match one of the section's topics
    topicsQuery := bleve.NewDisjunctionQuery()
    for _, topic := range section.Topics {
        topicsQuery.AddQuery(bleve.NewTermQuery(topic).SetField("Topics"))
    }
    query.AddMust(topicsQuery)
    
    // Must NOT be shown by default
    query.AddMust(bleve.NewTermQuery("false").SetField("ShowPerDefault"))
    
    // Exclude the current section
    query.AddMustNot(bleve.NewTermQuery(section.Slug).SetField("Slug"))
    
    return query
}
```

#### Default Applications and Tutorials

**Current Implementation:**
```go
// Similar patterns for applications and tutorials
func (s *Section) DefaultApplications() []*Section {
    return NewSectionQuery().
        ReturnApplications().
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}

func (s *Section) DefaultTutorials() []*Section {
    return NewSectionQuery().
        ReturnTutorials().
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}
```

**Bleve Implementation:**
```go
func (bhi *BleveHelpIndex) DefaultContentByType(section *Section, sectionType string) *bleve.Query {
    query := bleve.NewBooleanQuery()
    
    // Must be of the specified type
    query.AddMust(bleve.NewTermQuery(sectionType).SetField("SectionType"))
    
    // Must match one of the section's topics
    topicsQuery := bleve.NewDisjunctionQuery()
    for _, topic := range section.Topics {
        topicsQuery.AddQuery(bleve.NewTermQuery(topic).SetField("Topics"))
    }
    query.AddMust(topicsQuery)
    
    // Must be shown by default
    query.AddMust(bleve.NewTermQuery("true").SetField("ShowPerDefault"))
    
    // Exclude the current section
    query.AddMustNot(bleve.NewTermQuery(section.Slug).SetField("Slug"))
    
    return query
}
```

These existing query patterns are crucial for maintaining the help system's current functionality while enhancing it with Bleve's full-text search capabilities. The Bleve implementations preserve the exact matching behavior while adding the potential for:

1. Better relevance scoring
2. Full-text search within the filtered results
3. Field boosting for better result ordering
4. More flexible query combinations
5. Better performance on large help collections
