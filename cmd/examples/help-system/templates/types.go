package templates

import "github.com/go-go-golems/glazed/pkg/help"

// QueryResult represents the data passed to the templates
type QueryResult struct {
	HelpPage *help.HelpPage
	Query    *help.SectionQuery
}
