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
		return integerLiteral(s)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", fmt.Errorf("canonical: invalid number literal %q", s)
	}
	return esNumber(f)
}

func integerLiteral(s string) (string, error) {
	integer, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return "", fmt.Errorf("canonical: invalid integer literal %q", s)
	}
	// ES Number::toString(-0) is "0", including the literal fast path.
	if integer.Sign() == 0 {
		return "0", nil
	}
	return s, nil
}

// esNumber renders f per ECMAScript Number::toString(x, 10).
// Shortest round-trip digits come from strconv; the decimal/exponent
// notation boundaries follow ES: decimal when -6 < n <= 21, else
// mantissa × 10^exponent with a single digit before the point.
func esNumber(f float64) (string, error) {
	if text, err, exact := exactNumber(f); exact {
		return text, err
	}
	sign, digits, exp, err := scientificDigits(f)
	if err != nil {
		return "", err
	}
	return decimalNotation(sign, digits, exp), nil
}

func exactNumber(f float64) (string, error, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "", fmt.Errorf("canonical: non-finite number %v is not representable in canonical JSON", f), true
	}
	if f == 0 {
		return "0", nil, true
	}
	// Integral floats within exact int64 range render as integers.
	if f == math.Trunc(f) && math.Abs(f) <= 1<<53 {
		return strconv.FormatInt(int64(f), 10), nil, true
	}
	return "", nil, false
}

func scientificDigits(f float64) (string, string, int, error) {
	e := strconv.FormatFloat(f, 'e', -1, 64) // e.g. "-1.2345e+08"
	em := strings.IndexByte(e, 'e')
	mant, expStr := e[:em], e[em+1:]
	exp, err := strconv.Atoi(expStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("canonical: bad exponent %q", expStr)
	}
	sign, digits := mantissaDigits(mant)
	return sign, digits, exp, nil
}

func mantissaDigits(mantissa string) (string, string) {
	sign := ""
	if strings.HasPrefix(mantissa, "-") {
		sign = "-"
		mantissa = mantissa[1:]
	}
	digits := strings.ReplaceAll(mantissa, ".", "")
	// Strip leading zeros in the shortest representation (e.g. 0.5 -> "5").
	return sign, strings.TrimLeft(digits, "0")
}

func decimalNotation(sign, digits string, exp int) string {
	if digits == "" {
		return "0"
	}
	return signedDecimalNotation(sign, digits, exp)
}

func signedDecimalNotation(sign, digits string, exp int) string {
	n := exp + 1 // digits before the decimal point
	switch {
	case n > 21 || n <= -6:
		return sign + exponentNotation(digits, exp)
	case n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	case n >= len(digits):
		return sign + digits + strings.Repeat("0", n-len(digits))
	default:
		return sign + digits[:n] + "." + digits[n:]
	}
}

func formatExp(e int) string {
	if e >= 0 {
		return "+" + strconv.Itoa(e)
	}
	return strconv.Itoa(e)
}

func exponentNotation(digits string, exp int) string {
	mantissa := digits[:1]
	if len(digits) > 1 {
		mantissa += "." + digits[1:]
	}
	return mantissa + "e" + formatExp(exp)
}
