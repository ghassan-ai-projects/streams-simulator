import { describe, expect, it } from "vitest";
import { busyState, emptyState, fanWaits, replyOf, sensorWaits } from "../../../test/interruptFixture";
import {
  argsWords,
  arrivalWords,
  badgeKind,
  deadlineWords,
  modeLine,
  outcomeLine,
  pendingTitle,
  requestWords,
  waitingWords,
} from "./interruptWords";
import { scopeKey, scopeLabel } from "./interruptScope";
import type { ActionRequest, ControlScope, WordedRequest } from "./interrupts";

const board: ControlScope = { kind: "board", id: "mega-01" };
const everything: ControlScope = { kind: "all", id: "" };
const digest = `sha256:${"a".repeat(64)}`;

describe("interrupt words", () => {
  it.each<[WordedRequest, string]>([
    [{ op: "hold", scope: board }, "Hold on board mega-01"],
    [{ op: "release", scope: everything }, "Release of all devices"],
    [{ op: "estop", scope: { kind: "device", id: "fan" } }, "Emergency stop on device fan"],
    [{ op: "approve_mode", scope: board, enabled: true }, "Approve each command on board mega-01"],
    [{ op: "approve_mode", scope: board, enabled: false }, "Stop approving each command on board mega-01"],
    [{ op: "approve", commandId: "cmd-7" }, "Approve cmd-7"],
    [{ op: "deny", commandId: "cmd-7" }, "Deny cmd-7"],
  ])("words the request %j", (request, words) => {
    expect(requestWords(request)).toBe(words);
  });

  it("names a scope for a person and for a key", () => {
    expect(scopeLabel(everything)).toBe("All devices");
    expect(scopeLabel(board)).toBe("Board mega-01");
    expect(scopeKey(board)).toBe("board:mega-01");
    expect(scopeKey(everything)).toBe("all:");
  });

  it("tells a hold, a latch and an approval mode apart by words and by kind", () => {
    expect(modeLine({ mode: "hold", latched: false })).toBe("Held");
    expect(modeLine({ mode: "hold", latched: true })).toBe("Held · latched by emergency stop");
    expect(modeLine({ mode: "approve", latched: false })).toBe("Approving each command");
    expect(badgeKind({ mode: "hold", latched: true })).toBe("latched");
    expect(badgeKind({ mode: "hold", latched: false })).toBe("hold");
    expect(badgeKind({ mode: "approve", latched: true })).toBe("approve");
  });

  it("says what the node answered, in plain words, with the refusal reason from the one table", () => {
    const hold: ActionRequest = { op: "hold", scope: board };
    expect(outcomeLine(hold, { kind: "reply", reply: replyOf() }, "none")).toEqual({
      tone: "ok",
      text: "Hold on board mega-01 was applied.",
    });
    const refused = replyOf({ result: "refused", reason: "unknown_board" });
    expect(outcomeLine(hold, { kind: "reply", reply: refused }, "none")).toEqual({
      tone: "refused",
      text: "Hold on board mega-01 was refused: The node has no such board.",
    });
  });

  it("words a reason it has no words for without reading an inherited property of the table", () => {
    const hold: ActionRequest = { op: "hold", scope: board };
    for (const reason of ["constructor", "toString"]) {
      const reply = replyOf({ result: "refused", reason });
      expect(outcomeLine(hold, { kind: "reply", reply }, "none").text).toBe(
        `Hold on board mega-01 was refused: ${reason}.`,
      );
    }
    const underscored = replyOf({ result: "refused", reason: "__proto__" });
    expect(outcomeLine(hold, { kind: "reply", reply: underscored }, "none").text).toBe(
      "Hold on board mega-01 was refused:   proto  .",
    );
  });

  it("states what an emergency stop did and that the scope stays held", () => {
    const stop: ActionRequest = { op: "estop", scope: everything };
    const reply = replyOf({ op: "estop", stopped: ["fan", "led"], skipped: ["sensor", "probe"] });
    expect(outcomeLine(stop, { kind: "reply", reply }, "none").text).toBe(
      "Emergency stop on all devices was applied. Safe stop sent to fan, led. sensor, probe declare no safe stop. The scope stays held until it is released.",
    );
    expect(outcomeLine(stop, { kind: "reply", reply: replyOf({ op: "estop" }) }, "none").text).toContain(
      "No safe stop was sent.",
    );
    const one = replyOf({ op: "estop", stopped: ["fan"], skipped: ["sensor"] });
    expect(outcomeLine(stop, { kind: "reply", reply: one }, "none").text).toContain("sensor declares no safe stop.");
  });

  it("says that the stop could not be sent to a device it could not reach, as a failure", () => {
    const stop: ActionRequest = { op: "estop", scope: everything };
    const reply = replyOf({ op: "estop", stopped: ["fan"], unreachable: ["pump", "valve"] });
    const line = outcomeLine(stop, { kind: "reply", reply }, "none");
    expect(line.tone).toBe("failed");
    expect(line.text).toContain("The stop could NOT be sent to pump, valve: they may still be energised.");
    expect(line.text).toContain("Safe stop sent to fan.");
  });

  it("words a release from the state the node returned, naming the holds that remain", () => {
    const release = (scope: ControlScope, reply = replyOf({ op: "release" })) =>
      outcomeLine({ op: "release", scope }, { kind: "reply", reply }, "none");
    expect(release(everything)).toEqual({
      tone: "ok",
      text: "Released all devices. 2 narrower holds remain: board uno-07, device fan.",
    });
    const clear = replyOf({ op: "release", state: emptyState });
    expect(release(everything, clear).text).toBe("Released all devices.");
    const oneLeft = replyOf({ op: "release", state: { ...busyState, holds: [busyState.holds[0] ?? never()] } });
    expect(release(everything, oneLeft).text).toBe("Released all devices. 1 narrower hold remains: board uno-07.");
    expect(release({ kind: "device", id: "sensor" }).text).toBe(
      "Released device sensor. Device sensor is still held through board uno-07.",
    );
    expect(
      release({ kind: "board", id: "uno-07" }, replyOf({ op: "release", state: { ...busyState, holds: [] } })).text,
    ).toBe("Released board uno-07.");
  });

  it("shows a failure plainly and says what was done to read the state again", () => {
    const hold: ActionRequest = { op: "hold", scope: board };
    const failed = (failure: "timeout" | "gone" | "unavailable" | "no-session" | "rejected", detail = "") =>
      ({ kind: "failed", failure, detail }) as const;
    const timeout = outcomeLine(hold, failed("timeout"), "pending");
    expect(timeout.tone).toBe("failed");
    expect(timeout.text).toBe(
      "Hold on board mega-01: The node did not answer in time. The control may still apply afterwards. Reading the node's state again…",
    );
    expect(outcomeLine(hold, failed("gone"), "read").text).toMatch(/read from the node just now\.$/);
    expect(outcomeLine(hold, failed("unavailable"), "failed").text).toMatch(/may be out of date\.$/);
    expect(outcomeLine(hold, failed("no-session"), "none").text).toMatch(/nothing was sent\.$/);
    const busy = replyOf({ result: "refused", reason: "node_busy" });
    expect(outcomeLine(hold, { kind: "reply", reply: busy }, "pending").text).toContain("may still apply this later");
  });

  it("does not call a request the bridge refused an unreadable answer, and passes on the bridge's words", () => {
    const hold: ActionRequest = { op: "hold", scope: board };
    const line = outcomeLine(hold, { kind: "failed", failure: "rejected", detail: "unknown action" }, "none");
    expect(line.text).toBe(
      "Hold on board mega-01: The bridge refused the request, so nothing was confirmed. (unknown action)",
    );
    expect(line.text).not.toContain("could not be read");
  });

  it("says plainly that a control was not sent because another is still waiting", () => {
    const approve: ActionRequest = { op: "approve", commandId: "cmd-7", contentDigest: digest };
    expect(outcomeLine(approve, { kind: "not-sent" }, "none")).toEqual({
      tone: "refused",
      text: "Approve cmd-7: not sent: another control is still waiting.",
    });
  });

  it("counts the commands that wait and shows their deadline from the binary's numbers, as an estimate", () => {
    expect(waitingWords(0)).toBe("No commands wait for approval.");
    expect(waitingWords(1)).toBe("1 command waiting for approval");
    expect(waitingWords(2)).toBe("2 commands waiting for approval");
    expect(pendingTitle(fanWaits)).toBe("set_cooling on fan");
    expect(deadlineWords(fanWaits, { worldNs: 100e9, wallMs: 0 })).toBe("expires in about 30 s (run time)");
    expect(deadlineWords(fanWaits, { worldNs: 130e9, wallMs: 0 })).toBe(
      "deadline reached; the node settles it at its next step",
    );
    expect(deadlineWords(sensorWaits, { worldNs: 0, wallMs: 1_700_000_000_000 })).toBe("expires in about 45 s");
    expect(deadlineWords(sensorWaits, { worldNs: 0, wallMs: 1_700_000_100_000 })).toBe(
      "deadline reached; the node settles it at its next step",
    );
  });

  it("shows the arguments of a waiting command, bounded, and says when the binary left them out", () => {
    expect(argsWords(fanWaits)).toBe("level_percent: 30");
    expect(argsWords(sensorWaits)).toBe("no arguments");
    expect(argsWords({ ...sensorWaits, argsTruncated: true })).toBe("arguments too large to show");
    const many = Array.from({ length: 8 }, (_, index) => ({ name: `k${String(index)}`, value: "v" }));
    expect(argsWords({ ...sensorWaits, args: many })).toBe("k0: v, k1: v, k2: v, k3: v, k4: v, k5: v and 2 more");
    const long = { name: "note", value: "x".repeat(60) };
    expect(argsWords({ ...sensorWaits, args: [long] })).toBe(`note: ${"x".repeat(40)}…`);
  });

  it("announces a command that arrived in words", () => {
    expect(arrivalWords(fanWaits)).toBe("A command is waiting for approval: set_cooling on fan (cmd-7).");
  });
});

function never(): never {
  throw new Error("the fixture holds this hold");
}
