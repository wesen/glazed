package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
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

type TemplatedAnalyzerCommand struct {
	*cmds.CommandDescription
}

type TemplateSettings struct {
	ReportType     string `glazed.parameter:"report-type"`
	TemplateName   string `glazed.parameter:"report-template"`
	IncludeCharts  bool   `glazed.parameter:"include-charts"`
	ShowMetadata   bool   `glazed.parameter:"show-metadata"`
	CustomTemplate string `glazed.parameter:"custom-template"`
}

var _ cmds.GlazeCommand = &TemplatedAnalyzerCommand{}

func (c *TemplatedAnalyzerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &TemplateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Generate sample report data
	reportData := c.generateReportData(settings.ReportType)

	switch settings.TemplateName {
	case "executive":
		return c.generateExecutiveReport(ctx, gp, reportData, settings)
	case "detailed":
		return c.generateDetailedReport(ctx, gp, reportData, settings)
	case "dashboard":
		return c.generateDashboardReport(ctx, gp, reportData, settings)
	case "custom":
		return c.generateCustomReport(ctx, gp, reportData, settings)
	default:
		return c.generateStandardReport(ctx, gp, reportData, settings)
	}
}

type ReportData struct {
	Title       string
	Summary     map[string]interface{}
	Metrics     []Metric
	Trends      []Trend
	Insights    []string
	Metadata    map[string]string
	GeneratedAt string
}

type Metric struct {
	Name        string
	Value       float64
	Unit        string
	Change      float64
	Status      string
	Description string
}

type Trend struct {
	Period string
	Value  float64
	Change float64
}

func (c *TemplatedAnalyzerCommand) generateReportData(reportType string) *ReportData {
	data := &ReportData{
		Title: fmt.Sprintf("%s Analysis Report", reportType),
		Summary: map[string]interface{}{
			"total_revenue":    125000.50,
			"total_orders":     1250,
			"avg_order_value":  100.00,
			"customer_count":   450,
			"growth_rate":      12.5,
		},
		Metrics: []Metric{
			{Name: "Revenue", Value: 125000.50, Unit: "USD", Change: 15.2, Status: "up", Description: "Total revenue this period"},
			{Name: "Orders", Value: 1250, Unit: "count", Change: 8.5, Status: "up", Description: "Total number of orders"},
			{Name: "Conversion Rate", Value: 3.2, Unit: "%", Change: -2.1, Status: "down", Description: "Website conversion rate"},
			{Name: "Customer Satisfaction", Value: 4.7, Unit: "/5", Change: 0.3, Status: "up", Description: "Average customer rating"},
		},
		Trends: []Trend{
			{Period: "Week 1", Value: 25000, Change: 5.0},
			{Period: "Week 2", Value: 28000, Change: 12.0},
			{Period: "Week 3", Value: 32000, Change: 14.3},
			{Period: "Week 4", Value: 40000, Change: 25.0},
		},
		Insights: []string{
			"Revenue growth is accelerating, up 15.2% from last period",
			"Mobile conversion rates need improvement",
			"Customer satisfaction is at an all-time high",
			"Premium segment showing strongest growth",
		},
		Metadata: map[string]string{
			"period":       "Q3 2023",
			"analyst":      "AI System",
			"data_quality": "95%",
			"last_updated": "2023-09-15",
		},
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	return data
}

func (c *TemplatedAnalyzerCommand) generateExecutiveReport(ctx context.Context, gp middlewares.Processor, data *ReportData, settings *TemplateSettings) error {
	// Executive summary template
	summaryTemplate := `
EXECUTIVE SUMMARY
================
{{.Title}}
Generated: {{.GeneratedAt}}

Key Metrics:
- Revenue: ${{printf "%.2f" .Summary.total_revenue}} (↑{{printf "%.1f" .Summary.growth_rate}}%)
- Orders: {{.Summary.total_orders}}
- Customers: {{.Summary.customer_count}}
- Avg Order: ${{printf "%.2f" .Summary.avg_order_value}}

Top Insights:
{{range $i, $insight := .Insights}}{{add $i 1}}. {{$insight}}
{{end}}`

	tmpl, err := template.New("executive").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(summaryTemplate)
	if err != nil {
		return err
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("report_type", "executive"),
		types.MRP("title", data.Title),
		types.MRP("content", output.String()),
		types.MRP("format", "text"),
	)

	return gp.AddRow(ctx, row)
}

func (c *TemplatedAnalyzerCommand) generateDetailedReport(ctx context.Context, gp middlewares.Processor, data *ReportData, settings *TemplateSettings) error {
	// Generate detailed rows for each metric
	for _, metric := range data.Metrics {
		statusIcon := "📈"
		if metric.Status == "down" {
			statusIcon = "📉"
		} else if metric.Status == "stable" {
			statusIcon = "➡️"
		}

		row := types.NewRow(
			types.MRP("report_type", "detailed"),
			types.MRP("metric_name", metric.Name),
			types.MRP("value", metric.Value),
			types.MRP("unit", metric.Unit),
			types.MRP("change_percent", metric.Change),
			types.MRP("status", metric.Status),
			types.MRP("status_icon", statusIcon),
			types.MRP("description", metric.Description),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Add trend data
	for _, trend := range data.Trends {
		row := types.NewRow(
			types.MRP("report_type", "trend"),
			types.MRP("period", trend.Period),
			types.MRP("value", trend.Value),
			types.MRP("change_percent", trend.Change),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *TemplatedAnalyzerCommand) generateDashboardReport(ctx context.Context, gp middlewares.Processor, data *ReportData, settings *TemplateSettings) error {
	// Dashboard-style compact format
	dashboardTemplate := `
📊 {{.Title}} Dashboard
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💰 Revenue: ${{printf "%.0f" .Summary.total_revenue}} {{if gt .Summary.growth_rate 0.0}}↗️ +{{printf "%.1f" .Summary.growth_rate}}%{{else}}↘️ {{printf "%.1f" .Summary.growth_rate}}%{{end}}
📦 Orders: {{.Summary.total_orders}}
👥 Customers: {{.Summary.customer_count}}
💳 Avg Order: ${{printf "%.2f" .Summary.avg_order_value}}

🔍 Key Insights:
{{range .Insights}}• {{.}}
{{end}}

📈 Trending Metrics:
{{range .Metrics}}{{if eq .Status "up"}}📈{{else if eq .Status "down"}}📉{{else}}➡️{{end}} {{.Name}}: {{printf "%.1f" .Value}}{{.Unit}} ({{if gt .Change 0.0}}+{{end}}{{printf "%.1f" .Change}}%)
{{end}}`

	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return err
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("report_type", "dashboard"),
		types.MRP("title", data.Title),
		types.MRP("content", output.String()),
		types.MRP("format", "dashboard"),
		types.MRP("generated_at", data.GeneratedAt),
	)

	if settings.ShowMetadata {
		row.Set("metadata", data.Metadata)
	}

	return gp.AddRow(ctx, row)
}

func (c *TemplatedAnalyzerCommand) generateCustomReport(ctx context.Context, gp middlewares.Processor, data *ReportData, settings *TemplateSettings) error {
	if settings.CustomTemplate == "" {
		return fmt.Errorf("custom template is required when using custom report type")
	}

	tmpl, err := template.New("custom").Parse(settings.CustomTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse custom template: %w", err)
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return fmt.Errorf("failed to execute custom template: %w", err)
	}

	row := types.NewRow(
		types.MRP("report_type", "custom"),
		types.MRP("title", data.Title),
		types.MRP("content", output.String()),
		types.MRP("format", "custom"),
	)

	return gp.AddRow(ctx, row)
}

func (c *TemplatedAnalyzerCommand) generateStandardReport(ctx context.Context, gp middlewares.Processor, data *ReportData, settings *TemplateSettings) error {
	// Standard tabular format
	row := types.NewRow(
		types.MRP("report_type", "standard"),
		types.MRP("title", data.Title),
		types.MRP("total_revenue", data.Summary["total_revenue"]),
		types.MRP("total_orders", data.Summary["total_orders"]),
		types.MRP("avg_order_value", data.Summary["avg_order_value"]),
		types.MRP("customer_count", data.Summary["customer_count"]),
		types.MRP("growth_rate", data.Summary["growth_rate"]),
		types.MRP("generated_at", data.GeneratedAt),
	)

	return gp.AddRow(ctx, row)
}

func NewTemplatedAnalyzerCommand() (*TemplatedAnalyzerCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"report",
		cmds.WithShort("Generate templated analysis reports"),
		cmds.WithLong(`
Generate analysis reports using custom templates and formatters.

Template Types:
- standard: Basic tabular format
- executive: High-level summary for executives
- detailed: Comprehensive metrics breakdown
- dashboard: Visual dashboard-style layout
- custom: User-defined template

Examples:
  report --template executive --report-type sales
  report --template dashboard --show-metadata --output json
  report --template detailed --include-charts
  report --template custom --custom-template "Revenue: {{.Summary.total_revenue}}"
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"report-type",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Type of report to generate"),
				parameters.WithChoices("sales", "marketing", "operations", "financial"),
				parameters.WithDefault("sales"),
			),
			parameters.NewParameterDefinition(
				"report-template",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Report template to use"),
				parameters.WithChoices("standard", "executive", "detailed", "dashboard", "custom"),
				parameters.WithDefault("standard"),
			),
			parameters.NewParameterDefinition(
				"include-charts",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include chart data in report"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"show-metadata",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include metadata in output"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"custom-template",
				parameters.ParameterTypeString,
				parameters.WithHelp("Custom Go template string (required for custom template type)"),
				parameters.WithDefault(""),
			),
		),

		cmds.WithLayersList(glazedLayer),
	)

	return &TemplatedAnalyzerCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "templated-analyzer",
		Short: "Analysis with custom templates and formatters",
	}

	reportCmd, err := NewTemplatedAnalyzerCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(reportCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
