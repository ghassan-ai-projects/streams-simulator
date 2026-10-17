import type {
  BridgeAdapters,
  ComposeSummary,
  EditorFailure,
  EditorGateway,
  FileContent,
  LaunchFailure,
  LaunchGateway,
  Outcome,
  PlantFailure,
  PlantGateway,
  PlantSource,
  ReviewGateway,
  RunFollower,
  SessionFailure,
  SessionGateway,
  WorkspaceGateway,
} from "../app/ports";
import { failure as failed, success } from "../app/ports";
import type { Composition } from "../model/editor";
import type { JournalEntry } from "../model/review";
import type { Workspace } from "../model/workspace";
import {
  editorFailure,
  cannedFailure,
  launchFailure,
  plantFailure,
  sessionFailure,
  workspaceFailure,
} from "./fakeFailures";
import type { Canned, CannedFailure, EditorAction, LaunchStep, PlantAction } from "./fakeFailures";
import { collect, settle } from "./fakeFailures";
import {
  acknowledgementOf,
  compositionOf,
  examplesOf,
  libraryOf,
  presetsOf,
  specOf,
  summaryOf,
  worldOf,
} from "./fakeValues";
import { newPlant } from "./plantBuilder";
import { reviewDomain, reviewReport, reviewVerify } from "./fakeReview";
import { sampleWorkspace } from "./workspaceFixture";

export { cannedFailure } from "./fakeFailures";

type Handler = (args: Readonly<Record<string, unknown>>) => Canned<unknown> | Promise<Canned<unknown>>;
interface Written {
  readonly path: string;
  readonly etag: string;
}
interface Compose {
  readonly composition: string;
  readonly output: string;
  readonly etag: string;
}

export class FakeBridge {
  paired = false;
  sessionOutcome: Canned<boolean> | null = null;
  pairOutcomes: Canned<boolean>[] = [];
  workspaceOutcome: Canned<Workspace> = success(sampleWorkspace);
  files = new Map<string, Canned<FileContent>>();
  actions = new Map<string, Canned<unknown>>();
  handlers = new Map<string, Handler>();
  calls: { name: string; args: Readonly<Record<string, unknown>> }[] = [];
  reads: string[] = [];
  writes: { path: string; content: string; etag: string }[] = [];
  writeOutcomes = new Map<string, Canned<Written>>();
  watchers = new Set<() => void>();
  followers = new Map<string, RunFollower>();
  follows: RunFollower[] = [];
  pairedCodes: string[] = [];

  readonly session: SessionGateway = {
    check: () => Promise.resolve(settle(this.sessionOutcome ?? success(this.paired), sessionFailure)),
    pair: (code) => Promise.resolve(this.pairWith(code)),
  };

  readonly workspace: WorkspaceGateway = {
    load: () => Promise.resolve(settle(this.workspaceOutcome, workspaceFailure)),
    watch: (onChange) => this.watchWorkspace(onChange),
  };

  readonly plant: PlantGateway = {
    readFile: (path) => this.readFile(path).then((file) => settle(file, (problem) => plantFailure("read", problem))),
    component: (ref) => this.plantRead("component.show", { ref }, specOf),
    examples: () => this.plantRead("component.examples", {}, examplesOf),
    generateComposition: (input) => this.generate(input),
    writeFile: (path, content, etag) =>
      this.writeFile(path, content, etag).then((written) =>
        settle(written, (problem) => plantFailure("write", problem)),
      ),
    compose: (input) => this.composed("domain.compose", input, (problem) => plantFailure("compose", problem)),
  };

  readonly editor: EditorGateway = {
    readFile: (path) => this.readFile(path).then((file) => settle(file, (problem) => editorFailure("read", problem))),
    library: () => this.editorRead("component.list", {}, libraryOf),
    component: (ref) => this.editorRead("component.show", { ref }, specOf),
    presets: (ref) => this.editorRead("component.show", { ref }, presetsOf),
    writeFile: (path, content, etag) =>
      this.writeFile(path, content, etag).then((written) =>
        settle(written, (problem) => editorFailure("write", problem)),
      ),
    compose: (input) => this.composed("domain.compose", input, (problem) => editorFailure("compose", problem)),
  };

  readonly launch: LaunchGateway = {
    createWorld: (input) =>
      this.launchStep(
        "create",
        "world.create",
        { domain: input.domain, seed: input.seed, explore: input.explore },
        worldOf,
      ),
    advance: (worldId, byNs) => this.voidStep("advance", "clock.advance", { world_id: worldId, by_ns: byNs }),
    inject: (worldId, entity, fault) =>
      this.voidStep("inject", "fault.inject", { world_id: worldId, entity_id: entity, fault }),
    invoke: (worldId, entity, effector, args) =>
      this.launchStep(
        "invoke",
        "effector.invoke",
        { world_id: worldId, entity_id: entity, effector, args },
        acknowledgementOf,
      ),
    endRun: (worldId) => this.voidStep("end", "run.end", { world_id: worldId }),
    destroyWorld: (worldId) => this.destroy(worldId),
  };

  readonly review: ReviewGateway = {
    report: (runId) => reviewReport(this, runId),
    verify: (runId) => reviewVerify(this, runId),
    runDomain: (runId) => reviewDomain(this, runId),
    follow: (runId, follower) => this.followRun(runId, follower),
  };

  adapters(): BridgeAdapters {
    return {
      session: this.session,
      workspace: this.workspace,
      plant: this.plant,
      editor: this.editor,
      launch: this.launch,
      review: this.review,
    };
  }

  addFile(path: string, content: unknown, etag = "e"): void {
    const text = typeof content === "string" ? content : JSON.stringify(content);
    this.files.set(path, success({ path, content: text, etag }));
  }

  readFile(path: string): Promise<Canned<FileContent>> {
    this.reads.push(path);
    return Promise.resolve(this.files.get(path) ?? failed(cannedFailure("not_found", 404, `no ${path}`)));
  }

  writeFile(path: string, content: string, etag: string): Promise<Canned<Written>> {
    this.writes.push({ path, content, etag });
    const outcome = this.writeOutcomes.get(path) ?? success({ path, etag: `e${String(this.writes.length)}` });
    if (outcome.ok) this.files.set(path, success({ path, content, etag: outcome.value.etag }));
    return Promise.resolve(outcome);
  }

  action(name: string, args: Readonly<Record<string, unknown>>): Promise<Canned<unknown>> {
    this.calls.push({ name, args });
    const handler = this.handlers.get(name);
    if (handler !== undefined) return Promise.resolve(handler(args));
    const canned = this.actions.get(name);
    if (canned === undefined && name === "component.new")
      return Promise.resolve(newPlant(this.actions, this.files, args));
    return Promise.resolve(canned ?? failed(cannedFailure("unknown_action", 404)));
  }

  followRun(runId: string, follower: RunFollower): () => void {
    this.followers.set(runId, follower);
    this.follows.push(follower);
    return () => this.followers.delete(runId);
  }

  watchWorkspace(onChange: () => void): () => void {
    this.watchers.add(onChange);
    return () => this.watchers.delete(onChange);
  }

  stream(runId: string, entries: readonly JournalEntry[], end: "closed" | "failed" | null = "closed"): void {
    this.send(this.followers.get(runId), entries, end);
  }

  streamFollow(index: number, entries: readonly JournalEntry[], end: "closed" | "failed" | null = "closed"): void {
    this.send(this.follows[index], entries, end);
  }

  changeOnDisk(): void {
    this.watchers.forEach((watcher) => {
      watcher();
    });
  }

  private send(
    follower: RunFollower | undefined,
    entries: readonly JournalEntry[],
    end: "closed" | "failed" | null,
  ): void {
    entries.forEach((entry) => follower?.entry(entry));
    if (end === "closed") follower?.closed();
    if (end === "failed") follower?.failed();
  }

  private pairWith(code: string): Outcome<boolean, SessionFailure> {
    this.pairedCodes.push(code);
    const next = this.pairOutcomes.shift() ?? success(true);
    this.paired = next.ok;
    return settle(next, sessionFailure);
  }

  private plantRead<T>(
    name: string,
    args: Readonly<Record<string, unknown>>,
    read: (value: unknown) => T | null,
    action: PlantAction = "read",
  ): Promise<Outcome<T, PlantFailure>> {
    return this.action(name, args).then((outcome) =>
      collect(outcome, (problem) => plantFailure(action, problem), read),
    );
  }

  private editorRead<T>(
    name: string,
    args: Readonly<Record<string, unknown>>,
    read: (value: unknown) => T | null,
    action: EditorAction = "read",
  ): Promise<Outcome<T, EditorFailure>> {
    return this.action(name, args).then((outcome) =>
      collect(outcome, (problem) => editorFailure(action, problem), read),
    );
  }

  private generate(input: {
    readonly source: PlantSource;
    readonly id: string;
    readonly title: string;
    readonly mode: "copy" | "empty";
  }): Promise<Outcome<Composition, PlantFailure>> {
    const from = input.source.kind === "example" ? { example: input.source.id } : { from: input.source.path };
    const args = { id: input.id, title: input.title, ...from };
    const withMode = input.mode === "empty" ? { ...args, empty: true } : args;
    return this.action("component.new", withMode).then((outcome) =>
      collect(outcome, (problem) => plantFailure("generate", problem), compositionOf),
    );
  }

  private composed<F>(
    name: string,
    input: Compose,
    map: (problem: CannedFailure) => F,
  ): Promise<Outcome<ComposeSummary, F>> {
    return this.action(name, { composition: input.composition, out: input.output, etag: input.etag }).then((outcome) =>
      collect(outcome, map, summaryOf),
    );
  }

  private launchStep<T>(
    step: LaunchStep,
    name: string,
    args: Readonly<Record<string, unknown>>,
    read: (value: unknown) => T | null,
  ): Promise<Outcome<T, LaunchFailure>> {
    return this.action(name, args).then((outcome) => collect(outcome, (problem) => launchFailure(step, problem), read));
  }

  private voidStep(
    step: LaunchStep,
    name: string,
    args: Readonly<Record<string, unknown>>,
  ): Promise<Outcome<void, LaunchFailure>> {
    return this.action(name, args).then((outcome) =>
      outcome.ok ? success(undefined) : failed(launchFailure(step, outcome.failure)),
    );
  }

  private async destroy(worldId: string): Promise<void> {
    await this.voidStep("end", "world.destroy", { world_id: worldId });
  }
}
