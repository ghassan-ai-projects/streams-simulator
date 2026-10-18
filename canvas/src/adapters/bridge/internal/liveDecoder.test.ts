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

  it.each([
    ["host", { host: "lost" }],
    ["gateway", { gateway: "lost" }],
    ["device list", { devices: {} }],
    ["mark kind", { marks: [{ at_ns: 1, kind: "note", label: "x", entity: "" }] }],
    ["mark row", { marks: ["x"] }],
  ])("rejects an unknown %s", (_, change) => {
    expect(journalEntry(JSON.stringify({ ...liveWire(), ...change }))).toBeNull();
  });

  it.each([
    ["environment", { environment: "cloud" }],
    ["health", { health: "asleep" }],
    ["readings", { readings: "none" }],
    ["reading state", { readings: [{ name: "t", unit: "", value: 1, quality: "good", state: "fine" }] }],
  ])("rejects a device with an unknown %s", (_, change) => {
    const wire = liveWire();
    expect(journalEntry(JSON.stringify({ ...wire, devices: [{ ...wire.devices[0], ...change }] }))).toBeNull();
  });

  it("rejects a command that was never sent", () => {
    const wire = liveWire();
    expect(journalEntry(JSON.stringify({ ...wire, commands: [{ ...wire.commands[0], sent: false }] }))).toBeNull();
  });

  it("keeps text and boolean readings and reads explicit truncation", () => {
    const wire = liveWire();
    const readings = [
      { name: "mode", unit: "", value: "auto", quality: "good", state: "good" },
      { name: "fan", unit: "", value: true, quality: "stale", state: "stale" },
    ];
    const entry = journalEntry(
      JSON.stringify({ ...wire, commands_truncated: true, devices: [{ ...wire.devices[0], readings }] }),
    );
    expect(entry).toMatchObject({
      kind: "live",
      live: { commandsTruncated: true, devices: [{ readings: [{ value: "auto" }, { value: true, state: "stale" }] }] },
    });
  });

  it("reads the recorded session plan's simulated instances and refuses a session without a plan", () => {
    const envelope = { type: "session", stream: "observer", seq: 2, at_ns: 0, mode: "connected" };
    const plan = {
      instances: [
        { instance: "fan-01", origin: "simulated" },
        { instance: "zone-01", origin: "real" },
      ],
    };
    expect(journalEntry(JSON.stringify({ ...envelope, plan }))).toEqual({
      kind: "session",
      atNs: 0,
      simulated: ["fan-01"],
    });
    expect(journalEntry(JSON.stringify({ ...envelope, plan: {} }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...envelope, mode: "evaluate", plan }))).toBeNull();
  });
});
