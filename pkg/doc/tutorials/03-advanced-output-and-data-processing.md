---
Title: Advanced Output and Data Processing
Slug: advanced-output-and-data-processing
Short: Master complex data structures, custom formatters, templates, and data transformations
Topics:
- output
- formatting
- templates
- data-processing
- tutorial
Commands:
Flags:
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Advanced Output and Data Processing

This tutorial teaches you how to handle complex data structures, create custom output formats, use templates for data transformation, and build sophisticated data processing pipelines with glazed.

## Learning Objectives

- Work with complex nested data structures
- Create custom output formatters and middleware
- Use Go templates for data transformation
- Implement data aggregation and analysis
- Build streaming data processors
- Handle large datasets efficiently

## Prerequisites

- Completed Tutorials 1 and 2
- Understanding of Go templates
- Basic knowledge of data processing concepts
- Familiarity with JSON, CSV, and other data formats

## Tutorial Overview

We'll build a "data-analyzer" tool that demonstrates:
- Processing complex nested JSON data
- Custom data transformations and aggregations
- Template-based output formatting
- Streaming data processing
- Custom middleware for data enrichment

## Setting Up

```bash
mkdir glazed-tutorial-03
cd glazed-tutorial-03
go mod init glazed-tutorial-03
go get github.com/go-go-golems/glazed
```

## Part 1: Complex Data Structures and Transformations

Let's start with a command that processes complex nested data structures.

Create `cmd/analyzer/main.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
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

// DataAnalyzerCommand processes complex data structures
type DataAnalyzerCommand struct {
	*cmds.CommandDescription
}

// Settings for data analysis
type AnalysisSettings struct {
	InputFile      string   `glazed.parameter:"input-file"`
	Analysis       []string `glazed.parameter:"analysis"`
	GroupBy        string   `glazed.parameter:"group-by"`
	FilterField    string   `glazed.parameter:"filter-field"`
	FilterValue    string   `glazed.parameter:"filter-value"`
	SortBy         string   `glazed.parameter:"sort-by"`
	Limit          int      `glazed.parameter:"limit"`
	IncludeRaw     bool     `glazed.parameter:"include-raw"`
	DateFormat     string   `glazed.parameter:"date-format"`
	PrecisionDigits int     `glazed.parameter:"precision"`
}

// Sample complex data structure
type SalesRecord struct {
	ID          string                 `json:"id"`
	Timestamp   string                 `json:"timestamp"`
	Customer    CustomerInfo           `json:"customer"`
	Products    []ProductInfo          `json:"products"`
	Total       float64                `json:"total"`
	Discount    float64                `json:"discount"`
	Tax         float64                `json:"tax"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
}

type CustomerInfo struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Segment  string  `json:"segment"`
	Location Location `json:"location"`
}

type Location struct {
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
}

type ProductInfo struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

var _ cmds.GlazeCommand = &DataAnalyzerCommand{}

func (c *DataAnalyzerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &AnalysisSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Load and parse data
	records, err := c.loadSalesData(settings.InputFile)
	if err != nil {
		return err
	}

	// Apply filters
	if settings.FilterField != "" && settings.FilterValue != "" {
		records = c.filterRecords(records, settings.FilterField, settings.FilterValue)
	}

	// Perform analysis
	for _, analysisType := range settings.Analysis {
		switch strings.ToLower(analysisType) {
		case "summary":
			if err := c.generateSummary(ctx, gp, records, settings); err != nil {
				return err
			}
		case "customer":
			if err := c.analyzeCustomers(ctx, gp, records, settings); err != nil {
				return err
			}
		case "product":
			if err := c.analyzeProducts(ctx, gp, records, settings); err != nil {
				return err
			}
		case "geography":
			if err := c.analyzeGeography(ctx, gp, records, settings); err != nil {
				return err
			}
		case "time":
			if err := c.analyzeTimePatterns(ctx, gp, records, settings); err != nil {
				return err
			}
		default:
			fmt.Printf("Unknown analysis type: %s\n", analysisType)
		}
	}

	return nil
}

func (c *DataAnalyzerCommand) loadSalesData(filename string) ([]SalesRecord, error) {
	// If no file specified, generate sample data
	if filename == "" {
		return c.generateSampleData(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var records []SalesRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (c *DataAnalyzerCommand) generateSampleData() []SalesRecord {
	customers := []CustomerInfo{
		{ID: "C001", Name: "Alice Johnson", Email: "alice@example.com", Segment: "premium",
			Location: Location{Country: "USA", State: "CA", City: "San Francisco", Zipcode: "94105"}},
		{ID: "C002", Name: "Bob Smith", Email: "bob@example.com", Segment: "standard",
			Location: Location{Country: "USA", State: "NY", City: "New York", Zipcode: "10001"}},
		{ID: "C003", Name: "Carol Davis", Email: "carol@example.com", Segment: "premium",
			Location: Location{Country: "UK", State: "England", City: "London", Zipcode: "SW1A"}},
		{ID: "C004", Name: "David Wilson", Email: "david@example.com", Segment: "basic",
			Location: Location{Country: "Canada", State: "ON", City: "Toronto", Zipcode: "M5H"}},
	}

	products := []ProductInfo{
		{SKU: "P001", Name: "Laptop Pro", Category: "Electronics", Price: 1299.99, Quantity: 1},
		{SKU: "P002", Name: "Wireless Mouse", Category: "Electronics", Price: 29.99, Quantity: 2},
		{SKU: "P003", Name: "Office Chair", Category: "Furniture", Price: 199.99, Quantity: 1},
		{SKU: "P004", Name: "Coffee Maker", Category: "Appliances", Price: 89.99, Quantity: 1},
		{SKU: "P005", Name: "Book Set", Category: "Books", Price: 49.99, Quantity: 3},
	}

	var records []SalesRecord
	for i := 0; i < 20; i++ {
		customer := customers[i%len(customers)]
		selectedProducts := []ProductInfo{products[i%len(products)]}
		
		if i%3 == 0 {
			// Add second product for some orders
			selectedProducts = append(selectedProducts, products[(i+1)%len(products)])
		}

		total := 0.0
		for _, p := range selectedProducts {
			total += p.Price * float64(p.Quantity)
		}

		discount := 0.0
		if customer.Segment == "premium" {
			discount = total * 0.1
		}

		tax := (total - discount) * 0.08
		
		timestamp := time.Now().AddDate(0, 0, -i).Format(time.RFC3339)
		
		record := SalesRecord{
			ID:        fmt.Sprintf("ORD-%03d", i+1),
			Timestamp: timestamp,
			Customer:  customer,
			Products:  selectedProducts,
			Total:     total,
			Discount:  discount,
			Tax:       tax,
			Status:    []string{"pending", "completed", "shipped"}[i%3],
			Tags:      []string{"online", "mobile", "store"}[i%3:i%3+1],
			Metadata: map[string]interface{}{
				"source":    []string{"web", "mobile", "store"}[i%3],
				"campaign":  fmt.Sprintf("CAMP-%d", (i%5)+1),
				"device":    []string{"desktop", "mobile", "tablet"}[i%3],
			},
		}

		records = append(records, record)
	}

	return records
}

func (c *DataAnalyzerCommand) filterRecords(records []SalesRecord, field, value string) []SalesRecord {
	var filtered []SalesRecord
	
	for _, record := range records {
		match := false
		
		switch strings.ToLower(field) {
		case "status":
			match = record.Status == value
		case "segment":
			match = record.Customer.Segment == value
		case "country":
			match = record.Customer.Location.Country == value
		case "category":
			for _, product := range record.Products {
				if product.Category == value {
					match = true
					break
				}
			}
		}
		
		if match {
			filtered = append(filtered, record)
		}
	}
	
	return filtered
}

func (c *DataAnalyzerCommand) generateSummary(ctx context.Context, gp middlewares.Processor, records []SalesRecord, settings *AnalysisSettings) error {
	if len(records) == 0 {
		return nil
	}

	totalRevenue := 0.0
	totalDiscount := 0.0
	totalTax := 0.0
	statusCounts := make(map[string]int)
	segmentCounts := make(map[string]int)

	for _, record := range records {
		totalRevenue += record.Total
		totalDiscount += record.Discount
		totalTax += record.Tax
		statusCounts[record.Status]++
		segmentCounts[record.Customer.Segment]++
	}

	avgOrderValue := totalRevenue / float64(len(records))
	
	row := types.NewRow(
		types.MRP("analysis_type", "summary"),
		types.MRP("total_orders", len(records)),
		types.MRP("total_revenue", c.formatFloat(totalRevenue, settings.PrecisionDigits)),
		types.MRP("total_discount", c.formatFloat(totalDiscount, settings.PrecisionDigits)),
		types.MRP("total_tax", c.formatFloat(totalTax, settings.PrecisionDigits)),
		types.MRP("avg_order_value", c.formatFloat(avgOrderValue, settings.PrecisionDigits)),
		types.MRP("net_revenue", c.formatFloat(totalRevenue-totalDiscount, settings.PrecisionDigits)),
		types.MRP("status_breakdown", c.mapToString(statusCounts)),
		types.MRP("segment_breakdown", c.mapToString(segmentCounts)),
	)

	return gp.AddRow(ctx, row)
}

func (c *DataAnalyzerCommand) analyzeCustomers(ctx context.Context, gp middlewares.Processor, records []SalesRecord, settings *AnalysisSettings) error {
	customerStats := make(map[string]struct {
		Name     string
		Email    string
		Segment  string
		Country  string
		Orders   int
		Revenue  float64
		Discount float64
	})

	for _, record := range records {
		stats := customerStats[record.Customer.ID]
		stats.Name = record.Customer.Name
		stats.Email = record.Customer.Email
		stats.Segment = record.Customer.Segment
		stats.Country = record.Customer.Location.Country
		stats.Orders++
		stats.Revenue += record.Total
		stats.Discount += record.Discount
		customerStats[record.Customer.ID] = stats
	}

	// Convert to slice for sorting
	type CustomerStat struct {
		ID       string
		Name     string
		Email    string
		Segment  string
		Country  string
		Orders   int
		Revenue  float64
		Discount float64
		AvgOrder float64
	}

	var customers []CustomerStat
	for id, stats := range customerStats {
		customers = append(customers, CustomerStat{
			ID:       id,
			Name:     stats.Name,
			Email:    stats.Email,
			Segment:  stats.Segment,
			Country:  stats.Country,
			Orders:   stats.Orders,
			Revenue:  stats.Revenue,
			Discount: stats.Discount,
			AvgOrder: stats.Revenue / float64(stats.Orders),
		})
	}

	// Sort by revenue (descending)
	sort.Slice(customers, func(i, j int) bool {
		return customers[i].Revenue > customers[j].Revenue
	})

	// Apply limit
	if settings.Limit > 0 && settings.Limit < len(customers) {
		customers = customers[:settings.Limit]
	}

	for _, customer := range customers {
		row := types.NewRow(
			types.MRP("analysis_type", "customer"),
			types.MRP("customer_id", customer.ID),
			types.MRP("name", customer.Name),
			types.MRP("email", customer.Email),
			types.MRP("segment", customer.Segment),
			types.MRP("country", customer.Country),
			types.MRP("orders", customer.Orders),
			types.MRP("revenue", c.formatFloat(customer.Revenue, settings.PrecisionDigits)),
			types.MRP("discount", c.formatFloat(customer.Discount, settings.PrecisionDigits)),
			types.MRP("avg_order", c.formatFloat(customer.AvgOrder, settings.PrecisionDigits)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DataAnalyzerCommand) analyzeProducts(ctx context.Context, gp middlewares.Processor, records []SalesRecord, settings *AnalysisSettings) error {
	productStats := make(map[string]struct {
		Name     string
		Category string
		Revenue  float64
		Quantity int
		Orders   int
	})

	for _, record := range records {
		for _, product := range record.Products {
			stats := productStats[product.SKU]
			stats.Name = product.Name
			stats.Category = product.Category
			stats.Revenue += product.Price * float64(product.Quantity)
			stats.Quantity += product.Quantity
			stats.Orders++
			productStats[product.SKU] = stats
		}
	}

	type ProductStat struct {
		SKU      string
		Name     string
		Category string
		Revenue  float64
		Quantity int
		Orders   int
		AvgPrice float64
	}

	var products []ProductStat
	for sku, stats := range productStats {
		products = append(products, ProductStat{
			SKU:      sku,
			Name:     stats.Name,
			Category: stats.Category,
			Revenue:  stats.Revenue,
			Quantity: stats.Quantity,
			Orders:   stats.Orders,
			AvgPrice: stats.Revenue / float64(stats.Quantity),
		})
	}

	// Sort by revenue
	sort.Slice(products, func(i, j int) bool {
		return products[i].Revenue > products[j].Revenue
	})

	if settings.Limit > 0 && settings.Limit < len(products) {
		products = products[:settings.Limit]
	}

	for _, product := range products {
		row := types.NewRow(
			types.MRP("analysis_type", "product"),
			types.MRP("sku", product.SKU),
			types.MRP("name", product.Name),
			types.MRP("category", product.Category),
			types.MRP("revenue", c.formatFloat(product.Revenue, settings.PrecisionDigits)),
			types.MRP("quantity", product.Quantity),
			types.MRP("orders", product.Orders),
			types.MRP("avg_price", c.formatFloat(product.AvgPrice, settings.PrecisionDigits)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DataAnalyzerCommand) analyzeGeography(ctx context.Context, gp middlewares.Processor, records []SalesRecord, settings *AnalysisSettings) error {
	geoStats := make(map[string]struct {
		Orders  int
		Revenue float64
		Customers map[string]bool
	})

	for _, record := range records {
		key := record.Customer.Location.Country
		if settings.GroupBy == "state" {
			key = fmt.Sprintf("%s-%s", record.Customer.Location.Country, record.Customer.Location.State)
		} else if settings.GroupBy == "city" {
			key = fmt.Sprintf("%s-%s-%s", 
				record.Customer.Location.Country, 
				record.Customer.Location.State, 
				record.Customer.Location.City)
		}

		stats := geoStats[key]
		if stats.Customers == nil {
			stats.Customers = make(map[string]bool)
		}
		stats.Orders++
		stats.Revenue += record.Total
		stats.Customers[record.Customer.ID] = true
		geoStats[key] = stats
	}

	type GeoStat struct {
		Location      string
		Orders        int
		Revenue       float64
		Customers     int
		AvgOrderValue float64
	}

	var geos []GeoStat
	for location, stats := range geoStats {
		geos = append(geos, GeoStat{
			Location:      location,
			Orders:        stats.Orders,
			Revenue:       stats.Revenue,
			Customers:     len(stats.Customers),
			AvgOrderValue: stats.Revenue / float64(stats.Orders),
		})
	}

	sort.Slice(geos, func(i, j int) bool {
		return geos[i].Revenue > geos[j].Revenue
	})

	for _, geo := range geos {
		row := types.NewRow(
			types.MRP("analysis_type", "geography"),
			types.MRP("location", geo.Location),
			types.MRP("orders", geo.Orders),
			types.MRP("revenue", c.formatFloat(geo.Revenue, settings.PrecisionDigits)),
			types.MRP("customers", geo.Customers),
			types.MRP("avg_order_value", c.formatFloat(geo.AvgOrderValue, settings.PrecisionDigits)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DataAnalyzerCommand) analyzeTimePatterns(ctx context.Context, gp middlewares.Processor, records []SalesRecord, settings *AnalysisSettings) error {
	timeStats := make(map[string]struct {
		Orders  int
		Revenue float64
	})

	for _, record := range records {
		timestamp, err := time.Parse(time.RFC3339, record.Timestamp)
		if err != nil {
			continue
		}

		var key string
		switch settings.GroupBy {
		case "hour":
			key = timestamp.Format("2006-01-02-15")
		case "day":
			key = timestamp.Format("2006-01-02")
		case "week":
			year, week := timestamp.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case "month":
			key = timestamp.Format("2006-01")
		default:
			key = timestamp.Format("2006-01-02")
		}

		stats := timeStats[key]
		stats.Orders++
		stats.Revenue += record.Total
		timeStats[key] = stats
	}

	type TimeStat struct {
		Period    string
		Orders    int
		Revenue   float64
		AvgOrder  float64
	}

	var times []TimeStat
	for period, stats := range timeStats {
		times = append(times, TimeStat{
			Period:   period,
			Orders:   stats.Orders,
			Revenue:  stats.Revenue,
			AvgOrder: stats.Revenue / float64(stats.Orders),
		})
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].Period < times[j].Period
	})

	for _, t := range times {
		row := types.NewRow(
			types.MRP("analysis_type", "time"),
			types.MRP("period", t.Period),
			types.MRP("orders", t.Orders),
			types.MRP("revenue", c.formatFloat(t.Revenue, settings.PrecisionDigits)),
			types.MRP("avg_order", c.formatFloat(t.AvgOrder, settings.PrecisionDigits)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DataAnalyzerCommand) formatFloat(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

func (c *DataAnalyzerCommand) mapToString(m map[string]int) string {
	var parts []string
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s:%d", k, v))
	}
	return strings.Join(parts, ", ")
}

func NewDataAnalyzerCommand() (*DataAnalyzerCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"analyze",
		cmds.WithShort("Analyze complex data structures"),
		cmds.WithLong(`
Advanced data analysis tool demonstrating complex data processing.

Analysis Types:
- summary: Overall statistics and totals
- customer: Customer-level analysis
- product: Product performance analysis
- geography: Geographic distribution analysis
- time: Time-based pattern analysis

Examples:
  analyze --analysis summary,customer --output table
  analyze --analysis product --limit 5 --sort-by revenue --output json
  analyze --analysis geography --group-by country --output csv
  analyze --analysis time --group-by month --filter-field status --filter-value completed
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"input-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("JSON file containing sales data (uses sample data if not specified)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"analysis",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Types of analysis to perform"),
				parameters.WithDefault([]string{"summary"}),
			),
			parameters.NewParameterDefinition(
				"group-by",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Grouping for geography/time analysis"),
				parameters.WithChoices("country", "state", "city", "hour", "day", "week", "month"),
				parameters.WithDefault("country"),
			),
			parameters.NewParameterDefinition(
				"filter-field",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Field to filter by"),
				parameters.WithChoices("status", "segment", "country", "category"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"filter-value",
				parameters.ParameterTypeString,
				parameters.WithHelp("Value to filter by"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"sort-by",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Field to sort results by"),
				parameters.WithChoices("revenue", "orders", "customers", "name"),
				parameters.WithDefault("revenue"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of results per analysis"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"precision",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Decimal precision for floating point numbers"),
				parameters.WithDefault(2),
			),
			parameters.NewParameterDefinition(
				"include-raw",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include raw data in output"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"date-format",
				parameters.ParameterTypeString,
				parameters.WithHelp("Date format for time analysis"),
				parameters.WithDefault("2006-01-02"),
			),
		),

		cmds.WithLayersList(glazedLayer),
	)

	return &DataAnalyzerCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "data-analyzer",
		Short: "Advanced data analysis with complex structures",
	}

	analyzerCmd, err := NewDataAnalyzerCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(analyzerCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

### Test Complex Data Analysis

```bash
# Basic summary analysis
go run cmd/analyzer/main.go analyze --analysis summary

# Multiple analysis types
go run cmd/analyzer/main.go analyze --analysis summary,customer,product --output json

# Geographic analysis with grouping
go run cmd/analyzer/main.go analyze --analysis geography --group-by state --output table

# Filtered analysis
go run cmd/analyzer/main.go analyze --analysis customer --filter-field segment --filter-value premium --limit 3

# Time analysis
go run cmd/analyzer/main.go analyze --analysis time --group-by month
```

## Part 2: Custom Templates and Formatters

Now let's add custom templates for sophisticated output formatting.

Create `cmd/templated-analyzer/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"text/template"

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
	TemplateName   string `glazed.parameter:"template"`
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
			"period":     "Q3 2023",
			"analyst":    "AI System",
			"data_quality": "95%",
			"last_updated": "2023-09-15",
		},
		GeneratedAt: "2023-09-15 14:30:00",
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
{{end}}
`

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

💰 Revenue: ${{printf "%.0f" .Summary.total_revenue}} {{if gt .Summary.growth_rate 0}}↗️ +{{printf "%.1f" .Summary.growth_rate}}%{{else}}↘️ {{printf "%.1f" .Summary.growth_rate}}%{{end}}
📦 Orders: {{.Summary.total_orders}}
👥 Customers: {{.Summary.customer_count}}
💳 Avg Order: ${{printf "%.2f" .Summary.avg_order_value}}

🔍 Key Insights:
{{range .Insights}}• {{.}}
{{end}}

📈 Trending Metrics:
{{range .Metrics}}{{if eq .Status "up"}}📈{{else if eq .Status "down"}}📉{{else}}➡️{{end}} {{.Name}}: {{printf "%.1f" .Value}}{{.Unit}} ({{if gt .Change 0}}+{{end}}{{printf "%.1f" .Change}}%)
{{end}}
`

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
				"template",
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
```

Add the missing import:

```go
import (
	// ... existing imports
	"strings"  // Add this import
)
```

### Test Template-Based Reports

```bash
# Executive summary
go run cmd/templated-analyzer/main.go report --template executive

# Dashboard view
go run cmd/templated-analyzer/main.go report --template dashboard --show-metadata

# Detailed breakdown
go run cmd/templated-analyzer/main.go report --template detailed --output csv

# Custom template
go run cmd/templated-analyzer/main.go report --template custom --custom-template "{{.Title}}: ${{.Summary.total_revenue}}"
```

## Key Concepts Learned

### Complex Data Processing
- **Nested Structures**: Handle complex JSON with embedded objects and arrays
- **Data Aggregation**: Group and summarize data across multiple dimensions
- **Filtering and Sorting**: Apply business logic to data selection

### Template Systems
- **Go Templates**: Use template/text for custom output formatting
- **Template Functions**: Add custom functions for data processing
- **Conditional Logic**: Dynamic content based on data values

### Output Customization
- **Multiple Formats**: Support different output styles for different audiences
- **Structured Data**: Maintain data structure while customizing presentation
- **Metadata**: Include contextual information in output

### Performance Considerations
- **Memory Usage**: Handle large datasets efficiently
- **Streaming**: Process data without loading everything into memory
- **Caching**: Avoid redundant calculations

## What's Next

In the next tutorial, you'll learn about:
- Production-ready CLI applications
- Error handling and logging
- Testing strategies
- Performance optimization
- Deployment considerations

## Exercise

Create a "log-analyzer" command that:

1. **Data Processing**:
   - Parse log files with multiple formats (Apache, Nginx, custom)
   - Extract IP addresses, status codes, response times, user agents
   - Handle malformed log entries gracefully

2. **Analysis Features**:
   - Traffic patterns by hour/day/week
   - Error rate analysis (4xx, 5xx responses)
   - Geographic distribution of requests (IP-based)
   - Most requested URLs and response times
   - User agent analysis (browsers, bots, mobile)

3. **Custom Templates**:
   - Security report (highlighting potential attacks)
   - Performance report (slow requests, peak times)
   - Summary dashboard with key metrics
   - Detailed technical breakdown

4. **Advanced Features**:
   - Real-time streaming analysis of growing log files
   - Anomaly detection (unusual traffic patterns)
   - Export results in multiple formats
   - Configuration file for custom log formats

Test your implementation with sample web server logs and verify the analysis results!
