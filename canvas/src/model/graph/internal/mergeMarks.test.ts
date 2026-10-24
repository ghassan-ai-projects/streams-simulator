import { describe, expect, it } from "vitest";
import { mergeMarks } from "./mergeMarks";
import type { NodeStatus } from "./types";

interface Mark {
  status: NodeStatus;
  count: number;
  fault: string;
  effector: string;
}

const mark = (overrides: Partial<Mark>): Mark => ({
  status: null,
  count: 0,
  fault: "",
  effector: "",
  ...overrides,
});

describe("merging what a run is showing with what the observer reported", () => {
  it("keeps what was reported when the plan adds nothing", () => {
    const live = mark({ status: "warn", count: 2 });
    expect(mergeMarks(live, mark({}))).toEqual(live);
  });

  it("adds the plan's fault and effector call to a node that reported nothing remarkable", () => {
    const planned = mark({ status: "fault", fault: "fan failure", effector: "start · accepted" });
    expect(mergeMarks(mark({}), planned)).toEqual(planned);
  });

  it("lets an alarm from the observer outrank a planned fault, and a fault outrank a warning", () => {
    expect(mergeMarks(mark({ status: "alarm" }), mark({ status: "fault", fault: "x" }))).toMatchObject({
      status: "alarm",
      fault: "x",
    });
    expect(mergeMarks(mark({ status: "warn" }), mark({ status: "fault", fault: "x" }))).toMatchObject({
      status: "fault",
    });
  });

  it("counts the faulted members of a hall that also has a warning", () => {
    expect(mergeMarks(mark({ status: "warn", count: 3 }), mark({ status: "fault", count: 1 }))).toMatchObject({
      status: "fault",
      count: 1,
    });
  });
});
