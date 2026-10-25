import { describe, expect, it } from "vitest";
import { chooseInstrument, renameKind, serialiseChoices } from "../../../model/devices";
import type { DeviceFailure } from "../../ports";
import { coolerDocument, coolerPath, thermistor, twinBridge } from "../../../test/twinFixture";
import { checkFacts } from "./checkFacts";
import { createDeviceSession } from "./deviceStore";
import { stepStatuses, twinReadiness } from "./stepStatus";
import type { TwinDraft } from "./twinDraft";
import { interfaceNeedsOf, rangeNeedsOf } from "./twinNeeds";
import { createBlock } from "./twinSelectors";
import { twinStatusOf } from "./twinStatus";

const refused: DeviceFailure = {
  kind: "refused",
  diagnostic: { message: "refused", displayCode: "kind_failed" },
  issues: [{ pointer: "/x", message: "m" }],
};

async function opened(withRange: boolean) {
  const fake = twinBridge();
  const temperature = { name: "temperature", data_type: "number", unit: "Cel", resolution: 0.1 };
  const ranged = { ...temperature, eu_range: { min: 0, max: 100 } };
  const metrics = [withRange ? ranged : temperature, ...coolerDocument.metrics.slice(1)];
  fake.addFile(coolerPath, { ...coolerDocument, metrics }, "k1");
  const session = createDeviceSession(fake.devices);
  await session.open(coolerPath);
  const device = session.snapshot();
  const twin = session.twin.snapshot();
  if (device.phase !== "ready" || twin.phase !== "ready") throw new Error("not ready");
  return { session, device, twin };
}

function chosen(twin: TwinDraft, instrument: string): TwinDraft {
  const doc = chooseInstrument(twin.doc, coolerPath, "m1", instrument);
  return { ...twin, doc, touched: true };
}

describe("what the chosen instruments say the interface lacks", () => {
  it("asks for the range of a reading only when the chosen instrument says it needs one", async () => {
    const { device, twin } = await opened(false);
    expect(rangeNeedsOf(twin)).toEqual([]);
    const picked = chosen(twin, thermistor.ref);
    expect(rangeNeedsOf(picked)).toEqual([
      expect.objectContaining({ part: "metric", key: "m1", field: "eu_range", twin: true }) as unknown,
    ]);
    expect(interfaceNeedsOf(device, picked)).toHaveLength(1);
    expect(interfaceNeedsOf(device, chosen(twin, "none"))).toEqual([]);
    expect(rangeNeedsOf({ phase: "idle" })).toEqual([]);
    expect(twinReadiness(device)).toEqual({ text: "Ready for a twin", tone: "ok" });
    expect(twinReadiness(device, 1)).toEqual({ text: "Needed for a twin: 1", tone: "warn" });
  });

  it("asks for nothing when the reading states its range", async () => {
    const { device, twin } = await opened(true);
    expect(interfaceNeedsOf(device, chosen(twin, thermistor.ref))).toEqual([]);
  });

  it("blocks Create, says why in the banner, and shows the same in the step badge", async () => {
    const { device, twin } = await opened(false);
    const picked = chosen(twin, thermistor.ref);
    const why = "Set what the twin needs in the Interface step first.";
    expect(createBlock(picked, device)).toBe(why);
    expect(twinStatusOf(picked, device)).toMatchObject({ text: why, tone: "warn" });
    expect(stepStatuses(device, picked).behaviour).toEqual({ text: "Needs attention", tone: "warn" });
    expect(stepStatuses(device, chosen(twin, "none")).behaviour).toEqual({ text: "Ready", tone: "" });
  });
});

describe("the badge of the behaviour step agrees with its banner", () => {
  it("is never Ready while the banner asks for something to be fixed, and is Ready again once it is", async () => {
    const { device, twin } = await opened(true);
    const touched = { ...twin, touched: true };
    const badge = (draft: TwinDraft, over = device) => stepStatuses(over, draft).behaviour;
    const banner = (draft: TwinDraft, over = device) => twinStatusOf(draft, over);
    expect(badge(touched)).toEqual({ text: "Ready", tone: "" });
    expect(banner(touched).tone).toBe("");
    const dirty = { ...device, doc: renameKind(device.doc, "other") };
    expect(banner(touched, dirty).tone).toBe("warn");
    expect(badge(touched, dirty)).toEqual({ text: "Needs attention", tone: "warn" });
    const issue = { scope: "elsewhere" as const, kind: "", key: "", field: "", spot: "", pointer: "", keyword: "" };
    const refusedDraft: TwinDraft = {
      ...touched,
      build: { kind: "refused", failure: refused },
      issues: [{ ...issue, message: "m", text: "m" }],
    };
    expect(banner(refusedDraft).tone).toBe("warn");
    expect(badge(refusedDraft)).toEqual({ text: "Needs attention", tone: "warn" });
    const fixed: TwinDraft = { ...refusedDraft, issues: [] };
    expect(banner(fixed).tone).toBe("");
    expect(badge(fixed)).toEqual({ text: "Ready", tone: "" });
  });

  it("says Needs choices for a form that is not finished, and Needs attention for a device that is held", async () => {
    const { device, twin } = await opened(true);
    const touched = { ...twin, touched: true };
    const unfinished = { ...touched, doc: { ...touched.doc, title: "" } };
    expect(stepStatuses(device, unfinished).behaviour).toEqual({ text: "Needs choices", tone: "warn" });
    const held = { path: coolerPath, name: "pump", words: "pump still needs a unit", issues: [], adding: false };
    const blocked = { ...touched, held: [held] };
    expect(stepStatuses(device, blocked).behaviour).toEqual({ text: "Needs attention", tone: "warn" });
    expect(twinStatusOf(blocked, device).text).toBe("pump still needs a unit.");
    const waiting = { ...touched, held: [{ ...held, adding: true }] };
    expect(stepStatuses(device, waiting).behaviour).toEqual({ text: "Ready", tone: "" });
  });
});

describe("the facts of the check", () => {
  it("name the choices file, the readings and commands, and the instrument of each reading", async () => {
    const { twin } = await opened(true);
    expect(checkFacts(chosen(twin, thermistor.ref))).toEqual([
      { label: "Choices file", values: ["twins/cooler.twin.json"] },
      { label: "Readings and commands", values: ["3 readings · 2 commands"] },
      {
        label: "Instruments",
        values: [
          "temperature: Thermistor · -40 to 125 · measured in degC, converted (your choice)",
          "duty: No instrument: a plain reading (suggested)",
        ],
      },
    ]);
  });

  it("add the plant that was made, and say so when no reading has an instrument", async () => {
    const { twin } = await opened(true);
    const made = { id: "t", digest: "d", channels: 3, faults: 2, effectors: 1, implements: ["cooler"] };
    const built: TwinDraft = {
      ...twin,
      build: {
        kind: "built",
        made,
        path: "domains/t.domain.json",
        fingerprint: serialiseChoices(twin.doc, twin.members),
        implemented: { kind: "passed", checked: { id: "t", digest: "d" } },
        lint: { kind: "unavailable" },
      },
    };
    const facts = checkFacts(built);
    expect(facts.map((fact) => fact.label)).toEqual([
      "Choices file",
      "Readings and commands",
      "Instruments",
      "Plant file",
      "Digest",
      "Parts",
      "Devices",
    ]);
    expect(facts.find((fact) => fact.label === "Plant file")?.values).toEqual(["domains/t.domain.json"]);
    const without = checkFacts({ ...twin, doc: { ...twin.doc, kinds: [] } });
    expect(without.find((fact) => fact.label === "Instruments")?.values).toEqual([
      "None: every reading is a plain reading",
    ]);
  });
});
