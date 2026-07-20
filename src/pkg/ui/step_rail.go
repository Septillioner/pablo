package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Step rail layout and motion.
const (
	stepRailConnectorWidth = 12
	stepRailPulseInterval  = 120 * time.Millisecond
	stepRailMarkPending    = "○"
	stepRailMarkActive     = "●"
	stepRailMarkDone       = "●"
	stepRailMarkFailed     = "●"
	stepRailMarkSkipped    = "·"
	// Erase the footer line that is currently the last printed line.
	stepRailEraseFooter = "\033[1A\r\033[K"
)

type stepRailState int

const (
	stepRailPending stepRailState = iota
	stepRailActive
	stepRailDone
	stepRailFailed
	stepRailSkipped
)

// StepRail is a TTY progress footer showing pipeline or sequence phases.
// Non-interactive sessions get a static one-line snapshot (no animation).
type StepRail struct {
	labels  []string
	states  []stepRailState
	current int

	inert bool

	mu      sync.Mutex
	stopCh  chan struct{}
	doneCh  chan struct{}
	stopped bool
	frame   int
}

var (
	activeRailMu sync.Mutex
	activeRail   *StepRail

	// footerPainted and externalHold are guarded by chromeMu.
	footerPainted bool
	externalHold  int
)

// StartStepRail begins a step rail with the first label active.
// Labels should already omit phases that do not apply to the run.
func StartStepRail(labels []string) *StepRail {
	cleaned := cleanStepLabels(labels)
	if len(cleaned) == 0 {
		return &StepRail{inert: true}
	}

	r := &StepRail{
		labels:  cleaned,
		states:  make([]stepRailState, len(cleaned)),
		current: 0,
	}
	r.states[0] = stepRailActive

	if !Interactive() {
		r.inert = true
		fmt.Println(r.format(false))
		return r
	}

	chromeMu.Lock()
	clearActiveSpinnerLineLocked()
	progressLive = false
	eraseFooterLocked()
	fmt.Println(r.format(false))
	footerPainted = true
	chromeMu.Unlock()

	setActiveRail(r)

	r.stopCh = make(chan struct{})
	r.doneCh = make(chan struct{})
	go r.loop()
	return r
}

// Succeed marks the active step done and activates the next pending step.
func (r *StepRail) Succeed() {
	if r == nil || r.inert {
		return
	}
	r.mu.Lock()
	if r.stopped || r.current < 0 || r.current >= len(r.states) {
		r.mu.Unlock()
		return
	}
	if r.states[r.current] == stepRailActive {
		r.states[r.current] = stepRailDone
	}
	r.activateNextLocked()
	r.mu.Unlock()
	r.redraw()
}

// Skip marks the active step skipped and activates the next pending step.
func (r *StepRail) Skip() {
	if r == nil || r.inert {
		return
	}
	r.mu.Lock()
	if r.stopped || r.current < 0 || r.current >= len(r.states) {
		r.mu.Unlock()
		return
	}
	if r.states[r.current] == stepRailActive {
		r.states[r.current] = stepRailSkipped
	}
	r.activateNextLocked()
	r.mu.Unlock()
	r.redraw()
}

// Fail marks the active step as failed and stops pulse animation.
func (r *StepRail) Fail() {
	if r == nil || r.inert {
		return
	}
	r.mu.Lock()
	if r.stopped {
		r.mu.Unlock()
		return
	}
	if r.current >= 0 && r.current < len(r.states) && r.states[r.current] == stepRailActive {
		r.states[r.current] = stepRailFailed
	}
	r.mu.Unlock()
	r.redraw()
	r.stopPulse()
}

// Close stops animation and unregisters the rail. Safe to call more than once.
// The last footer line stays on screen as a final snapshot.
func (r *StepRail) Close() {
	if r == nil || r.inert {
		return
	}
	r.stopPulse()
	clearActiveRail(r)
}

// FailActiveStepRail marks the active step failed on the current rail, if any.
// Call immediately before Result(false) so the rail reflects failure first.
func FailActiveStepRail() {
	activeRailMu.Lock()
	r := activeRail
	activeRailMu.Unlock()
	if r != nil {
		r.Fail()
	}
}

// WithExternalOutput temporarily removes the sticky footer so subprocess
// stdout/stderr can scroll freely without erasing the rail. The footer is
// restored when fn returns. Nested calls are reference-counted.
func WithExternalOutput(fn func() error) error {
	suspendExternalOutput()
	defer resumeExternalOutput()
	return fn()
}

func suspendExternalOutput() {
	chromeMu.Lock()
	defer chromeMu.Unlock()
	clearActiveSpinnerLineLocked()
	progressLive = false
	eraseFooterLocked()
	externalHold++
}

func resumeExternalOutput() {
	chromeMu.Lock()
	defer chromeMu.Unlock()
	if externalHold > 0 {
		externalHold--
	}
	if externalHold == 0 {
		paintFooterLocked()
	}
}

func (r *StepRail) activateNextLocked() {
	for i := r.current + 1; i < len(r.states); i++ {
		if r.states[i] == stepRailPending {
			r.states[i] = stepRailActive
			r.current = i
			return
		}
	}
	r.current = len(r.states)
}

func (r *StepRail) stopPulse() {
	r.mu.Lock()
	if r.stopped || r.stopCh == nil {
		r.mu.Unlock()
		return
	}
	r.stopped = true
	close(r.stopCh)
	done := r.doneCh
	r.mu.Unlock()
	if done != nil {
		<-done
	}
}

func (r *StepRail) loop() {
	defer close(r.doneCh)
	ticker := time.NewTicker(stepRailPulseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.mu.Lock()
			r.frame++
			active := r.current >= 0 && r.current < len(r.states) && r.states[r.current] == stepRailActive
			r.mu.Unlock()
			// Spinner/progress own the live \r line; external output owns the
			// scroll region. Pulsing the footer while either is active races
			// the cursor on Windows/PowerShell.
			if active && !liveLineBusy() {
				r.redraw()
			}
		}
	}
}

func (r *StepRail) redraw() {
	if r == nil || r.inert {
		return
	}

	chromeMu.Lock()
	defer chromeMu.Unlock()

	// Keep in-memory state; paint when the live line / external stream frees.
	if externalHold > 0 || liveLineBusyLocked() {
		return
	}

	eraseFooterLocked()
	paintFooterLocked()
}

func (r *StepRail) format(pulse bool) string {
	if len(r.labels) == 0 {
		return ""
	}

	connector := GrayColor(strings.Repeat("─", stepRailConnectorWidth))
	parts := make([]string, 0, len(r.labels)*2-1)
	for i, label := range r.labels {
		parts = append(parts, r.formatStep(i, label, pulse))
		if i < len(r.labels)-1 {
			parts = append(parts, connector)
		}
	}
	return strings.Join(parts, "")
}

func (r *StepRail) formatStep(index int, label string, pulse bool) string {
	state := stepRailPending
	if index < len(r.states) {
		state = r.states[index]
	}

	mark := stepRailMarkPending
	var paint func(a ...interface{}) string
	paint = GrayColor

	switch state {
	case stepRailActive:
		mark = stepRailMarkActive
		paint = ThemeColor
		if pulse && r.frame%2 == 1 {
			paint = ActionColor
		}
	case stepRailDone:
		mark = stepRailMarkDone
		paint = SuccessColor
	case stepRailFailed:
		mark = stepRailMarkFailed
		paint = ErrorColor
	case stepRailSkipped:
		mark = stepRailMarkSkipped
		paint = GrayColor
	}

	return fmt.Sprintf("%s %s", paint(mark), paint(label))
}

func cleanStepLabels(labels []string) []string {
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		out = append(out, label)
	}
	return out
}

func setActiveRail(r *StepRail) {
	activeRailMu.Lock()
	prev := activeRail
	activeRail = r
	activeRailMu.Unlock()
	if prev != nil && prev != r {
		prev.Close()
	}
}

func clearActiveRail(r *StepRail) {
	activeRailMu.Lock()
	defer activeRailMu.Unlock()
	if activeRail == r {
		activeRail = nil
	}
}

// eraseFooterLocked removes the footer line when it is the last printed line.
// Caller must hold chromeMu.
func eraseFooterLocked() {
	if !footerPainted {
		return
	}
	fmt.Print(stepRailEraseFooter)
	footerPainted = false
}

// paintFooterLocked prints the active rail as the last line. No-op while a
// spinner/progress bar owns the live line or external output is suspended.
// Caller must hold chromeMu.
func paintFooterLocked() {
	if footerPainted || externalHold > 0 || liveLineBusyLocked() {
		return
	}
	activeRailMu.Lock()
	r := activeRail
	activeRailMu.Unlock()
	if r == nil || r.inert {
		return
	}
	r.mu.Lock()
	line := r.format(true)
	r.mu.Unlock()
	if line == "" {
		return
	}
	fmt.Println(line)
	footerPainted = true
}

// prepareContentWriteLocked clears live chrome so a normal log line can print
// above the footer. Caller must hold chromeMu.
func prepareContentWriteLocked() {
	clearActiveSpinnerLineLocked()
	if progressLive {
		clearLine()
	}
	progressLive = false
	eraseFooterLocked()
}

// finishContentWriteLocked restores the footer after a normal log line.
// Caller must hold chromeMu.
func finishContentWriteLocked() {
	paintFooterLocked()
}
