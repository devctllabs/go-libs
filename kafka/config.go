package kafka

import (
	"errors"
	"fmt"
	"strings"
	"time"

	libretry "github.com/devctllabs/go-libs/retry"
)

// RejectAction selects the terminal handling of permanently rejected records.
type RejectAction uint8

const (
	RejectUnspecified RejectAction = iota
	RejectStop
	RejectDrop
	RejectDLQ
)

// DLQFailureAction selects durability or processing availability after DLQ
// delivery retries are exhausted.
type DLQFailureAction uint8

const (
	DLQFailureUnspecified DLQFailureAction = iota
	DLQFailureStop
	DLQFailureDrop
)

// BatchConfig controls when an accumulated raw Kafka batch is sealed.
type BatchConfig struct {
	MaxSize       int
	FlushInterval time.Duration
}

// RetryConfig bounds a retry cycle by calls, elapsed time, or both.
type RetryConfig struct {
	Policy         libretry.Policy
	MaxAttempts    uint
	MaxElapsedTime time.Duration
}

// DLQConfig routes permanently rejected records within the consumer cluster.
type DLQConfig struct {
	Topic     string
	OnFailure DLQFailureAction
	// Retry overrides ConsumerConfig.Retry for DLQ delivery. Nil inherits it.
	Retry *RetryConfig
}

// ConsumerConfig contains required consumer routing, batching, failure, and
// lifecycle budgets.
type ConsumerConfig struct {
	Brokers []string
	Group   string
	Topics  []string
	Batch   BatchConfig
	Retry   RetryConfig
	// CommitRetry overrides Retry for offset commits. Nil inherits it.
	CommitRetry           *RetryConfig
	OnReject              RejectAction
	DLQ                   *DLQConfig
	RebalanceTimeout      time.Duration
	RebalanceDrainTimeout time.Duration
	ShutdownTimeout       time.Duration
	// Observer receives processing, retry, and disposition signals. Nil disables
	// custom observation.
	Observer Observer
}

func (config ConsumerConfig) validate() error {
	if err := validateStrings("broker", config.Brokers); err != nil {
		return err
	}
	if strings.TrimSpace(config.Group) == "" {
		return errors.New("kafka: consumer group must not be blank")
	}
	if err := validateStrings("topic", config.Topics); err != nil {
		return err
	}
	if err := config.Batch.validate(); err != nil {
		return err
	}
	if err := config.Retry.validate(); err != nil {
		return fmt.Errorf("kafka: retry config: %w", err)
	}
	if config.CommitRetry != nil {
		if err := config.CommitRetry.validate(); err != nil {
			return fmt.Errorf("kafka: commit retry config: %w", err)
		}
	}
	if config.RebalanceTimeout <= 0 {
		return errors.New("kafka: rebalance timeout must be positive")
	}
	if config.RebalanceDrainTimeout <= 0 || config.RebalanceDrainTimeout >= config.RebalanceTimeout {
		return errors.New("kafka: rebalance drain timeout must be positive and shorter than rebalance timeout")
	}
	if config.ShutdownTimeout <= 0 {
		return errors.New("kafka: shutdown timeout must be positive")
	}
	return config.validateReject()
}

func (config BatchConfig) validate() error {
	if config.MaxSize <= 0 {
		return errors.New("kafka: maximum batch size must be positive")
	}
	if config.MaxSize > 1 && config.FlushInterval <= 0 {
		return errors.New("kafka: flush interval must be positive when maximum batch size is greater than one")
	}
	if config.FlushInterval < 0 {
		return errors.New("kafka: flush interval must not be negative")
	}
	return nil
}

func (config RetryConfig) validate() error {
	if config.MaxAttempts == 0 && config.MaxElapsedTime <= 0 {
		return errors.New("at least one retry limit must be positive")
	}
	if config.MaxElapsedTime < 0 {
		return errors.New("maximum elapsed time must not be negative")
	}
	if config.MaxAttempts != 1 && config.Policy == nil {
		return errors.New("retry policy must not be nil when retries are possible")
	}
	return nil
}

func (config ConsumerConfig) validateReject() error {
	switch config.OnReject {
	case RejectStop, RejectDrop:
		if config.DLQ != nil {
			return errors.New("kafka: DLQ config is only valid with RejectDLQ")
		}
		return nil
	case RejectDLQ:
		if config.DLQ == nil {
			return errors.New("kafka: DLQ config is required with RejectDLQ")
		}
		if strings.TrimSpace(config.DLQ.Topic) == "" {
			return errors.New("kafka: DLQ topic must not be blank")
		}
		if config.DLQ.OnFailure != DLQFailureStop && config.DLQ.OnFailure != DLQFailureDrop {
			return errors.New("kafka: DLQ failure action must be Stop or Drop")
		}
		if config.DLQ.Retry != nil {
			if err := config.DLQ.Retry.validate(); err != nil {
				return fmt.Errorf("kafka: DLQ retry config: %w", err)
			}
		}
		return nil
	default:
		return errors.New("kafka: reject action must be explicitly configured")
	}
}

func validateStrings(name string, values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("kafka: at least one %s is required", name)
	}
	for index, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("kafka: %s %d must not be blank", name, index)
		}
	}
	return nil
}
