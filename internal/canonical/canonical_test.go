package canonical

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

// timeValue adapts a nanosecond epoch to the type canonical.Marshal renders.
func timeValue(ns int64) time.Time {
	return time.Unix(0, ns)
}

// must is a helper that fails the test if err != nil.
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRFC8785WorkedExample(t *testing.T) {
	// The worked example from RFC 8785 §3.2, verbatim. The input string
	// contains two backslashes (one from \u005c, one from \\), so the
	// canonical output escapes both.
	in := `{"numbers": [333333333.33333329, 1E30, 4.50, 2e-3, 0.000000000000000000000000001], "string": "\u20ac$\u000F\u000aA'\u0042\u0022\u005c\\\"/", "literals": [null, true, false]}`
	want := "{\"literals\":[null,true,false],\"numbers\":[333333333.3333333,1e+30,4.5,0.002,1e-27],\"string\":\"€$\\u000f\\nA'B\\\"\\\\\\\\\\\"/\"}"
	var v any
	if err := json.Unmarshal([]byte(in), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, err := Marshal(v)
	must(t, err)
	if string(got) != want {
		t.Fatalf("RFC 8785 example mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestKeySorting(t *testing.T) {
	// Keys sort by UTF-16 code unit; uppercase precedes lowercase.
	got, err := MarshalString(map[string]any{"b": 1, "A": 2, "a": 3})
	must(t, err)
	if got != `{"A":2,"a":3,"b":1}` {
		t.Fatalf("sort mismatch: %s", got)
	}
}

func TestUTF16KeyOrdering(t *testing.T) {
	// U+20BB7 encodes as a surrogate pair (D842 DFB7); the first code unit
	// D842 sorts before U+E000, so it comes first even though its code point
	// is larger. UTF-8 byte order would get this wrong.
	got, err := MarshalString(map[string]any{"\uE000": 1, "\U00020BB7": 2})
	must(t, err)
	if got != "{\"\U00020BB7\":2,\"\uE000\":1}" {
		t.Fatalf("utf-16 ordering mismatch: %s", got)
	}
}

func TestNumberForms(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`1`, `1`},
		{`-1`, `-1`},
		{`0`, `0`},
		{`-0`, `0`},
		{`1.0`, `1`},
		{`0.5`, `0.5`},
		{`0.10`, `0.1`},
		{`1e+06`, `1000000`},
		{`1.5e+06`, `1500000`},
		{`1e21`, `1e+21`},
		{`1e-7`, `1e-7`},
		{`2e-3`, `0.002`},
		{`123456789012345678901`, `123456789012345678901`}, // big int literal, kept verbatim
		{`18446744073709551615`, `18446744073709551615`},   // uint64 max (seed range)
		{`-1.23e+45`, `-1.23e+45`},
		{`100000000000000000000`, `100000000000000000000`}, // 1e20 integral, decimal form
	}
	for _, tc := range cases {
		// Decode with UseNumber so integer literals beyond float64 precision
		// reach Marshal exactly as written.
		dec := json.NewDecoder(strings.NewReader(tc.in))
		dec.UseNumber()
		var v any
		if err := dec.Decode(&v); err != nil {
			t.Fatalf("decode %s: %v", tc.in, err)
		}
		got, err := MarshalString(v)
		must(t, err)
		if got != tc.want {
			t.Errorf("number %s: got %s, want %s", tc.in, got, tc.want)
		}
	}
}

func TestStringEscapes(t *testing.T) {
	got, err := MarshalString("a\"b\\c\nd\u0001\u2028\u2029e")
	must(t, err)
	if got != `"a\"b\\c\nd\u0001\u2028\u2029e"` {
		t.Fatalf("escape mismatch: %s", got)
	}
}

func TestDeterministicAcrossMapOrders(t *testing.T) {
	a := map[string]any{"x": 1, "y": []any{1, 2, map[string]any{"z": "s"}}}
	b := map[string]any{"y": []any{1, 2, map[string]any{"z": "s"}}, "x": 1}
	sa, err := MarshalString(a)
	must(t, err)
	sb, err := MarshalString(b)
	must(t, err)
	if sa != sb {
		t.Fatalf("canonical JSON depends on map construction order:\n%s\n%s", sa, sb)
	}
}

func TestDigestStable(t *testing.T) {
	d1, err := Digest(map[string]any{"v": 6.5, "t": int64(1785315600000000000)})
	must(t, err)
	d2, err := Digest(map[string]any{"t": int64(1785315600000000000), "v": 6.5})
	must(t, err)
	if d1 != d2 {
		t.Fatalf("digest not canonical: %s != %s", d1, d2)
	}
	if !strings.HasPrefix(d1, "sha256:") || len(d1) != len("sha256:")+64 {
		t.Fatalf("bad digest shape: %s", d1)
	}
}

func TestRejectsNonFinite(t *testing.T) {
	if _, err := MarshalString(math.Inf(1)); err == nil {
		t.Fatal("Inf should be rejected")
	}
	if _, err := MarshalString(math.NaN()); err == nil {
		t.Fatal("NaN should be rejected")
	}
}

func TestRejectsUnsupportedTypes(t *testing.T) {
	if _, err := MarshalString(struct{ A int }{1}); err == nil {
		t.Fatal("struct should be rejected")
	}
	if _, err := MarshalString([]int{1}); err == nil {
		t.Fatal("[]int should be rejected")
	}
}

func TestTimeFormatting(t *testing.T) {
	// A fixed instant must render as RFC 3339 with nanoseconds, UTC.
	ns := int64(1785315600123456789)
	got, err := MarshalString(timeValue(ns))
	must(t, err)
	want := `"` + time.Unix(0, ns).UTC().Format(time.RFC3339Nano) + `"`
	if got != want {
		t.Fatalf("time mismatch: got %s, want %s", got, want)
	}
}
