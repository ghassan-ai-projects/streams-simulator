package domain

import (
	"errors"
	"testing"
)

func TestOnlyRequestsFromTheBridgeOriginAreAdmitted(t *testing.T) {
	t.Parallel()
	bound := "127.0.0.1:4100"
	cases := []struct {
		name    string
		request Request
		ok      bool
	}{
		{"a same-origin fetch", Request{Host: bound, Origin: "http://" + bound, SecFetchSite: "same-origin"}, true},
		{"localhost alias", Request{Host: "localhost:4100", Origin: "http://localhost:4100"}, true},
		{"a navigation without origin", Request{Host: bound, SecFetchSite: "none"}, true},
		{"DNS rebinding", Request{Host: "evil.example:4100"}, false},
		{"another port", Request{Host: "127.0.0.1:4101"}, false},
		{"a cross-origin page", Request{Host: bound, Origin: "http://evil.example"}, false},
		{"a cross-site fetch without origin", Request{Host: bound, SecFetchSite: "cross-site"}, false},
		{"a same-site but other-origin fetch", Request{Host: bound, SecFetchSite: "same-site"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := Admit(tc.request, bound)
			if tc.ok != (err == nil) || (err != nil && !errors.Is(err, ErrNotAdmitted)) {
				t.Fatalf("Admit(%+v) = %v, want admitted %v", tc.request, err, tc.ok)
			}
		})
	}
}

func TestTheBridgeListensOnLoopbackOnly(t *testing.T) {
	t.Parallel()
	for addr, ok := range map[string]bool{"127.0.0.1:0": true, "[::1]:9000": true, "localhost:9000": true,
		"0.0.0.0:9000": false, "192.168.1.4:9000": false, ":9000": false, "nonsense": false} {
		if err := AdmitListenAddress(addr); ok != (err == nil) || (err != nil && !errors.Is(err, ErrNonLoopback)) {
			t.Errorf("AdmitListenAddress(%q) = %v, want ok %v", addr, err, ok)
		}
	}
}

func TestTheSessionCookieIsReadFromTheCookieHeader(t *testing.T) {
	t.Parallel()
	if got := CookieValue("a=1; streamsim_canvas=s-9; b=2", "streamsim_canvas"); got != "s-9" {
		t.Fatalf("CookieValue = %q", got)
	}
	if got := CookieValue("streamsim_canvas_x=s-9", "streamsim_canvas"); got != "" {
		t.Fatalf("a cookie with a longer name matched: %q", got)
	}
}

func TestAListenAddressNeedsAValidPortAndALoopbackHost(t *testing.T) {
	t.Parallel()
	cases := []struct {
		addr string
		ok   bool
	}{
		{"127.0.0.1:65535", true},
		{"127.255.255.254:80", true},
		{"[::1]:0", true},
		{"127.0.0.1:65536", false},
		{"127.0.0.1:-1", false},
		{"127.0.0.1:http", false},
		{"127.0.0.1:", false},
		{"127.0.0.300:80", false},
		{"127.0.0.x:80", false},
		{"127.0.0:80", false},
		{"127.0.0.1.1:80", false},
		{"128.0.0.1:80", false},
		{"[::2]:80", false},
		{"127.0.0.1", false},
	}
	for _, tc := range cases {
		t.Run(tc.addr, func(t *testing.T) {
			t.Parallel()
			err := AdmitListenAddress(tc.addr)
			if tc.ok != (err == nil) || (err != nil && !errors.Is(err, ErrNonLoopback)) {
				t.Fatalf("AdmitListenAddress(%q) = %v, want ok %v", tc.addr, err, tc.ok)
			}
		})
	}
}

func TestNothingIsAdmittedWhenTheBoundAddressHasNoValidPort(t *testing.T) {
	t.Parallel()
	for _, bound := range []string{"127.0.0.1", "127.0.0.1:port", "127.0.0.1:99999"} {
		t.Run(bound, func(t *testing.T) {
			t.Parallel()
			if err := Admit(Request{Host: bound}, bound); !errors.Is(err, ErrNotAdmitted) {
				t.Fatalf("Admit against bound %q = %v, want ErrNotAdmitted", bound, err)
			}
		})
	}
}
