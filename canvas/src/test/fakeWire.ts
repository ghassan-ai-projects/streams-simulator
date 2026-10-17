import type { Workspace } from "../model/workspace";
import type { Canned } from "./fakeFailures";

export class FakeSource extends EventTarget {
  readyState = 1;

  constructor(private readonly onClose: (source: FakeSource) => void) {
    super();
  }

  close(): void {
    this.readyState = 2;
    this.onClose(this);
  }

  record(entry: object): void {
    this.dispatchEvent(new MessageEvent("record", { data: JSON.stringify(entry) }));
  }

  end(how: "closed" | "failed"): void {
    if (how === "failed") this.readyState = 2;
    this.dispatchEvent(new Event(how === "closed" ? "closed" : "error"));
  }
}

export function answer(outcome: Canned<unknown>): Response {
  if (outcome.ok) return Response.json(outcome.value);
  const { status, code, message, issues, result } = outcome.failure;
  if (code === "unreachable") throw new TypeError(message);
  return Response.json({ error: { code, message, issues, result } }, { status });
}

export function workspaceWire(workspace: Workspace) {
  return {
    name: workspace.name,
    domains: workspace.domains,
    compositions: workspace.compositions,
    scenarios: workspace.scenarios,
    runs: workspace.runs.map((run) => ({
      world_id: run.worldId,
      ended: run.ended,
      complete: run.complete,
      incomplete_reason: run.incompleteReason,
      explore: run.explore,
      seed: run.seed,
      plant: run.plant,
      duration_ns: run.durationS * 1e9,
      recorded_at: run.recordedAt,
      faults: run.faults,
    })),
  };
}
