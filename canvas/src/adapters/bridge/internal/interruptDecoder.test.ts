import { describe, expect, it } from "vitest";
import { interruptWire, interruptsWire, liveWire } from "../../../test/liveFixture";
import { controlReply, interruptRecord, optionalInterrupts } from "./interruptDecoder";
import { journalEntry } from "./journalDecoder";

const line = (fields: Record<string, unknown>) => journalEntry(JSON.stringify(interruptWire(fields)));
const unreadable = { kind: "live-unreadable", atNs: 10e9 };

describe("interrupt journal records", () => {
  it("reads an operator control with its scope", () => {
    expect(line({ event: "control", op: "hold", scope: { kind: "board", id: "mega-01" } })).toEqual({
      kind: "interrupt",
      atNs: 10e9,
      record: {
        event: "control",
        nodeNs: 10e9,
        clock: "world",
        op: "hold",
        scope: { kind: "board", id: "mega-01" },
        enabled: null,
        commandId: "",
        device: "",
        command: "",
        verdict: null,
        resolution: null,
        result: null,
        reason: "",
      },
    });
    expect(line({ event: "control", op: "approve_mode", scope: { kind: "all", id: "" }, enabled: true })).toMatchObject(
      {
        record: { op: "approve_mode", enabled: true, scope: { kind: "all", id: "" } },
      },
    );
  });

  it("reads the node's decisions about commands and stops", () => {
    expect(line({ event: "verdict", command_id: "cmd-7", verdict: "refuse", reason: "local_control" })).toMatchObject({
      record: { verdict: "refuse", commandId: "cmd-7", reason: "local_control" },
    });
    expect(line({ event: "resolve", command_id: "cmd-7", resolution: "forward", waiting: [] })).toMatchObject({
      record: { resolution: "forward" },
    });
    expect(line({ event: "stop", device: "fan", command: "stop_fan" })).toMatchObject({
      record: { event: "stop", device: "fan", command: "stop_fan" },
    });
    expect(line({ event: "reply", op: "hold", result: "refused", reason: "unknown_board", state: {} })).toMatchObject({
      record: { result: "refused", reason: "unknown_board" },
    });
    expect(line({ event: "inventory", devices: [] })).toMatchObject({ record: { event: "inventory" } });
  });

  it.each([
    ["event", { event: "restart" }],
    ["clock", { clock: "sundial" }],
    ["op", { op: "reset" }],
    ["result", { result: "maybe" }],
    ["verdict", { verdict: "wait" }],
    ["resolution", { resolution: "defer" }],
    ["scope kind", { scope: { kind: "plant", id: "x" } }],
    ["scope without an id", { scope: { kind: "board" } }],
    ["scope with a stray id", { scope: { kind: "all", id: "x" } }],
    ["scope id type", { scope: { kind: "board", id: 7 } }],
    ["scope shape", { scope: "board" }],
    ["enabled flag", { enabled: "yes" }],
    ["node time", { node_ns: "now" }],
  ])("drops a record with an unknown %s instead of guessing", (_, change) => {
    expect(line({ event: "control", op: "hold", scope: { kind: "all", id: "" }, ...change })).toEqual(unreadable);
  });

  it("refuses a row that is not a record", () => {
    expect(interruptRecord({})).toBeNull();
  });
});

describe("the interrupt state in a live snapshot", () => {
  const snapshot = (change: Record<string, unknown>) => journalEntry(JSON.stringify({ ...liveWire(), ...change }));

  it("reads holds, approval scopes, waiting commands and covered devices", () => {
    expect(snapshot({ interrupts: interruptsWire() })).toMatchObject({
      kind: "live",
      live: {
        interrupts: {
          holds: [{ scope: { kind: "board", id: "uno-07" }, latched: false }],
          approve: [{ kind: "all", id: "" }],
          pending: [
            { commandId: "cmd-7", device: "fan", command: "set_fan", clock: "world", deadlineNs: 1767225780e9 },
          ],
          devices: [
            { device: "sensor", mode: "hold" },
            { device: "fan", mode: "approve" },
          ],
        },
      },
    });
  });

  it("reads a snapshot without the member, or with null, as one without controls", () => {
    expect(snapshot({})).toMatchObject({ live: { interrupts: null } });
    expect(snapshot({ interrupts: null })).toMatchObject({ live: { interrupts: null } });
    expect(optionalInterrupts(undefined)).toEqual({ value: null });
  });

  it.each([
    ["not an object", "held"],
    ["without holds", { ...interruptsWire(), holds: undefined }],
    ["a hold without its latch", { ...interruptsWire(), holds: [{ kind: "all", id: "" }] }],
    ["an unknown scope kind", { ...interruptsWire(), approve: [{ kind: "plant", id: "x" }] }],
    [
      "a waiting command without an id",
      { ...interruptsWire(), pending: [{ ...interruptsWire().pending[0], command_id: "" }] },
    ],
    [
      "a waiting command on an unknown clock",
      { ...interruptsWire(), pending: [{ ...interruptsWire().pending[0], clock: "lunar" }] },
    ],
    [
      "a waiting command with a text deadline",
      { ...interruptsWire(), pending: [{ ...interruptsWire().pending[0], deadline_ns: "soon" }] },
    ],
    [
      "an unknown mode",
      { ...interruptsWire(), devices: [{ device: "fan", board: "m", mode: "frozen", latched: false }] },
    ],
    [
      "a device without a name",
      { ...interruptsWire(), devices: [{ device: "", board: "m", mode: "hold", latched: false }] },
    ],
    ["a device without its latch", { ...interruptsWire(), devices: [{ device: "fan", board: "m", mode: "hold" }] }],
  ])("drops the snapshot when the interrupt state is %s", (_, interrupts) => {
    expect(snapshot({ interrupts })).toEqual(unreadable);
  });

  it("gives states that differ only in their interrupts different keys", () => {
    const key = (entry: ReturnType<typeof snapshot>) => (entry?.kind === "live" ? entry.live.key : "");
    const plain = key(snapshot({}));
    const held = key(snapshot({ interrupts: interruptsWire() }));
    const again = key(snapshot({ interrupts: interruptsWire(), at_ns: 99e9 }));
    expect(held).not.toBe(plain);
    expect(again).toBe(held);
  });
});

describe("the node's answer to a control", () => {
  it("reads a reply, its stops and its state", () => {
    expect(
      controlReply({ result: "applied", op: "estop", stopped: ["fan"], skipped: ["led"], state: interruptsWire() }),
    ).toMatchObject({ result: "applied", op: "estop", stopped: ["fan"], skipped: ["led"], reason: "", commandId: "" });
    expect(controlReply({ result: "refused", reason: "node_busy", op: "hold", state: interruptsWire() })).toMatchObject(
      {
        result: "refused",
        reason: "node_busy",
      },
    );
  });

  it("reads a refusal that names no request, and keeps a reason it has no words for", () => {
    expect(controlReply({ result: "refused", reason: "malformed", state: interruptsWire() })).toMatchObject({
      op: null,
    });
    expect(controlReply({ result: "refused", reason: "brand_new", op: "hold", state: interruptsWire() })).toMatchObject(
      {
        reason: "brand_new",
      },
    );
  });

  it.each([
    ["not an object", 7],
    ["an unknown result", { result: "done", state: interruptsWire() }],
    ["an unknown op", { result: "applied", op: "reset", state: interruptsWire() }],
    ["a missing state", { result: "applied", op: "hold" }],
    ["stopped devices that are not names", { result: "applied", op: "estop", stopped: [1], state: interruptsWire() }],
    [
      "skipped devices that are not a list",
      { result: "applied", op: "estop", skipped: "led", state: interruptsWire() },
    ],
  ])("refuses a reply with %s", (_, value) => {
    expect(controlReply(value)).toBeNull();
  });
});
