package sink

// Write appends a line.
func (s *Inproc) Write(line []byte) error {
	return s.impl.Write(line)
}

// Close returns the buffered bytes.
func (s *Inproc) Close() ([]byte, error) {
	return s.impl.Close()
}

// Write appends a line.
func (s *File) Write(line []byte) error {
	if s == nil || s.file == nil {
		return ErrNotConstructed
	}
	return s.file.Write(line)
}

// Flush pushes buffered bytes to the file descriptor without syncing, so the
// delivery ledger and the trace stay consistent at command boundaries; Close
// syncs them to disk.
func (s *File) Flush() error {
	if s == nil || s.file == nil {
		return ErrNotConstructed
	}
	return s.file.Flush()
}

// Close flushes, syncs and returns the file contents.
func (s *File) Close() ([]byte, error) {
	if s == nil || s.file == nil {
		return nil, ErrNotConstructed
	}
	return s.file.Close()
}

// Write posts one line, recording it for the digest.
func (s *HTTPPush) Write(line []byte) error {
	if s == nil || s.push == nil {
		return ErrNotConstructed
	}
	return s.push.Write(line)
}

// Close returns the delivered bytes in delivery order.
func (s *HTTPPush) Close() ([]byte, error) {
	if s == nil || s.push == nil {
		return nil, ErrNotConstructed
	}
	return s.push.Close()
}
