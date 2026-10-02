package canonical

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

func writeValue(b *strings.Builder, v any) error {
	switch x := v.(type) {
	case []any:
		return writeArray(b, x)
	case map[string]any:
		return writeObject(b, x)
	default:
		return writeScalar(b, v)
	}
}

func writeScalar(b *strings.Builder, v any) error {
	switch x := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(x))
	case string:
		writeString(b, x)
	case time.Time:
		writeString(b, x.UTC().Format(time.RFC3339Nano))
	default:
		return writeNumber(b, v)
	}
	return nil
}

func writeNumber(b *strings.Builder, v any) error {
	text, err := numberText(v)
	if err != nil {
		return fmt.Errorf("canonical: %w", err)
	}
	b.WriteString(text)
	return nil
}

func numberText(v any) (string, error) {
	switch x := v.(type) {
	case json.Number:
		return canonicalNumber(x.String())
	case float64:
		return esNumber(x)
	case int, int8, int16, int32, int64:
		return signedNumber(v), nil
	case uint, uint8, uint16, uint32, uint64:
		return unsignedNumber(v), nil
	default:
		return "", fmt.Errorf("unsupported value type %T", v)
	}
}

func signedNumber(v any) string {
	switch x := v.(type) {
	case int:
		return strconv.Itoa(x)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	default:
		return strconv.FormatInt(v.(int64), 10)
	}
}

func unsignedNumber(v any) string {
	switch x := v.(type) {
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	default:
		return strconv.FormatUint(v.(uint64), 10)
	}
}

func writeArray(b *strings.Builder, values []any) error {
	b.WriteByte('[')
	for i, value := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := writeValue(b, value); err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

func writeObject(b *strings.Builder, m map[string]any) error {
	keys, err := objectKeys(m)
	if err != nil {
		return err
	}
	b.WriteByte('{')
	if err := writeObjectMembers(b, m, keys); err != nil {
		return err
	}
	b.WriteByte('}')
	return nil
}

func objectKeys(m map[string]any) ([]string, error) {
	keys := make([]string, 0, len(m))
	// determinism-safe: collected keys are sorted before writing output.
	for k := range m {
		if !utf8.ValidString(k) {
			return nil, fmt.Errorf("canonical: object key is not valid UTF-8")
		}
		keys = append(keys, k)
	}
	// RFC 8785 §3.2.3: sort by UTF-16 code units, not code points.
	sort.Slice(keys, func(i, j int) bool { return utf16Less(keys[i], keys[j]) })
	return keys, nil
}

func writeObjectMembers(b *strings.Builder, m map[string]any, keys []string) error {
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
	return nil
}

func utf16Less(a, b string) bool {
	au, bu := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
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
		writeStringRune(b, r)
	}
	b.WriteByte('"')
}

func writeStringRune(b *strings.Builder, r rune) {
	if escaped, ok := escapedRunes[r]; ok {
		b.WriteString(escaped)
	} else if r < 0x20 {
		fmt.Fprintf(b, `\u%04x`, r)
	} else {
		b.WriteRune(r)
	}
}

// Escapes preserve the existing digest representation, including U+2028/2029.
var escapedRunes = map[rune]string{
	'"': `\"`, '\\': `\\`, '\b': `\b`, '\f': `\f`, '\n': `\n`,
	'\r': `\r`, '\t': `\t`, 0x2028: `\u2028`, 0x2029: `\u2029`,
}
