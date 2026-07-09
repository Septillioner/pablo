package lsp

import (
	"pablo/pkg/schema"
	"pablo/pkg/validate"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	_ "github.com/tliron/go-kutil/terminal"
)

const serverName = "pablo"

var serverVersion string
var handler protocol.Handler

func RunStdio(version string) error {
	serverVersion = version
	handler = protocol.Handler{
		Initialize:             initialize,
		Initialized:            initialized,
		Shutdown:               shutdown,
		SetTrace:               setTrace,
		TextDocumentDidOpen:    textDocumentDidOpen,
		TextDocumentDidChange:  textDocumentDidChange,
		TextDocumentDidSave:    textDocumentDidSave,
		TextDocumentCompletion: textDocumentCompletion,
		TextDocumentHover:      textDocumentHover,
	}

	s := server.NewServer(&handler, serverName, false)
	return s.RunStdio()
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := protocol.ServerCapabilities{
		TextDocumentSync: protocol.TextDocumentSyncKindFull,
		CompletionProvider: &protocol.CompletionOptions{
			TriggerCharacters: []string{":", " "},
		},
		HoverProvider: true,
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    serverName,
			Version: &serverVersion,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

func publishValidation(context *glsp.Context, uri, content, baseDir string) {
	diags, _, _ := validate.ValidateYAML([]byte(content), baseDir)
	context.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: toProtocolDiagnostics(diags),
	})
}

func toProtocolDiagnostics(diags []validate.Diagnostic) []protocol.Diagnostic {
	result := make([]protocol.Diagnostic, 0, len(diags))
	source := serverName
	for _, d := range diags {
		severity := protocol.DiagnosticSeverityError
		if d.Severity == validate.SeverityWarning {
			severity = protocol.DiagnosticSeverityWarning
		}

		endLine := uint32(d.EndLine - 1)
		endColumn := uint32(d.EndColumn - 1)
		if endLine < uint32(d.Line-1) {
			endLine = uint32(d.Line - 1)
		}
		if endColumn <= uint32(d.Column-1) {
			endColumn = uint32(d.Column)
		}

		result = append(result, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(d.Line - 1),
					Character: uint32(d.Column - 1),
				},
				End: protocol.Position{
					Line:      endLine,
					Character: endColumn,
				},
			},
			Severity: &severity,
			Source:   &source,
			Message:  d.Message,
		})
	}
	return result
}

var kindProperty = protocol.CompletionItemKindProperty
var kindEnumMember = protocol.CompletionItemKindEnumMember

func strPtr(s string) *string {
	return &s
}

func getFieldAtPath(path []string) *schema.Field {
	return schema.GetFieldAtPath(path)
}
