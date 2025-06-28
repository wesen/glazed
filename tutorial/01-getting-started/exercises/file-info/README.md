# Exercise: File Info Command

Create a "file-info" command using any command type that:

1. Accepts a file path as an argument
2. Has a flag for including hidden information
3. Outputs file size, modification time, and permissions
4. Test it with different files on your system

## Requirements

- Use any of the three command types (BareCommand, WriterCommand, or GlazeCommand)
- Handle errors gracefully (file not found, permission denied, etc.)
- Format output appropriately for the command type chosen
- Include helpful parameter descriptions

## Bonus Challenges

1. Support multiple files as arguments
2. Add filtering options (by size, date, etc.)
3. Add recursive directory processing
4. Compare different command type implementations

## Tips

- Use `os.Stat()` to get file information
- Consider using `filepath.Walk()` for recursive processing
- Remember to handle permission errors gracefully
- Think about which command type best fits your output needs

Try implementing this yourself before looking at the solutions!
