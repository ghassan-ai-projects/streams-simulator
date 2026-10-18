package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func withProvenance(captured Capture, scan Scan, message Message) Capture {
	captured.BirthsSeen = scan.Seen
	captured.Taken = Provenance{Line: scan.Taken.Line, BDSeq: message.BDSeq}
	captured.LaterBirths = laterBirths(scan)
	captured.Issues = provenanceIssues(scan, message)
	return captured
}

func laterBirths(scan Scan) []LaterBirth {
	var later []LaterBirth
	for _, sighting := range scan.Later {
		later = append(later, LaterBirth{Line: sighting.Line, BDSeq: sighting.BDSeq, SameContent: sighting.Digest == scan.Taken.Digest})
	}
	return later
}

func provenanceIssues(scan Scan, message Message) []model.Issue {
	issues := unreadableIssues(scan)
	issues = append(issues, capabilityIssues(message)...)
	issues = append(issues, rebirthIssues(scan)...)
	if scan.Truncated {
		issues = append(issues, model.UnlocatedRuleIssue("scan_truncated",
			fmt.Sprintf("the scan stopped after %d bytes: births_seen counts only what was read", MaxScanBytes)))
	}
	return issues
}

func unreadableIssues(scan Scan) []model.Issue {
	var issues []model.Issue
	for _, skipped := range scan.Unreadable {
		issues = append(issues, model.UnlocatedRuleIssue("birth_line_unreadable", fmt.Sprintf("line %d names a birth but cannot be read: %v", skipped.Line, skipped.Cause)))
	}
	return issues
}

func capabilityIssues(message Message) []model.Issue {
	if CapabilityDigest(message.Kinds) == message.CapabilityDigest {
		return nil
	}
	text := fmt.Sprintf("the birth announces capability_digest %q, its kinds digest to %s", clip(message.CapabilityDigest), CapabilityDigest(message.Kinds))
	return []model.Issue{model.RuleIssue("/capability_digest", "capability_digest_mismatch", "capability_digest", text)}
}

func rebirthIssues(scan Scan) []model.Issue {
	var issues []model.Issue
	for _, later := range scan.Later {
		if later.Digest != scan.Taken.Digest {
			text := fmt.Sprintf("the birth at line %d (bd_seq %d) announces other kinds than the one taken at line %d", later.Line, later.BDSeq, scan.Taken.Line)
			issues = append(issues, model.UnlocatedRuleIssue("rebirth_differs", text))
		}
	}
	return issues
}
