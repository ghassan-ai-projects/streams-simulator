import { issueText } from "./issues";
import { neededPhrase } from "./issueWords";
import type { RawIssue } from "./types";

function joined(parts: readonly string[]): string {
  const last = parts.at(-1);
  if (parts.length < 2 || last === undefined) return parts.join("");
  return `${parts.slice(0, -1).join(", ")} and ${last}`;
}

export function heldWords(device: string, issues: readonly RawIssue[]): string {
  const [first] = issues;
  if (first === undefined) return `${device} is not ready for a twin`;
  const phrases = issues.map((issue) => neededPhrase(issue.keyword ?? "", issue.member ?? ""));
  const asked = phrases.filter((phrase): phrase is string => phrase !== null);
  if (asked.length === issues.length) return `${device} still needs ${joined([...new Set(asked)])}`;
  return `${device} is not ready for a twin: ${issueText(first)}`;
}
