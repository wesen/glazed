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
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Segment  string   `json:"segment"`
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
			Country:  strings.ReplaceAll(stats.Country, "-", "_"), // Safe for field names
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
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "0"
	}
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
				parameters.ParameterTypeString,
				parameters.WithHelp("Field to filter by (status, segment, country, category)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"filter-value",
				parameters.ParameterTypeString,
				parameters.WithHelp("Value to filter by"),
				parameters.WithDefault(""),
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
