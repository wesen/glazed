# Bleve Help Search

A command-line tool for indexing and searching glazed help documentation using the Bleve search engine.

## Building

```bash
go build -o bleve-search main.go
```

## Usage

### Indexing Help Documentation

First, create a search index of all help documentation:

```bash
# Index using default location (/tmp/bleve-help-index)
./bleve-search index

# Index using custom location
./bleve-search --index ./my-help-index index
```

### Basic Search

Search across all fields using full-text search:

```bash
# Search for "template" in all fields
./bleve-search search "template"

# Limit results to 5 entries
./bleve-search search --limit 5 "template"

# Use boolean operators
./bleve-search search "template AND format"
./bleve-search search "template OR format"
./bleve-search search "template NOT json"
```

### Field-Specific Search

Search with field filters:

```bash
# Search for "format" in GeneralTopic sections
./bleve-search field-search --type GeneralTopic "format"

# Search for "example" in top-level sections
./bleve-search field-search --top-level "example"

# Combine filters
./bleve-search field-search --type Example --top-level --limit 3 "format"
```

## Search Query Syntax

The search command supports Bleve's query string syntax:

- Field-specific search: `Title:template`
- Phrase search: `"exact phrase"`
- Required terms: `+required`
- Excluded terms: `-excluded`
- Fuzzy search: `template~2`
- Boosting: `template^2`

Examples:

```bash
# Search for "template" in title with higher relevance
./bleve-search search "Title:template^2"

# Search for exact phrase "help system" in content
./bleve-search search "Content:\"help system\""

# Search with fuzzy matching
./bleve-search search "templat~2"
``` 