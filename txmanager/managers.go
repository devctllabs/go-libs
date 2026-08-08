package txmanager

import "errors"

// Managers provides transaction managers bound to reader and writer roles.
type Managers interface {
	// Reader returns the read-only transaction manager.
	Reader() Manager
	// Writer returns the read-write transaction manager.
	Writer() Manager
}

// ManagerSet is an immutable pair of reader and writer transaction managers.
type ManagerSet struct {
	reader Manager
	writer Manager
}

// NewManagers constructs an immutable reader/writer manager pair.
func NewManagers(reader Manager, writer Manager) (*ManagerSet, error) {
	if reader == nil {
		return nil, errors.New("txmanager: reader manager must not be nil")
	}
	if writer == nil {
		return nil, errors.New("txmanager: writer manager must not be nil")
	}
	return &ManagerSet{reader: reader, writer: writer}, nil
}

// Reader returns the configured reader manager.
func (m *ManagerSet) Reader() Manager {
	return m.reader
}

// Writer returns the configured writer manager.
func (m *ManagerSet) Writer() Manager {
	return m.writer
}
