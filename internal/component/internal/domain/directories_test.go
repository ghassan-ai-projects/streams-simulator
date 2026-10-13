package domain

import (
	"slices"
	"testing"
)

func TestCustomerDirectoriesComeFromFlagsThenTheEnvironmentWithoutEmptyEntries(t *testing.T) {
	t.Parallel()
	got := CustomerDirectories([]string{"/acme/lib", " "}, "/acme/shared::/acme/site:")
	if want := []string{"/acme/lib", "/acme/shared", "/acme/site"}; !slices.Equal(got, want) {
		t.Fatalf("directories = %v, want %v", got, want)
	}
	if got := CustomerDirectories(nil, ""); got == nil || len(got) != 0 {
		t.Fatalf("no flags and no environment give an empty list, got %#v", got)
	}
}
