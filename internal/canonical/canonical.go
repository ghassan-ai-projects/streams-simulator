// Package canonical implements RFC 8785 canonical JSON serialization,
// which every digest the simulator computes is based on. Canonical JSON
// makes the digest a pure function of the data, independent of map
// iteration order, number formatting, or whitespace.
//
// Supported input values are exactly what the simulator's own JSON path
// produces: nil, bool, json.Number, float64, int/uint of any width, string,
// []any, map[string]any, and time.Time (rendered RFC 3339 with nanoseconds).
// Anything else is rejected rather than guessed at.
package canonical

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

// Marshal returns the RFC 8785 canonical serialization of v.
//
// Rules applied (RFC 8785 §3.2):
//   - object keys sorted lexicographically by UTF-16 code units;
//   - numbers serialized per ECMAScript Number::toString (shortest round-trip,
//     no exponent outside the ES bounds, no trailing ".0");
//   - strings escaped per JSON with U+2028/U+2029 additionally escaped;
//   - no insignificant whitespace;
//   - only finite numbers are representable; NaN and Inf are errors.
func Marshal(v any) ([]byte, error) {
	var b strings.Builder
	if err := writeValue(&b, v); err != nil {
		return nil, fmt.Errorf("canonical: %w", err)
	}
	return []byte(b.String()), nil
}

// MarshalString is Marshal returning a string.
func MarshalString(v any) (string, error) {
	b, err := Marshal(v)
	return string(b), err
}

// Digest returns the RFC 8785 canonical digest of v as "sha256:<hex>".
func Digest(v any) (string, error) {
	b, err := Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// DigestBytes returns "sha256:<hex>" of raw bytes.
func DigestBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func writeValue(b *strings.Builder, v any) error {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case json.Number:
		s, err := canonicalNumber(x.String())
		if err != nil {
			return fmt.Errorf("canonical: %w", err)
		}
		b.WriteString(s)
	case float64:
		s, err := esNumber(x)
		if err != nil {
			return fmt.Errorf("canonical: %w", err)
		}
		b.WriteString(s)
	case int:
		b.WriteString(strconv.Itoa(x))
	case int8:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int16:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int32:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case uint:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint8:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint16:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint32:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint64:
		b.WriteString(strconv.FormatUint(x, 10))
	case string:
		writeString(b, x)
	case time.Time:
		writeString(b, x.UTC().Format(time.RFC3339Nano))
	case []any:
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := writeValue(b, e); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]any:
		if err := writeObject(b, x); err != nil {
			return err
		}
	default:
		return fmt.Errorf("canonical: unsupported value type %T", v)
	}
	return nil
}

func writeObject(b *strings.Builder, m map[string]any) error {
	keys := make([]string, 0, len(m))
	// determinism-safe: keys are collected here but sorted by UTF-16 code
	// units before any output is written.
	for k := range m {
		if !utf8.ValidString(k) {
			return fmt.Errorf("canonical: object key is not valid UTF-8")
		}
		keys = append(keys, k)
	}
	// RFC 8785 §3.2.3: sort keys by UTF-16 code units, not code points.
	sort.Slice(keys, func(i, j int) bool {
		return utf16Less(keys[i], keys[j])
	})
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		writeString(b, k)
		b.WriteByte(':')
		if err := writeValue(b, m[k]); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func utf16Less(a, b string) bool {
	au := utf16.Encode([]rune(a))
	bu := utf16.Encode([]rune(b))
	for i := 0; i < len(au) && i < len(bu); i++ {
		if au[i] != bu[i] {
			return au[i] < bu[i]
		}
	}
	return len(au) < len(bu)
}

func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case 0x2028:
			b.WriteString(`\u2028`)
		case 0x2029:
			b.WriteString(`\u2029`)
		default:
			if r < 0x20 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

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
