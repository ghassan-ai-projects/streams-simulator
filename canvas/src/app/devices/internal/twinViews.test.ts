import { describe, expect, it } from "vitest";
import { localIssue, renameKind, serialiseChoices } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";
import { coolerPath, domainPath, twinBridge } from "../../../test/twinFixture";
import { checkRows } from "./checkRows";
import type { Draft } from "./deviceDraft";
import { createDeviceSession } from "./deviceStore";
import { stepStatuses, twinReadiness } from "./stepStatus";
import type { Built, TwinDraft } from "./twinDraft";
import { canCreate, createBlock } from "./twinSelectors";
import { twinStatusOf } from "./twinStatus";

const unavailable: DeviceFailure = { kind: "unavailable", diagnostic: { message: "down", displayCode: "unreachable" } };
const conflict: DeviceFailure = { kind: "conflict", diagnostic: { message: "changed", displayCode: "conflict" } };
const refused: DeviceFailure = {
  kind: "refused",
  diagnostic: { message: "refused", displayCode: "kind_failed" },
  issues: [{ pointer: "/x", message: "differs", keyword: "implements_drift" }],
};
const bare: DeviceFailure = {
  kind: "refused",
  diagnostic: { message: "refused", displayCode: "kind_failed" },
  issues: [],
};

async function drafts() {
  const fake = twinBridge();
  const session = createDeviceSession(fake.devices);
  await session.open(coolerPath);
  const device = session.snapshot();
  const twin = session.twin.snapshot();
  if (device.phase !== "ready" || twin.phase !== "ready") throw new Error("not ready");
  return { device, twin };
}

function built(twin: TwinDraft, over: Partial<Built> = {}): TwinDraft {
  const made = { id: "cooler-twin", digest: "d", channels: 1, faults: 2, effectors: 3, implements: ["cooler"] };
  const value: Built = {
    kind: "built",
    made,
    path: domainPath,
    fingerprint: serialiseChoices(twin.doc, twin.members),
    implemented: { kind: "passed", checked: { id: "cooler-twin", digest: "d" } },
    lint: { kind: "done", report: { errors: 0, warnings: 0, findings: [] } },
    ...over,
  };
  return { ...twin, build: value };
}

describe("the state of each step", () => {
  it("says whether the interface is valid, has issues, is being checked or has unsaved changes", async () => {
    const { device, twin } = await drafts();
    const status = (draft: Draft) => stepStatuses(draft, twin).interface;
    expect(status(device)).toEqual({ text: "Valid", tone: "ok" });
    expect(status({ ...device, check: { kind: "refused", failure: refused } })).toMatchObject({ text: "Has issues" });
    const flagged = { ...device.review, schema: [localIssue("kind", "", "kind", "x")] };
    expect(status({ ...device, review: flagged })).toEqual({ text: "Has issues", tone: "warn" });
    expect(status({ ...device, check: { kind: "checking" } })).toEqual({ text: "Checking", tone: "busy" });
    expect(status({ ...device, check: { kind: "unchecked" } })).toEqual({ text: "Not checked", tone: "" });
    expect(status({ ...device, doc: renameKind(device.doc, "other") })).toEqual({ text: "Unsaved", tone: "warn" });
    expect(stepStatuses({ phase: "idle" }, twin).interface).toEqual({ text: "Opening", tone: "busy" });
  });

  it("says whether the behaviour is not started, ready, in need of choices, created or refused", async () => {
    const { device, twin } = await drafts();
    const status = (state: Parameters<typeof stepStatuses>[1]) => stepStatuses(device, state).behaviour;
    expect(status({ phase: "idle" })).toEqual({ text: "Not started", tone: "" });
    expect(status({ phase: "loading", path: coolerPath })).toEqual({ text: "Opening", tone: "busy" });
    expect(status({ phase: "failed", path: coolerPath, failure: unavailable, held: [] })).toEqual({
      text: "Not started",
      tone: "",
    });
    expect(status(twin)).toEqual({ text: "Not started", tone: "" });
    expect(status({ ...twin, touched: true })).toEqual({ text: "Ready", tone: "" });
    expect(status({ ...twin, build: { kind: "building" } })).toEqual({ text: "Creating", tone: "busy" });
    expect(status(built(twin))).toEqual({ text: "Created", tone: "ok" });
    const issue = { scope: "elsewhere" as const, kind: "", key: "", field: "", spot: "", pointer: "", keyword: "" };
    const refusedWith = { ...twin, build: { kind: "refused" as const, failure: refused } };
    expect(status({ ...refusedWith, issues: [{ ...issue, message: "m", text: "m" }] })).toEqual({
      text: "Needs attention",
      tone: "warn",
    });
    expect(status({ ...twin, build: { kind: "exists" } })).toMatchObject({ text: "Needs attention" });
    const open = { ...twin, doc: { ...twin.doc, title: " " } };
    expect(status(open)).toEqual({ text: "Needs choices", tone: "warn" });
  });

  it("says whether the check passed, failed, is running or has not run", async () => {
    const { device, twin } = await drafts();
    const status = (state: Parameters<typeof stepStatuses>[1]) => stepStatuses(device, state).check;
    expect(status({ phase: "idle" })).toEqual({ text: "Not run", tone: "" });
    expect(status(twin)).toEqual({ text: "Not run", tone: "" });
    expect(status(built(twin))).toEqual({ text: "Checks passed", tone: "ok" });
    expect(status(built(twin, { implemented: { kind: "checking" } }))).toEqual({ text: "Checking", tone: "busy" });
    const errors = { errors: 2, warnings: 0, findings: [] };
    expect(status(built(twin, { lint: { kind: "done", report: errors } }))).toEqual({ text: "Failed", tone: "warn" });
    expect(status(built(twin, { lint: { kind: "running" } }))).toEqual({ text: "Checking", tone: "busy" });
    expect(status(built(twin, { implemented: { kind: "failed", failure: refused } }))).toEqual({
      text: "Failed",
      tone: "warn",
    });
    expect(status(built({ ...twin, doc: { ...twin.doc, id: "changed" } }, { fingerprint: "old" }))).toEqual({
      text: "Not run",
      tone: "",
    });
  });
});

describe("what a twin needs of the interface", () => {
  it("is a second status beside the first: valid, and still lacking something for a twin", async () => {
    const { device } = await drafts();
    expect(twinReadiness(device)).toEqual({ text: "Ready for a twin", tone: "ok" });
    const need = { ...localIssue("metric", "m1", "resolution", "Needs a resolution"), twin: true };
    const lacking = { ...device, review: { state: "done" as const, schema: [], needs: [need, need] } };
    expect(twinReadiness(lacking)).toEqual({ text: "Needed for a twin: 2", tone: "warn" });
    expect(stepStatuses(lacking, { phase: "idle" }).interface).toEqual({ text: "Valid", tone: "ok" });
    expect(twinReadiness({ ...device, review: { ...device.review, state: "checking" } })).toEqual({
      text: "Checking what a twin needs",
      tone: "busy",
    });
    expect(twinReadiness({ ...device, review: { state: "failed", schema: [], needs: [] } })).toEqual({
      text: "",
      tone: "",
    });
    expect(twinReadiness({ phase: "idle" })).toEqual({ text: "", tone: "" });
  });
});

describe("what blocks the creation of a twin", () => {
  it("names the reason, in the order it matters", async () => {
    const { device, twin } = await drafts();
    expect(createBlock(twin, device)).toBeNull();
    expect(canCreate(twin, device)).toBe(true);
    expect(createBlock(twin, { ...device, doc: renameKind(device.doc, "other") })).toMatch(/Save the interface first/);
    expect(createBlock({ ...twin, build: { kind: "building" } }, device)).toMatch(/being created/);
    const stuck = { ...twin, save: { kind: "write-failed" as const, failure: conflict } };
    expect(createBlock(stuck, device)).toMatch(/changed on disk/);
    expect(createBlock({ ...twin, save: { kind: "write-failed", failure: unavailable } }, device)).toBeNull();
    const unfinished = { ...twin, doc: { ...twin.doc, title: "" } };
    expect(createBlock(unfinished, device)).toMatch(/Finish the choices/);
    expect(canCreate({ ...twin, save: { kind: "saving" } }, device)).toBe(true);
  });

  it("names what the interface lacks for a twin before the form's own gaps", async () => {
    const { device, twin } = await drafts();
    const need = { ...localIssue("metric", "m1", "resolution", "Needs a resolution"), twin: true };
    const lacking = { ...device, review: { state: "done" as const, schema: [], needs: [need] } };
    expect(createBlock(twin, lacking)).toBe("Set what the twin needs in the Interface step first.");
    expect(canCreate(twin, lacking)).toBe(false);
    expect(createBlock({ ...twin, doc: { ...twin.doc, title: "" } }, lacking)).toMatch(/Interface step/);
    expect(twinStatusOf(twin, lacking)).toEqual({
      text: "Set what the twin needs in the Interface step first.",
      tone: "warn",
      detail: null,
    });
  });
});

describe("the headline of the behaviour step", () => {
  it("tells what happened to the choices and to the twin, in plain words", async () => {
    const { device, twin } = await drafts();
    const text = (draft: TwinDraft) => twinStatusOf(draft, device);
    expect(text({ ...twin, save: { kind: "write-failed", failure: conflict } }).text).toMatch(/Reload it/);
    expect(text({ ...twin, save: { kind: "write-failed", failure: unavailable } }).text).toBe(
      "Choices not saved: down",
    );
    expect(twinStatusOf(twin, { ...device, doc: renameKind(device.doc, "other") }).text).toMatch(
      /Save the interface first/,
    );
    expect(text(twin)).toMatchObject({ text: "Choices are ready. Create the twin when you are.", tone: "" });
    expect(text({ ...twin, onDisk: true }).text).toBe("Choices are ready. Replace the twin when you are.");
    expect(text({ ...twin, doc: { ...twin.doc, title: "" } }).text).toBe(
      "Finish the choices marked in the form first.",
    );
    expect(text({ ...twin, build: { kind: "building" } }).text).toBe("Creating the twin…");
    expect(text({ ...twin, build: { kind: "exists" } }).tone).toBe("warn");
    expect(text({ ...twin, build: { kind: "failed", failure: unavailable } }).text).toBe(
      "The twin could not be created: down",
    );
    expect(text({ ...twin, build: { kind: "refused", failure: bare } }).text).toBe(
      "Choices are ready. Create the twin when you are.",
    );
    const issue = { scope: "elsewhere" as const, kind: "", key: "", field: "", spot: "" };
    const issues = [{ ...issue, pointer: "", keyword: "", message: "m", text: "m" }];
    expect(text({ ...twin, build: { kind: "refused", failure: refused }, issues }).text).toBe(
      "The binary refuses the choices: 1 issue.",
    );
  });

  it("tells that the twin is ready, being checked, not matching its device, or out of date", async () => {
    const { device, twin } = await drafts();
    expect(twinStatusOf(built(twin), device)).toEqual({
      text: "Twin created. Interface check passed; plausibility check passed; not calibrated; not compared with a board.",
      tone: "ok",
      detail: `Twin created. Interface check passed; plausibility check passed; not calibrated; not compared with a board. (${domainPath})`,
    });
    expect(twinStatusOf(built(twin, { implemented: { kind: "checking" } }), device).text).toBe("Checking the twin…");
    const failed = built(twin, { implemented: { kind: "failed", failure: refused } });
    expect(twinStatusOf(failed, device)).toMatchObject({ tone: "warn", detail: "refused" });
    expect(twinStatusOf(failed, device).text).toMatch(/does not match its device/);
    const errors = { errors: 3, warnings: 0, findings: [] };
    const unplausible = twinStatusOf(built(twin, { lint: { kind: "done", report: errors } }), device);
    expect(unplausible).toMatchObject({
      tone: "warn",
      text: "Twin created, but the plausibility check found 3 errors.",
    });
    expect(twinStatusOf(built(twin, { lint: { kind: "running" } }), device).text).toBe("Checking the twin…");
    const stale = built({ ...twin, doc: { ...twin.doc, id: "other-twin" } }, { fingerprint: "old" });
    expect(twinStatusOf(stale, device).text).toMatch(/Choices changed since/);
  });
});

describe("the rows of the check", () => {
  it("starts with nothing created", async () => {
    const { device, twin } = await drafts();
    expect(checkRows(device, twin).map((row) => [row.id, row.state])).toEqual([
      ["device", "ok"],
      ["twin", "idle"],
      ["implements", "idle"],
      ["lint", "idle"],
      ["calibration", "neutral"],
      ["hardware", "neutral"],
    ]);
    expect(checkRows(device, twin)[1]?.text).toBe("No twin has been created yet");
  });

  it("tells the state of the device definition", async () => {
    const { device, twin } = await drafts();
    const row = (state: Parameters<typeof checkRows>[0]) => checkRows(state, twin)[0];
    expect(row({ phase: "idle" })).toMatchObject({ state: "pending" });
    expect(row({ ...device, doc: renameKind(device.doc, "other") })).toMatchObject({ state: "idle" });
    expect(row({ ...device, check: { kind: "refused", failure: refused } })).toMatchObject({ state: "bad" });
    expect(row({ ...device, check: { kind: "unchecked" } })).toMatchObject({
      state: "idle",
      text: expect.stringMatching(/not been checked/) as unknown,
    });
  });

  it("tells how the creation went", async () => {
    const { device, twin } = await drafts();
    const row = (draft: TwinDraft) => checkRows(device, draft)[1];
    expect(row({ ...twin, build: { kind: "building" } })).toMatchObject({ state: "pending" });
    expect(row({ ...twin, build: { kind: "exists" } })).toMatchObject({ state: "bad" });
    expect(row({ ...twin, build: { kind: "failed", failure: unavailable } })).toMatchObject({
      state: "bad",
      notes: ["down"],
    });
    expect(row({ ...twin, build: { kind: "refused", failure: refused } })).toMatchObject({
      notes: ["The twin no longer matches this device"],
      details: ["differs"],
    });
    expect(row({ ...twin, build: { kind: "refused", failure: bare } })).toMatchObject({ notes: ["refused"] });
    expect(row(built(twin))).toMatchObject({ state: "ok", text: "Twin created" });
    const errors = { errors: 1, warnings: 0, findings: [] };
    expect(row(built(twin, { lint: { kind: "done", report: errors } }))).toMatchObject({
      state: "bad",
      text: "The twin was created, but it has problems",
    });
    expect(row(built(twin, { implemented: { kind: "checking" } }))).toMatchObject({ state: "pending" });
    const stale = built({ ...twin, doc: { ...twin.doc, id: "other-twin" } }, { fingerprint: "old" });
    expect(row(stale)).toMatchObject({ state: "idle", text: "The choices changed since the twin was created" });
  });

  it("tells whether the twin matches its device and whether its readings stay plausible", async () => {
    const { device, twin } = await drafts();
    const rows = (over: Partial<Built>) => checkRows(device, built(twin, over));
    expect(rows({})[2]).toMatchObject({ id: "implements", state: "ok" });
    expect(rows({ implemented: { kind: "checking" } })[2]).toMatchObject({ state: "pending" });
    expect(rows({ implemented: { kind: "failed", failure: refused } })[2]).toMatchObject({
      state: "bad",
      notes: ["The twin no longer matches this device"],
      details: ["differs"],
    });
    expect(rows({})[3]).toMatchObject({
      id: "lint",
      state: "ok",
      status: "Passed",
      text: "Plausibility: readings stay plausible in a nominal day",
    });
    expect(rows({ lint: { kind: "running" } })[3]).toMatchObject({ state: "pending" });
    expect(rows({ lint: { kind: "unavailable" } })[3]).toMatchObject({ state: "idle", status: "Not run" });
    const warned = { errors: 0, warnings: 2, findings: [{ severity: "warning", entity: "a", detail: "slow" }] };
    expect(rows({ lint: { kind: "done", report: warned } })[3]).toMatchObject({
      state: "warn",
      status: "Passed with 2 warnings",
      notes: ["a: slow"],
    });
    const findings = Array.from({ length: 8 }, (_unused, index) => ({
      severity: "error",
      entity: "e",
      detail: String(index),
    }));
    const failed = rows({ lint: { kind: "done", report: { errors: 8, warnings: 0, findings } } })[3];
    expect(failed).toMatchObject({ state: "bad", status: "Failed", text: "Plausibility: the check found 8 errors" });
    expect(failed?.notes).toHaveLength(7);
    expect(failed?.notes.at(-1)).toBe("…and 2 more");
  });
});
