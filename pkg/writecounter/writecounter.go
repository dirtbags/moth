package writecounter

// WriteCounter provides io.Counter, counting how much was written to it.
// The zero value for WriteCounter is ready to use.
type WriteCounter struct {
	n int64
}

// Write provides io.Writer.
func (w *WriteCounter) Write(b []byte) (int, error) {
	w.n += int64(len(b))
	return len(b), nil
}

// Len returns the number of bytes written.
func (w *WriteCounter) Len() int64 {
	return w.n
}
