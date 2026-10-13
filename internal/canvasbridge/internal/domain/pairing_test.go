package domain

import (
	"errors"
	"testing"
)

func TestAPairingCodeIsUsedOnceAndAdmitsOnlyItsSession(t *testing.T) {
	t.Parallel()
	p := NewPairing("ABCDEFGHJKLM")
	if err := p.Pair("abcd-efgh-jklm", "s-1"); err != nil {
		t.Fatalf("the right code, typed in lower case with dashes, is refused: %v", err)
	}
	if !p.Admits("s-1") || p.Admits("s-2") || p.Admits("") {
		t.Fatal("only the paired session is admitted")
	}
	if err := p.Pair("ABCDEFGHJKLM", "s-2"); !errors.Is(err, ErrPairingUsed) {
		t.Fatalf("a second pairing with the same code: %v, want ErrPairingUsed", err)
	}
}

func TestFiveWrongCodesLockPairing(t *testing.T) {
	t.Parallel()
	p := NewPairing("ABCDEFGHJKLM")
	for range maxPairingFailures {
		if err := p.Pair("WRONGWRONGWR", "s"); !errors.Is(err, ErrWrongCode) {
			t.Fatalf("a wrong code: %v, want ErrWrongCode", err)
		}
	}
	if err := p.Pair("ABCDEFGHJKLM", "s"); !errors.Is(err, ErrPairingLocked) || p.Admits("s") {
		t.Fatalf("the right code after five wrong ones: %v, want ErrPairingLocked", err)
	}
}

func TestCodesAreNormalizedAndFormattedInGroupsOfFour(t *testing.T) {
	t.Parallel()
	if got := NormalizeCode(" ab2c-d7e8 9! "); got != "AB2CD7E" {
		t.Fatalf("NormalizeCode = %q, want only base32 letters and digits, upper case", got)
	}
	if got := FormatCode("ABCDEFGHJKLM"); got != "ABCD-EFGH-JKLM" {
		t.Fatalf("FormatCode = %q", got)
	}
}
