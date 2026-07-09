package lsp

import (
	"fmt"
	"os"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"gopkg.in/yaml.v3"
)

const runWithArgsCommand = "pablo.runWithArgs"

type envLocation struct {
	profileName string
	envName     string
	keyNode     *yaml.Node
}

func textDocumentCodeLens(context *glsp.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	uri := params.TextDocument.URI
	data, err := readDocumentBytes(uri)
	if err != nil {
		return nil, err
	}

	locations, err := findEnvironmentLocations(data)
	if err != nil {
		return nil, err
	}

	lenses := make([]protocol.CodeLens, 0, len(locations))
	for _, loc := range locations {
		title := fmt.Sprintf("Run %s/%s", loc.profileName, loc.envName)
		line := uint32(loc.keyNode.Line - 1)
		col := uint32(loc.keyNode.Column - 1)
		endCol := col + uint32(len(loc.envName))

		lenses = append(lenses, protocol.CodeLens{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: endCol},
			},
			Command: &protocol.Command{
				Title:     title,
				Command:   runWithArgsCommand,
				Arguments: []any{uri, loc.profileName, loc.envName},
			},
		})
	}

	return lenses, nil
}

func readDocumentBytes(uri string) ([]byte, error) {
	if doc, ok := documents[uri]; ok {
		return []byte(doc.content), nil
	}

	filePath, err := filePathFromURI(uri)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filePath)
}

func findEnvironmentLocations(data []byte) ([]envLocation, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	doc := documentRoot(&root)
	if doc == nil {
		return nil, nil
	}

	profilesNode := mappingValue(doc, "profiles")
	if profilesNode == nil || profilesNode.Kind != yaml.MappingNode {
		return nil, nil
	}

	var locations []envLocation
	for i := 0; i+1 < len(profilesNode.Content); i += 2 {
		profileKey := profilesNode.Content[i]
		profileVal := profilesNode.Content[i+1]
		if profileVal.Kind != yaml.MappingNode {
			continue
		}

		envsNode := mappingValue(profileVal, "environments")
		if envsNode == nil || envsNode.Kind != yaml.MappingNode {
			continue
		}

		for j := 0; j+1 < len(envsNode.Content); j += 2 {
			envKey := envsNode.Content[j]
			locations = append(locations, envLocation{
				profileName: profileKey.Value,
				envName:     envKey.Value,
				keyNode:     envKey,
			})
		}
	}

	return locations, nil
}

func documentRoot(root *yaml.Node) *yaml.Node {
	if root.Kind == yaml.DocumentNode {
		for _, child := range root.Content {
			if child.Kind == yaml.MappingNode {
				return child
			}
		}
	}
	if root.Kind == yaml.MappingNode {
		return root
	}
	return nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
