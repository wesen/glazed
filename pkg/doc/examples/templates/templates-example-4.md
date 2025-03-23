---
Title: Use named templates from command definitions
Slug: templates-named
Short: |
  ```
  glaze examples template-demo --glazed-template-name markdown
  ```
Topics:
- templates
Commands:
- examples
Flags:
- glazed-template-name
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
Glazed commands can define their own named templates, allowing users to select
different output formats using the `--glazed-template-name` flag.

The example command below has three built-in templates: "default", "json", and "markdown".
You can choose between them using the template name flag:

```
❯ glaze examples template-demo
# Template Demo Results
- Name: Item 1, Value: 100
- Name: Item 2, Value: 200
- Name: Item 3, Value: 300

❯ glaze examples template-demo --glazed-template-name json
{
  "results": [
    {
      "name": "Item 1",
      "value": 100
    },
    {
      "name": "Item 2",
      "value": 200
    },
    {
      "name": "Item 3",
      "value": 300
    }
  ]
}

❯ glaze examples template-demo --glazed-template-name markdown
# Template Demo Results

| Name | Value |
|------|-------|
| Item 1 | 100 |
| Item 2 | 200 |
| Item 3 | 300 |
```

This feature allows commands to provide multiple output formats without requiring users
to specify template content manually. Command authors can define sensible defaults while
still allowing users to override with custom templates if needed.