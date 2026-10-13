package domain

import "fmt"

type Digests struct {
	Observer     string
	Director     string
	ObserverFile string
	DirectorFile string
}

type Journal struct {
	observer *Stream
	director *Stream
}

func New(observer, director *Stream) *Journal {
	return &Journal{observer: observer, director: director}
}

func NewMemory(retain bool) *Journal {
	return New(NewStream(StreamObserver, nil, retain), NewStream(StreamDirector, nil, retain))
}

func (j *Journal) Flush() error {
	observerErr, directorErr := j.observer.Flush(), j.director.Flush()
	return firstError(observerErr, directorErr)
}

func (j *Journal) Close(atNS int64, incomplete bool, runError string) (Digests, error) {
	j.observer.Append(TypeRunClosed, atNS, map[string]any{"incomplete": incomplete})
	j.director.Append(TypeRunClosed, atNS, closedFields(incomplete, runError))
	observer, observerErr := j.observer.Close()
	director, directorErr := j.director.Close()
	digests := Digests{Observer: observer, Director: director, ObserverFile: j.observer.FileDigest(), DirectorFile: j.director.FileDigest()}
	return digests, firstError(observerErr, directorErr)
}

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

func (j *Journal) Digests() Digests {
	return Digests{Observer: j.observer.Digest(), Director: j.director.Digest(), ObserverFile: j.observer.FileDigest(), DirectorFile: j.director.FileDigest()}
}

func (j *Journal) Bytes(stream string) ([]byte, error) {
	switch stream {
	case StreamObserver:
		return j.observer.Bytes(), nil
	case StreamDirector:
		return j.director.Bytes(), nil
	}
	return nil, fmt.Errorf("journal: unknown stream %q", stream)
}
