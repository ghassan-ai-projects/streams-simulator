package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrDuplicateMember = errors.New("plantwire: a member name appears twice in one object")

func decodeKind(raw []byte) (map[string]any, error) {
	var entry map[string]any
	if err := model.DecodeBytes(raw, &entry); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if err := refuseDuplicateMembers(raw); err != nil {
		return nil, err
	}
	return entry, nil
}

func refuseDuplicateMembers(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := walkValue(decoder); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func walkValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	switch token {
	case json.Delim('{'):
		return walkObject(decoder)
	case json.Delim('['):
		return walkArray(decoder)
	}
	return nil
}

func walkObject(decoder *json.Decoder) error {
	seen := map[string]bool{}
	for decoder.More() {
		if err := walkMember(decoder, seen); err != nil {
			return err
		}
	}
	return closing(decoder)
}

func walkMember(decoder *json.Decoder, seen map[string]bool) error {
	name, err := memberName(decoder)
	if err != nil {
		return err
	}
	if seen[name] {
		return fmt.Errorf("%w: %q", ErrDuplicateMember, name)
	}
	seen[name] = true
	return walkValue(decoder)
}

func walkArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := walkValue(decoder); err != nil {
			return err
		}
	}
	return closing(decoder)
}

func memberName(decoder *json.Decoder) (string, error) {
	token, err := decoder.Token()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	name, _ := token.(string)
	return name, nil
}

func closing(decoder *json.Decoder) error {
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
