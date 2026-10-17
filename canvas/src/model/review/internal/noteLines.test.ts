import { describe, expect, it } from "vitest";
import { noteLinesByEntity, noteLinesOf } from "./noteLines";
import type { Report } from "./report";

describe("note lines", () => {
  it("keeps one line for a note with nothing, and a short cause or outcome", () => {
    expect(noteLinesOf("", "")).toBe(1);
    expect(noteLinesOf("scaling", "")).toBe(1);
    expect(noteLinesOf("", "stop · accepted")).toBe(1);
  });

  it("adds the wrapped lines of the fault line and of the effector line", () => {
    expect(noteLinesOf("condenser fouling", "")).toBe(2);
    expect(noteLinesOf("", "set speed · applied · 360 ms")).toBe(2);
    expect(noteLinesOf("condenser fouling", "set speed · applied · 360 ms")).toBe(4);
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
