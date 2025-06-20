package help

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"strings"


	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed static/*
var staticFiles embed.FS

// AddServeCommand adds the "serve" subcommand to the help command
func (hs *HelpSystem) AddServeCommand(helpCmd *cobra.Command) {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a web server for browsing help",
		Long:  "Start a web server that provides a nice web UI for browsing the help system",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			return hs.StartWebServer(port)
		},
	}

	serveCmd.Flags().IntP("port", "p", 8080, "Port to serve on")
	helpCmd.AddCommand(serveCmd)
}

func (hs *HelpSystem) StartWebServer(port int) error {
	server := &WebServer{
		HelpSystem: hs,
		Port:       port,
	}

	return server.Start()
}

type WebServer struct {
	HelpSystem *HelpSystem
	Port       int
}

func (ws *WebServer) Start() error {
	mux := http.NewServeMux()

	// Static files
	staticHandler := http.FileServer(http.FS(staticFiles))
	mux.Handle("/static/", staticHandler)

	// Routes
	mux.HandleFunc("/", ws.handleIndex)
	mux.HandleFunc("/section/", ws.handleSection)
	mux.HandleFunc("/search", ws.handleSearch)
	mux.HandleFunc("/api/sections", ws.handleAPISections)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.Port),
		Handler: mux,
	}

	log.Info().Int("port", ws.Port).Msg("Starting web server")
	fmt.Printf("Help web UI available at http://localhost:%d\n", ws.Port)

	return server.ListenAndServe()
}

func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	helpPage := ws.HelpSystem.GetTopLevelHelpPage()
	
	data := map[string]interface{}{
		"Title":       "Help System",
		"HelpPage":    helpPage,
		"HelpSystem":  ws.HelpSystem,
	}

	component := IndexPage(data)
	component.Render(context.Background(), w)
}

func (ws *WebServer) handleSection(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[len("/section/"):]
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	section, err := ws.HelpSystem.GetSectionWithSlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":   section.Title,
		"Section": section,
	}

	component := SectionPage(data)
	component.Render(context.Background(), w)
}

func (ws *WebServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	sectionType := r.URL.Query().Get("type")
	
	var results []*Section
	
	if query != "" {
		// Simple search implementation
		qb := NewSectionQuery().ReturnAllTypes()
		
		if sectionType != "" {
			switch sectionType {
			case "topics":
				qb = qb.ReturnTopics()
			case "examples":
				qb = qb.ReturnExamples()
			case "applications":
				qb = qb.ReturnApplications()
			case "tutorials":
				qb = qb.ReturnTutorials()
			}
		}
		
		// Search by title, content, or slug
		for _, section := range ws.HelpSystem.Sections {
			queryLower := strings.ToLower(query)
			if strings.Contains(strings.ToLower(section.Title), queryLower) ||
				strings.Contains(strings.ToLower(section.Content), queryLower) ||
				strings.Contains(strings.ToLower(section.Slug), queryLower) ||
				strings.Contains(strings.ToLower(section.Short), queryLower) {
				
				// Apply type filter
				if sectionType == "" || ws.sectionMatchesType(section, sectionType) {
					results = append(results, section)
				}
			}
		}
	}

	data := map[string]interface{}{
		"Title":   "Search Results",
		"Query":   query,
		"Type":    sectionType,
		"Results": results,
	}

	component := SearchPage(data)
	component.Render(context.Background(), w)
}

func (ws *WebServer) sectionMatchesType(section *Section, sectionType string) bool {
	switch sectionType {
	case "topics":
		return section.SectionType == SectionGeneralTopic
	case "examples":
		return section.SectionType == SectionExample
	case "applications":
		return section.SectionType == SectionApplication
	case "tutorials":
		return section.SectionType == SectionTutorial
	default:
		return true
	}
}

func (ws *WebServer) handleAPISections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Simple JSON response for AJAX requests
	fmt.Fprintf(w, `{"count": %d}`, len(ws.HelpSystem.Sections))
}
