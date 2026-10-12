package domain

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrRefused = errors.New("plantwire: the host refused the session")

func Negotiate(state map[string]any, nodeMax int) (int, error) {
	if online, _ := state["online"].(bool); !online {
		return 0, fmt.Errorf("%w: %v", ErrRefused, state["refusal"])
	}
	version := Int(state["protocol_version"])
	low, high := max(Version, Int(state["protocol_min"])), min(int64(nodeMax), Int(state["protocol_max"]))
	if version < low || version > high {
		return 0, fmt.Errorf("%w: version %d outside %d-%d", ErrRefused, version, low, high)
	}
	return int(version), nil
}

func Int(value any) int64 {
	switch number := value.(type) {
	case interface{ Int64() (int64, error) }:
		parsed, _ := number.Int64()
		return parsed
	case int:
		return int64(number)
	case int64:
		return number
	case float64:
		return int64(number)
	}
	return 0
}

func Seq(value any) uint64 {
	switch number := value.(type) {
	case uint64:
		return number
	case float64:
		return uint64(max(number, 0))
	case fmt.Stringer:
		parsed, _ := strconv.ParseUint(number.String(), 10, 64)
		return parsed
	}
	return 0
}
