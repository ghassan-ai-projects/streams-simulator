package domain

import "fmt"

// DefaultSampleNS is the channel sampling period a published suite is
// generated with: two minutes.
const DefaultSampleNS int64 = 120 * 1e9

// FileName names the suite file of a domain and a profile.
func FileName(domainID, profile string) string {
	return fmt.Sprintf("%s-%s-suite.json", domainID, profile)
}

// UnknownProfileError refuses a suite for a profile its domain does not
// declare.
type UnknownProfileError struct {
	Domain, Profile string
}

// Error names the domain and the profile.
func (e *UnknownProfileError) Error() string {
	return fmt.Sprintf("suite: domain %q has no profile %q", e.Domain, e.Profile)
}
