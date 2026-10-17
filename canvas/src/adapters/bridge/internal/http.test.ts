import { describe, expect, it } from "vitest";
import { bridgeTransport, respond } from "../../../test/bridgeTransport";
import { action, getFile, isRecord, putFile, request } from "./http";
import type { RequestFailure, WireProblem } from "./http";

function problem(failure: RequestFailure | undefined): WireProblem {
  if (failure?.kind !== "http") throw new Error("expected an http problem");
  return failure.problem;
}

function record(seen: RequestFailure[]): (failure: RequestFailure) => RequestFailure {
  return (failure) => {
    seen.push(failure);
    return failure;
  };
}

const decodeObject = (value: unknown) => (isRecord(value) ? value : null);

describe("bridge requests", () => {
  it("reads the error envelope, or names the status when there is none", async () => {
    const seen: RequestFailure[] = [];
    const envelope = bridgeTransport(() =>
      respond(422, {
        error: {
          code: "composition_invalid",
          message: "bad",
          issues: [{ pointer: "/connections/29", message: "types differ" }],
        },
      }),
    );
    await request(envelope, "/api/session", () => true, record(seen));
    expect(problem(seen[0])).toEqual({
      status: 422,
      code: "composition_invalid",
      message: "bad",
      issues: [{ pointer: "/connections/29", message: "types differ" }],
      result: null,
    });

    const verdict = bridgeTransport(() =>
      respond(422, { error: { code: "replay_mismatch", message: "no", result: { matches: false } } }),
    );
    await request(verdict, "/api/session", () => true, record(seen));
    expect(problem(seen[1])).toMatchObject({ code: "replay_mismatch", result: { matches: false } });

    const opaque = bridgeTransport(() => respond(502, "oops"));
    await request(opaque, "/api/session", () => true, record(seen));
    expect(problem(seen[2])).toEqual({
      status: 502,
      code: "bridge_failed",
      message: "The bridge answered 502.",
      issues: [],
      result: null,
    });
  });

  it("separates a transport that never answered from a body it cannot decode", async () => {
    const seen: RequestFailure[] = [];
    await request(
      bridgeTransport(() => Promise.reject(new Error("down"))),
      "/api/session",
      () => true,
      record(seen),
    );
    expect(seen[0]).toEqual({ kind: "transport" });
    const onlyTrue = (value: unknown) => (value === true ? true : null);
    await request(
      bridgeTransport(() => respond(200, "not json")),
      "/api/session",
      onlyTrue,
      record(seen),
    );
    expect(seen[1]).toEqual({ kind: "unreadable" });
  });

  it("recovers a verdict the bridge returned with an unsuccessful status", async () => {
    const transport = bridgeTransport(() =>
      respond(422, { error: { code: "replay_mismatch", message: "no", result: { matches: false } } }),
    );
    const readVerdict = (value: unknown) =>
      isRecord(value) && typeof value.matches === "boolean" ? { matches: value.matches } : null;
    const outcome = await request(
      transport,
      "/api/actions/run.verify",
      readVerdict,
      () => "failed",
      undefined,
      (p) => readVerdict(p.result),
    );
    expect(outcome).toEqual({ ok: true, value: { matches: false } });
  });

  it("reads files by encoded path, unwraps action results and sends the write body", async () => {
    const transport = bridgeTransport((url) =>
      url.startsWith("/api/file")
        ? respond(200, { path: "runs/w-1/run.json", content: "{}", etag: "e1" })
        : respond(200, { result: { ok: 1 } }),
    );
    const failed = () => "failed";
    expect(await getFile(transport, "runs/w-1/run.json", decodeObject, failed)).toEqual({
      ok: true,
      value: { path: "runs/w-1/run.json", content: "{}", etag: "e1" },
    });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/file?path=runs%2Fw-1%2Frun.json");
    expect(await action(transport, "run.report", { run: "w-1" }, decodeObject, failed)).toEqual({
      ok: true,
      value: { ok: 1 },
    });
    expect(transport.fetch.mock.calls[1]?.[1]).toMatchObject({ body: '{"args":{"run":"w-1"}}' });
    await putFile(transport, "compositions/lab.json", "{}", "e0", decodeObject, failed);
    expect(transport.fetch.mock.calls[2]?.[1]).toMatchObject({
      method: "PUT",
      body: '{"path":"compositions/lab.json","content":"{}","etag":"e0"}',
    });
  });
});
