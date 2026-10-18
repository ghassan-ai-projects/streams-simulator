import { describe, expect, it } from "vitest";
import { liveWire } from "../../../test/liveFixture";
import { journalEntry } from "./journalDecoder";

describe("binary session evidence", () => {
  it("preserves independent stages without turning a board report into an observed effect", () => {
    expect(journalEntry(JSON.stringify(liveWire()))).toMatchObject({
      kind: "live",
      live: {
        host: "online",
        gateway: "online",
        devices: [{ bootId: "61", readings: [{ value: 28, state: "good" }] }],
        commands: [{ receipt: "accepted", report: "succeeded", effect: "unknown" }],
      },
    });
  });

  it.each(["receipt", "report", "effect"])("rejects an unknown %s stage", (field) => {
    const wire = liveWire();
    const command = wire.commands[0];
    expect(journalEntry(JSON.stringify({ ...wire, commands: [{ ...command, [field]: "done" }] }))).toBeNull();
  });

  it("rejects a non-scalar good reading and accepts it only with explicit invalid quality", () => {
    const wire = liveWire();
    const device = wire.devices[0];
    const reading = { name: "temperature_c", unit: "degC", value: {}, quality: "bad", state: "good" };
    const invalid = { ...wire, devices: [{ ...device, readings: [reading] }] };
    expect(journalEntry(JSON.stringify(invalid))).toBeNull();
    expect(
      journalEntry(JSON.stringify({ ...wire, devices: [{ ...device, readings: [{ ...reading, state: "invalid" }] }] })),
    ).toMatchObject({ kind: "live", live: { devices: [{ readings: [{ value: null, state: "invalid" }] }] } });
  });
});
