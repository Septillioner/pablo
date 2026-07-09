package lsp

import (
	"strings"

	"pablo/pkg/schema"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentHover(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	docURI := params.TextDocument.URI
	doc, ok := documents[docURI]
	if !ok {
		return nil, nil
	}

	lines := strings.Split(doc.content, "\n")
	if params.Position.Line >= uint32(len(lines)) {
		return nil, nil
	}

	line := lines[params.Position.Line]
	word := getWordAt(line, int(params.Position.Character))
	if word == "" {
		return nil, nil
	}

	path := schema.GetYAMLPath(lines, int(params.Position.Line), int(params.Position.Character))
	field := getFieldAtPath(path)
	if field == nil || field.Description == "" {
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
