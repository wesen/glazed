#!/bin/bash

# Script to validate the README examples and functionality

set -e

echo "🧪 Validating glazed README examples..."

# Go to the glazed directory
cd "$(dirname "$0")/.."

echo "📁 Working directory: $(pwd)"

echo "🔨 Building glaze tool..."
go build -o ./bin/glaze ./cmd/glaze

echo "🔨 Building simple command example..."
(cd examples/simple-command && go build -o ../../bin/simple-command main.go)

echo "📝 Creating test data..."
mkdir -p testdata
echo '{"id": 1, "name": "Alice", "email": "alice@example.com", "role": "admin"}' > testdata/user1.json
echo '{"id": 2, "name": "Bob", "email": "bob@example.com", "role": "user"}' > testdata/user2.json

echo "✅ Testing glaze tool basic functionality..."

# Test table output
echo "  📊 Testing table output..."
./bin/glaze json testdata/user1.json testdata/user2.json > /dev/null

# Test JSON output
echo "  📄 Testing JSON output..."
./bin/glaze json testdata/user1.json --output json > /dev/null

# Test CSV output
echo "  📊 Testing CSV output..."
./bin/glaze json testdata/user1.json testdata/user2.json --output csv > /dev/null

# Test field selection
echo "  🎯 Testing field selection..."
./bin/glaze json testdata/user1.json testdata/user2.json --fields name,role > /dev/null

echo "✅ Testing simple command example..."

# Test simple command
echo "  📊 Testing default output..."
./bin/simple-command list-users > /dev/null

# Test JSON output
echo "  📄 Testing JSON output..."
./bin/simple-command list-users --output json --count 2 > /dev/null

# Test field selection
echo "  🎯 Testing field selection..."
./bin/simple-command list-users --search admin --fields username,role > /dev/null

# Test help system
echo "  ❓ Testing help system..."
./bin/glaze help help-system > /dev/null

echo "✅ All tests passed!"
echo "🎉 README examples are working correctly"

# Clean up
rm -rf ./bin testdata

echo "🧹 Cleanup complete"
