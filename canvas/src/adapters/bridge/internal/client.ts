import type { BridgeAdapters, Outcome, PlantSource, ReviewFailure } from "../../../app/ports";
import type { Replay } from "../../../model/review";
import { connectGateway } from "./connect";
import { jobGateway } from "./jobs";
import { controlSession } from "./control";
import { devicesGateway } from "./devices";
import { followRun } from "./follow";
import { requestTwinReport } from "./twinReportGateway";
import type { FailureMapper, Transport, WireProblem } from "./http";
import { action, getFile, post, putFile, request } from "./http";
import {
  composeSummary,
  component,
  composition,
  examples,
  fileContent,
  library,
  paired,
  presets,
  replay,
  report,
  runDomain,
  written,
  workspace,
} from "./decoders";
import { editorFailure, plantFailure, reviewFailure, sessionFailure, workspaceFailure } from "./failures";

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
    devices: devicesGateway(transport),
    launch: launchGateway(transport),
    review: reviewGateway(transport),
  };
}

type ComposeInput = Parameters<BridgeAdapters["plant"]["compose"]>[0];

function fileActions<F>(transport: Transport, failure: (step: "read" | "write" | "compose") => FailureMapper<F>) {
  return {
    readFile: (path: string) => getFile(transport, path, fileContent, failure("read")),
    component: (ref: string) => action(transport, "component.show", { ref }, component, failure("read")),
    writeFile: (path: string, content: string, etag: string) =>
      putFile(transport, path, content, etag, written, failure("write")),
    compose: ({ composition: from, output, etag }: ComposeInput) =>
      action(transport, "domain.compose", { composition: from, out: output, etag }, composeSummary, failure("compose")),
  };
}

function plantGateway(transport: Transport): BridgeAdapters["plant"] {
  return {
    ...fileActions(transport, plantFailure),
    examples: () => action(transport, "component.examples", {}, examples, plantFailure("read")),
    generateComposition: (input) =>
      action(
        transport,
        "component.new",
        generationArgs(input.source, input.id, input.title, input.mode),
        composition,
        plantFailure("generate"),
      ),
  };
}

function editorGateway(transport: Transport): BridgeAdapters["editor"] {
  return {
    ...fileActions(transport, editorFailure),
    library: () => action(transport, "component.list", {}, library, editorFailure("read")),
    presets: (ref) => action(transport, "component.show", { ref }, presets, editorFailure("read")),
  };
}

function launchGateway(transport: Transport): BridgeAdapters["launch"] {
  return { ...jobGateway(transport), ...connectGateway(transport) };
}

function reviewGateway(transport: Transport): BridgeAdapters["review"] {
  return {
    report: (runId) => action(transport, "run.report", { run: runId }, report, reviewFailure),
    verify: (runId) => verify(transport, runId),
    runDomain: (runId) => getFile(transport, `runs/${runId}/run.json`, runDomain, reviewFailure),
    follow: (runId, follower) => followRun(transport, runId, follower),
    control: (request) => controlSession(transport, request),
    twinReport: (request) => requestTwinReport(transport, request),
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

function verify(transport: Transport, runId: string): Promise<Outcome<Replay, ReviewFailure>> {
  return action(transport, "run.verify", { run: runId }, replay, reviewFailure, (problem: WireProblem) =>
    replay(problem.result),
  );
}

function watchWorkspace(transport: Transport, onChange: () => void): () => void {
  const source = transport.events("/api/events");
  source.addEventListener("workspace", onChange);
  return () => {
    source.close();
  };
}
