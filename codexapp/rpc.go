package codexapp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

const (
	defaultMaxMessageBytes = 64 << 20
	jsonRPCVersion         = "2.0"
)

type initializeParams struct {
	Capabilities *initializeCapabilities `json:"capabilities,omitempty"`
	ClientInfo   ClientInfo              `json:"clientInfo"`
}

type initializeCapabilities struct {
	ExperimentalAPI *bool `json:"experimentalApi,omitempty"`
}

type modelListParams struct {
	Cursor        *string `json:"cursor,omitempty"`
	Limit         *int64  `json:"limit,omitempty"`
	IncludeHidden *bool   `json:"includeHidden,omitempty"`
}

type permissionProfileListParams struct{}

type rpcRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      int64             `json:"id,omitempty"`
	Method  string            `json:"method"`
	Params  any               `json:"params,omitempty"`
	Trace   map[string]string `json:"trace,omitempty"`
}

type rpcResponse struct {
	result json.RawMessage
	err    error
}

// RPCError is an error object returned by App Server.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e RPCError) Error() string {
	return fmt.Sprintf("json-rpc error %d: %s", e.Code, e.Message)
}

func (c *Client) call(ctx context.Context, method string, params any, result any) (err error) {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	ctx, finish := c.instrumentation.startRPC(ctx, method)
	defer func() { finish(err) }()
	id := c.nextID.Add(1)
	response := make(chan rpcResponse, 1)
	c.pendingMu.Lock()
	c.pending[id] = response
	c.pendingMu.Unlock()

	traceCarrier := c.instrumentation.inject(ctx)
	if err := c.write(rpcRequest{JSONRPC: jsonRPCVersion, ID: id, Method: method, Params: params, Trace: traceCarrier}); err != nil {
		c.removePending(id)
		return err
	}
	select {
	case received := <-response:
		if received.err != nil {
			return received.err
		}
		if result == nil {
			return nil
		}
		if err := json.Unmarshal(received.result, result); err != nil {
			return fmt.Errorf("decode %s response: %w", method, err)
		}
		return nil
	case <-ctx.Done():
		c.removePending(id)
		return &CallError{Method: method, Cause: ctx.Err(), OutcomeUnknown: true}
	case <-c.done:
		c.removePending(id)
		return c.sessionError()
	}
}

func (c *Client) notify(method string, params any) error {
	return c.write(rpcRequest{JSONRPC: jsonRPCVersion, Method: method, Params: params})
}

func (c *Client) write(message rpcRequest) error {
	return c.writeJSON(message, message.Method+" request")
}

func (c *Client) writeJSON(message any, operation string) error {
	encoded, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("encode %s: %w", operation, err)
	}
	encoded = append(encoded, '\n')
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if _, err := c.stdin.Write(encoded); err != nil {
		return fmt.Errorf("write %s: %w", operation, err)
	}
	return nil
}

func (c *Client) readLoop(stdout io.Reader, maxMessageBytes int) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64<<10), maxMessageBytes)
	for scanner.Scan() {
		var envelope struct {
			ID     json.RawMessage   `json:"id"`
			Method string            `json:"method"`
			Params json.RawMessage   `json:"params"`
			Trace  map[string]string `json:"trace"`
			Result json.RawMessage   `json:"result"`
			Error  *RPCError         `json:"error"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &envelope); err != nil {
			c.finish(&ProtocolError{Operation: "decode app server message", Cause: err})
			return
		}
		if len(envelope.ID) == 0 || string(envelope.ID) == "null" {
			c.dispatchNotification(envelope.Method, envelope.Params)
			continue
		}
		if envelope.Method != "" {
			c.dispatchServerRequest(envelope.ID, envelope.Method, envelope.Params, envelope.Trace)
			continue
		}
		var responseID int64
		if err := json.Unmarshal(envelope.ID, &responseID); err != nil {
			c.finish(&ProtocolError{Operation: "decode response id", Cause: err})
			return
		}
		c.pendingMu.Lock()
		response, ok := c.pending[responseID]
		if ok {
			delete(c.pending, responseID)
		}
		c.pendingMu.Unlock()
		if !ok {
			continue
		}
		if envelope.Error != nil {
			response <- rpcResponse{err: envelope.Error}
			continue
		}
		response <- rpcResponse{result: envelope.Result}
	}
	if err := scanner.Err(); err != nil {
		c.finish(&ProtocolError{Operation: "read app server message", Cause: err})
		return
	}
	// The process waiter owns clean EOF classification so an immediately following non-zero
	// exit is not lost to a race with stdout closure.
}

func (c *Client) removePending(id int64) {
	c.pendingMu.Lock()
	delete(c.pending, id)
	c.pendingMu.Unlock()
}

func (c *Client) finish(err error) {
	c.doneOnce.Do(func() {
		c.handlerCancel()
		c.terminalMu.Lock()
		c.terminalErr = err
		c.terminalMu.Unlock()
		c.failTurns(err)
		close(c.done)
	})
}

func (c *Client) sessionError() error {
	c.terminalMu.Lock()
	defer c.terminalMu.Unlock()
	if c.terminalErr == nil {
		return io.EOF
	}
	return c.terminalErr
}
