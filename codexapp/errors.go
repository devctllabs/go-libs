package codexapp

import (
	"errors"
	"fmt"
)

// ErrOutcomeUnknown marks a call that was written but whose response was not observed.
var ErrOutcomeUnknown = errors.New("codexapp: call outcome is unknown")

// ErrEventOverflow reports that a consumer did not drain its bounded turn stream in time.
var ErrEventOverflow = errors.New("codexapp: turn event buffer overflow")

// ErrClosed reports an explicitly closed event stream.
var ErrClosed = errors.New("codexapp: closed")

// ErrTurnStartInProgress reports a concurrent turn/start for the same thread.
var ErrTurnStartInProgress = errors.New("codexapp: turn start already in progress")

// ErrPermissionProfileMismatch reports a turn profile different from its thread binding.
var ErrPermissionProfileMismatch = errors.New("codexapp: permission profile does not match thread")

// CallError reports an RPC failure and whether App Server may have applied the request.
type CallError struct {
	Method         string
	Cause          error
	OutcomeUnknown bool
}

// ProcessExitError reports an unexpected App Server process exit.
type ProcessExitError struct {
	Cause      error
	StderrTail string
}

func (e *ProcessExitError) Error() string {
	if e.StderrTail != "" {
		return fmt.Sprintf("codexapp: app server exited: %v; stderr: %s", e.Cause, e.StderrTail)
	}
	return fmt.Sprintf("codexapp: app server exited: %v", e.Cause)
}

// Unwrap returns the process wait error.
func (e *ProcessExitError) Unwrap() error { return e.Cause }

// SessionLostError reports that process-owned in-flight state cannot survive a session exit.
type SessionLostError struct {
	Generation int64
	Cause      error
}

func (e *SessionLostError) Error() string {
	return fmt.Sprintf("codexapp: session generation %d was lost: %v", e.Generation, e.Cause)
}

// Unwrap returns the process or protocol failure that ended the session.
func (e *SessionLostError) Unwrap() error { return e.Cause }

// ProtocolError reports malformed or oversized JSONL protocol data.
type ProtocolError struct {
	Operation string
	Cause     error
}

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("codexapp: protocol %s: %v", e.Operation, e.Cause)
}

// Unwrap returns the decoding or framing error.
func (e *ProtocolError) Unwrap() error { return e.Cause }

// UnsupportedCapabilityError reports a failed required capability probe.
type UnsupportedCapabilityError struct {
	Capability Capability
	Cause      error
}

func (e *UnsupportedCapabilityError) Error() string {
	return fmt.Sprintf("codexapp: required capability %s is unavailable: %v", e.Capability, e.Cause)
}

// Unwrap returns the probe error.
func (e *UnsupportedCapabilityError) Unwrap() error { return e.Cause }

func (e *CallError) Error() string {
	if e.OutcomeUnknown {
		return fmt.Sprintf("codexapp: %s call outcome is unknown: %v", e.Method, e.Cause)
	}
	return fmt.Sprintf("codexapp: %s call failed: %v", e.Method, e.Cause)
}

// Unwrap returns the underlying transport or context error.
func (e *CallError) Unwrap() error { return e.Cause }

// Is reports the outcome-unknown error category while preserving the underlying cause.
func (e *CallError) Is(target error) bool {
	return e.OutcomeUnknown && target == ErrOutcomeUnknown
}
