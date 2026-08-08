package codexapp

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"
)

// RestartConfig controls process restart backoff.
type RestartConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
	ResetAfter   time.Duration
}

// Supervisor explicitly runs and replaces failed App Server process generations.
type Supervisor struct {
	config  Config
	restart RestartConfig

	mu         sync.Mutex
	current    *Client
	generation int64
	changed    chan struct{}
	started    bool
	runCancel  context.CancelFunc
	runDone    chan struct{}
	runErr     error
}

// NewSupervisor validates configuration without starting background work.
func NewSupervisor(config Config, restart RestartConfig) (*Supervisor, error) {
	var err error
	config, err = normalizeConfig(config)
	if err != nil {
		return nil, err
	}
	if restart == (RestartConfig{}) {
		restart = RestartConfig{
			InitialDelay: 250 * time.Millisecond,
			MaxDelay:     5 * time.Second,
			Multiplier:   2,
			Jitter:       0.2,
			ResetAfter:   time.Minute,
		}
	}
	if restart.InitialDelay <= 0 {
		return nil, fmt.Errorf("codexapp: restart initial delay must be positive")
	}
	if restart.MaxDelay < restart.InitialDelay {
		return nil, fmt.Errorf("codexapp: restart max delay must not be less than initial delay")
	}
	if restart.Multiplier < 1 {
		return nil, fmt.Errorf("codexapp: restart multiplier must be at least one")
	}
	if restart.Jitter < 0 || restart.Jitter > 1 {
		return nil, fmt.Errorf("codexapp: restart jitter must be between zero and one")
	}
	if restart.ResetAfter <= 0 {
		return nil, fmt.Errorf("codexapp: restart reset-after must be positive")
	}
	return &Supervisor{
		config:  config,
		restart: restart,
		changed: make(chan struct{}),
		runDone: make(chan struct{}),
	}, nil
}

// Run owns the restart loop until ctx ends or a terminal startup error occurs.
func (s *Supervisor) Run(ctx context.Context) (runErr error) {
	if ctx == nil {
		return fmt.Errorf("codexapp: context is required")
	}
	runCtx, cancel, err := s.beginRun(ctx)
	if err != nil {
		return err
	}
	defer func() { s.finishRun(cancel, runErr) }()

	delay := s.restart.InitialDelay
	for {
		if runCtx.Err() != nil {
			return nil
		}
		startedAt := time.Now()
		client, err := Open(runCtx, s.config)
		if err != nil {
			if runCtx.Err() != nil {
				return nil
			}
			return fmt.Errorf("codexapp: supervisor start: %w", err)
		}
		s.publishClient(client)
		waitForClient(runCtx, client)
		s.clearClient(client)
		if runCtx.Err() != nil {
			return nil
		}
		if time.Since(startedAt) >= s.restart.ResetAfter {
			delay = s.restart.InitialDelay
		}
		if !waitForRestart(runCtx, withJitter(delay, s.restart.Jitter)) {
			return nil
		}
		delay = s.nextDelay(delay)
	}
}

func (s *Supervisor) beginRun(ctx context.Context) (context.Context, context.CancelFunc, error) {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil, nil, fmt.Errorf("codexapp: supervisor is already running")
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.started = true
	s.runCancel = cancel
	s.notifyLocked()
	s.mu.Unlock()
	return runCtx, cancel, nil
}

func (s *Supervisor) finishRun(cancel context.CancelFunc, runErr error) {
	cancel()
	s.mu.Lock()
	s.runErr = runErr
	close(s.runDone)
	s.notifyLocked()
	s.mu.Unlock()
}

func (s *Supervisor) publishClient(client *Client) {
	s.mu.Lock()
	s.generation++
	client.generation = s.generation
	s.current = client
	s.notifyLocked()
	s.mu.Unlock()
}

func waitForClient(runCtx context.Context, client *Client) {
	select {
	case <-runCtx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = client.Close(shutdownCtx)
		shutdownCancel()
	case <-client.done:
	}
}

func (s *Supervisor) clearClient(client *Client) {
	s.mu.Lock()
	if s.current == client {
		s.current = nil
		s.notifyLocked()
	}
	s.mu.Unlock()
}

func waitForRestart(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return false
	case <-timer.C:
		return true
	}
}

func (s *Supervisor) nextDelay(delay time.Duration) time.Duration {
	next := time.Duration(float64(delay) * s.restart.Multiplier)
	if next > s.restart.MaxDelay {
		return s.restart.MaxDelay
	}
	return next
}

// Client waits for and returns the current ready process generation.
func (s *Supervisor) Client(ctx context.Context) (*Client, error) {
	if ctx == nil {
		return nil, fmt.Errorf("codexapp: context is required")
	}
	for {
		s.mu.Lock()
		client := s.current
		changed := s.changed
		started := s.started
		runDone := s.runDone
		runErr := s.runErr
		s.mu.Unlock()
		if client != nil {
			select {
			case <-client.done:
			default:
				return client, nil
			}
		}
		if started {
			select {
			case <-runDone:
				if runErr != nil {
					return nil, runErr
				}
				return nil, errors.New("codexapp: supervisor stopped")
			default:
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-changed:
		}
	}
}

// Shutdown stops Run and closes the current client.
func (s *Supervisor) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("codexapp: context is required")
	}
	s.mu.Lock()
	started := s.started
	cancel := s.runCancel
	client := s.current
	runDone := s.runDone
	s.mu.Unlock()
	if !started {
		return nil
	}
	cancel()
	var closeErr error
	if client != nil {
		closeErr = client.Close(ctx)
	}
	select {
	case <-runDone:
		return closeErr
	case <-ctx.Done():
		return errors.Join(closeErr, ctx.Err())
	}
}

func (s *Supervisor) notifyLocked() {
	close(s.changed)
	s.changed = make(chan struct{})
}

func withJitter(delay time.Duration, fraction float64) time.Duration {
	if fraction == 0 {
		return delay
	}
	var random [8]byte
	if _, err := cryptorand.Read(random[:]); err != nil {
		return delay
	}
	ratio := float64(binary.LittleEndian.Uint64(random[:]))/float64(^uint64(0))*2 - 1
	return time.Duration(float64(delay) * (1 + ratio*fraction))
}
