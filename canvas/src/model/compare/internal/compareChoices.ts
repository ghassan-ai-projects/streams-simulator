import { durationOf } from "../../shared/duration";
import { seedWords } from "../../shared/words";

export const referenceRole = "Reference (A)";
export const thisRunRole = "This run (B)";

interface ListedRun {
  readonly worldId: string;
  readonly ended: boolean;
  readonly complete: boolean;
  readonly explore: boolean;
  readonly seed: number | null;
  readonly plant: string;
  readonly durationS: number;
  readonly name?: string;
  readonly requestedS?: number;
  readonly terminationReason?: string;
}

export interface ReferenceChoice {
  readonly worldId: string;
  readonly title: string;
  readonly facts: string;
}

function lengthWords(run: ListedRun): string {
  const reached = durationOf(run.durationS);
  const asked = run.requestedS === undefined ? "" : durationOf(run.requestedS);
  return asked !== "" && asked !== reached ? `${reached} of ${asked}` : reached;
}

function endingWords(run: ListedRun): string {
  if (run.terminationReason === "stopped_by_operator") return "stopped by operator";
  return run.complete ? "complete" : "did not finish";
}

function choiceOf(run: ListedRun): ReferenceChoice {
  const named = run.name !== undefined && run.name !== "";
  const facts = [
    named ? run.worldId : "",
    run.explore ? "Explore" : "Evaluate",
    seedWords(run.seed),
    lengthWords(run),
    endingWords(run),
  ];
  return {
    worldId: run.worldId,
    title: named ? (run.name ?? run.worldId) : run.worldId,
    facts: facts.filter((part) => part !== "").join(" · "),
  };
}

export function referenceChoices(
  runs: readonly ListedRun[],
  { current, plant }: { readonly current: string; readonly plant: string },
): readonly ReferenceChoice[] {
  return runs.filter((run) => run.ended && run.plant === plant && run.worldId !== current).map(choiceOf);
}
