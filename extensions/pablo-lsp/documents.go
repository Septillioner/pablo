package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var documents = make(map[string]string)

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	documents[params.TextDocument.URI] = params.TextDocument.Text
	validate(context, params.TextDocument.URI, params.TextDocument.Text)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) > 0 {
		// With TextDocumentSyncKindFull, we always get the whole text in the last change
		lastChange := params.ContentChanges[len(params.ContentChanges)-1]
		switch c := lastChange.(type) {
		case protocol.TextDocumentContentChangeEvent:
			documents[params.TextDocument.URI] = c.Text
			validate(context, params.TextDocument.URI, c.Text)
		case protocol.TextDocumentContentChangeEventWhole:
			documents[params.TextDocument.URI] = c.Text
			validate(context, params.TextDocument.URI, c.Text)
		}
	}
	return nil
}

func textDocumentDidSave(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func validate(context *glsp.Context, uri string, content string) {
	diagnostics := validateYAML(content)
	context.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}
