package main

import (
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var kindProperty = protocol.CompletionItemKindProperty
var kindEnumMember = protocol.CompletionItemKindEnumMember

func strPtr(s string) *string {
	return &s
}
