package domain

import "fmt"

// Digests are the SHA-256 digests of a journal's two streams.
type Digests struct {
	Observer string
	Director string
}

// Journal is the pair of streams of one run. Every record the run writes goes
// through one of its verbs, which fix the record's type, stream and fields.
type Journal struct {
	observer *Stream
	director *Stream
}

// New pairs an observer stream with a director stream.
func New(observer, director *Stream) *Journal {
	return &Journal{observer: observer, director: director}
}

// NewMemory is a journal with no files: it keeps digests, and with retain the
// bytes of both streams.
func NewMemory(retain bool) *Journal {
	return New(NewStream(StreamObserver, nil, retain), NewStream(StreamDirector, nil, retain))
}

// Flush pushes both streams to their destinations at a command boundary and
// returns the first failure either has met.
func (j *Journal) Flush() error {
	observerErr, directorErr := j.observer.Flush(), j.director.Flush()
	return firstError(observerErr, directorErr)
}

// Close writes run.closed to both streams, syncs and closes them, and returns
// their digests. The error is the first failure of either stream.
func (j *Journal) Close(atNS int64, incomplete bool, runError string) (Digests, error) {
	j.observer.Append(TypeRunClosed, atNS, map[string]any{"incomplete": incomplete})
	j.director.Append(TypeRunClosed, atNS, closedFields(incomplete, runError))
	observer, observerErr := j.observer.Close()
	director, directorErr := j.director.Close()
	return Digests{Observer: observer, Director: director}, firstError(observerErr, directorErr)
}

// firstError is the first failure of the streams, nil when none failed.
func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func closedFields(incomplete bool, runError string) map[string]any {
	fields := map[string]any{"incomplete": incomplete}
	if runError != "" {
		fields["error"] = runError
	}
	return fields
}

// Digests returns the digests of the replayable lines written so far.
func (j *Journal) Digests() Digests {
	return Digests{Observer: j.observer.Digest(), Director: j.director.Digest()}
}

// Bytes returns the retained lines of the named stream (nil unless retained).
func (j *Journal) Bytes(stream string) ([]byte, error) {
	switch stream {
	case StreamObserver:
		return j.observer.Bytes(), nil
	case StreamDirector:
		return j.director.Bytes(), nil
	}
	return nil, fmt.Errorf("journal: unknown stream %q", stream)
}
