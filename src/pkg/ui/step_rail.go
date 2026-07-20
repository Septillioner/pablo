package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
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
)

type stepRailState int

const (
	stepRailPending stepRailState = iota
	stepRailActive
	stepRailDone
	stepRailFailed
	stepRailSkipped
)

// StepRail is a TTY progress header showing pipeline or sequence phases.
// Non-interactive sessions get a static one-line snapshot (no animation).
type StepRail struct {
	labels  []string
	states  []stepRailState
	current int

	inert    bool
	fallback bool

	mu           sync.Mutex
	renderMu     sync.Mutex
	stopCh       chan struct{}
	doneCh       chan struct{}
	stopped      bool
	frame        int
	contentLines int
	termHeight   int
}

var (
	activeRailMu sync.Mutex
	activeRail   *StepRail
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

	if _, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		r.termHeight = height
	}

	clearActiveSpinnerLine()
	fmt.Println()
	fmt.Println(r.format(false))
	r.contentLines = 0
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
			if active {
				r.redraw()
			}
		}
	}
}

func (r *StepRail) redraw() {
	if r == nil || r.inert {
		return
	}

	r.renderMu.Lock()
	defer r.renderMu.Unlock()

	clearActiveSpinnerLine()

	r.mu.Lock()
	line := r.format(true)
	contentLines := r.contentLines
	height := r.termHeight
	fallback := r.fallback
	if !fallback && height > 0 && contentLines+1 >= height {
		r.fallback = true
		r.contentLines = 0
		fallback = true
		contentLines = 0
	}
	r.mu.Unlock()

	if fallback {
		fmt.Println(line)
		return
	}

	up := contentLines + 1
	fmt.Printf("\033[%dA\r\033[K%s", up, line)
	fmt.Printf("\033[%dB\r", up)
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

func noteRailContentLines(lines int) {
	if lines <= 0 {
		return
	}
	activeRailMu.Lock()
	r := activeRail
	activeRailMu.Unlock()
	if r == nil || r.inert {
		return
	}
	r.mu.Lock()
	r.contentLines += lines
	r.mu.Unlock()
}
