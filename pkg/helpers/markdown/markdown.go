package markdown

import (
	"bufio"
	"strings"
)

type State int

const (
	OutsideBlock State = iota
	InsideBlock
)

type BlockType string

const (
	Normal BlockType = "Normal"
	Code   BlockType = "Code"
)

type MarkdownBlock struct {
	Type     BlockType
	Language string
	Content  string
}

// ExtractAllBlocks extracts blocks enclosed by ``` using a state machine.
func ExtractAllBlocks(input string) []MarkdownBlock {
	var result []MarkdownBlock
	state := OutsideBlock
	var blockLines []string

	scanner := bufio.NewScanner(strings.NewReader(input))
	language := ""
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case OutsideBlock:
			if strings.HasPrefix(line, "```") {
				state = InsideBlock
				language = strings.TrimPrefix(line, "```")
				blockLines = nil // reset blockLines
			} else if len(blockLines) > 0 {
				// For normal blocks
				result = append(result, MarkdownBlock{Type: Normal, Content: strings.Join(blockLines, "\n")})
				blockLines = nil // reset blockLines
			}
			blockLines = append(blockLines, line)
		case InsideBlock:
			if strings.HasPrefix(line, "```") {
				state = OutsideBlock
				if len(blockLines) > 2 {
					content := strings.Join(blockLines[1:], "\n") // excluding the language line
					result = append(result, MarkdownBlock{Type: Code, Language: language, Content: content})
				}
				blockLines = nil // reset blockLines
				language = ""
			} else {
				blockLines = append(blockLines, line)
			}
		}
	}

	// Handle any remaining normal content outside of code blocks
	if len(blockLines) > 0 && state == OutsideBlock {
		result = append(result, MarkdownBlock{Type: Normal, Content: strings.Join(blockLines, "\n")})
	}

	return result
}

func ExtractQuotedBlocks(input string, withQuotes bool) []string {
	blocks := ExtractAllBlocks(input)
	var result []string
	for _, block := range blocks {
		if block.Type != Code {
			continue
		}
		if withQuotes {
			result = append(result, "```"+block.Language+"\n"+block.Content+"\n```")
		} else {
			result = append(result, block.Content)
		}
	}

	return result
}
