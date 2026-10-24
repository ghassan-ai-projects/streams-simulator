import { isRecord, list, record, text } from "../model/shared/json";
import type { Canned } from "./fakeFailures";
import { cannedFailure } from "./fakeFailures";
import { failure, success } from "./outcome";

interface Issue {
  readonly pointer: string;
  readonly message: string;
  readonly keyword: string;
  readonly member?: string;
}

function metricIssues(metric: Record<string, unknown>, at: number, twin: boolean): readonly Issue[] {
  const pointer = `/metrics/${String(at)}`;
  const name = text(metric.name);
  const numeric = metric.data_type === "number" || metric.data_type === "integer";
  const issues: Issue[] = [];
  if (/\s/.test(name))
    issues.push({ pointer: `${pointer}/name`, message: "does not match pattern", keyword: "pattern" });
  if (twin && numeric && text(metric.unit) === "") {
    issues.push({ pointer: `${pointer}/unit`, message: "needs a unit", keyword: "twin_unit_required", member: "unit" });
  }
  if (twin && metric.resolution === undefined) {
    const message = 'needs "resolution"';
    issues.push({ pointer: `${pointer}/resolution`, message, keyword: "twin_member_required", member: "resolution" });
  }
  if (twin && metric.scan_class !== undefined) {
    const message = "is not part of the interface a twin reproduces";
    issues.push({
      pointer: `${pointer}/scan_class`,
      message,
      keyword: "twin_member_unsupported",
      member: "scan_class",
    });
  }
  return issues;
}

export function kindCheck(args: Readonly<Record<string, unknown>>): Canned<unknown> {
  const twin = args.twin === true;
  let parsed: unknown;
  try {
    parsed = JSON.parse(text(args.content));
  } catch {
    return failure({
      ...cannedFailure("kind_failed", 422, "not JSON"),
      issues: [{ message: "not JSON", keyword: "json" }],
    });
  }
  const doc = isRecord(parsed) ? parsed : {};
  const metrics = list(doc.metrics).map(record);
  const commands = list(doc.commands).map(record);
  const named = /\s/.test(text(doc.kind))
    ? [{ pointer: "/kind", message: "does not match pattern", keyword: "pattern" }]
    : [];
  const issues = [...named, ...metrics.flatMap((metric, at) => metricIssues(metric, at, twin))];
  if (issues.length > 0) return failure({ ...cannedFailure("kind_failed", 422, "the kind is refused"), issues });
  const summary = { path: "<stdin>", kind: text(doc.kind), kind_digest: "sha256:abc", metrics: metrics.length };
  return success({ valid: true, kinds: [{ ...summary, commands: commands.length }] });
}
