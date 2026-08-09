package txmanager

import "fmt"

// Isolation identifies a portable transaction isolation level.
type Isolation uint8

const (
	// ReadCommitted allows each statement to observe data committed before that statement began.
	ReadCommitted Isolation = iota + 1
	// RepeatableRead keeps rows read by a transaction stable for its lifetime.
	RepeatableRead
	// Serializable requires transactions to behave as though executed sequentially.
	Serializable
)

type isolationOption struct {
	isolation Isolation
}

// WithIsolation requests isolation for a transaction.
func WithIsolation(isolation Isolation) Option {
	return isolationOption{isolation: isolation}
}

func (o isolationOption) apply(options *txOptions) error {
	switch o.isolation {
	case ReadCommitted, RepeatableRead, Serializable:
		isolation := o.isolation
		options.isolation = &isolation
		return nil
	default:
		return fmt.Errorf("%w: %d", ErrInvalidIsolation, o.isolation)
	}
}
