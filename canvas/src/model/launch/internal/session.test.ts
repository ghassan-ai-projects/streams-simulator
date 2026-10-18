import { describe, expect, it } from "vitest";
import {
  firstSessionDraft,
  freeInstance,
  memberIssues,
  rowIssues,
  sessionPlanText,
  unplacedIssues,
  withRealRow,
  withRowField,
  withoutRealRow,
} from "./session";
import type { PlanIssue } from "./session";
import { bindingWords, checkWords, issueDetail, issueWords, plannedWords } from "./sessionWords";

describe("session plan", () => {
  it("writes the form as the plan document the binary reads, with no gateway or real board when none is given", () => {
    const plan = JSON.parse(
      sessionPlanText("arduino-thermal", { ...firstSessionDraft, hostSocket: " /tmp/h.sock " }),
    ) as unknown;
    expect(plan).toEqual({ schema_version: "0.1", domain: "arduino-thermal", seed: 1, host_socket: "/tmp/h.sock" });
  });

  it("sends the seed as typed: a whole number as a number, anything else as text the binary refuses", () => {
    const seed = (typed: string): unknown =>
      (JSON.parse(sessionPlanText("d", { ...firstSessionDraft, seed: typed })) as { seed: unknown }).seed;
    expect(seed(" 42 ")).toBe(42);
    expect(seed("seven")).toBe("seven");
    expect(seed("007")).toBe("007");
    expect(seed("1.5")).toBe("1.5");
    const huge = sessionPlanText("d", { ...firstSessionDraft, seed: "18446744073709551615" });
    expect(huge).toContain('"seed":18446744073709551615');
  });

  it("adds, edits and removes real boards and writes them with the gateway", () => {
    let draft = withRealRow({ ...firstSessionDraft, gatewaySocket: "/tmp/g.sock" }, "zone-01");
    draft = withRowField(withRowField(draft, 0, "board", "uno-07"), 0, "device", "zone-01");
    draft = withRealRow(draft, "fan-01");
    expect(draft.real.map((row) => row.key)).toEqual([0, 1]);
    const kept = withoutRealRow(draft, 1);
    expect(JSON.parse(sessionPlanText("d", kept))).toEqual({
      schema_version: "0.1",
      domain: "d",
      seed: 1,
      host_socket: "",
      gateway_socket: "/tmp/g.sock",
      real: [{ instance: "zone-01", board: "uno-07", device: "zone-01" }],
    });
    expect(withRealRow(kept, "pump-01").real.map((row) => row.key)).toEqual([0, 2]);
  });

  it("offers the next instance not yet bound, and none when every instance is", () => {
    const rows = withRealRow(firstSessionDraft, "zone-01").real;
    expect(freeInstance(["zone-01", "fan-01"], rows)).toBe("fan-01");
    expect(freeInstance(["zone-01"], rows)).toBeNull();
  });
});

describe("issue placement", () => {
  const issues: readonly PlanIssue[] = [
    { pointer: "/real/0/device", message: "missing device", keyword: "device_missing", member: "device" },
    { pointer: "/real/0", message: "kind differs", keyword: "kind_mismatch", member: "instance" },
    { pointer: "/real", message: "needs a gateway", keyword: "gateway_required", member: "gateway_socket" },
    { pointer: "/host_socket", message: "not absolute", keyword: "pattern" },
    { pointer: "/host_socket", message: "no control character", keyword: "socket_control_character" },
    { pointer: "/real/3/board", message: "a row no longer there", keyword: "board_missing" },
    { pointer: "", message: "the domain is not known" },
    { pointer: "/kinds", message: "kinds unreadable" },
  ];

  it("places every issue on the row and field its pointer or member names", () => {
    expect(rowIssues(issues, 0, "device").map((issue) => issue.keyword)).toEqual(["device_missing"]);
    expect(rowIssues(issues, 0, "instance").map((issue) => issue.keyword)).toEqual(["kind_mismatch"]);
    expect(rowIssues(issues, 0, "board")).toEqual([]);
    expect(rowIssues(issues, 1, "device")).toEqual([]);
  });

  it("places every issue on the member it points at or names, and lists all the rest", () => {
    expect(memberIssues(issues, "host_socket").map((issue) => issue.message)).toEqual([
      "not absolute",
      "no control character",
    ]);
    expect(memberIssues(issues, "gateway_socket").map((issue) => issue.keyword)).toEqual(["gateway_required"]);
    expect(memberIssues(issues, "seed")).toEqual([]);
    expect(unplacedIssues(issues, 1).map((issue) => issue.message)).toEqual([
      "a row no longer there",
      "the domain is not known",
      "kinds unreadable",
    ]);
    expect(unplacedIssues(issues, 4).map((issue) => issue.message)).not.toContain("a row no longer there");
  });

  it("words each issue plainly by its rule or member, keeping the binary's text for developers", () => {
    expect(issueWords(issues[0] ?? { pointer: "", message: "" })).toBe("The board holds no such device.");
    expect(issueWords({ pointer: "/seed", message: "want integer", keyword: "type" })).toBe(
      "The seed is a whole number, such as 1.",
    );
    expect(issueWords({ pointer: "/real/0/board", message: "x", keyword: "required", member: "board" })).toBe(
      "Give the board's name as the gateway announces it.",
    );
    expect(issueWords({ pointer: "", message: "the domain is not known" })).toBe("the domain is not known");
    expect(issueDetail({ pointer: "/seed", message: "want integer", keyword: "type" })).toBe(
      "type · /seed · want integer",
    );
    expect(issueDetail({ pointer: "", message: "the domain is not known" })).toBe("/ · the domain is not known");
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
    expect(plannedWords(real)).toBe("zone-01 · Real · device zone-01 on board uno-07");
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

  it("says whether the gateway matched the plan, and the next step when it did not", () => {
    const refused = { matched: false, refusal: "inventory_mismatch", reason: "no board", issues: [] };
    expect(bindingWords({ matched: true, refusal: "", reason: "", issues: [] })).toContain("every real board");
    expect(bindingWords(refused)).toBe(
      "The gateway was refused (inventory_mismatch), so its real boards were not bound. " +
        "Stop the session, fix the board or device, then start it again.",
    );
    const real = { instance: "zone-01", role: "real", kind: "k", board: "uno-07", device: "zone-01" } as const;
    expect(plannedWords(real, refused)).toBe("zone-01 · Not bound: refused · not announced to the host");
    expect(plannedWords({ ...real, role: "simulated" }, refused)).toBe("zone-01 · Simulated");
  });
});
