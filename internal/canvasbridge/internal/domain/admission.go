package domain

import (
	"fmt"
	"strconv"
	"strings"
)

type Request struct {
	Host         string
	Origin       string
	SecFetchSite string
}

func Admit(request Request, bound string) error {
	if !admittedHost(request.Host, bound) {
		return fmt.Errorf("%w: host %q is not the bridge", ErrNotAdmitted, request.Host)
	}
	if request.Origin != "" && request.Origin != "http://"+request.Host {
		return fmt.Errorf("%w: origin %q is not the bridge", ErrNotAdmitted, request.Origin)
	}
	if site := request.SecFetchSite; site != "" && site != "same-origin" && site != "none" {
		return fmt.Errorf("%w: cross-site request (%s)", ErrNotAdmitted, site)
	}
	return nil
}

func admittedHost(host, bound string) bool {
	_, port, ok := splitHostPort(bound)
	return ok && (host == bound || host == "localhost:"+port || host == "127.0.0.1:"+port)
}

func AdmitListenAddress(addr string) error {
	host, _, ok := splitHostPort(addr)
	if !ok || !loopback(host) {
		return fmt.Errorf("%w: %q", ErrNonLoopback, addr)
	}
	return nil
}

func splitHostPort(addr string) (string, string, bool) {
	cut := strings.LastIndexByte(addr, ':')
	if cut < 0 {
		return "", "", false
	}
	host, port := strings.TrimSuffix(strings.TrimPrefix(addr[:cut], "["), "]"), addr[cut+1:]
	if number, err := strconv.Atoi(port); err != nil || number < 0 || number > 65535 {
		return "", "", false
	}
	return host, port, true
}

func loopback(host string) bool {
	if host == "localhost" || host == "::1" {
		return true
	}
	octets := strings.Split(host, ".")
	if len(octets) != 4 || octets[0] != "127" {
		return false
	}
	for _, octet := range octets {
		if number, err := strconv.Atoi(octet); err != nil || number < 0 || number > 255 {
			return false
		}
	}
	return true
}

func CookieValue(header, name string) string {
	for part := range strings.SplitSeq(header, ";") {
		if key, value, found := strings.Cut(strings.TrimSpace(part), "="); found && key == name {
			return value
		}
	}
	return ""
}
