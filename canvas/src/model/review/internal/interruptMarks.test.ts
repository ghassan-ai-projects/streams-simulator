import { describe, expect, it } from "vitest";
import { recordOf } from "../../../test/interruptFixture";
import { interruptMarks } from "./interruptMarks";

describe("interrupt marks on the timeline", () => {
  it("marks what the operator asked, once, in words", () => {
    const [mark] = interruptMarks(recordOf(), 12e9);
    expect(mark).toEqual({
      atNs: 12e9,
      label: "Operator: Hold on board mega-01",
      raw: "control hold",
      kind: "interrupt",
      tone: "other",
      entity: "",
      untilNs: null,
    });
    const labels = (record: Parameters<typeof recordOf>[0]) => interruptMarks(recordOf(record), 1).map((m) => m.label);
    expect(labels({ op: "approve_mode", enabled: false })).toEqual([
      "Operator: Stop approving each command on board mega-01",
    ]);
    expect(labels({ op: "deny", scope: null, commandId: "cmd-7" })).toEqual(["Operator: Deny cmd-7"]);
    expect(labels({ op: "estop", scope: { kind: "all", id: "" } })).toEqual([
      "Operator: Emergency stop on all devices",
    ]);
  });

  it("marks nothing for a read of the state or a request the record cannot name", () => {
    expect(interruptMarks(recordOf({ op: "state", scope: null }), 1)).toEqual([]);
    expect(interruptMarks(recordOf({ op: "hold", scope: null }), 1)).toEqual([]);
    expect(interruptMarks(recordOf({ op: "approve_mode", enabled: null }), 1)).toEqual([]);
    expect(interruptMarks(recordOf({ op: "approve", scope: null, commandId: "" }), 1)).toEqual([]);
    expect(interruptMarks(recordOf({ op: null }), 1)).toEqual([]);
  });

  it("marks a refused control, an approval handed on, a refusal at the node and a safe stop", () => {
    const refused = recordOf({ event: "reply", result: "refused", reason: "unknown_device", op: "hold" });
    expect(interruptMarks(refused, 1).map((m) => m.label)).toEqual(["Control refused: The node has no such device"]);
    expect(interruptMarks(recordOf({ event: "reply", result: "applied" }), 1)).toEqual([]);
    expect(
      interruptMarks(recordOf({ event: "reply", result: "refused", op: null, reason: "malformed" }), 1),
    ).toHaveLength(1);
    const forward = recordOf({ event: "resolve", resolution: "forward", commandId: "cmd-7" });
    expect(interruptMarks(forward, 1).map((m) => m.label)).toEqual(["cmd-7 approved and handed on"]);
    const denied = recordOf({ event: "resolve", resolution: "refuse", commandId: "cmd-7", reason: "local_control" });
    expect(interruptMarks(denied, 1).map((m) => m.label)).toEqual([
      "cmd-7 refused: Refused at the node: local control",
    ]);
    const purged = recordOf({ event: "purge", commandId: "c-1", device: "fan", reason: "local_control" });
    expect(interruptMarks(purged, 1)).toMatchObject([
      { label: "c-1 refused: Refused at the node: local control", raw: "purge c-1 local_control" },
    ]);
    const stop = recordOf({ event: "stop", device: "fan", command: "stop_fan" });
    expect(interruptMarks(stop, 1)).toMatchObject([{ label: "Safe stop sent to fan", raw: "stop fan stop_fan" }]);
  });

  it("leaves the node's own bookkeeping records off the timeline", () => {
    for (const event of ["inventory", "command", "clock", "flush", "verdict", "effect"] as const) {
      expect(interruptMarks(recordOf({ event }), 1)).toEqual([]);
    }
  });
});
