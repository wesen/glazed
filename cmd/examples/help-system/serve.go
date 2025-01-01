package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-go-golems/glazed/cmd/examples/help-system/templates"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

// QueryResult represents the data passed to the templates
type QueryResult struct {
	HelpPage *help.HelpPage
	Query    *help.SectionQuery
}

func createServeCommand(hs *help.HelpSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a web server to query help content",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetString("port")

			// Create handlers
			mux := http.NewServeMux()

			// Index handler
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				result := &templates.QueryResult{
					Query: help.NewSectionQuery(),
				}
				err := templates.QueryForm(result).Render(context.Background(), w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})

			// Section view handler
			mux.HandleFunc("/section/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				slug := strings.TrimPrefix(r.URL.Path, "/section/")
				if slug == "" {
					http.NotFound(w, r)
					return
				}

				section, err := hs.GetSectionWithSlug(slug)
				if err != nil {
					http.NotFound(w, r)
					return
				}

				data := &templates.SectionPageData{
					Section: section,
				}

				err = templates.SectionPage(data).Render(context.Background(), w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})

			// Query handler
			mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				err := r.ParseForm()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				query := help.NewSectionQuery()

				// Handle section types
				types := r.Form["type"]
				for _, t := range types {
					switch t {
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

				// Handle filters
				if command := r.FormValue("command"); command != "" {
					query = query.SearchForCommand(command)
				}
				if slug := r.FormValue("slug"); slug != "" {
					query = query.SearchForSlug(slug)
				}
				if r.FormValue("onlyTopLevel") == "on" {
					query = query.ReturnOnlyTopLevel()
				}
				if r.FormValue("onlyShownByDefault") == "on" {
					query = query.ReturnOnlyShownByDefault()
				}

				// If no types selected, return all types
				if len(types) == 0 {
					query = query.ReturnAllTypes()
				}

				data, noResultsFound := hs.ComputeRenderData(query)
				helpData := data["Help"].(*help.HelpPage)
				_ = noResultsFound

				result := &templates.QueryResult{
					HelpPage: helpData,
					Query:    query,
				}

				err = templates.Results(result).Render(context.Background(), w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})

			fmt.Printf("Starting server on http://localhost:%s\n", port)
			return http.ListenAndServe(":"+port, mux)
		},
	}

	cmd.Flags().String("port", "8080", "Port to listen on")
	return cmd
}
