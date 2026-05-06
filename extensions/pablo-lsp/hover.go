package main

import (
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentHover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	docURI := params.TextDocument.URI
	content, ok := documents[docURI]
	if !ok {
		return nil, nil
	}

	lines := strings.Split(content, "\n")
	if params.Position.Line >= uint32(len(lines)) {
		return nil, nil
	}

	// Get word at position
	line := lines[params.Position.Line]
	word := getWordAt(line, int(params.Position.Character))
	if word == "" {
		return nil, nil
	}

	path := getYAMLPath(lines, int(params.Position.Line), int(params.Position.Character))

	// If path currently ends with the word we found, it's a value or the key itself
	// The path from getYAMLPath for a value already includes the key.
	// We want to find the field definition for that key.

	field := getFieldAtPath(PabloSchema, path)
	if field == nil {
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: field.Description,
		},
	}, nil
}

func getWordAt(line string, char int) string {
	if char >= len(line) {
		return ""
	}

	// Simple word boundary check
	start := char
	for start > 0 && isWordChar(line[start-1]) {
		start--
	}

	end := char
	for end < len(line) && isWordChar(line[end]) {
		end++
	}

	if start == end {
		return ""
	}
	return line[start:end]
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
}
