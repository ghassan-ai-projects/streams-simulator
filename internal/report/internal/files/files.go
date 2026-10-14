// Package files is the file-system edge of the report module: the reads of a run folder.
package files

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/kernel"
	"github.com/ghassan-ai-projects/streams-simulator/internal/report/internal/domain"
)

const maxLine = 8 << 20

func Read(path string) (domain.Document, error) {
	raw, present, err := kernel.ReadIfPresent(path)
	if err != nil {
		return domain.Document{}, fmt.Errorf("report: %w", err)
	}
	return domain.Document{Raw: raw, Present: present}, nil
}

func Scan(ctx context.Context, path string, kind domain.Kind) (domain.Document, error) {
	file, present, err := kernel.OpenIfPresent(path)
	if err != nil {
		return domain.Document{}, fmt.Errorf("report: %w", err)
	}
	if !present {
		return domain.Document{}, nil
	}
	scan, err := scanThenClose(ctx, file, kind)
	if err != nil {
		return domain.Document{}, fmt.Errorf("report: %s: %w", path, err)
	}
	return domain.Document{Present: true, Scan: scan}, nil
}

func scanThenClose(ctx context.Context, file io.ReadCloser, kind domain.Kind) (*domain.Scan, error) {
	scan, err := scanStream(ctx, file, kind)
	_ = file.Close()
	return scan, err
}

type digest struct {
	sum  hash.Hash
	last byte
	n    int64
}

func (d *digest) Write(p []byte) (int, error) {
	_, _ = d.sum.Write(p)
	if len(p) > 0 {
		d.last, d.n = p[len(p)-1], d.n+int64(len(p))
	}
	return len(p), nil
}

func scanStream(ctx context.Context, r io.Reader, kind domain.Kind) (*domain.Scan, error) {
	seen := &digest{sum: sha256.New()}
	scanner := domain.NewScanner(kind)
	if err := feed(ctx, io.TeeReader(r, seen), scanner); err != nil {
		return nil, err
	}
	sum := "sha256:" + hex.EncodeToString(seen.sum.Sum(nil))
	return scanner.Finish(sum, seen.n > 0 && seen.last != '\n'), nil
}

func feed(ctx context.Context, r io.Reader, scanner *domain.Scanner) error {
	lines := bufio.NewScanner(r)
	lines.Buffer(make([]byte, 0, 64<<10), maxLine)
	lines.Split(committedLine)
	for lines.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("%w", err)
		}
		scanner.Line(lines.Bytes())
	}
	return describe(lines.Err())
}

func committedLine(data []byte, atEOF bool) (int, []byte, error) {
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), nil, nil
	}
	return 0, nil, nil
}

func describe(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, bufio.ErrTooLong) {
		return fmt.Errorf("a line is longer than %d bytes", maxLine)
	}
	return fmt.Errorf("%w", err)
}
