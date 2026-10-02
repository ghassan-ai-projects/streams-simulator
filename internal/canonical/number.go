package canonical

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// canonicalNumber validates a JSON number literal and returns its canonical
// form. Integer literals are returned verbatim (JSON already forbids leading
// zeros, plus signs and trailing junk); non-integer literals are re-rendered
// from their float64 value per ES Number::toString.
func canonicalNumber(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("canonical: empty number")
	}
	if !strings.ContainsAny(s, ".eE") {
		bi, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return "", fmt.Errorf("canonical: invalid integer literal %q", s)
		}
		// ES Number::toString(-0) == "0"; the literal fast path must not
		// reintroduce a negative zero.
		if bi.Sign() == 0 {
			return "0", nil
		}
		return s, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", fmt.Errorf("canonical: invalid number literal %q", s)
	}
	return esNumber(f)
}

// esNumber renders f per ECMAScript Number::toString(x, 10).
// Shortest round-trip digits come from strconv; the decimal/exponent
// notation boundaries follow ES: decimal when -6 < n <= 21, else
// mantissa × 10^exponent with a single digit before the point.
func esNumber(f float64) (string, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "", fmt.Errorf("canonical: non-finite number %v is not representable in canonical JSON", f)
	}
	if f == 0 {
		return "0", nil
	}
	// Integral floats within exact int64 range render as integers.
	if f == math.Trunc(f) && math.Abs(f) <= 1<<53 {
		return strconv.FormatInt(int64(f), 10), nil
	}
	e := strconv.FormatFloat(f, 'e', -1, 64) // e.g. "-1.2345e+08"
	em := strings.IndexByte(e, 'e')
	mant, expStr := e[:em], e[em+1:]
	exp, err := strconv.Atoi(expStr)
	if err != nil {
		return "", fmt.Errorf("canonical: bad exponent %q", expStr)
	}
	sign := ""
	digits := mant
	if strings.HasPrefix(digits, "-") {
		sign = "-"
		digits = digits[1:]
	}
	digits = strings.ReplaceAll(digits, ".", "")
	// Strip leading zeros the shortest-repr may have produced (e.g. 0.5 -> "5").
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return "0", nil
	}
	n := exp + 1 // digits before the decimal point
	switch {
	case n > 21 || n <= -6:
		m := digits[:1]
		if len(digits) > 1 {
			m += "." + digits[1:]
		}
		return sign + m + "e" + formatExp(exp), nil
	case n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits, nil
	case n >= len(digits):
		return sign + digits + strings.Repeat("0", n-len(digits)), nil
	default:
		return sign + digits[:n] + "." + digits[n:], nil
	}
}

func formatExp(e int) string {
	if e >= 0 {
		return "+" + strconv.Itoa(e)
	}
	return strconv.Itoa(e)
}
