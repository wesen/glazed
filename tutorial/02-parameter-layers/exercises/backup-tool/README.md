# Exercise: Backup Tool with Layered Configuration

Create a "backup-tool" command with these requirements:

## Layers

1. **Source Layer**: source directories, exclusion patterns
2. **Destination Layer**: backup location, compression options  
3. **Schedule Layer**: backup frequency, retention policy

## Configuration Loading

- Load from `backup-config.yaml`
- Override with `BACKUP_*` environment variables
- Command line flags take precedence

## Validation

- Source directories must exist
- Destination must be writable
- Retention days must be positive
- Compression level must be valid (0-9)

## Output

- Use GlazeCommand to show backup plan
- Include source paths, sizes, and estimated backup time

## Example Configuration

```yaml
# backup-config.yaml
source:
  directories: ["/home/user/documents", "/home/user/photos"]
  exclude_patterns: ["*.tmp", "*.log", ".git"]
  follow_symlinks: false

destination:
  path: "/backup/daily"
  compression: true
  compression_level: 6
  encryption: false

schedule:
  frequency: "daily"
  retention_days: 30
  max_backup_size: 107374182400  # 100GB
```

## Testing Scenarios

Test your implementation with:
- Various configuration sources and validation scenarios
- Invalid source directories
- Invalid compression levels
- Unwritable destination paths
- Environment variable overrides

Try implementing this yourself before looking at the solution!
