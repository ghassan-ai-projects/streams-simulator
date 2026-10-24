import type { Definition, PlantEntry, RunEntry, RunRow, Workspace } from "./types";
import { durationOf } from "../../shared/duration";
import { counted, seedWords } from "../../shared/words";

function firstText(...values: readonly (string | undefined)[]): string {
  return values.find((value) => value !== undefined && value !== "") ?? "";
}

function idOf(definition: Definition): string {
  return firstText(definition.id, (definition.path.split("/").at(-1) ?? "").split(".")[0], definition.path);
}

function entry(id: string, domain: Definition | undefined, composition: Definition | undefined, runs: RunEntry[]) {
  const mine = runs.filter((run) => run.plant === id);
  return {
    id,
    title: firstText(domain?.title, composition?.title, id),
    domain: domain?.path ?? null,
    composition: composition?.path ?? null,
    instances: domain?.instances ?? 0,
    runs: mine.length,
    lastRun: mine.map((run) => run.recordedAt).reduce((a, b) => (b > a ? b : a), ""),
  };
}

function faultWords(run: RunEntry): string {
  if (!run.ended || !run.complete) return "";
  return run.faults === 0 ? "no faults" : counted(run.faults, "fault");
}

export function unfinishedReason(run: RunEntry): string {
  return run.incompleteReason || "the run did not close.";
}

export function openRunWords(run: RunEntry): string {
  if (run.ended) return "";
  return run.running ? "Running" : "Did not finish";
}

export function outcomeWords(run: RunEntry): string {
  if (run.connected === "running") return "";
  return run.ended && !run.complete ? `Did not finish: ${unfinishedReason(run)}` : "";
}

export function runModeWords(run: RunEntry): string {
  if (run.connected === "running") return "Connected session · running";
  if (run.connected === "saved") return "Connected session";
  return run.explore ? "Explore" : "Evaluate";
}

export function runFacts(run: RunEntry): readonly string[] {
  return [durationOf(run.durationS), seedWords(run.seed), faultWords(run), outcomeWords(run)].filter(
    (part) => part !== "",
  );
}

export function plantFacts(plant: PlantEntry): string {
  if (plant.domain === null) return "Not runnable yet. Edit it, then save.";
  const size = plant.instances > 0 ? `${counted(plant.instances, "instance")} · ` : "";
  if (plant.runs === 0) return size === "" ? "No runs yet" : `${size}no runs yet`;
  const last = whenOf(plant.lastRun);
  return `${size}${counted(plant.runs, "run")}${last === "" ? "" : ` · last ${last}`}`;
}

export function plantsOf(workspace: Workspace): readonly PlantEntry[] {
  const ids = [...new Set([...workspace.domains, ...workspace.compositions].map(idOf))];
  const runs = [...workspace.runs];
  return ids
    .map((id) =>
      entry(
        id,
        workspace.domains.find((d) => idOf(d) === id),
        workspace.compositions.find((c) => idOf(c) === id),
        runs,
      ),
    )
    .sort((a, b) => a.title.localeCompare(b.title));
}

function newerFirst(a: RunEntry, b: RunEntry): number {
  if (a.recordedAt !== b.recordedAt) return b.recordedAt.localeCompare(a.recordedAt);
  return b.worldId.localeCompare(a.worldId, undefined, { numeric: true });
}

export function runRows(workspace: Workspace, plant: string | null): readonly RunRow[] {
  const titles = new Map(plantsOf(workspace).map((p) => [p.id, p.title] as const));
  return workspace.runs
    .filter((run) => plant === null || run.plant === plant)
    .sort(newerFirst)
    .map((run) => ({ run, plantTitle: titles.get(run.plant) ?? (run.plant || "Unknown plant") }));
}

export function workspaceSummary(workspace: Workspace): string {
  return [counted(plantsOf(workspace).length, "plant"), counted(workspace.runs.length, "run")].join(" · ");
}

export function workspaceEmpty(workspace: Workspace): boolean {
  return (
    workspace.domains.length + workspace.compositions.length + workspace.scenarios.length + workspace.runs.length === 0
  );
}

export function whenOf(iso: string): string {
  const date = new Date(iso);
  if (iso === "" || Number.isNaN(date.getTime())) return "";
  const day = date.toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" });
  const time = date.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
  return `${day}, ${time}`;
}

export function suggestedId(name: string, taken: readonly string[]): string {
  const base = name
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/^[^a-z]+/, "")
    .slice(0, 56);
  const stem = base.length >= 3 ? base : `plant-${base}`.replace(/-$/, "");
  let candidate = stem;
  for (let n = 2; taken.includes(candidate); n += 1) candidate = `${stem}-${String(n)}`;
  return candidate;
}
