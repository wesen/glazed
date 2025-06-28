# Exercise: Advanced Log Analyzer

Create a "log-analyzer" command with sophisticated data processing capabilities.

## Data Processing Requirements

1. **Log Format Support**:
   - Parse Apache access logs
   - Parse Nginx logs  
   - Parse custom application logs
   - Handle malformed log entries gracefully

2. **Data Extraction**:
   - IP addresses and geographic distribution
   - HTTP status codes and error rates
   - Response times and performance metrics
   - User agents (browsers, bots, mobile devices)
   - Most requested URLs and endpoints

## Analysis Features

1. **Traffic Patterns**:
   - Requests by hour/day/week
   - Peak traffic identification
   - Geographic distribution (IP-based)
   - User agent analysis

2. **Error Analysis**:
   - 4xx client error analysis
   - 5xx server error tracking
   - Error rate trends
   - Most common error pages

3. **Performance Analysis**:
   - Slow request identification
   - Response time percentiles
   - Bandwidth usage patterns
   - Cache hit/miss rates

## Custom Templates

Create templates for different report types:

1. **Security Report**:
   - Potential attack patterns
   - Suspicious IP addresses
   - Bot traffic analysis
   - Failed authentication attempts

2. **Performance Report**:
   - Slowest endpoints
   - Peak traffic times
   - Resource usage patterns
   - Optimization recommendations

3. **Summary Dashboard**:
   - Key metrics overview
   - Traffic trends
   - Error rate summary
   - Performance indicators

## Advanced Features

1. **Streaming Analysis**:
   - Real-time processing of growing log files
   - Tail-like functionality with analysis
   - Live dashboard updates

2. **Anomaly Detection**:
   - Unusual traffic patterns
   - Sudden error rate spikes
   - Performance degradation alerts
   - Security threat indicators

3. **Export Options**:
   - Multiple output formats (JSON, CSV, HTML reports)
   - Integration with monitoring systems
   - Historical trend analysis

## Example Usage

```bash
# Basic analysis
log-analyzer analyze access.log --format apache

# Security focus
log-analyzer analyze access.log --template security --filter-status 4xx,5xx

# Performance analysis
log-analyzer analyze access.log --template performance --time-window 1h

# Real-time monitoring
log-analyzer monitor access.log --template dashboard --follow

# Custom analysis
log-analyzer analyze access.log --custom-template security-report.tmpl --output html
```

## Sample Data

Create test data representing realistic web server logs with:
- Normal traffic patterns
- Security incidents (attempted attacks)
- Performance issues (slow responses)
- Various user agents and geographic sources

## Testing

Test your implementation with:
- Large log files (performance testing)
- Malformed log entries (error handling)
- Different log formats
- Real-time log streaming
- Template customization

This exercise combines everything learned in the first three tutorials!
