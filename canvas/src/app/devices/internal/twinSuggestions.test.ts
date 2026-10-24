import { describe, expect, it } from "vitest";
import { chooseInstrument, pickResponse, responseOf } from "../../../model/devices";
import { cooler, coolerPath, sensorPath, twinBridge } from "../../../test/twinFixture";
import type { TwinDraft } from "./twinDraft";
import { createTwinSession } from "./twinStore";
import { choiceId, isSuggested, suggestedRefs, suggestionNoteOf } from "./twinSuggestions";

function ready(twin: ReturnType<typeof createTwinSession>): TwinDraft {
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function opened() {
  const twin = createTwinSession(twinBridge().devices, () => cooler.kind);
  await twin.open(coolerPath);
  return twin;
}

describe("which choices are still the binary's suggestion", () => {
  it("counts the instrument of every number reading and the response of every command", async () => {
    const twin = await opened();
    const refs = suggestedRefs(ready(twin));
    expect(refs.map(choiceId)).toEqual([
      `${coolerPath}|instrument|m1`,
      `${coolerPath}|instrument|m2`,
      `${coolerPath}|response|c1`,
    ]);
    expect(suggestionNoteOf(ready(twin))).toBe(
      "3 choices are still the suggested defaults: review the instrument and response of each reading.",
    );
  });

  it("stops calling a choice suggested once the person picks something else", async () => {
    const twin = await opened();
    twin.edit((choices) => chooseInstrument(choices, coolerPath, "m1", "streamsim/thermistor@1"));
    const ids = suggestedRefs(ready(twin)).map(choiceId);
    expect(ids).not.toContain(`${coolerPath}|instrument|m1`);
    expect(ids).toHaveLength(2);
    const command = cooler.kind.commands[0];
    if (command === undefined) throw new Error("no command");
    twin.edit((choices) =>
      pickResponse(choices, { member: coolerPath, kind: cooler.kind, command }, responseOf("none")),
    );
    expect(isSuggested(ready(twin), { path: coolerPath, key: "c1", field: "response" })).toBe(false);
  });

  it("keeps a choice suggested when the person picks the same value, until it is reviewed", async () => {
    const twin = await opened();
    twin.edit((choices) => chooseInstrument(choices, coolerPath, "m1", "none"));
    const ref = { path: coolerPath, key: "m1", field: "instrument" } as const;
    expect(isSuggested(ready(twin), ref)).toBe(true);
    const before = ready(twin);
    twin.review([choiceId(ref)]);
    expect(ready(twin)).toMatchObject({ edited: before.edited, touched: before.touched, doc: before.doc });
    expect(isSuggested(ready(twin), ref)).toBe(false);
    expect(suggestedRefs(ready(twin))).toHaveLength(2);
    twin.review(suggestedRefs(ready(twin)).map(choiceId));
    expect(suggestionNoteOf(ready(twin))).toBeNull();
  });

  it("counts the choices of a device that joins the twin, and nothing when no suggestion is known", async () => {
    const twin = await opened();
    await twin.include(sensorPath);
    expect(suggestedRefs(ready(twin)).map(choiceId)).toContain(`${sensorPath}|instrument|m1`);
    const unknown = { ...ready(twin), suggested: null };
    expect(suggestedRefs(unknown)).toEqual([]);
  });
});
