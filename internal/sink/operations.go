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
	return s.file.Write(line)
}

// Flush pushes buffered bytes to the file descriptor without syncing, so the
// delivery ledger and the trace stay consistent at command boundaries; Close
// syncs them to disk.
func (s *File) Flush() error {
	return s.file.Flush()
}

// Close flushes, syncs and returns the file contents.
func (s *File) Close() ([]byte, error) {
	return s.file.Close()
}

// Write posts one line, recording it for the digest.
func (s *HTTPPush) Write(line []byte) error {
	return s.push.Write(line)
}

// Close returns the delivered bytes in delivery order.
func (s *HTTPPush) Close() ([]byte, error) {
	return s.push.Close()
}
