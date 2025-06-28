#!/bin/bash

# Script to generate demo GIFs using VHS

set -e

echo "🎬 Generating glazed demo GIFs..."

# Check if VHS is installed
if ! command -v vhs &> /dev/null; then
    echo "❌ VHS is not installed. Please install it first:"
    echo "   go install github.com/charmbracelet/vhs@latest"
    exit 1
fi

# Create demos directory if it doesn't exist
mkdir -p demos

# Go to the glazed directory
cd "$(dirname "$0")/.."

echo "📁 Working directory: $(pwd)"

# Build the glaze binary first to speed up demo recording
echo "🔨 Building glaze binary..."
go build -o ./bin/glaze ./cmd/glaze

# Create test data
echo "📝 Creating test data..."
mkdir -p testdata
echo '{"id": 1, "name": "Alice", "email": "alice@example.com", "role": "admin"}' > testdata/user1.json
echo '{"id": 2, "name": "Bob", "email": "bob@example.com", "role": "user"}' > testdata/user2.json
echo '{"id": 1, "name": "Alice", "details": {"role": "admin", "dept": "IT"}}' > testdata/complex1.json
echo '{"id": 2, "name": "Bob", "details": {"role": "user", "dept": "Sales"}}' > testdata/complex2.json
cat > testdata/config.yaml << EOF
name: glazed
version: 2.0
features:
  - commands
  - parameters
  - help-system
EOF
cat > testdata/users.csv << EOF
id,name,role
1,Alice,admin
2,Bob,user
3,Carol,editor
EOF

echo "🎥 Recording basic usage demo..."
if [ -f "demos/basic-usage.tape" ]; then
    vhs demos/basic-usage.tape
    echo "✅ Generated demos/basic-usage.gif"
else
    echo "❌ demos/basic-usage.tape not found"
fi

echo "🎥 Recording advanced features demo..."
if [ -f "demos/advanced-features.tape" ]; then
    vhs demos/advanced-features.tape
    echo "✅ Generated demos/advanced-features.gif"
else
    echo "❌ demos/advanced-features.tape not found"
fi

echo "🎉 Demo generation complete!"
echo "📁 GIF files are available in the demos/ directory"
