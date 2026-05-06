package main

import (
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	docURI := params.TextDocument.URI
	content, ok := documents[docURI]
	if !ok {
		return nil, nil
	}

	// Simple path extraction based on line indentation
	lines := strings.Split(content, "\n")
	if params.Position.Line >= uint32(len(lines)) {
		return nil, nil
	}

	path := getYAMLPath(lines, int(params.Position.Line), int(params.Position.Character))
	field := getFieldAtPath(PabloSchema, path)

	if field == nil || field.Children == nil {
		// Check for enum completion
		if len(path) > 0 {
			parentPath := path[:len(path)-1]
			parentField := getFieldAtPath(PabloSchema, parentPath)
			if parentField != nil && parentField.Children != nil {
				key := path[len(path)-1]
				if f, ok := parentField.Children[key]; ok && f.Enum != nil {
					var items []protocol.CompletionItem
					for _, val := range f.Enum {
						items = append(items, protocol.CompletionItem{
							Label: val,
							Kind:  &kindEnumMember,
						})
					}
					return items, nil
				}
			}
		}
		return nil, nil
	}

	var items []protocol.CompletionItem
	for key, f := range field.Children {
		if key == "*" {
			continue
		}
		items = append(items, protocol.CompletionItem{
			Label:  key,
			Kind:   &kindProperty,
			Detail: strPtr(f.Description),
		})
	}

	return items, nil
}

func getYAMLPath(lines []string, line int, char int) []string {
	var path []string
	if line >= len(lines) {
		return path
	}

	currentLine := lines[line]
	// Check if we are at a key position
	isKey := !strings.Contains(currentLine[:char], ":")

	currentIndent := getIndent(currentLine)

	for i := line - 1; i >= 0; i-- {
		l := lines[i]
		if strings.TrimSpace(l) == "" || strings.HasPrefix(strings.TrimSpace(l), "#") {
			continue
		}

		indent := getIndent(l)
		if indent < currentIndent {
			key := extractKey(l)
			if key != "" {
				path = append([]string{key}, path...)
				currentIndent = indent
			}
		}
	}

	if isKey {
		// We are typing a key, path should be to the parent
		return path
	}

	// We are typing a value, path should include the current key
	key := extractKey(currentLine)
	if key != "" {
		path = append(path, key)
	}

	return path
}

func getIndent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func extractKey(line string) string {
	parts := strings.Split(line, ":")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return ""
}

func getFieldAtPath(root *Field, path []string) *Field {
	current := root
	for _, segment := range path {
		if current.Children == nil {
			return nil
		}
		if next, ok := current.Children[segment]; ok {
			current = next
		} else if wildcard, ok := current.Children["*"]; ok {
			current = wildcard
		} else {
			return nil
		}
	}
	return current
}
