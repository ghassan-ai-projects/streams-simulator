import { bridgeAdapter } from "../adapters/bridge";
import type { FileContent } from "../app/ports";
import { failure as failed, success } from "../app/ports";
import type { Workspace } from "../model/workspace";
import type { Deferred } from "./deferred";
import { deferred } from "./deferred";
import { cannedFailure } from "./fakeFailures";
import type { Canned } from "./fakeFailures";
import { FakeSource, answer, workspaceWire } from "./fakeWire";
import { newPlant } from "./plantBuilder";
import { sampleWorkspace } from "./workspaceFixture";

export { cannedFailure } from "./fakeFailures";

export function cannedFile(path: string, content: unknown, etag = "e"): Canned<FileContent> {
  return success({ path, content: typeof content === "string" ? content : JSON.stringify(content), etag });
}

type Args = Readonly<Record<string, unknown>>;
type Reply<T = unknown> = Canned<T> | Promise<Canned<T>>;
type Handler = (args: Args) => Reply;
type Transport = Parameters<typeof bridgeAdapter>[0];
interface Body {
  readonly args?: Args;
  readonly code?: string;
  readonly path?: string;
  readonly content?: string;
  readonly etag?: string;
}
interface Written {
  readonly path: string;
  readonly etag: string;
}
interface Hold {
  remaining: number;
  readonly gates: Deferred<Canned<never>>[];
}

export class FakeBridge {
  paired = false;
  sessionOutcome: Canned<boolean> | null = null;
  pairOutcomes: Canned<boolean>[] = [];
  workspaceOutcome: Canned<Workspace> = success(sampleWorkspace);
  files = new Map<string, Canned<FileContent>>();
  actions = new Map<string, Canned<unknown>>();
  handlers = new Map<string, Handler>();
  calls: { name: string; args: Args }[] = [];
  reads: string[] = [];
  writes: { path: string; content: string; etag: string }[] = [];
  writeOutcomes = new Map<string, Canned<Written>>();
  watchers = new Set<FakeSource>();
  followers = new Map<string, FakeSource>();
  follows: FakeSource[] = [];
  pairedCodes: string[] = [];
  private readonly holds = new Map<string, Hold>();
  private readonly transport: Transport = {
    fetch: (input, init) => this.serve(input, init),
    events: (url) => this.openSource(url) as unknown as EventSource,
  };

  readonly gateways = bridgeAdapter(this.transport);
  readonly session = this.gateways.session;
  readonly workspace = this.gateways.workspace;
  readonly plant = this.gateways.plant;
  readonly editor = this.gateways.editor;
  readonly launch = this.gateways.launch;
  readonly review = this.gateways.review;

  addFile(path: string, content: unknown, etag = "e"): void {
    this.files.set(path, cannedFile(path, content, etag));
  }

  holdRead(path: string, count = Infinity): Deferred<Canned<FileContent>>[] {
    return this.hold(`file:${path}`, count);
  }

  holdWrite(count = Infinity): Deferred<Canned<Written>>[] {
    return this.hold("write", count);
  }

  holdWorkspace(count = Infinity): Deferred<Canned<Workspace>>[] {
    return this.hold("workspace", count);
  }

  stream(runId: string, entries: readonly object[], end: "closed" | "failed" | null = "closed"): void {
    this.send(this.followers.get(runId), entries, end);
  }

  streamFollow(index: number, entries: readonly object[], end: "closed" | "failed" | null = "closed"): void {
    this.send(this.follows[index], entries, end);
  }

  changeOnDisk(): void {
    this.watchers.forEach((source) => source.dispatchEvent(new Event("workspace")));
  }

  private hold<T>(key: string, count: number): Deferred<Canned<T>>[] {
    const gates: Deferred<Canned<T>>[] = [];
    this.holds.set(key, { remaining: count, gates: gates as Deferred<Canned<never>>[] });
    return gates;
  }

  private gated<T>(key: string, read: () => Canned<T>): Reply<T> {
    const hold = this.holds.get(key);
    if (hold === undefined || hold.remaining === 0) return read();
    hold.remaining -= 1;
    const gate = deferred<Canned<T>>();
    hold.gates.push(gate as Deferred<Canned<never>>);
    return gate.promise;
  }

  private send(source: FakeSource | undefined, entries: readonly object[], end: "closed" | "failed" | null): void {
    entries.forEach((entry) => source?.record(entry));
    if (end !== null) source?.end(end);
  }

  private openSource(url: string): FakeSource {
    const runId = /^\/api\/runs\/([^/]+)\/events$/.exec(url)?.[1];
    const source = new FakeSource((closed) => {
      this.watchers.delete(closed);
      if (runId !== undefined) this.followers.delete(decodeURIComponent(runId));
    });
    if (runId === undefined) this.watchers.add(source);
    else this.followers.set(decodeURIComponent(runId), source);
    if (runId !== undefined) this.follows.push(source);
    return source;
  }

  private async serve(input: string, init?: RequestInit): Promise<Response> {
    const url = new URL(input, "http://bridge.test");
    const body = (typeof init?.body === "string" ? JSON.parse(init.body) : {}) as Body;
    return answer(await this.route(url, init?.method ?? "GET", body));
  }

  private route(url: URL, method: string, body: Body): Reply {
    if (url.pathname.startsWith("/api/actions/"))
      return this.action(decodeURIComponent(url.pathname.slice(13)), body.args ?? {});
    const routes: Record<string, () => Reply> = {
      "/api/workspace": () => this.gated("workspace", () => this.listing()),
      "/api/session": () => (method === "POST" ? this.pair(body) : this.check()),
      "/api/file": () => (method === "PUT" ? this.write(body) : this.read(url)),
    };
    return routes[url.pathname]?.() ?? failed(cannedFailure("not_found", 404, `no route ${url.pathname}`));
  }

  private listing(): Canned<unknown> {
    return this.workspaceOutcome.ok ? success(workspaceWire(this.workspaceOutcome.value)) : this.workspaceOutcome;
  }

  private check(): Canned<unknown> {
    const outcome = this.sessionOutcome ?? success(this.paired);
    return outcome.ok ? success({ paired: outcome.value }) : outcome;
  }

  private pair(body: Body): Canned<unknown> {
    this.pairedCodes.push(body.code ?? "");
    const next = this.pairOutcomes.shift() ?? success(true);
    this.paired = next.ok;
    return next.ok ? success({ paired: next.value }) : next;
  }

  private read(url: URL): Reply {
    const path = url.searchParams.get("path") ?? "";
    this.reads.push(path);
    return this.gated(
      `file:${path}`,
      () => this.files.get(path) ?? failed(cannedFailure("not_found", 404, `no ${path}`)),
    );
  }

  private write({ path = "", content = "", etag = "" }: Body): Reply {
    this.writes.push({ path, content, etag });
    return this.gated("write", () => this.stored(path, content));
  }

  private stored(path: string, content: string): Canned<Written> {
    const outcome = this.writeOutcomes.get(path) ?? success({ path, etag: `e${String(this.writes.length)}` });
    if (outcome.ok) this.files.set(path, cannedFile(path, content, outcome.value.etag));
    return outcome;
  }

  private action(name: string, args: Args): Promise<Canned<unknown>> {
    this.calls.push({ name, args });
    const reply = this.handlers.get(name)?.(args) ?? this.canned(name, args);
    return Promise.resolve(reply).then((outcome) => (outcome.ok ? success({ result: outcome.value }) : outcome));
  }

  private canned(name: string, args: Args): Canned<unknown> {
    const canned = this.actions.get(name);
    if (canned !== undefined) return canned;
    if (name === "component.new") return newPlant(this.actions, this.files, args);
    return failed(cannedFailure("unknown_action", 404));
  }
}
