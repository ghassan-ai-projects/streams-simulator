import { describe, expect, it } from "vitest";
import { noteLinesByEntity } from "./noteLines";
import type { Report } from "./report";

function linesFor(fault: string, call: { readonly effector: string; readonly outcome: string } | null): number {
  const report = {
    disturbances: fault === "" ? [] : [{ kind: "fault", target: "x", name: fault }],
    effectorCalls: call === null ? [] : [{ entity: "x", ...call }],
  } as unknown as Report;
  return noteLinesByEntity(report).get("x") ?? 1;
}

const stop = { effector: "stop", outcome: "accepted" };
const setSpeed = { effector: "set_speed", outcome: "applied · 360 ms" };

describe("note lines", () => {
  it("keeps one line for a note with nothing, and a short cause or outcome", () => {
    expect(linesFor("", null)).toBe(1);
    expect(linesFor("scaling", null)).toBe(1);
    expect(linesFor("", stop)).toBe(1);
  });

  it("adds the wrapped lines of the fault line and of the effector line", () => {
    expect(linesFor("condenser_fouling", null)).toBe(2);
    expect(linesFor("", setSpeed)).toBe(2);
    expect(linesFor("condenser_fouling", setSpeed)).toBe(4);
  });

  it("reserves, per entity, the lines of everything the run says about it", () => {
    const report = {
      disturbances: [
        { kind: "fault", target: "ch-u2", name: "chiller__condenser_fouling" },
        { kind: "perturbation", target: "pmp-u3", name: "drop" },
        { kind: "fault", target: "ch-u2", name: "chiller__low_charge" },
      ],
      effectorCalls: [
        { entity: "pmp-u3", effector: "pump__set_speed", outcome: "applied · 360 ms" },
        { entity: "pmp-u3", effector: "pump__stop", outcome: "accepted" },
      ],
    } as unknown as Report;
    const lines = noteLinesByEntity(report);
    expect(lines.get("ch-u2")).toBe(3);
    expect(lines.get("pmp-u3")).toBe(2);
    expect(lines.has("elsewhere")).toBe(false);
    expect(noteLinesByEntity(null).size).toBe(0);
  });
});
