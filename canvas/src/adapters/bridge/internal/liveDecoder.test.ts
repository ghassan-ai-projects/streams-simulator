import { describe, expect, it } from "vitest";
import { benchGatewayLost, benchWarmingUp } from "../../../test/benchLiveFixture";
import { liveWire } from "../../../test/liveFixture";
import { journalEntry } from "./journalDecoder";

const unreadable = { kind: "live-unreadable", atNs: 10e9 };

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
    expect(journalEntry(JSON.stringify({ ...wire, commands: [{ ...command, [field]: "done" }] }))).toEqual(unreadable);
  });

  it("rejects a non-scalar good reading and accepts it only with explicit invalid quality", () => {
    const wire = liveWire();
    const device = wire.devices[0];
    const reading = { name: "temperature_c", unit: "degC", value: {}, quality: "bad", state: "good" };
    const invalid = { ...wire, devices: [{ ...device, readings: [reading] }] };
    expect(journalEntry(JSON.stringify(invalid))).toEqual(unreadable);
    expect(
      journalEntry(JSON.stringify({ ...wire, devices: [{ ...device, readings: [{ ...reading, state: "invalid" }] }] })),
    ).toMatchObject({ kind: "live", live: { devices: [{ readings: [{ value: null, state: "invalid" }] }] } });
  });

  it.each([
    ["host", { host: "lost" }],
    ["gateway", { gateway: "lost" }],
    ["device list", { devices: {} }],
    ["mark kind", { marks: [{ at_ns: 1, kind: "note", subject: "host", label: "x", entity: "" }] }],
    ["mark subject", { marks: [{ at_ns: 1, kind: "connection", subject: "board", label: "x", entity: "" }] }],
    ["mark row", { marks: ["x"] }],
  ])("rejects an unknown %s", (_, change) => {
    expect(journalEntry(JSON.stringify({ ...liveWire(), ...change }))).toEqual(unreadable);
  });

  it.each([
    ["environment", { environment: "cloud" }],
    ["health", { health: "asleep" }],
    ["readings", { readings: "none" }],
    ["reading state", { readings: [{ name: "t", unit: "", value: 1, quality: "good", state: "fine" }] }],
  ])("rejects a device with an unknown %s", (_, change) => {
    const wire = liveWire();
    expect(journalEntry(JSON.stringify({ ...wire, devices: [{ ...wire.devices[0], ...change }] }))).toEqual(unreadable);
  });

  it.each([
    ["never sent", { sent: false }],
    ["without a side", { environment: "elsewhere" }],
    ["observed while refused", { effect: "observed", receipt: "refused", effect_evidence: "Host checked flow" }],
    ["observed without a board report", { effect: "observed", report: "pending", effect_evidence: "Host checked" }],
    ["observed without evidence", { effect: "observed" }],
  ])("marks a command %s as unreadable instead of showing it", (_, change) => {
    const wire = liveWire();
    expect(journalEntry(JSON.stringify({ ...wire, commands: [{ ...wire.commands[0], ...change }] }))).toEqual(
      unreadable,
    );
  });

  it("reads a host observation that follows acceptance and a successful board report", () => {
    const wire = liveWire();
    const observed = { ...wire.commands[0], effect: "observed", effect_evidence: "Host checked flow: observed" };
    expect(journalEntry(JSON.stringify({ ...wire, commands: [observed] }))).toMatchObject({
      live: { commands: [{ effect: "observed", environment: "simulated" }] },
    });
  });

  it("gives identical states one key, whatever their time and marks", () => {
    const first = journalEntry(JSON.stringify(liveWire(10e9)));
    const later = journalEntry(JSON.stringify({ ...liveWire(11e9), marks: [] }));
    const changed = journalEntry(JSON.stringify({ ...liveWire(12e9), host: "offline" }));
    const key = (entry: typeof first) => (entry?.kind === "live" ? entry.live.key : "");
    expect(key(first)).not.toBe("");
    expect(key(later)).toBe(key(first));
    expect(key(changed)).not.toBe(key(first));
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
    expect(journalEntry(JSON.stringify({ ...envelope, plan: { ...plan, domain: "arduino-thermal" } }))).toEqual({
      kind: "session",
      atNs: 0,
      domain: "arduino-thermal",
      simulated: ["fan-01"],
    });
    expect(journalEntry(JSON.stringify({ ...envelope, plan: {} }))).toBeNull();
    expect(journalEntry(JSON.stringify({ ...envelope, mode: "evaluate", plan }))).toBeNull();
  });

  it("reads the bench's own records: warming-up and simulated readings, and a lost gateway with no marks", () => {
    expect(journalEntry(benchWarmingUp)).toMatchObject({
      kind: "live",
      live: {
        marks: [],
        devices: [
          { health: "simulated", readings: [{ value: 0, state: "simulated" }] },
          { health: "simulated", readings: [{ value: null, state: "untrusted", quality: "warming_up" }, {}] },
          { environment: "hardware", health: "online", readings: [{ value: null, state: "untrusted" }, {}] },
        ],
      },
    });
    expect(journalEntry(benchGatewayLost)).toMatchObject({
      kind: "live",
      live: {
        gateway: "offline",
        marks: [
          { subject: "gateway", label: "Gateway link offline" },
          { subject: "device", entity: "zone-01" },
        ],
        devices: [{}, {}, { health: "offline", readings: [{ state: "disconnected" }, { state: "disconnected" }] }],
      },
    });
  });
});
