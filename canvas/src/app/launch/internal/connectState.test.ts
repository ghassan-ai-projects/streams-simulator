import { describe, expect, it } from "vitest";
import { isBusy, markedStale, refusedState, stateOf } from "./connectState";

const diagnostic = { message: "refused", displayCode: "plantwire_failed" };
const check = { domain: "d", mixed: false, instances: [], policies: [] };
const facts = { run: "c-1", domain: "d", check: null, controls: true };

describe("connect states", () => {
  it("is busy while checking, starting or live", () => {
    expect(isBusy({ phase: "checking", domain: "d" })).toBe(true);
    expect(isBusy({ phase: "ended", domain: "d", run: "c-1" })).toBe(false);
  });

  it("reads each session status as a state of the panel, with the session's own domain and plan", () => {
    expect(stateOf("d", check, { state: "running", ...facts })).toEqual({
      phase: "live",
      domain: "d",
      run: "c-1",
      check,
      binding: null,
      stopping: false,
      controls: true,
      note: null,
    });
    expect(stateOf("d", check, { state: "running", ...facts, controls: false })).toMatchObject({ controls: false });
    expect(stateOf("", null, { state: "stopping", ...facts, domain: "other", check })).toMatchObject({
      phase: "live",
      domain: "other",
      check,
      stopping: true,
    });
    expect(stateOf("d", null, { state: "starting", ...facts })).toEqual({
      phase: "starting",
      domain: "d",
      check: null,
    });
    expect(stateOf("d", null, { state: "ended", ...facts })).toEqual({ phase: "ended", domain: "d", run: "c-1" });
    expect(stateOf("d", null, { state: "none" })).toEqual({ phase: "idle" });
    expect(stateOf("d", null, { state: "failed", ...facts, diagnostic, issues: [] })).toMatchObject({
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
      stale: false,
    });
    expect(refusedState("d", { kind: "unavailable", diagnostic })).toEqual({
      phase: "failed",
      domain: "d",
      run: null,
      diagnostic,
      issues: [],
    });
  });

  it("marks a check out of date when the plan of its own plant changes, and nothing else", () => {
    const checked = { phase: "checked", domain: "d", check, stale: false } as const;
    expect(markedStale(checked, "d")).toEqual({ ...checked, stale: true });
    expect(markedStale(checked, "other")).toBe(checked);
    const idle = { phase: "idle" } as const;
    expect(markedStale(idle, "d")).toBe(idle);
  });
});
