import type { Bridge, RunFollower } from "../app/ports";
import type { BridgeFailure, FileContent, Outcome, Workspace, Written } from "../model/bridge";
import { failed, success } from "../model/bridge";

export const sampleWorkspace: Workspace = {
  name: "lab-workspace",
  domains: [{ path: "domains/factory-utilities.domain.json", id: "factory-utilities", title: "Factory utilities" }],
  compositions: [
    { path: "compositions/factory-utilities.composition.json", id: "factory-utilities", title: "Factory utilities" },
  ],
  scenarios: [],
  runs: [
    {
      worldId: "w-1",
      ended: true,
      complete: true,
      incompleteReason: "",
      explore: false,
      seed: 7,
      plant: "factory-utilities",
      durationS: 1800,
      recordedAt: "2026-10-14T10:00:00Z",
    },
    {
      worldId: "w-2",
      ended: false,
      complete: false,
      incompleteReason: "",
      explore: true,
      seed: null,
      plant: "factory-utilities",
      durationS: 0,
      recordedAt: "2026-10-14T11:00:00Z",
    },
  ],
};

export function failure(code: string, status = 401, message = "refused"): BridgeFailure {
  return { status, code, message, issues: [] };
}

export class FakeBridge implements Bridge {
  paired = false;
  sessionOutcome: Outcome<boolean> | null = null;
  pairOutcomes: Outcome<boolean>[] = [];
  workspaceOutcome: Outcome<Workspace> = success(sampleWorkspace);
  files = new Map<string, Outcome<FileContent>>();
  actions = new Map<string, Outcome<unknown>>();
  handlers = new Map<string, (args: Readonly<Record<string, unknown>>) => Outcome<unknown>>();
  calls: { name: string; args: Readonly<Record<string, unknown>> }[] = [];
  reads: string[] = [];
  watchers = new Set<() => void>();
  pairedCodes: string[] = [];

  session(): Promise<Outcome<boolean>> {
    return Promise.resolve(this.sessionOutcome ?? success(this.paired));
  }

  pair(code: string): Promise<Outcome<boolean>> {
    this.pairedCodes.push(code);
    const next = this.pairOutcomes.shift() ?? success(true);
    this.paired = next.ok;
    return Promise.resolve(next);
  }

  workspace(): Promise<Outcome<Workspace>> {
    return Promise.resolve(this.workspaceOutcome);
  }

  readFile(path: string): Promise<Outcome<FileContent>> {
    this.reads.push(path);
    return Promise.resolve(this.files.get(path) ?? failed(failure("not_found", 404, `no ${path}`)));
  }

  writes: { path: string; content: string; etag: string }[] = [];
  writeOutcomes = new Map<string, Outcome<Written>>();

  writeFile(path: string, content: string, etag: string): Promise<Outcome<Written>> {
    this.writes.push({ path, content, etag });
    const outcome = this.writeOutcomes.get(path) ?? success({ path, etag: `e${String(this.writes.length)}` });
    if (outcome.ok) this.files.set(path, success({ path, content, etag: outcome.value.etag }));
    return Promise.resolve(outcome);
  }

  action(name: string, args: Readonly<Record<string, unknown>>): Promise<Outcome<unknown>> {
    this.calls.push({ name, args });
    const handler = this.handlers.get(name);
    if (handler !== undefined) return Promise.resolve(handler(args));
    const set = this.actions.get(name);
    if (set === undefined && name === "component.new") return Promise.resolve(this.newPlant(args));
    return Promise.resolve(set ?? failed(failure("unknown_action", 404)));
  }

  private basePlant(args: Readonly<Record<string, unknown>>): Outcome<unknown> {
    if (typeof args.from !== "string") {
      const shown = this.actions.get("component.example") ?? failed(failure("component_failed", 422, "no example"));
      return shown.ok ? success((shown.value as { example?: unknown }).example) : shown;
    }
    const file = this.files.get(args.from) ?? failed(failure("not_found", 404, `no ${args.from}`));
    if (!file.ok) return file;
    try {
      return success(JSON.parse(file.value.content) as unknown);
    } catch {
      return success(null);
    }
  }

  private newPlant(args: Readonly<Record<string, unknown>>): Outcome<unknown> {
    const base = this.basePlant(args);
    if (!base.ok) return base;
    if (typeof base.value !== "object" || base.value === null || Array.isArray(base.value))
      return failed(failure("composition_invalid", 422, "the base is not a composition"));
    const plant: Record<string, unknown> = { ...base.value, id: args.id, title: args.title };
    if (args.empty !== true) return success({ composition: plant });
    const frame = Object.fromEntries(
      Object.entries(plant).filter(([key]) => !["description", "equipment", "connections", "areas"].includes(key)),
    );
    return success({ composition: { ...frame, areas: [{ name: siteArea(plant.site) }] } });
  }

  followers = new Map<string, RunFollower>();

  followRun(worldId: string, follower: RunFollower): () => void {
    this.followers.set(worldId, follower);
    return () => this.followers.delete(worldId);
  }

  stream(worldId: string, records: readonly unknown[], end: "closed" | "failed" | null = "closed"): void {
    const follower = this.followers.get(worldId);
    records.forEach((record) => follower?.record(JSON.stringify(record)));
    if (end === "closed") follower?.closed();
    if (end === "failed") follower?.failed();
  }

  watchWorkspace(onChange: () => void): () => void {
    this.watchers.add(onChange);
    return () => this.watchers.delete(onChange);
  }

  changeOnDisk(): void {
    this.watchers.forEach((watcher) => {
      watcher();
    });
  }
}

function siteArea(site: unknown): string {
  if (typeof site !== "object" || site === null || !("name" in site)) return "Site";
  return typeof site.name === "string" ? site.name : "Site";
}
