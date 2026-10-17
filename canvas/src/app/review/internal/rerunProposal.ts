import type { Rerun } from "../../../model/review";
import { rerunOf } from "../../../model/review";
import { readyReport } from "./readyReport";
import type { ReviewState } from "./reviewState";

export interface RerunProposal extends Rerun {
  readonly domainId: string;
}

export function rerunProposalOf(state: ReviewState): RerunProposal | null {
  const report = readyReport(state);
  return report === null ? null : { ...rerunOf(report), domainId: report.domainId };
}
