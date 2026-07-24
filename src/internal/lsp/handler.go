package lsp

import (
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type serverHandler struct {
	*protocol.Handler
}

func (h *serverHandler) Handle(context *glsp.Context) (r any, validMethod bool, validParams bool, err error) {
	if context.Method == methodListProfiles {
		return handleListProfiles(context)
	}
	return h.Handler.Handle(context)
}
