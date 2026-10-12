package domain

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

var ErrRefused = errors.New("plantwire: the host refused the session")

func Negotiate(state Message, nodeMax int) (int, error) {
	if !state.Online {
		return 0, fmt.Errorf("%w: %s", ErrRefused, state.Refusal)
	}
	version := state.ProtocolVersion
	low, high := max(Version, state.ProtocolMin), min(int64(nodeMax), state.ProtocolMax)
	if version != high || low > high {
		return 0, fmt.Errorf("%w: version %d outside %d-%d", ErrRefused, version, low, high)
	}
	return int(version), nil
}

func Seq(value any) uint64 {
	seq, _ := SeqOf(value)
	return seq
}

func SeqOf(value any) (uint64, bool) {
	switch number := value.(type) {
	case uint64:
		return number, true
	case float64:
		return wholeSeq(number)
	case fmt.Stringer:
		if parsed, err := strconv.ParseUint(number.String(), 10, 64); err == nil {
			return parsed, true
		}
		parsed, err := strconv.ParseFloat(number.String(), 64)
		return wholeSeq(map[bool]float64{true: parsed, false: -1}[err == nil])
	}
	return 0, false
}

func wholeSeq(number float64) (uint64, bool) {
	if number < 0 || number != math.Trunc(number) || number >= 1<<53 {
		return 0, false
	}
	return uint64(number), true
}
