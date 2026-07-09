package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type document struct {
	content string
	baseDir string
}

var documents = make(map[string]document)

func textDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	baseDir := baseDirFromURI(params.TextDocument.URI)
	documents[params.TextDocument.URI] = document{
		content: params.TextDocument.Text,
		baseDir: baseDir,
	}
	publishValidation(context, params.TextDocument.URI, params.TextDocument.Text, baseDir)
	return nil
}

func textDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	lastChange := params.ContentChanges[len(params.ContentChanges)-1]
	baseDir := baseDirFromURI(params.TextDocument.URI)

	switch c := lastChange.(type) {
	case protocol.TextDocumentContentChangeEvent:
		documents[params.TextDocument.URI] = document{content: c.Text, baseDir: baseDir}
		publishValidation(context, params.TextDocument.URI, c.Text, baseDir)
	case protocol.TextDocumentContentChangeEventWhole:
		documents[params.TextDocument.URI] = document{content: c.Text, baseDir: baseDir}
		publishValidation(context, params.TextDocument.URI, c.Text, baseDir)
	}

	return nil
}

func textDocumentDidSave(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	return nil
}

func baseDirFromURI(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "."
	}

	path := parsed.Path
	if strings.HasPrefix(parsed.Scheme, "file") && len(path) > 0 {
		if len(path) > 3 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		return filepath.Dir(path)
	}

	return "."
}

func filePathFromURI(uri string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	path := parsed.Path
	if strings.HasPrefix(parsed.Scheme, "file") && len(path) > 0 {
		if len(path) > 3 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		return path, nil
	}

	return "", fmt.Errorf("unsupported uri scheme: %s", parsed.Scheme)
}
