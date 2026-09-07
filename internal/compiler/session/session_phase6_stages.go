package session

import (
	"time"

	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// StageNames returns D-06-21's five fixed, sequential pipeline stages, in
// pipeline order. This is a closed, named set the compiler pipeline already
// crosses -- do not add a sixth, and do not derive it dynamically. It exists
// so FND-04's "a regression identifies the responsible stage" has a stable
// vocabulary to cite, without a tracing runtime, span model, or global
// registry (PROJECT.md's runtime-posture constraint).
func StageNames() []string {
	return []string{"parse", "check", "lower", "native_compile", "link"}
}

func stageOrdinal(stage string) (int, bool) {
	for index, name := range StageNames() {
		if name == stage {
			return index, true
		}
	}
	return -1, false
}

// StageRecordError is a typed refusal for a StageRecorder programming
// error: a Start or Stop naming a stage outside StageNames(), or a Stop
// with no matching prior Start. It refuses rather than silently recording
// a zero elapsed value, matching this project's {Code}-only typed-error
// convention (debugmap.Error, evidence.ValidationError).
type StageRecordError struct {
	Code string
}

func (e *StageRecordError) Error() string { return e.Code }

// StageRecorder records explicit start/stop timestamp pairs at D-06-21's
// five fixed, sequential stage boundaries into a fixed-size array indexed
// by stage ordinal -- no map, no span tree, no global registry, no
// goroutine-local state, no package-level mutable state. This is the
// entire mechanism PROJECT.md's "no mandatory global tracing heap or
// service runtime" constraint permits for stage attribution.
type StageRecorder struct {
	startedAt [5]time.Time
	elapsed   [5]time.Duration
	running   [5]bool
	recorded  [5]bool
}

// Start records stage's start time. Returns a *StageRecordError naming
// "session.stage_unknown" for a stage outside StageNames().
func (r *StageRecorder) Start(stage string) error {
	index, ok := stageOrdinal(stage)
	if !ok {
		return &StageRecordError{Code: "session.stage_unknown"}
	}
	r.startedAt[index] = time.Now()
	r.running[index] = true
	return nil
}

// Stop records stage's elapsed duration since its matching Start. Returns a
// *StageRecordError naming "session.stage_unknown" for a stage outside
// StageNames(), or "session.stage_not_started" for a Stop with no matching
// prior Start -- refusing rather than silently recording a zero elapsed
// value for a stage that never actually ran.
func (r *StageRecorder) Stop(stage string) error {
	index, ok := stageOrdinal(stage)
	if !ok {
		return &StageRecordError{Code: "session.stage_unknown"}
	}
	if !r.running[index] {
		return &StageRecordError{Code: "session.stage_not_started"}
	}
	r.elapsed[index] = time.Since(r.startedAt[index])
	r.running[index] = false
	r.recorded[index] = true
	return nil
}

// Breakdown returns one protocol.StageTiming per stage this recorder
// actually completed a Start/Stop pair for, in StageNames() pipeline
// order. It returns nil -- no breakdown at all, never an array of zeros --
// unless TimingObservationEnabled() reports true, so a run with
// LANG_OBSERVE_TIMING unset emits no stage_breakdown at all (FND-04's
// empty-input edge). A stage this recorder never started (for example
// native_compile/link on a cache-artifact-reused run, where no fresh
// compile happened this invocation) is simply omitted, not reported as a
// zero-elapsed entry.
func (r *StageRecorder) Breakdown() []protocol.StageTiming {
	if !TimingObservationEnabled() {
		return nil
	}
	var breakdown []protocol.StageTiming
	for index, name := range StageNames() {
		if !r.recorded[index] {
			continue
		}
		breakdown = append(breakdown, protocol.StageTiming{Stage: name, ElapsedNS: r.elapsed[index].Nanoseconds()})
	}
	return breakdown
}
