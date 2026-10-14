import type { Bridge, RunFollower } from "../app/ports";
import type { BridgeFailure, FileContent, Outcome, Workspace } from "../model/bridge";
import { failed, success } from "../model/bridge";

export const sampleWorkspace: Workspace = {
  name: "lab-workspace",
  domains: [{ path: "domains/factory-utilities.domain.json", id: "factory-utilities", title: "Factory utilities" }],
  compositions: [
    { path: "compositions/factory-utilities.composition.json", id: "factory-utilities", title: "Factory utilities" },
  ],
  scenarios: [],
  runs: [
    { worldId: "w-1", ended: true, explore: false },
    { worldId: "w-2", ended: false, explore: true },
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
  calls: { name: string; args: Readonly<Record<string, unknown>> }[] = [];
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
    return Promise.resolve(this.files.get(path) ?? failed(failure("not_found", 404, `no ${path}`)));
  }

  action(name: string, args: Readonly<Record<string, unknown>>): Promise<Outcome<unknown>> {
    this.calls.push({ name, args });
    return Promise.resolve(this.actions.get(name) ?? failed(failure("unknown_action", 404)));
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
