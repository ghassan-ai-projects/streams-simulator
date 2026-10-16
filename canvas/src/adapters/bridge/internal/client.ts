import type { BridgeAdapters, DomainDocument, Outcome } from "../../../app/ports/index";
import type { RunFollower } from "../../../app/ports/review";
import type { PlantSource } from "../../../app/ports/shared";
import type { ReviewFailure } from "../../../app/ports/review";
import type { Replay } from "../../../model/report";
import { action, getFile, putFile, request, type Transport, type WireProblem } from "./http";
import {
  acknowledgement,
  advanceResult,
  composeSummary,
  component,
  composition,
  destroyed,
  examples,
  faultInjected,
  fileContent,
  journalEntry,
  library,
  paired,
  presets,
  replay,
  report,
  runEnded,
  runDomain,
  written,
  workspace,
} from "./decoders";
import {
  editorFailure,
  launchFailure,
  plantFailure,
  reviewFailure,
  sessionFailure,
  workspaceFailure,
} from "./failures";

const streamClosed = 2;

export function bridgeAdapter(transport: Transport): BridgeAdapters {
  return {
    session: {
      check: () => request(transport, "/api/session", paired, sessionFailure),
      pair: (code) => request(transport, "/api/session", paired, sessionFailure, post({ code })),
    },
    workspace: {
      load: () => request(transport, "/api/workspace", workspace, workspaceFailure),
      watch: (onChange) => watchWorkspace(transport, onChange),
    },
    plant: plantGateway(transport),
    editor: editorGateway(transport),
    launch: launchGateway(transport),
    review: reviewGateway(transport),
  };
}

function plantGateway(transport: Transport): BridgeAdapters["plant"] {
  return {
    readFile: (path) => getFile(transport, path, fileContent, plantFailure("read")),
    component: (ref) => action(transport, "component.show", { ref }, component, plantFailure("read")),
    examples: () => action(transport, "component.examples", {}, examples, plantFailure("read")),
    generateComposition: (input) =>
      action(
        transport,
        "component.new",
        generationArgs(input.source, input.id, input.title, input.mode),
        composition,
        plantFailure("generate"),
      ),
    writeFile: (path, content, etag) => putFile(transport, path, content, etag, written, plantFailure("write")),
    compose: ({ composition: from, output, etag }) =>
      action(
        transport,
        "domain.compose",
        { composition: from, out: output, etag },
        composeSummary,
        plantFailure("compose"),
      ),
  };
}

function editorGateway(transport: Transport): BridgeAdapters["editor"] {
  return {
    readFile: (path) => getFile(transport, path, fileContent, editorFailure("read")),
    library: () => action(transport, "component.list", {}, library, editorFailure("read")),
    component: (ref) => action(transport, "component.show", { ref }, component, editorFailure("read")),
    presets: (ref) => action(transport, "component.show", { ref }, presets, editorFailure("read")),
    writeFile: (path, content, etag) => putFile(transport, path, content, etag, written, editorFailure("write")),
    compose: ({ composition: from, output, etag }) =>
      action(
        transport,
        "domain.compose",
        { composition: from, out: output, etag },
        composeSummary,
        editorFailure("compose"),
      ),
  };
}

function launchGateway(transport: Transport): BridgeAdapters["launch"] {
  return {
    createWorld: ({ domain, seed, explore }) =>
      action(transport, "world.create", { domain, seed, explore }, worldCreated, launchFailure("create")),
    advance: (worldId, byNs) =>
      action(
        transport,
        "clock.advance",
        { world_id: worldId, by_ns: byNs },
        advanceResult,
        launchFailure("advance"),
      ).then(toVoid),
    inject: (worldId, entity, fault) =>
      action(
        transport,
        "fault.inject",
        { world_id: worldId, entity_id: entity, fault },
        faultInjected,
        launchFailure("inject"),
      ).then(toVoid),
    invoke: (worldId, entity, effector, args) =>
      action(
        transport,
        "effector.invoke",
        { world_id: worldId, entity_id: entity, effector, args },
        acknowledgement,
        launchFailure("invoke"),
      ),
    endRun: (worldId) =>
      action(transport, "run.end", { world_id: worldId }, runEnded, launchFailure("end")).then(toVoid),
    destroyWorld: (worldId) => destroy(transport, worldId),
  };
}

function reviewGateway(transport: Transport): BridgeAdapters["review"] {
  return {
    report: (runId) => action(transport, "run.report", { run: runId }, report, reviewFailure),
    verify: (runId) => verify(transport, runId),
    runDomain: (runId) => getRunDomain(transport, runId),
    follow: (runId, follower) => followRun(transport, runId, follower),
  };
}

function generationArgs(
  source: PlantSource,
  id: string,
  title: string,
  mode: "copy" | "empty",
): Readonly<Record<string, unknown>> {
  const sourceArgs = source.kind === "example" ? { example: source.id } : { from: source.path };
  return { id, title, ...sourceArgs, ...(mode === "empty" ? { empty: true } : {}) };
}

function worldCreated(value: unknown): { readonly worldId: string } | null {
  if (typeof value !== "object" || value === null || !("world_id" in value)) return null;
  const worldId = value.world_id;
  return typeof worldId === "string" && worldId !== "" ? { worldId } : null;
}

function toVoid<F>(outcome: Outcome<true, F>): Outcome<void, F> {
  return outcome.ok ? { ok: true, value: undefined } : outcome;
}

function verify(transport: Transport, runId: string): Promise<Outcome<Replay, ReviewFailure>> {
  return action(transport, "run.verify", { run: runId }, replay, reviewFailure, (problem: WireProblem) =>
    replay(problem.result),
  );
}

function getRunDomain(transport: Transport, runId: string): Promise<Outcome<DomainDocument, ReviewFailure>> {
  return getFile(transport, `runs/${runId}/run.json`, runDomain, reviewFailure).then((outcome) =>
    outcome.ok ? { ok: true, value: outcome.value as DomainDocument } : outcome,
  );
}

function followRun(transport: Transport, runId: string, follower: RunFollower): () => void {
  const source = transport.events(`/api/runs/${encodeURIComponent(runId)}/events`);
  source.addEventListener("record", (event: Event) => {
    const data = event instanceof MessageEvent && typeof event.data === "string" ? event.data : "";
    const entry = journalEntry(data);
    if (entry === null) {
      source.close();
      follower.failed();
      return;
    }
    follower.entry(entry);
  });
  source.addEventListener("closed", () => {
    source.close();
    follower.closed();
  });
  source.addEventListener("error", () => {
    if (source.readyState === streamClosed) follower.failed();
  });
  return () => {
    source.close();
  };
}

function watchWorkspace(transport: Transport, onChange: () => void): () => void {
  const source = transport.events("/api/events");
  source.addEventListener("workspace", onChange);
  return () => {
    source.close();
  };
}

async function destroy(transport: Transport, worldId: string): Promise<void> {
  await action(transport, "world.destroy", { world_id: worldId }, destroyed, launchFailure("end"));
}

function post(payload: unknown): RequestInit {
  return { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload) };
}
