import type { Workspace } from "../model/workspace";
import type { Canned } from "./fakeFailures";

export class FakeSource extends EventTarget {
  readyState = 1;

  constructor(private readonly onClose: (source: FakeSource) => void = () => undefined) {
    super();
  }

  close(): void {
    this.readyState = 2;
    this.onClose(this);
  }

  record(entry: object): void {
    this.dispatchEvent(new MessageEvent("record", { data: JSON.stringify(entry) }));
  }

  reconnect(): void {
    this.readyState = 0;
    this.dispatchEvent(new Event("error"));
  }

  end(how: "closed" | "failed"): void {
    if (how === "failed") this.readyState = 2;
    this.dispatchEvent(new Event(how === "closed" ? "closed" : "error"));
  }
}

export function respond(status: number, body: unknown): Response {
  return typeof body === "string" ? new Response(body, { status }) : Response.json(body, { status });
}

export function answer(outcome: Canned<unknown>): Response {
  if (outcome.ok) return respond(200, outcome.value);
  const { status, code, message, issues, result } = outcome.failure;
  if (code === "unreachable") throw new TypeError(message);
  return respond(status, { error: { code, message, issues, result } });
}

export function workspaceWire(workspace: Workspace) {
  return {
    name: workspace.name,
    domains: workspace.domains,
    compositions: workspace.compositions,
    scenarios: workspace.scenarios,
    kinds: workspace.kinds,
    components: workspace.components,
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
