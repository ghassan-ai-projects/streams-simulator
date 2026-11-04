import { describe, expect, it } from "vitest";
import { firstEventNode } from "./reviewAnchor";
import type { Report } from "./report";

function reportWith(parts: Partial<Pick<Report, "disturbances" | "effectorCalls">>): Report {
  return { disturbances: [], effectorCalls: [], ...parts } as unknown as Report;
}

const fault = (target: string, onsetNs: number, kind = "fault") => ({ kind, target, onsetNs }) as never;
const call = (entity: string, atNs: number) => ({ entity, atNs }) as never;

describe("the node a review opens on", () => {
  const shown = new Set(["boiler-1", "pump-2", "hall-a"]);

  it("is the node of the earliest fault or effector call, whichever comes first", () => {
    const report = reportWith({
      disturbances: [fault("boiler-1", 600e9)],
      effectorCalls: [call("pump-2", 300e9), call("boiler-1", 900e9)],
    });
    expect(firstEventNode(report, shown)).toBe("pump-2");
  });

  it("skips perturbations and entities the view does not show, and opens a hall for one of its members", () => {
    const report = reportWith({
      disturbances: [fault("pump-2", 10e9, "perturbation"), fault("elsewhere", 20e9), fault("hall-a/fan-1", 30e9)],
    });
    expect(firstEventNode(report, shown)).toBe("hall-a");
  });

  it("is nothing when the run has no event in view or no report yet", () => {
    expect(firstEventNode(reportWith({}), shown)).toBeNull();
    expect(firstEventNode(null, shown)).toBeNull();
  });
});
