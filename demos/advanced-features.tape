Output demos/advanced-features.gif

Require echo
Require go

Set FontSize 16
Set Width 1200
Set Height 700

Type "# Glazed Advanced Features Demo"
Sleep 1500ms
Enter

Type "# Create complex test data"
Sleep 1000ms
Enter
Type "mkdir -p testdata"
Sleep 500ms
Enter
Type `echo '{"id": 1, "name": "Alice", "details": {"role": "admin", "dept": "IT"}}' > testdata/complex1.json`
Sleep 500ms
Enter
Type `echo '{"id": 2, "name": "Bob", "details": {"role": "user", "dept": "Sales"}}' > testdata/complex2.json`
Sleep 500ms
Enter

Type "# Multiple JSON objects with complex data"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze json testdata/complex1.json testdata/complex2.json"
Sleep 500ms
Enter
Sleep 3000ms

Type "# Flattening nested structures"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze json testdata/complex1.json --flatten"
Sleep 500ms
Enter
Sleep 2500ms

Type "# Create YAML test data"
Sleep 1000ms
Enter
Type `echo 'name: glazed
version: 2.0
features:
  - commands
  - parameters
  - help-system' > testdata/config.yaml`
Sleep 500ms
Enter

Type "# YAML processing with field selection"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze yaml testdata/config.yaml --fields name,version --output json"
Sleep 500ms
Enter
Sleep 2500ms

Type "# Create CSV test data"
Sleep 1000ms
Enter
Type `echo 'id,name,role
1,Alice,admin
2,Bob,user
3,Carol,editor' > testdata/users.csv`
Sleep 500ms
Enter

Type "# CSV input processing"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze csv testdata/users.csv --output yaml"
Sleep 500ms
Enter
Sleep 2500ms

Type "# Get help on specific topics"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze help --list"
Sleep 500ms
Enter
Sleep 3000ms

Type "# Glazed - Powerful, flexible, easy to use!"
Sleep 2000ms
