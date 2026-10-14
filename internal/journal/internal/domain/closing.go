package domain

import (
	"bytes"
	"encoding/json"
)

func ClosedIncomplete(raw []byte) (incomplete, closed bool) {
	var last struct {
		Type       string `json:"type"`
		Incomplete *bool  `json:"incomplete"`
	}
	if json.Unmarshal(lastCompleteLine(raw), &last) != nil || last.Type != TypeRunClosed || last.Incomplete == nil {
		return false, false
	}
	return *last.Incomplete, true
}

func lastCompleteLine(raw []byte) []byte {
	end := bytes.LastIndexByte(raw, '\n')
	if end < 0 {
		return nil
	}
	return raw[bytes.LastIndexByte(raw[:end], '\n')+1 : end]
}
