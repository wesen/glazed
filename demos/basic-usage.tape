Output demos/basic-usage.gif

Require echo
Require go

Set FontSize 18
Set Width 1200
Set Height 600

Type "# Glazed - Structured Data Output Framework"
Sleep 1500ms
Enter

Type "# Let's see the glaze tool in action"
Sleep 1000ms
Enter 2

Type "# First, let's see what commands are available"
Sleep 500ms
Enter
Type "go run ./cmd/glaze --help"
Sleep 500ms
Enter
Sleep 3000ms

Type "# First, let's create some test data"
Sleep 1000ms
Enter
Type "mkdir -p testdata"
Sleep 500ms
Enter
Type `echo '{"id": 1, "name": "Alice", "email": "alice@example.com", "role": "admin"}' > testdata/user1.json`
Sleep 500ms
Enter
Type `echo '{"id": 2, "name": "Bob", "email": "bob@example.com", "role": "user"}' > testdata/user2.json`
Sleep 500ms
Enter

Type "# Now let's process the JSON data"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze json testdata/user1.json testdata/user2.json"
Sleep 500ms
Enter
Sleep 2000ms

Type "# Let's try different output formats"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze json testdata/user1.json --output json"
Sleep 500ms
Enter
Sleep 2000ms

Type "go run ./cmd/glaze json testdata/user1.json testdata/user2.json --output csv"
Sleep 500ms
Enter
Sleep 2000ms

Type "# We can also select specific fields"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze json testdata/user1.json testdata/user2.json --fields name,role"
Sleep 500ms
Enter
Sleep 2000ms

Type "# And get comprehensive help"
Sleep 1000ms
Enter
Type "go run ./cmd/glaze help help-system"
Sleep 500ms
Enter
Sleep 4000ms

Type "# That's glazed - rich CLI tools made easy!"
Sleep 2000ms
