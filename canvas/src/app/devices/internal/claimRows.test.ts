import { describe, expect, it } from "vitest";
import type { DeviceFailure } from "../../ports";
import { serialiseChoices } from "../../../model/devices";
import { coolerPath, domainPath, twinBridge } from "../../../test/twinFixture";
import { checkFacts } from "./checkFacts";
import { checkRows } from "./checkRows";
import { checkSummary } from "./claimRows";
import { createDeviceSession } from "./deviceStore";
import type { Built, TwinDraft } from "./twinDraft";
import { twinStatusOf } from "./twinStatus";

const refused: DeviceFailure = {
  kind: "refused",
  diagnostic: { message: "refused", displayCode: "kind_failed" },
  issues: [{ pointer: "/x", message: "differs", keyword: "implements_drift" }],
};

async function drafts() {
  const session = createDeviceSession(twinBridge().devices);
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

describe("the four claims of the check", () => {
  it("keeps interface, plausibility, calibration and hardware validation apart, each with its own status word", async () => {
    const { device, twin } = await drafts();
    const rows = checkRows(device, built(twin));
    const claims = rows.filter((row) => ["implements", "lint", "calibration", "hardware"].includes(row.id));
    expect(claims.map((row) => [row.text, row.status])).toEqual([
      ["Interface: the twin matches its device", "Passed"],
      ["Plausibility: readings stay plausible in a nominal day", "Passed"],
      ["Calibration: not performed", "Not performed"],
      ["Hardware validation: not run", "Not run"],
    ]);
    expect(claims.slice(2).every((row) => row.state === "neutral")).toBe(true);
    expect(claims[2]?.notes.join(" ")).toMatch(/stand-in/);
    expect(claims[3]?.notes.join(" ")).toMatch(/agrees, diverges or cannot judge/);
  });

  it("says the interface does not match when the binary says so", async () => {
    const { device, twin } = await drafts();
    const failed = checkRows(device, built(twin, { implemented: { kind: "failed", failure: refused } }));
    expect(failed.find((row) => row.id === "implements")).toMatchObject({
      state: "bad",
      status: "Failed",
      text: "Interface: the twin does not match its device",
    });
  });

  it("summarises which claims passed without implying calibration", async () => {
    const { twin } = await drafts();
    expect(checkSummary(twin)).toBeNull();
    expect(checkSummary(built(twin))).toBe(
      "Interface check passed; plausibility check passed; not calibrated; not compared with a board.",
    );
    const warned = { errors: 0, warnings: 2, findings: [] };
    expect(checkSummary(built(twin, { lint: { kind: "done", report: warned } }))).toMatch(/passed with 2 warnings/);
    expect(checkSummary(built(twin, { lint: { kind: "running" } }))).toBeNull();
  });

  it("says in the headline that a plausibility warning is not a clean pass", async () => {
    const { device, twin } = await drafts();
    const warned = { errors: 0, warnings: 1, findings: [] };
    const status = twinStatusOf(built(twin, { lint: { kind: "done", report: warned } }), device);
    expect(status.text).toContain("plausibility check passed with 1 warning;");
    expect(status.tone).toBe("ok");
  });
});

describe("a partial twin", () => {
  const unsupported = {
    pointer: "/metrics/0/scan_class",
    message: "not supported",
    keyword: "twin_member_unsupported",
    member: "scan_class",
  };
  const held = (twin: TwinDraft): TwinDraft => ({
    ...twin,
    held: [
      {
        path: "kinds/indicator.kind.json",
        name: "indicator",
        words: "w",
        issues: [
          {
            owner: "reading",
            ownerName: "level_percent",
            setting: unsupported.member,
            sentence: "The setting scan_class of reading 'level_percent' is not reproduced by a twin.",
            original: unsupported.message,
            part: "metric",
            key: "m1",
            field: "scan_class",
          },
        ],
        adding: true,
      },
    ],
  });
  const excludes =
    "This twin excludes indicator (the setting scan_class of reading 'level_percent' is not reproduced by a twin).";

  it("keeps the exclusions in the headline and the check after the twin is created", async () => {
    const { device, twin } = await drafts();
    const partial = built(held(twin));
    expect(twinStatusOf(partial, device)).toMatchObject({ text: `Partial twin created. ${excludes}`, tone: "ok" });
    const rows = checkRows(device, partial);
    expect(rows.map((row) => row.id)).toEqual([
      "device",
      "twin",
      "scope",
      "implements",
      "lint",
      "calibration",
      "hardware",
    ]);
    expect(rows[2]).toMatchObject({
      state: "warn",
      status: "Partial twin",
      text: "Some devices are left out of the twin",
      notes: [excludes],
    });
    expect(checkRows(device, built(twin)).some((row) => row.id === "scope")).toBe(false);
  });

  it("names the devices left out among the facts of the check", async () => {
    const { twin } = await drafts();
    const labels = checkFacts(held(twin)).map((entry) => [entry.label, entry.values]);
    expect(labels).toContainEqual(["Left out of the twin", [expect.stringContaining("indicator") as unknown]]);
  });
});
