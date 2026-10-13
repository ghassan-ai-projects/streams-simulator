package sink

func (s *Inproc) Write(line []byte) error {
	return s.impl.Write(line)
}

func (s *Inproc) Close() ([]byte, error) {
	return s.impl.Close()
}

func (s *File) Write(line []byte) error {
	if s == nil || s.file == nil {
		return ErrNotConstructed
	}
	return s.file.Write(line)
}

func (s *File) Flush() error {
	if s == nil || s.file == nil {
		return ErrNotConstructed
	}
	return s.file.Flush()
}

func (s *File) Close() ([]byte, error) {
	if s == nil || s.file == nil {
		return nil, ErrNotConstructed
	}
	return s.file.Close()
}

func (s *HTTPPush) Write(line []byte) error {
	if s == nil || s.push == nil {
		return ErrNotConstructed
	}
	return s.push.Write(line)
}

func (s *HTTPPush) Close() ([]byte, error) {
	if s == nil || s.push == nil {
		return nil, ErrNotConstructed
	}
	return s.push.Close()
}
