package domain

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var wireID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

func WireID(id string) string {
	if wireID.MatchString(id) {
		return id
	}
	safe := strings.Map(wireRune, id)
	if safe == "" || !wireID.MatchString(safe[:1]) {
		safe = "d" + safe
	}
	return safe[:min(len(safe), 128)]
}

func wireRune(r rune) rune {
	switch {
	case r == '/':
		return ':'
	case r < 128 && wireID.MatchString(string(r)), r == '_' || r == '.' || r == '-' || r == ':':
		return r
	}
	return '_'
}

func Distinct(base string, taken func(string) bool) string {
	candidate := base
	for n := 2; taken(candidate); n++ {
		candidate = base[:min(len(base), 120)] + "." + strconv.Itoa(n)
	}
	return candidate
}

func Finite(value any) bool {
	number, isNumber := value.(float64)
	return !isNumber || (!math.IsNaN(number) && !math.IsInf(number, 0))
}
