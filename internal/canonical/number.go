package canonical

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

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

	if integer.Sign() == 0 {
		return "0", nil
	}
	return s, nil
}

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

	if f == math.Trunc(f) && math.Abs(f) <= 1<<53 {
		return strconv.FormatInt(int64(f), 10), nil, true
	}
	return "", nil, false
}

func scientificDigits(f float64) (string, string, int, error) {
	e := strconv.FormatFloat(f, 'e', -1, 64)
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

	return sign, strings.TrimLeft(digits, "0")
}

func decimalNotation(sign, digits string, exp int) string {
	if digits == "" {
		return "0"
	}
	return signedDecimalNotation(sign, digits, exp)
}

func signedDecimalNotation(sign, digits string, exp int) string {
	n := exp + 1
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
