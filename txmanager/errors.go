package txmanager

import "errors"

var (
	// ErrInvalidIsolation reports an isolation value outside the supported portable subset.
	ErrInvalidIsolation = errors.New("txmanager: invalid isolation")
	// ErrIsolationMismatch reports a nested transaction requesting a different isolation.
	ErrIsolationMismatch = errors.New("txmanager: nested isolation does not match active transaction")
	// ErrReadOnlyEscalation reports a writer transaction requested inside a reader transaction.
	ErrReadOnlyEscalation = errors.New("txmanager: cannot start writer transaction inside reader transaction")
)
