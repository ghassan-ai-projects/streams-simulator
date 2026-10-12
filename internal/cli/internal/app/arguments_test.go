package app

import (
	"io"
	"strings"
	"testing"
)

func TestFlagsMayFollowOrSeparatePositionalArguments(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, out  string
		args       []string
		positional []string
	}{
		{"flags first", "x", []string{"--out", "x", "a", "b"}, []string{"a", "b"}},
		{"flags last", "x", []string{"a", "b", "--out", "x"}, []string{"a", "b"}},
		{"flags between", "x", []string{"a", "--out", "x", "b"}, []string{"a", "b"}},
		{"terminator keeps dashes positional", "", []string{"a", "--", "--out", "x"}, []string{"a", "--out", "x"}},
		{"a flag's value is not the terminator", "--", []string{"--out", "--", "a"}, []string{"a"}},
		{"a value after = is not the terminator", "--", []string{"--out=--", "a", "--", "-b"}, []string{"a", "-b"}},
		{"a boolean flag takes no value", "", []string{"-v", "--", "-x"}, []string{"-x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fs := newFlagSet("test", io.Discard)
			out := fs.String("out", "", "")
			fs.Bool("v", false, "")
			if err := parseFlags(fs, tc.args); err != nil {
				t.Fatal(err)
			}
			if *out != tc.out || strings.Join(fs.Args(), "|") != strings.Join(tc.positional, "|") {
				t.Fatalf("out %q args %q, want %q and %q", *out, fs.Args(), tc.out, tc.positional)
			}
		})
	}
}

func TestAFlagAfterAPositionalArgumentIsStillChecked(t *testing.T) {
	t.Parallel()
	fs := newFlagSet("test", io.Discard)
	err := parseFlags(fs, []string{"a", "--no-such"})
	if exit, ok := asUsageExit(err); !ok || exit.code != 2 || !strings.Contains(exit.message, "no-such") {
		t.Fatalf("parseFlags = %v, want a usage exit naming the bad flag", err)
	}
}

// A command that takes no argument, or a verb given more than it takes,
// refuses the stray one instead of ignoring it.
func TestAStrayArgumentIsAUsageFailure(t *testing.T) {
	t.Parallel()
	cases := [][]string{
		{"run", "--domain", "aquaculture-pond", "extra"},
		{"suite", "--domain", "aquaculture-pond", "extra"},
		{"replay", "a.json", "b.json"},
		{"catalog", "list", "extra"},
		{"catalog", "describe", "a", "b"},
		{"scenario", "from-run", "a.json", "b.json", "--out", "s.json"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			got := invoke(t, args...)
			if got.code != 2 || got.failure(t)["code"] != "usage_invalid" || !strings.Contains(got.stderr, "unexpected argument") {
				t.Fatalf("%v: %+v, want usage_invalid naming the unexpected argument", args, got)
			}
		})
	}
}

// After a leading flag, a verb's arguments that follow "--" stay arguments.
func TestAVerbArgumentAfterTheTerminatorIsNotAFlag(t *testing.T) {
	t.Parallel()
	fs := newFlagSet("test", io.Discard)
	fs.String("dir", "", "")
	verb, err := parseCommandVerb(fs, []string{"--dir", "d", "--", "describe", "-x"}, "list")
	if err != nil || verb != "describe" || strings.Join(fs.Args(), "|") != "-x" {
		t.Fatalf("verb %q args %q err %v, want describe with -x", verb, fs.Args(), err)
	}
}
