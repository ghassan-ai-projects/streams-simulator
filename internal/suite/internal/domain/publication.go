package domain

import "fmt"

const DefaultSampleNS int64 = 120 * 1e9

func FileName(domainID, profile string) string {
	return fmt.Sprintf("%s-%s-suite.json", domainID, profile)
}

type UnknownProfileError struct {
	Domain, Profile string
}

func (e *UnknownProfileError) Error() string {
	return fmt.Sprintf("suite: domain %q has no profile %q", e.Domain, e.Profile)
}
