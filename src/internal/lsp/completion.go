package lsp

import (
	"strings"

	"pablo/pkg/schema"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func textDocumentCompletion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	docURI := params.TextDocument.URI
	doc, ok := documents[docURI]
	if !ok {
		return nil, nil
	}

	lines := strings.Split(doc.content, "\n")
	if params.Position.Line >= uint32(len(lines)) {
		return nil, nil
	}

	path := schema.GetYAMLPath(lines, int(params.Position.Line), int(params.Position.Character))
	field := getFieldAtPath(path)

	if field == nil || field.Children == nil {
		if len(path) > 0 {
			parentPath := path[:len(path)-1]
			parentField := getFieldAtPath(parentPath)
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
