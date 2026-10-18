import { describe, expect, it } from "vitest";
import {
  firstSessionDraft,
  memberIssue,
  rowIssue,
  sessionPlanText,
  unplacedIssues,
  withRealRow,
  withRowField,
  withoutRealRow,
} from "./session";
import type { PlanIssue } from "./session";
import { bindingWords, checkWords, plannedWords } from "./sessionWords";

describe("session plan", () => {
  it("writes the form as the plan document the binary reads, with no gateway or real board when none is given", () => {
    const plan = JSON.parse(
      sessionPlanText("arduino-thermal", { ...firstSessionDraft, hostSocket: " /tmp/h.sock " }),
    ) as unknown;
    expect(plan).toEqual({ schema_version: "0.1", domain: "arduino-thermal", seed: 1, host_socket: "/tmp/h.sock" });
  });

  it("keeps a seed that is not a number as text so the binary refuses it at its member", () => {
    const text = sessionPlanText("d", { ...firstSessionDraft, seed: "seven" });
    expect(JSON.parse(text)).toMatchObject({ seed: "seven" });
  });

  it("adds, edits and removes real boards and writes them with the gateway", () => {
    let draft = withRealRow({ ...firstSessionDraft, gatewaySocket: "/tmp/g.sock" }, "zone-01");
    draft = withRowField(withRowField(draft, 0, "board", "uno-07"), 0, "device", "zone-01");
    draft = withRealRow(draft, "fan-01");
    expect(JSON.parse(sessionPlanText("d", withoutRealRow(draft, 1)))).toMatchObject({
      gateway_socket: "/tmp/g.sock",
      real: [{ instance: "zone-01", board: "uno-07", device: "zone-01" }],
    });
  });
});

describe("issue placement", () => {
  const issues: readonly PlanIssue[] = [
    { pointer: "/real/0/device", message: "missing device", keyword: "device_missing", member: "device" },
    { pointer: "/real", message: "needs a gateway", keyword: "gateway_required", member: "gateway_socket" },
    { pointer: "/host_socket", message: "not absolute", keyword: "pattern" },
    { pointer: "", message: "the domain is not known" },
  ];

  it("places an issue on the row and field its pointer names", () => {
    expect(rowIssue(issues, 0, "device")?.keyword).toBe("device_missing");
    expect(rowIssue(issues, 0, "board")).toBeUndefined();
    expect(rowIssue(issues, 1, "device")).toBeUndefined();
  });

  it("places an issue on the member it points at or names, and keeps the rest for the panel", () => {
    expect(memberIssue(issues, "host_socket")?.message).toBe("not absolute");
    expect(memberIssue(issues, "gateway_socket")?.keyword).toBe("gateway_required");
    expect(memberIssue(issues, "seed")).toBeUndefined();
    expect(unplacedIssues(issues).map((issue) => issue.message)).toEqual(["the domain is not known"]);
  });
});

describe("session words", () => {
  it("names each instance real or simulated and counts what the plan holds", () => {
    const real = {
      instance: "zone-01",
      role: "real",
      kind: "thermal_sensor",
      board: "uno-07",
      device: "zone-01",
    } as const;
    const simulated = { instance: "fan-01", role: "simulated", kind: "fan", board: "", device: "" } as const;
    expect(plannedWords(real)).toBe("zone-01 · Real board · device zone-01 on board uno-07");
    expect(plannedWords(simulated)).toBe("fan-01 · Simulated");
    expect(checkWords({ domain: "d", mixed: true, instances: [real, simulated] })).toBe(
      "The plan holds: 1 real board and 1 simulated instance.",
    );
    expect(checkWords({ domain: "d", mixed: true, instances: [simulated, simulated] })).toBe(
      "The plan holds: 2 simulated instances beside the gateway's boards.",
    );
    expect(checkWords({ domain: "d", mixed: false, instances: [simulated] })).toBe(
      "The plan holds: 1 simulated instance.",
    );
    expect(checkWords({ domain: "d", mixed: false, instances: [real, real] })).toContain("2 real boards");
  });

  it("says whether the gateway matched the plan, with the binary's reason when it did not", () => {
    expect(bindingWords({ matched: true, refusal: "", reason: "", issues: [] })).toContain("every real board");
    expect(bindingWords({ matched: false, refusal: "inventory_mismatch", reason: "no board", issues: [] })).toBe(
      "The gateway was refused (inventory_mismatch): no board.",
    );
    expect(bindingWords({ matched: false, refusal: "environment_mismatch", reason: "", issues: [] })).toBe(
      "The gateway was refused (environment_mismatch).",
    );
  });
});
