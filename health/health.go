package health

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

//go:generate go tool mockgen -destination mocks/health.gen.go -package mocks . Checker,Observer

// Checker reports whether one registered component is healthy.
type Checker interface {
	// Check performs one bounded health observation and must honor ctx cancellation.
	Check(ctx context.Context) error
}

// CheckFunc adapts a function to Checker.
type CheckFunc func(ctx context.Context) error

// Check calls f with ctx.
func (f CheckFunc) Check(ctx context.Context) error {
	return f(ctx)
}

// Outcome classifies how one check invocation completed.
type Outcome string

const (
	OutcomeOK      Outcome = "ok"
	OutcomeError   Outcome = "error"
	OutcomeTimeout Outcome = "timeout"
)

// Observation contains the internal result supplied to diagnostic adapters.
type Observation struct {
	Name     string
	Critical bool
	Outcome  Outcome
	Duration time.Duration
	Err      error
}

// Observer receives completed component check observations.
type Observer interface {
	// Observe records observation and must return promptly.
	Observe(ctx context.Context, observation Observation)
}

// ObserverFunc adapts a function to Observer.
type ObserverFunc func(ctx context.Context, observation Observation)

// Observe calls f with observation.
func (f ObserverFunc) Observe(ctx context.Context, observation Observation) {
	f(ctx, observation)
}

// Status is the binary outcome of a Kubernetes probe or component check.
type Status string

const (
	StatusOK   Status = "ok"
	StatusFail Status = "fail"
)

// Report is the transport-neutral result of a probe evaluation.
type Report struct {
	Status Status
	Checks []CheckResult
}

// CheckResult is the safe, transport-neutral status of one component check.
type CheckResult struct {
	Name     string
	Status   Status
	Critical bool
}

type registration struct {
	name     string
	critical bool
	checker  Checker
}

type config struct {
	checkTimeout time.Duration
	checks       []registration
	observers    []Observer
}

// Option configures Probes during construction.
type Option interface {
	apply(*config) error
}

type optionFunc func(*config) error

func (f optionFunc) apply(cfg *config) error {
	return f(cfg)
}

// Critical registers checker as required for readiness.
func Critical(name string, checker Checker) Option {
	return register(name, true, checker)
}

// NonCritical registers checker for diagnostics without affecting readiness.
func NonCritical(name string, checker Checker) Option {
	return register(name, false, checker)
}

// WithCheckTimeout replaces the one-second deadline shared by readiness checks.
func WithCheckTimeout(timeout time.Duration) Option {
	return optionFunc(func(cfg *config) error {
		if timeout <= 0 {
			return errors.New("health: check timeout must be positive")
		}
		cfg.checkTimeout = timeout
		return nil
	})
}

// WithObserver appends observer to the completed-check notification list.
func WithObserver(observer Observer) Option {
	return optionFunc(func(cfg *config) error {
		if observer == nil {
			return errors.New("health: observer must not be nil")
		}
		cfg.observers = append(cfg.observers, observer)
		return nil
	})
}

func register(name string, critical bool, checker Checker) Option {
	return optionFunc(func(cfg *config) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return errors.New("health: check name must not be blank")
		}
		if checker == nil {
			return fmt.Errorf("health: checker %q must not be nil", name)
		}
		for _, check := range cfg.checks {
			if check.name == name {
				return fmt.Errorf("health: duplicate check name %q", name)
			}
		}
		cfg.checks = append(cfg.checks, registration{name: name, critical: critical, checker: checker})
		return nil
	})
}

// Probes owns one instance's health state and checks.
type Probes struct {
	checkTimeout time.Duration
	checks       []registration
	observers    []Observer
}

// New constructs an independent probe set.
func New(options ...Option) (*Probes, error) {
	cfg := config{checkTimeout: time.Second}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option.apply(&cfg); err != nil {
			return nil, err
		}
	}
	return &Probes{
		checkTimeout: cfg.checkTimeout,
		checks:       append([]registration(nil), cfg.checks...),
		observers:    append([]Observer(nil), cfg.observers...),
	}, nil
}

// Liveness reports whether the health handler itself can respond.
func (*Probes) Liveness() Report {
	return Report{Status: StatusOK}
}

// CheckTimeout returns the deadline applied to one readiness aggregation.
func (p *Probes) CheckTimeout() time.Duration {
	return p.checkTimeout
}

// Readiness checks every registered component.
func (p *Probes) Readiness(ctx context.Context) Report {
	checkCtx, cancel := context.WithTimeout(ctx, p.checkTimeout)
	defer cancel()

	startedAt := time.Now()
	if err := checkCtx.Err(); err != nil {
		invocations := make([]invocation, len(p.checks))
		for index, check := range p.checks {
			invocations[index] = canceledInvocation(index, check, err, time.Since(startedAt))
		}
		return p.finalizeReadiness(ctx, invocations)
	}

	invocations := make([]invocation, len(p.checks))
	completed := make([]bool, len(p.checks))
	results := make(chan invocation, len(p.checks))
	for index, check := range p.checks {
		go func() {
			results <- invoke(checkCtx, index, check)
		}()
	}

	for received := 0; received < len(p.checks); {
		select {
		case result := <-results:
			if completed[result.index] {
				continue
			}
			invocations[result.index] = result
			completed[result.index] = true
			received++
		case <-checkCtx.Done():
			for index, check := range p.checks {
				if completed[index] {
					continue
				}
				invocations[index] = canceledInvocation(index, check, checkCtx.Err(), time.Since(startedAt))
				completed[index] = true
				received++
			}
		}
	}

	return p.finalizeReadiness(ctx, invocations)
}

func (p *Probes) finalizeReadiness(ctx context.Context, invocations []invocation) Report {
	report := Report{Status: StatusOK, Checks: make([]CheckResult, 0, len(invocations))}
	for _, result := range invocations {
		status := StatusOK
		if result.outcome != OutcomeOK {
			status = StatusFail
		}
		if result.critical && status == StatusFail {
			report.Status = StatusFail
		}
		report.Checks = append(report.Checks, CheckResult{
			Name:     result.name,
			Status:   status,
			Critical: result.critical,
		})
		observation := Observation{
			Name:     result.name,
			Critical: result.critical,
			Outcome:  result.outcome,
			Duration: result.duration,
			Err:      result.err,
		}
		for _, observer := range p.observers {
			observer.Observe(ctx, observation)
		}
	}
	return report
}

type invocation struct {
	index    int
	name     string
	critical bool
	outcome  Outcome
	duration time.Duration
	err      error
}

func invoke(ctx context.Context, index int, check registration) (result invocation) {
	startedAt := time.Now()
	result = invocation{index: index, name: check.name, critical: check.critical}
	defer func() {
		result.duration = time.Since(startedAt)
		if recovered := recover(); recovered != nil {
			result.outcome = OutcomeError
			result.err = fmt.Errorf(
				"health: checker %q panicked: %v\n%s",
				check.name,
				recovered,
				debug.Stack(),
			)
		}
	}()

	result.err = check.checker.Check(ctx)
	result.outcome = classifyOutcome(result.err)
	return result
}

func canceledInvocation(index int, check registration, err error, duration time.Duration) invocation {
	return invocation{
		index:    index,
		name:     check.name,
		critical: check.critical,
		outcome:  classifyOutcome(err),
		duration: duration,
		err:      err,
	}
}

func classifyOutcome(err error) Outcome {
	switch {
	case err == nil:
		return OutcomeOK
	case errors.Is(err, context.DeadlineExceeded):
		return OutcomeTimeout
	default:
		return OutcomeError
	}
}
