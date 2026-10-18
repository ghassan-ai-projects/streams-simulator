import { describe, expect, it } from "vitest";
import { isBusy, refusedState, stateOf } from "./connectState";

const diagnostic = { message: "refused", displayCode: "plantwire_failed" };

describe("connect states", () => {
  it("is busy while checking, starting or live", () => {
    expect(isBusy({ phase: "checking", domain: "d" })).toBe(true);
    expect(isBusy({ phase: "ended", domain: "d", run: "c-1" })).toBe(false);
  });

  it("reads each session status as a state of the panel", () => {
    expect(stateOf("d", null, { state: "running", run: "c-1" })).toMatchObject({
      phase: "live",
      run: "c-1",
      binding: null,
    });
    expect(stateOf("d", null, { state: "ended", run: "c-1" })).toEqual({ phase: "ended", domain: "d", run: "c-1" });
    expect(stateOf("d", null, { state: "none" })).toEqual({ phase: "idle" });
    expect(stateOf("d", null, { state: "failed", run: "c-1", diagnostic, issues: [] })).toMatchObject({
      phase: "failed",
    });
  });

  it("keeps a refused plan's issues and turns any other failure into a failed session without a run", () => {
    const issues = [{ pointer: "/seed", message: "not an integer" }];
    expect(refusedState("d", { kind: "refused", diagnostic, issues })).toEqual({
      phase: "refused",
      domain: "d",
      diagnostic,
      issues,
    });
    expect(refusedState("d", { kind: "unavailable", diagnostic })).toEqual({
      phase: "failed",
      domain: "d",
      run: null,
      diagnostic,
      issues: [],
    });
  });
});
