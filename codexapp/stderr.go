package codexapp

import "sync"

type tailWriter struct {
	mu   sync.Mutex
	max  int
	data []byte
}

func newTailWriter(maxBytes int) *tailWriter {
	return &tailWriter{max: maxBytes, data: make([]byte, 0, maxBytes)}
}

func (w *tailWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	written := len(p)
	if len(p) >= w.max {
		w.data = append(w.data[:0], p[len(p)-w.max:]...)
		return written, nil
	}
	w.data = append(w.data, p...)
	if overflow := len(w.data) - w.max; overflow > 0 {
		copy(w.data, w.data[overflow:])
		w.data = w.data[:w.max]
	}
	return written, nil
}

func (w *tailWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return string(w.data)
}
