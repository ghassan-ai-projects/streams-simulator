import { describe, expect, it } from "vitest";
import { heldIssues } from "../../../model/devices";
import { cooler, coolerPath, kindOf, sensorPath, twinBridge } from "../../../test/twinFixture";
import { createTwinSession } from "./twinStore";
import type { Held, TwinDraft } from "./twinDraft";
import { partialOf, scopeOf } from "./twinScope";

const indicator = kindOf({
  kind: "indicator",
  metrics: [{ name: "level_percent", data_type: "number", unit: "%", scan_class: "fast" }],
  commands: [],
});
const unsupported = {
  pointer: "/metrics/0/scan_class",
  message: "not supported",
  keyword: "twin_member_unsupported",
  member: "scan_class",
};

const indicatorHeld: Held = {
  path: "kinds/indicator.kind.json",
  name: "indicator",
  words: "indicator is not ready for a twin: A twin does not reproduce the setting scan_class",
  issues: heldIssues(indicator, [unsupported]),
  adding: true,
};

async function opened(): Promise<TwinDraft> {
  const twin = createTwinSession(twinBridge().devices, () => cooler.kind);
  await twin.open(coolerPath);
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

describe("what goes into the twin and what stays out", () => {
  it("lists the devices in the twin, with nothing excluded or blocked when all are ready", async () => {
    const draft = await opened();
    expect(scopeOf(draft)).toEqual({ included: ["cooler"], excluded: [], blocked: [] });
    expect(partialOf(draft)).toBe("");
  });

  it("names a device that is waiting out of the twin with the setting that keeps it out", async () => {
    const draft = { ...(await opened()), held: [indicatorHeld] };
    expect(scopeOf(draft).excluded).toEqual([
      {
        name: "indicator",
        reason: "the setting scan_class of reading 'level_percent' is not reproduced by a twin",
      },
    ]);
    expect(partialOf(draft)).toBe(
      "This twin excludes indicator (the setting scan_class of reading 'level_percent' is not reproduced by a twin).",
    );
  });

  it("separates a device that blocks the twin, and says a device that cannot be read in plain words", async () => {
    const unreadable: Held = {
      path: sensorPath,
      name: "sensor",
      words: "sensor could not be read",
      issues: [],
      adding: true,
    };
    const draft = { ...(await opened()), held: [{ ...indicatorHeld, adding: false }, unreadable] };
    expect(scopeOf(draft).blocked.map((entry) => entry.name)).toEqual(["indicator"]);
    expect(scopeOf(draft).excluded).toEqual([{ name: "sensor", reason: "could not be read" }]);
  });
});
