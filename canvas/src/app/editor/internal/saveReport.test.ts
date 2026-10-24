import { describe, expect, it } from "vitest";
import type { Draft, Plausibility } from "./editorDraft";
import { failureWords, plausibilityRow, saveReportOf } from "./saveReport";

const doc = { id: "lab" };
const base: Draft = {
  phase: "ready",
  path: "compositions/lab.composition.json",
  domain: null,
  saved: doc,
  doc,
  save: { kind: "idle" },
  issues: null,
  past: [],
  future: [],
  specs: new Map(),
  library: [],
  selected: null,
};

function composed(plausibility: Plausibility, digest = `sha256:${"a".repeat(64)}`): Draft {
  return {
    ...base,
    save: { kind: "composed", summary: "Composed 3 channels, 2 faults, 1 effectors.", doc, digest, plausibility },
  };
}

const finding = (severity: string, detail: string) => ({ severity, entity: "CH-1", detail });

describe("the result of a save as separate claims", () => {
  it("has nothing to say before a successful save", () => {
    expect(saveReportOf(base)).toBeNull();
    expect(saveReportOf({ ...base, save: { kind: "saving" } })).toBeNull();
  });

  it("states the composition, the digest in full on hover, the plausibility, and that the interface does not apply", () => {
    const rows = saveReportOf(composed({ kind: "done", report: { errors: 0, warnings: 0, findings: [] } }));
    expect(rows?.map((row) => [row.id, row.state, row.text])).toEqual([
      ["composition", "ok", "Composed 3 channels, 2 faults, 1 effectors."],
      ["digest", "neutral", "sha256:aaaaaaaa…aaaaaaaa"],
      ["plausibility", "ok", "Readings stay plausible in a nominal day"],
      ["interface", "neutral", "Not applicable: a composed plant has no device interface to check"],
    ]);
    expect(rows?.[1]?.title).toBe(`sha256:${"a".repeat(64)}`);
  });

  it("says the digest was not reported by the binary", () => {
    expect(saveReportOf(composed({ kind: "running" }, ""))?.[1]).toMatchObject({ text: "not reported", title: "" });
  });
});

describe("the plausibility row", () => {
  it.each([
    [{ kind: "running" } as const, "pending", "Checking…", []],
    [{ kind: "failed", message: "lint broke" } as const, "warn", "Could not be checked: lint broke", []],
    [
      {
        kind: "done",
        report: { errors: 0, warnings: 2, findings: [finding("warning", "flat"), finding("warning", "drift")] },
      } as const,
      "warn",
      "Plausible, with 2 warnings",
      ["warning: CH-1: flat", "warning: CH-1: drift"],
    ],
    [
      { kind: "done", report: { errors: 1, warnings: 0, findings: [finding("error", "out of range")] } } as const,
      "bad",
      "1 issue",
      ["error: CH-1: out of range"],
    ],
  ])("reads %j", (plausibility, state, text, notes) => {
    expect(plausibilityRow(plausibility)).toMatchObject({ state, text, notes });
  });

  it("writes a refusal with its code", () => {
    expect(failureWords({ kind: "unavailable", diagnostic: { message: "down", displayCode: "bridge_failed" } })).toBe(
      "down (bridge_failed)",
    );
  });
});
