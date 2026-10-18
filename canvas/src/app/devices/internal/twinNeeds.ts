import { catalogueKey, unstatedMember } from "../../../model/devices";
import type { FieldIssue } from "../../../model/devices";
import type { Draft } from "./deviceDraft";
import { twinNeedsOf } from "./deviceSelectors";
import type { TwinState } from "./twinDraft";

const rangeNote = "A range is needed for the instrument chosen for this reading";

function rangeNeed(key: string, member: string): FieldIssue {
  return { part: "metric", key, field: member, keyword: "", message: rangeNote, text: rangeNote, twin: true };
}

export function rangeNeedsOf(twin: TwinState): readonly FieldIssue[] {
  const lead = twin.phase === "ready" ? twin.members[0] : undefined;
  if (twin.phase !== "ready" || lead === undefined) return [];
  const choice = twin.doc.kinds.find((entry) => entry.kind === lead.path);
  return lead.kind.metrics.flatMap((metric) => {
    const ref = choice?.metrics[metric.key]?.instrument;
    const chosen = twin.instruments[catalogueKey(metric)]?.find((instrument) => instrument.ref === ref);
    const member = unstatedMember(chosen);
    return member === null ? [] : [rangeNeed(metric.key, member)];
  });
}

export function interfaceNeedsOf(device: Draft, twin: TwinState): readonly FieldIssue[] {
  return [...twinNeedsOf(device), ...rangeNeedsOf(twin)];
}
