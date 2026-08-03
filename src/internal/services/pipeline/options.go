package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"pablo/pkg/domain"
	"pablo/pkg/ui"
)

// RunOptions controls presentation and optional machine-readable output for run.
type RunOptions struct {
	AllowProtected bool
	Verbose        bool
	JSONSummary    bool
}

// RunSummary is emitted as a single JSON object when --json-summary is set.
// Printed to stdout after the human Result line (or alone under --quiet).
type RunSummary struct {
	Project    string            `json:"project"`
	Version    string            `json:"version"`
	Profile    string            `json:"profile,omitempty"`
	Env        string            `json:"env,omitempty"`
	Type       string            `json:"type,omitempty"`
	Mode       string            `json:"mode,omitempty"`
	Sequence   string            `json:"sequence,omitempty"`
	Paths      map[string]string `json:"paths,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	OK         bool              `json:"ok"`
}

func failReturn(start time.Time, err error) error {
	resultFail(start)
	return ui.Logged(err)
}

func emitJSONSummary(summary RunSummary) error {
	encoded, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, string(encoded))
	return nil
}

func environmentMode(env domain.Environment) string {
	if env.Remote != nil {
		return "remote"
	}
	return "local"
}
