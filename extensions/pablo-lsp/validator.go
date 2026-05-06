package main

import (
	"fmt"
	"strings"

	protocol "github.com/tliron/glsp/protocol_3_16"
	"gopkg.in/yaml.v3"
)

func validateYAML(content string) []protocol.Diagnostic {
	diagnostics := []protocol.Diagnostic{}

	var node yaml.Node
	err := yaml.Unmarshal([]byte(content), &node)
	if err != nil {
		// Parse error message for line/column
		// Example: yaml: line 5: found character that cannot start any token
		line := 1
		col := 1
		msg := err.Error()

		if strings.Contains(msg, "line ") {
			fmt.Sscanf(msg, "yaml: line %d:", &line)
		}

		source := lsName
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: uint32(line - 1), Character: uint32(col - 1)},
				End:   protocol.Position{Line: uint32(line - 1), Character: uint32(col)},
			},
			Severity: &diagSeverityError,
			Source:   &source,
			Message:  msg,
		})
	}

	// TODO: Add schema-based validation here

	return diagnostics
}

var diagSeverityError = protocol.DiagnosticSeverityError
var diagSeverityWarning = protocol.DiagnosticSeverityWarning
