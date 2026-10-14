package domain

import "testing"

func TestTheSessionCookieNameIsSpecificToTheBoundPort(t *testing.T) {
	t.Parallel()
	a, b := SessionCookieName("127.0.0.1:4100"), SessionCookieName("127.0.0.1:4101")
	if a == b || a != "streamsim_canvas_4100" || SessionCookieName("[::1]:4100") != a {
		t.Fatalf("names %q, %q", a, b)
	}
}
