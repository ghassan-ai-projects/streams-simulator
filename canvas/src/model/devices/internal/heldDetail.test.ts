import { describe, expect, it } from "vitest";
import { kindOf } from "../../../test/twinFixture";
import { blockingWords, heldIssues, heldReason, heldSelection, leftOutWords } from "./heldDetail";

const kind = kindOf({
  kind: "indicator",
  metrics: [
    { name: "level_percent", data_type: "number", unit: "%", scan_class: "fast" },
    { name: "state", data_type: "bool", description: "x" },
  ],
  commands: [{ name: "reset", risk_class: "low", args_schema: { type: "object", properties: {} } }],
});

const scan = {
  pointer: "/metrics/0/scan_class",
  message: "is not part of the interface a twin reproduces",
  keyword: "twin_member_unsupported",
  member: "scan_class",
};
const described = { ...scan, pointer: "/metrics/1/description", member: "description" };
const risk = { pointer: "/commands/0/safe_stop", message: "m", keyword: "twin_member_required", member: "safe_stop" };

describe("naming what a twin does not reproduce", () => {
  it("names the setting and the reading it belongs to, from the pointer and the member", () => {
    const [issue] = heldIssues(kind, [scan]);
    expect(issue).toMatchObject({
      owner: "reading",
      ownerName: "level_percent",
      setting: "scan_class",
      sentence: "The setting scan_class of reading 'level_percent' is not reproduced by a twin.",
      original: "is not part of the interface a twin reproduces",
      part: "metric",
      field: "scan_class",
    });
  });

  it("keeps every issue of the device, each with its own owner", () => {
    const issues = heldIssues(kind, [scan, described, risk]);
    expect(issues.map((issue) => [issue.owner, issue.ownerName, issue.setting])).toEqual([
      ["reading", "level_percent", "scan_class"],
      ["reading", "state", "description"],
      ["command", "reset", "safe_stop"],
    ]);
    expect(issues[2]?.sentence).toBe("Command 'reset' needs a safe-stop choice.");
  });

  it("speaks of the device when the pointer names no reading or command", () => {
    const [issue] = heldIssues(kind, [{ pointer: "/kind", message: "m", keyword: "pattern", member: "" }]);
    expect(issue).toMatchObject({ owner: "device", sentence: "The device: This value is not in the accepted form." });
    const [loose] = heldIssues(kind, [{ ...scan, pointer: "/metrics/9/scan_class" }]);
    expect(loose?.sentence).toBe("The setting scan_class of the device is not reproduced by a twin.");
    const [unnamed] = heldIssues(kind, [{ ...scan, member: "" }]);
    expect(unnamed?.sentence).toBe("A setting of reading 'level_percent' is not reproduced by a twin.");
  });

  it("gives a short reason for a list of issues", () => {
    const issues = heldIssues(kind, [scan, described, risk, scan]);
    expect(heldReason(issues.slice(0, 1))).toBe(
      "the setting scan_class of reading 'level_percent' is not reproduced by a twin",
    );
    expect(heldReason(issues)).toBe(
      "the setting scan_class of reading 'level_percent' is not reproduced by a twin; the setting description of reading 'state' is not reproduced by a twin; command 'reset' needs a safe-stop choice; and 1 more",
    );
  });

  it("says that the whole device is left out, or blocks the twin", () => {
    expect(leftOutWords("indicator")).toMatch(/^indicator is left out of the twin as a whole/);
    expect(blockingWords("indicator")).toMatch(/cannot be created until it is fixed or left out/);
  });

  it("points at the reading or command, or at the device", () => {
    const [reading, , command] = heldIssues(kind, [scan, described, risk]);
    const [device] = heldIssues(kind, [{ pointer: "/kind", message: "m", keyword: "pattern", member: "" }]);
    expect(reading && heldSelection(reading)).toEqual({ part: "metric", key: "m1" });
    expect(command && heldSelection(command)).toEqual({ part: "command", key: "c1" });
    expect(device && heldSelection(device)).toEqual({ part: "kind" });
  });
});
