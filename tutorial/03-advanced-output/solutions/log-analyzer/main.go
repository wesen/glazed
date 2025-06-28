package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
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

// LogAnalyzerCommand processes web server logs
type LogAnalyzerCommand struct {
	*cmds.CommandDescription
}

type LogAnalysisSettings struct {
	LogFile      string   `glazed.parameter:"log-file"`
	LogFormat    string   `glazed.parameter:"log-format"`
	Analysis     []string `glazed.parameter:"analysis"`
	Template     string   `glazed.parameter:"report-template"`
	FilterStatus string   `glazed.parameter:"filter-status"`
	TimeRange    string   `glazed.parameter:"time-range"`
	Limit        int      `glazed.parameter:"limit"`
	GroupBy      string   `glazed.parameter:"group-by"`
}

// LogEntry represents a parsed log entry
type LogEntry struct {
	IP        string
	Timestamp time.Time
	Method    string
	URL       string
	Status    int
	Size      int
	UserAgent string
	Referer   string
	Duration  float64
}

var _ cmds.GlazeCommand = &LogAnalyzerCommand{}

func (c *LogAnalyzerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &LogAnalysisSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Load and parse log entries
	entries, err := c.loadLogEntries(settings.LogFile, settings.LogFormat)
	if err != nil {
		return err
	}

	// Apply filters
	entries = c.filterEntries(entries, settings)

	// Perform analysis
	for _, analysisType := range settings.Analysis {
		switch strings.ToLower(analysisType) {
		case "traffic":
			if err := c.analyzeTraffic(ctx, gp, entries, settings); err != nil {
				return err
			}
		case "errors":
			if err := c.analyzeErrors(ctx, gp, entries, settings); err != nil {
				return err
			}
		case "performance":
			if err := c.analyzePerformance(ctx, gp, entries, settings); err != nil {
				return err
			}
		case "security":
			if err := c.analyzeSecurity(ctx, gp, entries, settings); err != nil {
				return err
			}
		case "geography":
			if err := c.analyzeGeography(ctx, gp, entries, settings); err != nil {
				return err
			}
		case "useragents":
			if err := c.analyzeUserAgents(ctx, gp, entries, settings); err != nil {
				return err
			}
		case "summary":
			if err := c.generateSummary(ctx, gp, entries, settings); err != nil {
				return err
			}
		default:
			fmt.Printf("Unknown analysis type: %s\n", analysisType)
		}
	}

	return nil
}

func (c *LogAnalyzerCommand) loadLogEntries(filename, format string) ([]LogEntry, error) {
	if filename == "" {
		return c.generateSampleLogEntries(), nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if entry, err := c.parseLogLine(line, format); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries, scanner.Err()
}

func (c *LogAnalyzerCommand) parseLogLine(line, format string) (LogEntry, error) {
	var entry LogEntry
	
	// Apache Common Log Format regex
	apacheRegex := regexp.MustCompile(`(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*)" (\d+) (\d+)`)
	
	// Extended format with user agent and referer
	extendedRegex := regexp.MustCompile(`(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]*)" (\d+) (\d+) "([^"]*)" "([^"]*)"`)
	
	var matches []string
	
	switch format {
	case "apache", "common":
		matches = apacheRegex.FindStringSubmatch(line)
		if len(matches) < 7 {
			return entry, fmt.Errorf("invalid log format")
		}
	case "extended", "combined":
		matches = extendedRegex.FindStringSubmatch(line)
		if len(matches) < 9 {
			return entry, fmt.Errorf("invalid log format")
		}
	default:
		// Try extended first, then apache
		matches = extendedRegex.FindStringSubmatch(line)
		if len(matches) < 9 {
			matches = apacheRegex.FindStringSubmatch(line)
			if len(matches) < 7 {
				return entry, fmt.Errorf("invalid log format")
			}
		}
	}

	entry.IP = matches[1]
	
	// Parse timestamp
	timeStr := matches[2]
	if timestamp, err := time.Parse("02/Jan/2006:15:04:05 -0700", timeStr); err == nil {
		entry.Timestamp = timestamp
	}
	
	entry.Method = matches[3]
	entry.URL = matches[4]
	
	if status, err := strconv.Atoi(matches[5]); err == nil {
		entry.Status = status
	}
	
	if size, err := strconv.Atoi(matches[6]); err == nil {
		entry.Size = size
	}
	
	// Extended format fields
	if len(matches) >= 9 {
		entry.Referer = matches[7]
		entry.UserAgent = matches[8]
	}
	
	return entry, nil
}

func (c *LogAnalyzerCommand) generateSampleLogEntries() []LogEntry {
	entries := []LogEntry{
		{IP: "192.168.1.1", Timestamp: time.Now().Add(-1 * time.Hour), Method: "GET", URL: "/", Status: 200, Size: 1234, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
		{IP: "192.168.1.2", Timestamp: time.Now().Add(-55 * time.Minute), Method: "GET", URL: "/api/users", Status: 200, Size: 2345, UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"},
		{IP: "192.168.1.3", Timestamp: time.Now().Add(-50 * time.Minute), Method: "POST", URL: "/login", Status: 401, Size: 123, UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"},
		{IP: "10.0.0.1", Timestamp: time.Now().Add(-45 * time.Minute), Method: "GET", URL: "/admin", Status: 403, Size: 456, UserAgent: "curl/7.68.0"},
		{IP: "192.168.1.1", Timestamp: time.Now().Add(-40 * time.Minute), Method: "GET", URL: "/products", Status: 200, Size: 3456, UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X)"},
		{IP: "203.0.113.1", Timestamp: time.Now().Add(-35 * time.Minute), Method: "GET", URL: "/wp-admin", Status: 404, Size: 789, UserAgent: "Mozilla/5.0 (compatible; Googlebot/2.1)"},
		{IP: "192.168.1.4", Timestamp: time.Now().Add(-30 * time.Minute), Method: "GET", URL: "/slow-page", Status: 200, Size: 4567, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", Duration: 2.5},
		{IP: "198.51.100.1", Timestamp: time.Now().Add(-25 * time.Minute), Method: "POST", URL: "/api/upload", Status: 413, Size: 0, UserAgent: "Python-requests/2.25.1"},
		{IP: "192.168.1.2", Timestamp: time.Now().Add(-20 * time.Minute), Method: "GET", URL: "/images/logo.png", Status: 200, Size: 12345, UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"},
		{IP: "10.0.0.2", Timestamp: time.Now().Add(-15 * time.Minute), Method: "GET", URL: "/", Status: 500, Size: 1000, UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
	}
	
	return entries
}

func (c *LogAnalyzerCommand) filterEntries(entries []LogEntry, settings *LogAnalysisSettings) []LogEntry {
	var filtered []LogEntry
	
	for _, entry := range entries {
		// Filter by status if specified
		if settings.FilterStatus != "" {
			statusRange := strings.Split(settings.FilterStatus, "-")
			if len(statusRange) == 2 {
				min, _ := strconv.Atoi(statusRange[0])
				max, _ := strconv.Atoi(statusRange[1])
				if entry.Status < min || entry.Status > max {
					continue
				}
			} else {
				status, _ := strconv.Atoi(settings.FilterStatus)
				if entry.Status != status {
					continue
				}
			}
		}
		
		// TODO: Add time range filtering based on settings.TimeRange
		
		filtered = append(filtered, entry)
	}
	
	return filtered
}

func (c *LogAnalyzerCommand) analyzeTraffic(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	// Group by hour, day, or other time periods
	timeStats := make(map[string]int)
	
	for _, entry := range entries {
		var key string
		switch settings.GroupBy {
		case "hour":
			key = entry.Timestamp.Format("2006-01-02 15:00")
		case "day":
			key = entry.Timestamp.Format("2006-01-02")
		case "week":
			year, week := entry.Timestamp.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		default:
			key = entry.Timestamp.Format("2006-01-02 15:00")
		}
		
		timeStats[key]++
	}
	
	// Convert to sorted slice
	type TrafficStat struct {
		Period   string
		Requests int
	}
	
	var stats []TrafficStat
	for period, count := range timeStats {
		stats = append(stats, TrafficStat{
			Period:   period,
			Requests: count,
		})
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Period < stats[j].Period
	})
	
	for _, stat := range stats {
		row := types.NewRow(
			types.MRP("analysis_type", "traffic"),
			types.MRP("period", stat.Period),
			types.MRP("requests", stat.Requests),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	return nil
}

func (c *LogAnalyzerCommand) analyzeErrors(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	errorStats := make(map[int]int)
	errorURLs := make(map[string]int)
	
	for _, entry := range entries {
		if entry.Status >= 400 {
			errorStats[entry.Status]++
			errorURLs[entry.URL]++
		}
	}
	
	// Output error status counts
	for status, count := range errorStats {
		row := types.NewRow(
			types.MRP("analysis_type", "error_status"),
			types.MRP("status_code", status),
			types.MRP("count", count),
			types.MRP("error_type", c.getErrorType(status)),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	// Output top error URLs
	type URLError struct {
		URL   string
		Count int
	}
	
	var urlErrors []URLError
	for url, count := range errorURLs {
		urlErrors = append(urlErrors, URLError{URL: url, Count: count})
	}
	
	sort.Slice(urlErrors, func(i, j int) bool {
		return urlErrors[i].Count > urlErrors[j].Count
	})
	
	limit := settings.Limit
	if limit == 0 || limit > len(urlErrors) {
		limit = len(urlErrors)
	}
	
	for i := 0; i < limit; i++ {
		row := types.NewRow(
			types.MRP("analysis_type", "error_url"),
			types.MRP("url", urlErrors[i].URL),
			types.MRP("error_count", urlErrors[i].Count),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	return nil
}

func (c *LogAnalyzerCommand) analyzePerformance(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	var totalSize int64
	var slowRequests []LogEntry
	urlSizes := make(map[string][]int)
	
	for _, entry := range entries {
		totalSize += int64(entry.Size)
		urlSizes[entry.URL] = append(urlSizes[entry.URL], entry.Size)
		
		// Consider requests with duration > 1s as slow (if duration is available)
		if entry.Duration > 1.0 {
			slowRequests = append(slowRequests, entry)
		}
	}
	
	// Calculate average response sizes
	for url, sizes := range urlSizes {
		if len(sizes) > 0 {
			total := 0
			for _, size := range sizes {
				total += size
			}
			avgSize := total / len(sizes)
			
			row := types.NewRow(
				types.MRP("analysis_type", "performance"),
				types.MRP("url", url),
				types.MRP("avg_response_size", avgSize),
				types.MRP("request_count", len(sizes)),
			)
			
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	
	// Output slow requests
	for _, entry := range slowRequests {
		row := types.NewRow(
			types.MRP("analysis_type", "slow_request"),
			types.MRP("url", entry.URL),
			types.MRP("duration", entry.Duration),
			types.MRP("timestamp", entry.Timestamp.Format(time.RFC3339)),
			types.MRP("ip", entry.IP),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	return nil
}

func (c *LogAnalyzerCommand) analyzeSecurity(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	suspiciousIPs := make(map[string]int)
	attackPatterns := make(map[string]int)
	
	// Define suspicious patterns
	suspiciousPatterns := []string{
		"wp-admin", "wp-login", "admin", "phpmyadmin", "mysql",
		"../", "etc/passwd", "eval(", "script>", "union select",
		"1=1", "or 1=1", "drop table", "insert into",
	}
	
	for _, entry := range entries {
		// Track IPs with many failed requests
		if entry.Status == 401 || entry.Status == 403 {
			suspiciousIPs[entry.IP]++
		}
		
		// Check for attack patterns in URLs
		urlLower := strings.ToLower(entry.URL)
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(urlLower, pattern) {
				attackPatterns[pattern]++
				suspiciousIPs[entry.IP]++
			}
		}
	}
	
	// Output suspicious IPs
	type SuspiciousIP struct {
		IP    string
		Count int
	}
	
	var suspIPs []SuspiciousIP
	for ip, count := range suspiciousIPs {
		if count >= 3 { // Threshold for suspicious activity
			suspIPs = append(suspIPs, SuspiciousIP{IP: ip, Count: count})
		}
	}
	
	sort.Slice(suspIPs, func(i, j int) bool {
		return suspIPs[i].Count > suspIPs[j].Count
	})
	
	for _, suspIP := range suspIPs {
		row := types.NewRow(
			types.MRP("analysis_type", "security_threat"),
			types.MRP("ip_address", suspIP.IP),
			types.MRP("suspicious_requests", suspIP.Count),
			types.MRP("threat_level", c.getThreatLevel(suspIP.Count)),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	// Output attack patterns
	for pattern, count := range attackPatterns {
		if count > 0 {
			row := types.NewRow(
				types.MRP("analysis_type", "attack_pattern"),
				types.MRP("pattern", pattern),
				types.MRP("occurrences", count),
			)
			
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (c *LogAnalyzerCommand) analyzeGeography(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	// Simple IP-based geographic analysis (would need GeoIP database for real implementation)
	ipCounts := make(map[string]int)
	
	for _, entry := range entries {
		ipCounts[entry.IP]++
	}
	
	for ip, count := range ipCounts {
		country := c.getCountryFromIP(ip)
		
		row := types.NewRow(
			types.MRP("analysis_type", "geography"),
			types.MRP("ip_address", ip),
			types.MRP("country", country),
			types.MRP("request_count", count),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	return nil
}

func (c *LogAnalyzerCommand) analyzeUserAgents(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	uaCounts := make(map[string]int)
	browsers := make(map[string]int)
	bots := make(map[string]int)
	
	for _, entry := range entries {
		if entry.UserAgent != "" {
			uaCounts[entry.UserAgent]++
			
			// Simple browser detection
			ua := strings.ToLower(entry.UserAgent)
			if strings.Contains(ua, "chrome") {
				browsers["Chrome"]++
			} else if strings.Contains(ua, "firefox") {
				browsers["Firefox"]++
			} else if strings.Contains(ua, "safari") {
				browsers["Safari"]++
			} else if strings.Contains(ua, "bot") || strings.Contains(ua, "crawler") {
				bots[entry.UserAgent]++
			}
		}
	}
	
	// Output browser statistics
	for browser, count := range browsers {
		row := types.NewRow(
			types.MRP("analysis_type", "browser"),
			types.MRP("browser_name", browser),
			types.MRP("request_count", count),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	// Output bot statistics
	for bot, count := range bots {
		row := types.NewRow(
			types.MRP("analysis_type", "bot"),
			types.MRP("bot_name", bot),
			types.MRP("request_count", count),
		)
		
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	
	return nil
}

func (c *LogAnalyzerCommand) generateSummary(ctx context.Context, gp middlewares.Processor, entries []LogEntry, settings *LogAnalysisSettings) error {
	if len(entries) == 0 {
		return nil
	}
	
	totalRequests := len(entries)
	uniqueIPs := make(map[string]bool)
	statusCounts := make(map[int]int)
	var totalBytes int64
	
	for _, entry := range entries {
		uniqueIPs[entry.IP] = true
		statusCounts[entry.Status]++
		totalBytes += int64(entry.Size)
	}
	
	// Calculate error rate
	errorRequests := 0
	for status, count := range statusCounts {
		if status >= 400 {
			errorRequests += count
		}
	}
	
	errorRate := float64(errorRequests) / float64(totalRequests) * 100
	
	// Generate summary based on template
	if settings.Template == "dashboard" {
		return c.generateDashboardSummary(ctx, gp, totalRequests, len(uniqueIPs), errorRate, totalBytes)
	}
	
	// Standard summary
	row := types.NewRow(
		types.MRP("analysis_type", "summary"),
		types.MRP("total_requests", totalRequests),
		types.MRP("unique_ips", len(uniqueIPs)),
		types.MRP("error_rate_percent", fmt.Sprintf("%.2f", errorRate)),
		types.MRP("total_bytes", totalBytes),
		types.MRP("avg_request_size", totalBytes/int64(totalRequests)),
	)
	
	return gp.AddRow(ctx, row)
}

func (c *LogAnalyzerCommand) generateDashboardSummary(ctx context.Context, gp middlewares.Processor, totalRequests, uniqueIPs int, errorRate float64, totalBytes int64) error {
	dashboardTemplate := `
🌐 Web Server Log Analysis Dashboard
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Traffic Overview:
   • Total Requests: {{.TotalRequests}}
   • Unique Visitors: {{.UniqueIPs}}
   • Error Rate: {{printf "%.2f" .ErrorRate}}%
   • Data Transferred: {{.TotalMB}} MB

{{if gt .ErrorRate 5.0}}⚠️  High error rate detected!{{end}}
{{if gt .TotalRequests 10000}}📈 High traffic volume{{end}}
`
	
	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return err
	}
	
	data := struct {
		TotalRequests int
		UniqueIPs     int
		ErrorRate     float64
		TotalMB       float64
	}{
		TotalRequests: totalRequests,
		UniqueIPs:     uniqueIPs,
		ErrorRate:     errorRate,
		TotalMB:       float64(totalBytes) / 1024 / 1024,
	}
	
	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return err
	}
	
	row := types.NewRow(
		types.MRP("analysis_type", "dashboard_summary"),
		types.MRP("content", output.String()),
		types.MRP("format", "dashboard"),
	)
	
	return gp.AddRow(ctx, row)
}

// Helper functions
func (c *LogAnalyzerCommand) getErrorType(status int) string {
	switch {
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}

func (c *LogAnalyzerCommand) getThreatLevel(count int) string {
	switch {
	case count >= 20:
		return "HIGH"
	case count >= 10:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func (c *LogAnalyzerCommand) getCountryFromIP(ip string) string {
	// Simple heuristic - in real implementation, use GeoIP database
	if net.ParseIP(ip) == nil {
		return "Invalid IP"
	}
	
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "Local Network"
	}
	
	// Mock country detection based on IP ranges
	if strings.HasPrefix(ip, "203.") {
		return "Australia"
	} else if strings.HasPrefix(ip, "198.") {
		return "United States"
	}
	
	return "Unknown"
}

func NewLogAnalyzerCommand() (*LogAnalyzerCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"analyze-logs",
		cmds.WithShort("Analyze web server logs"),
		cmds.WithLong(`
Advanced log analysis tool for web server logs (Apache, Nginx).

Analysis Types:
- traffic: Request patterns over time
- errors: Error analysis and troubleshooting
- performance: Response times and sizes
- security: Threat detection and suspicious activity
- geography: Geographic distribution of requests
- useragents: Browser and bot analysis
- summary: Overall statistics

Examples:
  analyze-logs --analysis traffic,errors --group-by hour
  analyze-logs --analysis security --template dashboard
  analyze-logs --log-file access.log --analysis summary,performance
  analyze-logs --analysis errors --filter-status 400-499 --limit 10
		`),

		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"log-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to log file (uses sample data if not specified)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"log-format",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log format type"),
				parameters.WithChoices("apache", "nginx", "common", "combined", "extended", "auto"),
				parameters.WithDefault("auto"),
			),
			parameters.NewParameterDefinition(
				"analysis",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Types of analysis to perform"),
				parameters.WithDefault([]string{"summary"}),
			),
			parameters.NewParameterDefinition(
				"report-template",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Output template"),
				parameters.WithChoices("standard", "dashboard", "security", "performance"),
				parameters.WithDefault("standard"),
			),
			parameters.NewParameterDefinition(
				"filter-status",
				parameters.ParameterTypeString,
				parameters.WithHelp("Filter by status code (e.g., '404' or '400-499')"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"time-range",
				parameters.ParameterTypeString,
				parameters.WithHelp("Time range filter (e.g., '1h', '24h', '7d')"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"group-by",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Group results by time period"),
				parameters.WithChoices("hour", "day", "week", "month"),
				parameters.WithDefault("hour"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of results to show"),
				parameters.WithDefault(10),
			),
		),

		cmds.WithLayersList(glazedLayer),
	)

	return &LogAnalyzerCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "log-analyzer",
		Short: "Advanced web server log analysis tool",
	}

	analyzeCmd, err := NewLogAnalyzerCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(analyzeCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
