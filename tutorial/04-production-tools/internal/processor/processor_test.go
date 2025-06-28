package processor

import (
	"context"
	"os"
	"strings"
	"testing"

	"tutorial-04/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessor_ProcessReader(t *testing.T) {
	cfg := config.DefaultConfig()
	processor := NewProcessor(cfg)

	tests := []struct {
		name           string
		input          string
		parser         string
		expectedLines  int64
		expectedErrors int64
	}{
		{
			name:           "valid apache logs",
			input:          `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
			parser:         "apache",
			expectedLines:  1,
			expectedErrors: 0,
		},
		{
			name:           "invalid apache logs",
			input:          `invalid log line`,
			parser:         "apache",
			expectedLines:  0,
			expectedErrors: 1,
		},
		{
			name: "mixed valid and invalid",
			input: `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
invalid line
192.168.1.1 - - [10/Oct/2000:13:55:37 -0700] "POST /form HTTP/1.0" 404 1234`,
			parser:         "apache",
			expectedLines:  2,
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := processor.parsers[tt.parser]
			require.NotNil(t, parser, "Parser not found: %s", tt.parser)

			result, err := processor.processReader(context.Background(), reader, parser)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedLines, result.LinesProcessed)
			assert.Equal(t, tt.expectedErrors, result.LinesFailed)
		})
	}
}

func TestApacheLogParser_Parse(t *testing.T) {
	parser := &ApacheLogParser{}

	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedIP  string
		expectedStatus int
	}{
		{
			name:           "valid log line",
			input:          `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`,
			expectError:    false,
			expectedIP:     "127.0.0.1",
			expectedStatus: 200,
		},
		{
			name:        "invalid log line",
			input:       `invalid log format`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parser.Parse(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, entry)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entry)
				assert.Equal(t, tt.expectedIP, entry.Fields["ip"])
				assert.Equal(t, tt.expectedStatus, entry.Fields["status"])
			}
		})
	}
}

func TestProcessor_HealthCheck(t *testing.T) {
	// Create the data directory for the default config test
	cfg := config.DefaultConfig()
	os.MkdirAll(cfg.Storage.Path, 0755)
	defer os.RemoveAll(cfg.Storage.Path)

	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      cfg,
			expectError: false,
		},
		{
			name: "invalid storage path",
			config: &config.Config{
				Storage: config.StorageConfig{
					Type: "local",
					Path: "/nonexistent/path",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(tt.config)
			err := processor.HealthCheck()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkApacheLogParser_Parse(b *testing.B) {
	parser := &ApacheLogParser{}
	logLine := `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(logLine)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessor_ProcessReader(b *testing.B) {
	cfg := config.DefaultConfig()
	processor := NewProcessor(cfg)
	parser := &ApacheLogParser{}

	// Create a larger input
	logLines := make([]string, 1000)
	for i := range logLines {
		logLines[i] = `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`
	}
	input := strings.Join(logLines, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(input)
		_, err := processor.processReader(context.Background(), reader, parser)
		if err != nil {
			b.Fatal(err)
		}
	}
}
