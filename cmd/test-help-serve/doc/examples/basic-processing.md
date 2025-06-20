---
Title: Basic File Processing
Slug: basic-processing
SectionType: Example
Short: Simple example of processing a data file
Topics:
  - examples
  - processing
Commands:
  - process
Flags:
  - input
  - output
ShowPerDefault: true
Order: 10
---

# Basic File Processing

This example shows how to process a simple JSON file.

## Input File (data.json)

```json
{
  "users": [
    {"name": "Alice", "age": 30},
    {"name": "Bob", "age": 25}
  ]
}
```

## Command

```bash
test-help process --input data.json --output processed.json --verbose
```

## Expected Output

The command will process the input file and save the results to `processed.json`.

```json
{
  "processed_users": [
    {"name": "Alice", "age": 30, "processed": true},
    {"name": "Bob", "age": 25, "processed": true}
  ],
  "metadata": {
    "processed_at": "2024-01-01T12:00:00Z",
    "total_records": 2
  }
}
```

## Tips

- Use `--verbose` to see detailed processing information
- Check the output file exists before running the command
- Large files may take longer to process
