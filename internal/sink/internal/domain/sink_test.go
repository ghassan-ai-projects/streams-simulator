package domain

import "testing"

func TestInprocAppendsLinesInOrderWithNewlines(t *testing.T) {
	t.Parallel()
	var s Inproc
	for _, line := range []string{"a", "b"} {
		if err := s.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Close()
	if err != nil || string(got) != "a\nb\n" {
		t.Fatalf("stream = %q (%v)", got, err)
	}
}
