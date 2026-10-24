import { shortDigest } from "../../shared/words";
import type { LineagePlant, LineageRun, LineageSource, SourceStatus } from "./lineageTypes";

export interface ChainStep {
  readonly label: string;
  readonly value: string;
  readonly detail: string;
  readonly title: string;
  readonly tone: "ok" | "warn" | "";
}

export interface RunLine {
  readonly id: string;
  readonly facts: string;
  readonly verdict: string;
  readonly current: boolean;
}

const statusWords: Readonly<Record<SourceStatus, string>> = {
  in_sync: "in sync with this plant",
  out_of_sync: "out of sync: it no longer produces this plant",
  does_not_compose: "does not compose",
  does_not_scaffold: "does not scaffold a twin",
  kind_files_missing: "device files are missing",
};

function sourceStep(label: string, source: LineageSource): ChainStep {
  const words = statusWords[source.status];
  const reason = source.reason === "" ? "" : `: ${source.reason}`;
  return {
    label,
    value: `${source.id} ${source.version}`,
    detail: `${source.file} · ${words}${reason}`,
    title: source.file,
    tone: source.status === "in_sync" ? "ok" : "warn",
  };
}

function devicesStep(plant: LineagePlant): ChainStep {
  const missing = plant.kinds.filter((kind) => !kind.found).map((kind) => kind.kind);
  const names = plant.kinds.map((kind) => kind.kind).join(", ");
  const lost = missing.length === 0 ? "" : `missing: ${missing.join(", ")}`;
  return { label: "Device files", value: names, detail: lost, title: "", tone: missing.length === 0 ? "ok" : "warn" };
}

function plantStep(plant: LineagePlant): ChainStep {
  return {
    label: "This plant",
    value: `${plant.id} ${plant.version}`,
    detail: plant.file,
    title: plant.file,
    tone: "",
  };
}

export function chainOf(plant: LineagePlant): readonly ChainStep[] {
  return [
    ...(plant.kinds.length > 0 ? [devicesStep(plant)] : []),
    ...(plant.twin === null ? [] : [sourceStep("Twin choices", plant.twin)]),
    ...(plant.composition === null ? [] : [sourceStep("Composition", plant.composition)]),
    plantStep(plant),
  ];
}

export function digestOf(plant: LineagePlant): { readonly short: string; readonly full: string } {
  return { short: shortDigest(plant.digest), full: plant.digest };
}

export const earlierVersion = "ran on an earlier version of this plant";

function runVerdict(run: LineageRun): string {
  if (!run.recorded) return "no run record to compare";
  return run.sameAsCurrent ? "same as current" : earlierVersion;
}

export function runLines(plant: LineagePlant): readonly RunLine[] {
  return plant.runs.map((run) => ({
    id: run.id,
    facts: [run.mode, run.domainVersion === "" ? "" : `version ${run.domainVersion}`].filter(Boolean).join(" · "),
    verdict: runVerdict(run),
    current: run.sameAsCurrent,
  }));
}
