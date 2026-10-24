import { describe, expect, it } from "vitest";
import { cooler, coolerPath, twinBridge } from "../../../test/twinFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { success } from "../../../test/outcome";
import { checkRows } from "./checkRows";
import { createDeviceSession } from "./deviceStore";
import type { TwinDraft } from "./twinDraft";
import { stepStatuses } from "./stepStatus";
import { scopeOf } from "./twinScope";
import { canCreate } from "./twinSelectors";
import { twinStatusOf } from "./twinStatus";

const indicatorPath = "kinds/indicator.kind.json";
const indicator = {
  kind: "indicator",
  metrics: [
    {
      name: "level_percent",
      data_type: "number",
      unit: "%",
      eu_range: { min: 0, max: 100 },
      resolution: 1,
      scan_class: "fast",
    },
  ],
  commands: [],
};
const excludes =
  "This twin excludes indicator (the setting scan_class of reading 'level_percent' is not reproduced by a twin).";

function ready(state: ReturnType<ReturnType<typeof createDeviceSession>["twin"]["snapshot"]>): TwinDraft {
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function withIndicator() {
  const fake = twinBridge();
  fake.addFile(indicatorPath, indicator, "k5");
  const listing = fake.workspaceOutcome;
  if (!listing.ok) throw new Error("no listing");
  const kinds = [...listing.value.kinds, { path: indicatorPath, kind: "indicator", title: "" }];
  fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
  const session = createDeviceSession(fake.devices);
  await session.open(coolerPath);
  await session.twin.include(indicatorPath);
  return { session, fake };
}

describe("a twin that leaves a device out", () => {
  it("names the setting that keeps the device out, for every issue of the device", async () => {
    const { session } = await withIndicator();
    const [held] = ready(session.twin.snapshot()).held;
    expect(held).toMatchObject({ name: "indicator", adding: true });
    expect(held?.issues).toHaveLength(1);
    expect(held?.issues[0]).toMatchObject({
      owner: "reading",
      ownerName: "level_percent",
      setting: "scan_class",
      field: "scan_class",
      original: "is not part of the interface a twin reproduces",
    });
    expect(held?.words).toBe("indicator is not ready for a twin: A twin does not reproduce the setting scan_class");
  });

  it("can still be created, and keeps saying which device it excludes", async () => {
    const { session } = await withIndicator();
    const device = session.snapshot();
    expect(canCreate(ready(session.twin.snapshot()), device)).toBe(true);
    await session.twin.create();
    const draft = ready(session.twin.snapshot());
    expect(draft.members.map((member) => member.path)).toEqual([cooler.path]);
    expect(scopeOf(draft).excluded.map((entry) => entry.name)).toEqual(["indicator"]);
    expect(twinStatusOf(draft, device)).toMatchObject({ text: `Partial twin created. ${excludes}`, tone: "ok" });
    expect(stepStatuses(device, draft).behaviour).toEqual({ text: "Created (partial)", tone: "ok" });
    const scope = checkRows(device, draft).find((row) => row.id === "scope");
    expect(scope).toMatchObject({ status: "Partial twin", notes: [excludes] });
  });

  it("stops saying so once the device is left out on purpose", async () => {
    const { session } = await withIndicator();
    session.twin.leave(indicatorPath);
    expect(scopeOf(ready(session.twin.snapshot())).excluded).toEqual([]);
  });
});
