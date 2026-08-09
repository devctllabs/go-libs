package codexapp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	notificationThreadID = "thread"
	notificationTurnID   = "turn"
)

func TestNotificationFamiliesMapToOneTypedPayload(t *testing.T) {
	t.Parallel()
	handle := &TurnHandle{threadID: "thread", stream: newEventStream(4), done: make(chan struct{})}
	client := &Client{
		turns:    map[string]*TurnHandle{"turn": handle},
		starting: map[string]*TurnHandle{},
	}
	client.dispatchNotification(string(EventCommandOutputDelta), json.RawMessage(`{"threadId":"thread","turnId":"turn","delta":"ok"}`))
	client.dispatchNotification(string(EventTurnDiffUpdated), json.RawMessage(`{"threadId":"thread","turnId":"turn","diff":"+line"}`))
	client.dispatchNotification(string(EventWarning), json.RawMessage(`{"threadId":"thread","message":"careful"}`))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	event, err := handle.Events().Next(ctx)
	require.NoError(t, err)
	require.Equal(t, &CommandOutputDelta{Text: "ok"}, event.CommandOutput)
	require.Nil(t, event.Diff)
	event, err = handle.Events().Next(ctx)
	require.NoError(t, err)
	require.Equal(t, &DiffUpdate{Diff: "+line"}, event.Diff)
	require.Nil(t, event.CommandOutput)
	event, err = handle.Events().Next(ctx)
	require.NoError(t, err)
	require.Equal(t, &WarningEvent{Message: "careful"}, event.Warning)
}

func TestNotificationVariantsPreserveTypedPayloads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   EventType
		params   json.RawMessage
		expected Event
	}{
		{
			name:   "plan delta",
			method: EventPlanDelta,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","delta":"step"}`),
			expected: Event{
				Type: EventPlanDelta, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				PlanDelta: &PlanDelta{Text: "step"},
			},
		},
		{
			name:   "reasoning summary delta",
			method: EventReasoningSummaryDelta,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","delta":"summary"}`),
			expected: Event{
				Type: EventReasoningSummaryDelta, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				TextDelta: &TextDelta{Text: "summary"},
			},
		},
		{
			name:   "reasoning text delta",
			method: EventReasoningTextDelta,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","delta":"reasoning"}`),
			expected: Event{
				Type: EventReasoningTextDelta, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				TextDelta: &TextDelta{Text: "reasoning"},
			},
		},
		{
			name:   "plan updated",
			method: EventTurnPlanUpdated,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","explanation":"why","plan":[{"step":"do it","status":"pending"}]}`),
			expected: Event{
				Type: EventTurnPlanUpdated, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				Plan: &PlanUpdate{Explanation: "why", Steps: []PlanStep{{Step: "do it", Status: "pending"}}},
			},
		},
		{
			name:   "item started",
			method: EventItemStarted,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","item":{"id":"item","type":"message"}}`),
			expected: Event{
				Type: EventItemStarted, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				Item: &Item{ID: "item", Type: "message", Raw: json.RawMessage(`{"id":"item","type":"message"}`)},
			},
		},
		{
			name:   "item completed",
			method: EventItemCompleted,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","item":{"id":"item","type":"message"}}`),
			expected: Event{
				Type: EventItemCompleted, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				Item: &Item{ID: "item", Type: "message", Raw: json.RawMessage(`{"id":"item","type":"message"}`)},
			},
		},
		{
			name:   "error",
			method: EventError,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","willRetry":true,"error":{"message":"failed"}}`),
			expected: Event{
				Type: EventError, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				Failure: &FailureEvent{Message: "failed", WillRetry: true},
			},
		},
		{
			name:   "token usage",
			method: EventTokenUsageUpdated,
			params: json.RawMessage(`{"threadId":"thread","turnId":"turn","tokenUsage":{"total":42}}`),
			expected: Event{
				Type: EventTokenUsageUpdated, ThreadID: notificationThreadID, TurnID: notificationTurnID,
				TokenUsage: &TokenUsageEvent{Raw: json.RawMessage(`{"total":42}`)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handle := &TurnHandle{threadID: notificationThreadID, stream: newEventStream(1), done: make(chan struct{})}
			client := &Client{
				turns:    map[string]*TurnHandle{notificationTurnID: handle},
				starting: map[string]*TurnHandle{},
			}

			client.dispatchNotification(string(test.method), test.params)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			event, err := handle.Events().Next(ctx)
			require.NoError(t, err)
			require.Equal(t, test.expected, event)
		})
	}
}

func TestNotificationDispatcherIgnoresMalformedAndUnknownMessages(t *testing.T) {
	t.Parallel()

	handle := &TurnHandle{threadID: notificationThreadID, stream: newEventStream(1), done: make(chan struct{})}
	client := &Client{
		turns:    map[string]*TurnHandle{notificationTurnID: handle},
		starting: map[string]*TurnHandle{},
	}

	client.dispatchNotification(string(EventWarning), json.RawMessage(`{"threadId":`))
	client.dispatchNotification("future/event", json.RawMessage(`{"threadId":"thread","turnId":"turn"}`))

	require.Empty(t, handle.stream.events)
}
