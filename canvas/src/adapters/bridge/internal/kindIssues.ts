import type { KindChecked } from "../../../app/ports";
import { isRecord, list, record, text } from "../../../model/shared/json";

export function kindChecked(value: unknown): KindChecked | null {
  if (!isRecord(value) || value.valid !== true) return null;
  const first = record(list(value.kinds)[0]);
  return {
    kind: text(first.kind),
    digest: text(first.kind_digest),
    metrics: typeof first.metrics === "number" ? first.metrics : 0,
    commands: typeof first.commands === "number" ? first.commands : 0,
  };
}
