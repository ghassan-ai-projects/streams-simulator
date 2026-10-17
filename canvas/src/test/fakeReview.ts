import type { DomainDocument, FileContent, Outcome, ReviewFailure } from "../app/ports";
import { failure as failed, success } from "../app/ports";
import type { Replay, Report } from "../model/review";
import type { Canned } from "./fakeFailures";
import { collect, cannedFailure, reviewFailure } from "./fakeFailures";
import { isReport, replayOf, runDomainOf } from "./fakeValues";

export interface ReviewSource {
  action(name: string, args: Readonly<Record<string, unknown>>): Promise<Canned<unknown>>;
  readFile(path: string): Promise<Canned<FileContent>>;
}

export function reviewReport(source: ReviewSource, runId: string): Promise<Outcome<Report, ReviewFailure>> {
  return source
    .action("run.report", { run: runId })
    .then((outcome) => collect(outcome, reviewFailure, (value) => (isReport(value) ? value : null)));
}

export function reviewVerify(source: ReviewSource, runId: string): Promise<Outcome<Replay, ReviewFailure>> {
  return source.action("run.verify", { run: runId }).then((outcome) => {
    const verdict = outcome.ok ? null : replayOf(outcome.failure.result);
    return verdict === null ? collect(outcome, reviewFailure, replayOf) : success(verdict);
  });
}

export async function reviewDomain(
  source: ReviewSource,
  runId: string,
): Promise<Outcome<DomainDocument, ReviewFailure>> {
  const file = await source.readFile(`runs/${runId}/run.json`);
  if (!file.ok) return failed(reviewFailure(file.failure));
  const domain = runDomainOf(file.value.content);
  return domain === null
    ? failed(reviewFailure(cannedFailure("unreadable", 200, "the run recorded no domain")))
    : success(domain);
}
