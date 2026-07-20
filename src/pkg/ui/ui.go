package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// Terminal style tokens. Brand accent is hi-cyan (dark-friendly); avoid magenta/purple glow.
const (
	progressBarWidth   = 24
	sectionRuleWidth   = 36
	markColumnWidth    = 4
	spinnerInterval    = 80 * time.Millisecond
	progressPulseSteps = 3
)

var (
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	WarnColor    = color.New(color.FgYellow).SprintFunc()
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	ActionColor  = color.New(color.FgHiWhite).SprintFunc()
	BoldColor    = color.New(color.Bold).SprintFunc()
	GrayColor    = color.New(color.FgHiBlack).SprintFunc()
	ThemeColor   = color.New(color.FgHiCyan).SprintFunc()
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var (
	// chromeMu serializes sticky rail, spinner, and progress-bar writes so
	// cursor moves and \r redraws cannot interleave (especially on Windows).
	chromeMu sync.Mutex

	activeMu      sync.Mutex
	activeSpinner *Spinner
	progressPulse int
	// progressLive is true while an incomplete ProgressBar owns the \r line.
	// Guarded by chromeMu.
	progressLive bool
)

// Interactive reports whether animated chrome should run.
// Quiet when stdout is not a TTY, NO_COLOR is set, CI=1, or PABLO_PLAIN is set.
func Interactive() bool {
	if color.NoColor {
		return false
	}
	if envTruthy(os.Getenv("CI")) {
		return false
	}
	if envTruthy(os.Getenv("PABLO_PLAIN")) {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func envTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// Header prints a compact brand block with a light accent rule.
func Header(version string) {
	wordmark := ThemeColor(`    ┓ ┓
┏┓┏┓┣┓┃┏┓
┣┛┗┻┗┛┗┗┛
┛`)
	meta := fmt.Sprintf("%s  %s", BoldColor("pablo"), GrayColor("v"+version))
	accent := ThemeColor("─") + GrayColor(strings.Repeat("─", 18)) + ThemeColor("─")
	fmt.Println(wordmark)
	fmt.Printf("  %s\n", meta)
	fmt.Printf("  %s\n\n", accent)
}

// Log prints a structured status line: time, aligned mark, message.
func Log(mark string, message string) {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	clearActiveSpinnerLineLocked()
	progressLive = false
	timestamp := GrayColor(time.Now().Format("15:04:05"))
	fmt.Printf("%s  %s  %s\n", timestamp, formatMark(mark), message)
	noteRailContentLines(1)
}

func formatMark(mark string) string {
	switch mark {
	case "+":
		return SuccessColor(padMark("ok"))
	case "-":
		return ErrorColor(padMark("fail"))
	case "!":
		return WarnColor(padMark("warn"))
	case "*":
		return InfoColor(padMark("info"))
	case ">":
		return ActionColor(padMark("run"))
	case " ":
		return GrayColor(padMark("·"))
	default:
		return padMark(mark)
	}
}

func padMark(mark string) string {
	if len(mark) >= markColumnWidth {
		return mark
	}
	return mark + strings.Repeat(" ", markColumnWidth-len(mark))
}

// Section prints a light titled divider (title stays readable; no shouty uppercase).
func Section(title string) {
	label := strings.TrimSpace(title)
	if label == "" {
		return
	}

	chromeMu.Lock()
	defer chromeMu.Unlock()

	clearActiveSpinnerLineLocked()
	progressLive = false
	fmt.Println()
	prefix := ThemeColor("─")
	fmt.Printf("%s %s\n", prefix, BoldColor(label))
	fmt.Println(GrayColor(strings.Repeat("─", sectionRuleWidth)))
	noteRailContentLines(3)
}

// Result prints a single-line outcome without heavy separator bars.
func Result(success bool, duration time.Duration) {
	chromeMu.Lock()
	defer chromeMu.Unlock()

	clearActiveSpinnerLineLocked()
	progressLive = false
	fmt.Println()
	elapsed := GrayColor(duration.Round(time.Millisecond).String())
	if success {
		fmt.Printf("%s  %s  %s\n", SuccessColor(padMark("ok")), BoldColor("done"), elapsed)
	} else {
		fmt.Printf("%s  %s  %s\n", ErrorColor(padMark("fail")), BoldColor("failed"), elapsed)
	}
	noteRailContentLines(2)
}

// ProgressBar prints a compact in-place progress bar with a pulsing tip while incomplete.
// No-op when not Interactive (callers should keep surrounding Log lines).
func ProgressBar(percent int, label string) {
	if !Interactive() {
		return
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	chromeMu.Lock()
	defer chromeMu.Unlock()

	clearActiveSpinnerLineLocked()

	filled := percent * progressBarWidth / 100
	var bar string
	if percent >= 100 {
		bar = ThemeColor(strings.Repeat("█", progressBarWidth))
	} else {
		tipFrames := []string{"▒", "▓", "█"}
		tip := tipFrames[progressPulse%progressPulseSteps]
		progressPulse++
		solid := filled
		if solid > progressBarWidth-1 {
			solid = progressBarWidth - 1
		}
		rest := progressBarWidth - solid - 1
		bar = ThemeColor(strings.Repeat("█", solid)+tip) + GrayColor(strings.Repeat("░", rest))
	}

	fmt.Printf("\r%s  %s  %s %3d%%", GrayColor(label), bar, GrayColor("·"), percent)
	if percent >= 100 {
		progressLive = false
		fmt.Println()
		noteRailContentLines(1)
	} else {
		progressLive = true
	}
}

// FileProgress returns a callback that drives ProgressBar from done/total counts.
func FileProgress(label string) func(done, total int) {
	return func(done, total int) {
		if total <= 0 {
			return
		}
		percent := done * 100 / total
		if done >= total {
			percent = 100
		}
		ProgressBar(percent, label)
	}
}

// Spinner is a braille-frame in-place status indicator for indeterminate work.
type Spinner struct {
	label string
	inert bool

	mu       sync.Mutex
	message  string
	stopCh   chan struct{}
	doneCh   chan struct{}
	stopped  bool
	frameIdx int
}

// StartSpinner begins an animated spinner when Interactive; otherwise logs once.
func StartSpinner(label string) *Spinner {
	label = strings.TrimSpace(label)
	s := &Spinner{label: label, message: label}
	if !Interactive() {
		if label != "" {
			Log(">", label)
		}
		s.inert = true
		return s
	}

	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	setActiveSpinner(s)
	go s.loop()
	return s
}

// Update changes the spinner label without restarting the animation.
func (s *Spinner) Update(label string) {
	if s == nil || s.inert {
		return
	}
	s.mu.Lock()
	s.message = strings.TrimSpace(label)
	s.mu.Unlock()
}

// Stop clears the spinner line. Safe to call more than once.
func (s *Spinner) Stop() {
	if s == nil || s.inert {
		return
	}

	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	close(s.stopCh)
	s.mu.Unlock()

	<-s.doneCh

	chromeMu.Lock()
	clearLine()
	clearActiveSpinner(s)
	chromeMu.Unlock()
}

// WithSpinner runs fn while a spinner is active. Stops the spinner before returning.
func WithSpinner(label string, fn func() error) error {
	spin := StartSpinner(label)
	err := fn()
	spin.Stop()
	return err
}

func (s *Spinner) loop() {
	defer close(s.doneCh)
	ticker := time.NewTicker(spinnerInterval)
	defer ticker.Stop()

	s.render()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			s.frameIdx = (s.frameIdx + 1) % len(spinnerFrames)
			s.mu.Unlock()
			s.render()
		}
	}
}

func (s *Spinner) render() {
	chromeMu.Lock()
	defer chromeMu.Unlock()
	s.renderLocked()
}

// renderLocked writes the spinner live line. Caller must hold chromeMu.
func (s *Spinner) renderLocked() {
	if s == nil || s.inert {
		return
	}
	s.mu.Lock()
	stopped := s.stopped
	frame := spinnerFrames[s.frameIdx%len(spinnerFrames)]
	msg := s.message
	s.mu.Unlock()
	if stopped {
		return
	}
	fmt.Printf("\r%s  %s  %s", ThemeColor(frame), ActionColor(padMark("run")), GrayColor(msg))
}

func setActiveSpinner(s *Spinner) {
	activeMu.Lock()
	prev := activeSpinner
	activeSpinner = s
	activeMu.Unlock()
	if prev != nil && prev != s {
		prev.Stop()
	}
}

func clearActiveSpinner(s *Spinner) {
	activeMu.Lock()
	defer activeMu.Unlock()
	if activeSpinner == s {
		activeSpinner = nil
	}
}

func spinnerIsActive() bool {
	activeMu.Lock()
	s := activeSpinner
	activeMu.Unlock()
	if s == nil || s.inert {
		return false
	}
	s.mu.Lock()
	stopped := s.stopped
	s.mu.Unlock()
	return !stopped
}

// liveLineBusy reports whether a spinner or incomplete progress bar currently
// owns the in-place (\r) line. Rail pulse must not redraw while this is true.
func liveLineBusy() bool {
	if spinnerIsActive() {
		return true
	}
	chromeMu.Lock()
	live := progressLive
	chromeMu.Unlock()
	return live
}

// clearActiveSpinnerLineLocked clears the spinner live line. Caller must hold chromeMu.
func clearActiveSpinnerLineLocked() {
	activeMu.Lock()
	s := activeSpinner
	activeMu.Unlock()
	if s == nil || s.inert {
		return
	}
	clearLine()
}

// repaintActiveSpinnerLocked restores the spinner after a sticky rail redraw.
// Caller must hold chromeMu.
func repaintActiveSpinnerLocked() {
	activeMu.Lock()
	s := activeSpinner
	activeMu.Unlock()
	if s == nil {
		return
	}
	s.renderLocked()
}

func clearLine() {
	fmt.Print("\r\033[K")
}
