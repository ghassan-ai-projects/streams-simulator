package domain

import "crypto/subtle"

const maxPairingFailures = 5

type Pairing struct {
	code     string
	failures int
	used     bool
	sessions map[string]bool
}

func NewPairing(code string) *Pairing {
	return &Pairing{code: code, sessions: map[string]bool{}}
}

func (p *Pairing) Pair(offered, sessionID string) error {
	switch {
	case p.failures >= maxPairingFailures:
		return ErrPairingLocked
	case p.used:
		return ErrPairingUsed
	case subtle.ConstantTimeCompare([]byte(NormalizeCode(offered)), []byte(p.code)) != 1:
		p.failures++
		return ErrWrongCode
	}
	p.used = true
	p.sessions[sessionID] = true
	return nil
}

func (p *Pairing) Admits(sessionID string) bool {
	return sessionID != "" && p.sessions[sessionID]
}

func NormalizeCode(code string) string {
	out := make([]byte, 0, len(code))
	for _, c := range []byte(code) {
		switch {
		case c >= 'a' && c <= 'z':
			out = append(out, c-'a'+'A')
		case c >= 'A' && c <= 'Z', c >= '2' && c <= '7':
			out = append(out, c)
		}
	}
	return string(out)
}

func FormatCode(code string) string {
	if len(code) <= 4 {
		return code
	}
	return code[:4] + "-" + FormatCode(code[4:])
}
