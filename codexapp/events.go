package codexapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// EventType identifies the payload populated on Event.
type EventType string

const (
	// EventAgentMessageDelta carries an incremental assistant text fragment.
	EventAgentMessageDelta EventType = "item/agentMessage/delta"
	// EventTurnCompleted carries the terminal turn projection.
	EventTurnCompleted         EventType = "turn/completed"
	EventItemStarted           EventType = "item/started"
	EventItemCompleted         EventType = "item/completed"
	EventPlanDelta             EventType = "item/plan/delta"
	EventReasoningSummaryDelta EventType = "item/reasoning/summaryTextDelta"
	EventReasoningTextDelta    EventType = "item/reasoning/textDelta"
	EventCommandOutputDelta    EventType = "item/commandExecution/outputDelta"
	EventTurnDiffUpdated       EventType = "turn/diff/updated"
	EventTurnPlanUpdated       EventType = "turn/plan/updated"
	EventError                 EventType = "error"
	EventWarning               EventType = "warning"
	EventTokenUsageUpdated     EventType = "thread/tokenUsage/updated"
)

// Event is a tagged union. Exactly one matching payload pointer is populated.
type Event struct {
	Type          EventType
	ThreadID      string
	TurnID        string
	TextDelta     *TextDelta
	Turn          *Turn
	Item          *Item
	CommandOutput *CommandOutputDelta
	PlanDelta     *PlanDelta
	Diff          *DiffUpdate
	Plan          *PlanUpdate
	Failure       *FailureEvent
	Warning       *WarningEvent
	TokenUsage    *TokenUsageEvent
	Raw           json.RawMessage
}

// TextDelta is one incremental text fragment.
type TextDelta struct {
	Text string
}

// Item is the stable identity plus raw schema-versioned item payload.
type Item struct {
	ID   string
	Type string
	Raw  json.RawMessage
}

type CommandOutputDelta struct{ Text string }
type PlanDelta struct{ Text string }
type DiffUpdate struct{ Diff string }

type PlanUpdate struct {
	Explanation string
	Steps       []PlanStep
}

type PlanStep struct {
	Step   string
	Status string
}

type FailureEvent struct {
	Message   string
	WillRetry bool
}

type WarningEvent struct{ Message string }
type TokenUsageEvent struct{ Raw json.RawMessage }

// EventStream is a bounded stream owned by a TurnHandle.
type EventStream struct {
	events chan Event
	mu     sync.Mutex
	closed bool
	err    error
}

// Close stops delivery to this consumer. It does not interrupt the underlying turn.
func (s *EventStream) Close() error {
	s.finish(ErrClosed)
	return nil
}

func newEventStream(capacity int) *EventStream {
	return &EventStream{events: make(chan Event, capacity)}
}

// Next waits for the next event or returns io.EOF after normal completion.
func (s *EventStream) Next(ctx context.Context) (Event, error) {
	if ctx == nil {
		return Event{}, fmt.Errorf("codexapp: context is required")
	}
	select {
	case event, ok := <-s.events:
		if ok {
			return event, nil
		}
		s.mu.Lock()
		err := s.err
		s.mu.Unlock()
		if err == nil {
			return Event{}, io.EOF
		}
		return Event{}, err
	case <-ctx.Done():
		return Event{}, ctx.Err()
	}
}

func (s *EventStream) push(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return io.ErrClosedPipe
	}
	select {
	case s.events <- event:
		return nil
	default:
		err := ErrEventOverflow
		s.err = err
		s.closed = true
		close(s.events)
		return err
	}
}

func (s *EventStream) finish(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.err = err
	s.closed = true
	close(s.events)
}
