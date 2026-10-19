import type { Definition, Workspace } from "../../../model/workspace";
import type { PlantSession, PlantState, ReadyPlant } from "../../plant";
import { readyReport } from "./readyReport";
import type { ReviewState } from "./reviewState";

export interface ReviewPlant {
  readonly plant: ReadyPlant | null;
  readonly title: string;
  readonly path: string | null;
  readonly key: string;
  readonly fallback: string | undefined;
}

function shownFor(shown: PlantState, key: string, fallback: string | undefined): ReadyPlant | null {
  return shown.phase === "ready" && (shown.path === key || shown.path === fallback) ? shown : null;
}

function plantTitle(plant: ReadyPlant | null, entry: Definition | undefined, domainId: string): string {
  const title = plant?.plant.title ?? entry?.title ?? domainId;
  return title === "" ? "Plant" : title;
}

export function reviewPlantFor(state: ReviewState, shown: PlantState, workspace: Workspace | null): ReviewPlant {
  const key = `run:${state.worldId ?? ""}`;
  const report = readyReport(state);
  const reported = report?.domainId ?? "";
  const domainId = reported === "" ? state.sessionDomain : reported;
  const entry = domainId === "" ? undefined : workspace?.domains.find((domain) => domain.id === domainId);
  const fallback = entry?.path;
  const plant = shownFor(shown, key, fallback);
  return { plant, title: plantTitle(plant, entry, domainId), path: entry?.path ?? null, key, fallback };
}

type Recorded = Pick<ReviewState, "runDomain" | "connection">;

function fallbackWanted({ runDomain, connection }: Recorded): boolean {
  return runDomain.phase === "failed" || (runDomain.phase === "loading" && connection === "running");
}

export function showRunPlant(
  plants: PlantSession,
  target: Pick<ReviewPlant, "key" | "fallback">,
  recorded: Recorded,
  workspace: Workspace | null,
): void {
  if (recorded.runDomain.phase === "ready") void plants.show(target.key, recorded.runDomain.value);
  else if (fallbackWanted(recorded) && target.fallback !== undefined && workspace !== null)
    void plants.load(target.fallback, workspace);
}
