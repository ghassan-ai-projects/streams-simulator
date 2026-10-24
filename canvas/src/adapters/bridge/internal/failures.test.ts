import { describe, expect, it } from "vitest";
import { cannedFailure } from "../../../test/fakeFailures";
import {
  editorFailure,
  launchFailure,
  plantFailure,
  reviewFailure,
  sessionFailure,
  workspaceFailure,
} from "./failures";
import type { WireProblem } from "./http";

const http = (code: string, status = 400, issues: readonly { pointer: string; message: string }[] = []) => ({
  kind: "http" as const,
  problem: { ...cannedFailure(code, status), issues, result: null },
});

describe("protocol failures mapped to the owning feature", () => {
  it("explains each pairing refusal in words and blocks only the final ones", () => {
    expect(sessionFailure(http("wrong_code", 401))).toMatchObject({
      kind: "code-rejected",
      diagnostic: { displayCode: "wrong_code" },
    });
    expect(sessionFailure(http("pairing_locked", 401))).toMatchObject({
      kind: "pairing-closed",
      reason: "locked",
      diagnostic: { displayCode: "pairing_locked" },
    });
    expect(sessionFailure(http("pairing_used", 401))).toMatchObject({
      kind: "pairing-closed",
      reason: "already-used",
      diagnostic: { displayCode: "pairing_used" },
    });
    expect(sessionFailure(http("other", 401))).toMatchObject({ kind: "unreachable" });
    expect(sessionFailure({ kind: "transport" })).toMatchObject({
      kind: "unreachable",
      diagnostic: { displayCode: "unreachable" },
    });
    expect(sessionFailure({ kind: "unreadable" })).toMatchObject({ kind: "unreachable" });
  });

  it("keeps a workspace refusal the feature can recover from", () => {
    expect(workspaceFailure(http("unpaired", 401))).toMatchObject({ kind: "unpaired" });
    expect(workspaceFailure(http("bridge_failed", 500))).toMatchObject({ kind: "unavailable" });
    expect(workspaceFailure(http("bridge_failed", 400))).toMatchObject({ kind: "unreadable" });
    expect(workspaceFailure({ kind: "transport" })).toMatchObject({ kind: "unavailable" });
    expect(workspaceFailure({ kind: "unreadable" })).toMatchObject({ kind: "unreadable" });
  });

  it("names a plant file the reader must stop for, or one it can retry", () => {
    expect(plantFailure("read")(http("not_found", 404))).toMatchObject({ kind: "missing" });
    expect(plantFailure("read")(http("file_not_found", 404))).toMatchObject({ kind: "missing" });
    expect(plantFailure("write")(http("conflict", 409))).toMatchObject({ kind: "conflict" });
    expect(
      plantFailure("generate")(http("composition_invalid", 422, [{ pointer: "/id", message: "bad id" }])),
    ).toMatchObject({ kind: "invalid-composition", issues: [{ pointer: "/id", message: "bad id" }] });
    expect(
      plantFailure("compose")(http("composition_invalid", 422, [{ pointer: "/id", message: "bad id" }])),
    ).toMatchObject({ kind: "compose-refused", issues: [{ pointer: "/id", message: "bad id" }] });
    expect(plantFailure("read")(http("bridge_failed", 503))).toMatchObject({ kind: "unavailable" });
    expect(plantFailure("read")(http("bridge_failed", 400))).toMatchObject({ kind: "unreadable" });
    expect(plantFailure("read")({ kind: "transport" })).toMatchObject({ kind: "unavailable" });
  });

  it("stops an editor save for the same reasons the plant reader does", () => {
    expect(editorFailure("read")(http("not_found", 404))).toMatchObject({ kind: "missing" });
    expect(editorFailure("write")(http("conflict", 409))).toMatchObject({ kind: "conflict" });
    expect(
      editorFailure("compose")(http("composition_invalid", 422, [{ pointer: "/x", message: "bad" }])),
    ).toMatchObject({ kind: "compose-refused" });
    expect(editorFailure("compose")(http("composition_invalid", 422))).toMatchObject({ kind: "unreadable" });
    expect(editorFailure("read")(http("bridge_failed", 500))).toMatchObject({ kind: "unavailable" });
    expect(editorFailure("read")(http("bridge_failed", 400))).toMatchObject({ kind: "unreadable" });
    expect(editorFailure("read")({ kind: "transport" })).toMatchObject({ kind: "unavailable" });
  });

  it("says which launch step the binary refused", () => {
    expect(launchFailure("start")(http("domain_invalid", 422))).toMatchObject({
      kind: "rejected",
      step: "start",
      diagnostic: { displayCode: "domain_invalid", message: "refused" },
    });
    expect(launchFailure("stop")(http("job_active", 422))).toMatchObject({ kind: "rejected", step: "stop" });
    expect(launchFailure("status")(http("bridge_failed", 503))).toMatchObject({ kind: "unavailable" });
    expect(launchFailure("status")({ kind: "transport" })).toMatchObject({ kind: "unavailable" });
    expect(launchFailure("status")({ kind: "unreadable" })).toMatchObject({ kind: "unreadable" });
  });

  it("tells a review which part of a run is missing, and which answer it could not read", () => {
    expect(reviewFailure(http("not_found", 404))).toMatchObject({ kind: "missing" });
    expect(reviewFailure(http("file_not_found", 404))).toMatchObject({ kind: "missing" });
    expect(reviewFailure(http("run_incomplete", 422))).toMatchObject({
      kind: "run-incomplete",
      diagnostic: { displayCode: "run_incomplete" },
    });
    expect(reviewFailure(http("sealed", 403))).toMatchObject({ kind: "run-incomplete" });
    expect(reviewFailure(http("bridge_failed", 500))).toMatchObject({ kind: "unavailable" });
    expect(reviewFailure(http("bridge_failed", 422))).toMatchObject({ kind: "unreadable" });
    expect(reviewFailure({ kind: "transport" })).toMatchObject({ kind: "unavailable" });
    expect(reviewFailure({ kind: "unreadable" })).toMatchObject({ kind: "unreadable" });
  });

  it("keeps an unknown protocol code as display data, not as feature behaviour", () => {
    const unknown: WireProblem = {
      status: 418,
      code: "something_new",
      message: "never seen",
      issues: [],
      result: null,
    };
    expect(plantFailure("read")({ kind: "http", problem: unknown })).toMatchObject({
      kind: "unreadable",
      diagnostic: { displayCode: "something_new", message: "never seen" },
    });
    expect(reviewFailure({ kind: "http", problem: unknown })).toMatchObject({
      kind: "unreadable",
      diagnostic: { displayCode: "something_new", message: "never seen" },
    });
  });
});
