import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { request } from "./http";
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
});
