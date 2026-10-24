import { describe, expect, it } from "vitest";
import { reportWire } from "../../../test/runFixture";
import { parseReportPayload } from "./reportDecoder";

const question = { code: "stopped_early", message: "The operator stopped the run at +540s." };

describe("the replay status and open questions of a report", () => {
  it("reads the replay status with its detail and the open questions in the binary's order", () => {
    const wire = {
      ...reportWire,
      replay: { status: "does_not_reproduce", detail: "trace differs" },
      open_questions: [question, { code: "replay_does_not_reproduce", message: "Replaying does not reproduce." }],
    };
    expect(parseReportPayload(wire)).toMatchObject({
      replay: { status: "does_not_reproduce", detail: "trace differs" },
      openQuestions: [question, { code: "replay_does_not_reproduce", message: "Replaying does not reproduce." }],
    });
  });

  it.each(["not_checked", "reproduces", "incomplete"])("accepts the replay status %s", (status) => {
    expect(parseReportPayload({ ...reportWire, replay: { status } })?.replay).toEqual({ status, detail: "" });
  });

  it("reads a report from before these sections as not replayed with nothing open", () => {
    expect(parseReportPayload(reportWire)).toMatchObject({
      replay: { status: "not_checked", detail: "" },
      openQuestions: [],
    });
  });

  it("ignores fields it does not know, in the sections it reads", () => {
    const wire = {
      ...reportWire,
      replay: { status: "reproduces", future: 1 },
      open_questions: [{ ...question, severity: "high" }],
    };
    expect(parseReportPayload(wire)).toMatchObject({ replay: { status: "reproduces" }, openQuestions: [question] });
  });

  it.each([
    ["an unknown replay status", { replay: { status: "maybe" } }],
    ["a replay that is not an object", { replay: "reproduces" }],
    ["open questions that are not a list", { open_questions: "stopped_early" }],
    ["an open question without a code", { open_questions: [{ message: "x" }] }],
    ["an open question without a message", { open_questions: [{ code: "x" }] }],
  ])("refuses a report with %s rather than guess what is open", (_name, extra) => {
    expect(parseReportPayload({ ...reportWire, ...extra })).toBeNull();
  });

  it("marks an effector call applied only when the binary says the effect was applied", () => {
    const calls = parseReportPayload(reportWire)?.effectorCalls.map((call) => call.applied);
    expect(calls).toEqual([true, false, false, false]);
  });
});
