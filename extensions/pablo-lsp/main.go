package main

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
	_ "github.com/tliron/go-kutil/terminal"
)

const lsName = "pablo-lsp"

var version = "0.0.1"
var handler protocol.Handler

func main() {
	// Disable logging to stdout to avoid corrupting LSP protocol
	// commonlog.Configure(1, nil)
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

	s := server.NewServer(&handler, lsName, false)
	s.RunStdio()
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
			Name:    lsName,
			Version: &version,
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
