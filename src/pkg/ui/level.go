package ui

import (
	"errors"
	"os"
)

// Output verbosity for CLI chrome. Default is normal.
type Level int

const (
	LevelQuiet Level = iota - 1
	LevelNormal
	LevelVerbose
)

var outputLevel = LevelNormal

// Configure applies --quiet / --verbose (and optional env overrides).
// Quiet and verbose are mutually exclusive at the call site; if both end up
// true, quiet wins for chrome.
func Configure(quiet, verbose bool) {
	if !quiet {
		quiet = envTruthy(os.Getenv("PABLO_QUIET"))
	}
	if !verbose {
		verbose = envTruthy(os.Getenv("PABLO_VERBOSE"))
	}
	switch {
	case quiet:
		outputLevel = LevelQuiet
	case verbose:
		outputLevel = LevelVerbose
	default:
		outputLevel = LevelNormal
	}
}

// Quiet reports whether chrome should be minimized.
func Quiet() bool {
	return outputLevel <= LevelQuiet
}

// Verbose reports whether extra detail (artifact paths, cwd, hook commands) is on.
func Verbose() bool {
	return outputLevel >= LevelVerbose
}

// LevelOf returns the active output level.
func LevelOf() Level {
	return outputLevel
}

// loggedError marks an error whose failure was already shown via Log/Result.
// main skips reprinting these on stderr.
type loggedError struct {
	err error
}

func (e *loggedError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *loggedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Logged wraps err so callers that already printed the failure can avoid a
// duplicate stderr line from main.
func Logged(err error) error {
	if err == nil {
		return nil
	}
	var existing *loggedError
	if errors.As(err, &existing) {
		return err
	}
	return &loggedError{err: err}
}

// IsLogged reports whether err (or a wrapped cause) was already presented to the user.
func IsLogged(err error) bool {
	var logged *loggedError
	return errors.As(err, &logged)
}
