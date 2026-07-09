package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"pablo/pkg/inspect"

	"github.com/tliron/glsp"
)

const methodListProfiles = "pablo/listProfiles"

type listProfilesParams struct {
	URI string `json:"uri"`
}

func handleListProfiles(context *glsp.Context) (any, bool, bool, error) {
	var params listProfilesParams
	if err := json.Unmarshal(context.Params, &params); err != nil {
		return nil, true, false, err
	}
	if params.URI == "" {
		return nil, true, false, fmt.Errorf("uri is required")
	}

	baseDir := baseDirFromURI(params.URI)
	var data []byte
	var err error

	if doc, ok := documents[params.URI]; ok {
		data = []byte(doc.content)
		if doc.baseDir != "" {
			baseDir = doc.baseDir
		}
	} else {
		filePath, pathErr := filePathFromURI(params.URI)
		if pathErr != nil {
			return nil, true, false, pathErr
		}
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, true, true, err
		}
	}

	result, err := inspect.FromYAML(data, baseDir)
	if err != nil {
		return nil, true, true, err
	}
	return result, true, true, nil
}
