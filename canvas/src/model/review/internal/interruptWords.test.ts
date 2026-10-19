import { describe, expect, it } from "vitest";
import { fanWaits, replyOf, sensorWaits } from "../../../test/interruptFixture";
import {
  deadlineWords,
  modeLine,
  outcomeLine,
  pendingTitle,
  requestWords,
  scopeKey,
  scopeLabel,
  waitingWords,
} from "./interruptWords";
import type { ControlRequest, ControlScope } from "./interrupts";

const board: ControlScope = { kind: "board", id: "mega-01" };
const everything: ControlScope = { kind: "all", id: "" };

describe("interrupt words", () => {
  it.each<[ControlRequest, string]>([
    [{ op: "hold", scope: board }, "Hold on board mega-01"],
    [{ op: "release", scope: everything }, "Release of all devices"],
    [{ op: "estop", scope: { kind: "device", id: "fan" } }, "Emergency stop on device fan"],
    [{ op: "approve_mode", scope: board, enabled: true }, "Approve each command on board mega-01"],
    [{ op: "approve_mode", scope: board, enabled: false }, "Stop approving each command on board mega-01"],
    [{ op: "approve", commandId: "cmd-7" }, "Approve cmd-7"],
    [{ op: "deny", commandId: "cmd-7" }, "Deny cmd-7"],
    [{ op: "state" }, "Reading the state"],
  ])("words the request %j", (request, words) => {
    expect(requestWords(request)).toBe(words);
  });

  it("names a scope for a person and for a key", () => {
    expect(scopeLabel(everything)).toBe("All devices");
    expect(scopeLabel(board)).toBe("Board mega-01");
    expect(scopeKey(board)).toBe("board:mega-01");
    expect(scopeKey(everything)).toBe("all:");
  });

  it("tells a hold, a latch and an approval mode apart by words", () => {
    expect(modeLine({ mode: "hold", latched: false })).toBe("Held");
    expect(modeLine({ mode: "hold", latched: true })).toBe("Held · latched by emergency stop");
    expect(modeLine({ mode: "approve", latched: false })).toBe("Approving each command");
  });

  it("says what the node answered, in plain words, with the refusal reason from the one table", () => {
    const hold: ControlRequest = { op: "hold", scope: board };
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

  it("states what an emergency stop did and that the scope stays held", () => {
    const stop: ControlRequest = { op: "estop", scope: everything };
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

  it("shows a failure plainly and says what was done to read the state again", () => {
    const hold: ControlRequest = { op: "hold", scope: board };
    const timeout = outcomeLine(hold, { kind: "failed", failure: "timeout" }, "pending");
    expect(timeout.tone).toBe("failed");
    expect(timeout.text).toBe(
      "Hold on board mega-01: The node did not answer in time. The control may still apply afterwards. Reading the node's state again…",
    );
    expect(outcomeLine(hold, { kind: "failed", failure: "gone" }, "read").text).toMatch(
      /read from the node just now\.$/,
    );
    expect(outcomeLine(hold, { kind: "failed", failure: "unavailable" }, "failed").text).toMatch(
      /may be out of date\.$/,
    );
    expect(outcomeLine(hold, { kind: "failed", failure: "no-session" }, "none").text).toMatch(/nothing was sent\.$/);
    const busy = replyOf({ result: "refused", reason: "node_busy" });
    expect(outcomeLine(hold, { kind: "reply", reply: busy }, "pending").text).toContain("may still apply this later");
  });

  it("counts the commands that wait and shows their deadline from the binary's numbers", () => {
    expect(waitingWords(0)).toBe("No commands wait for approval.");
    expect(waitingWords(1)).toBe("1 command waiting for approval");
    expect(waitingWords(2)).toBe("2 commands waiting for approval");
    expect(pendingTitle(fanWaits)).toBe("set_fan on fan");
    expect(deadlineWords(fanWaits, { worldNs: 100e9, wallMs: 0 })).toBe("expires in 30 s (run time)");
    expect(deadlineWords(fanWaits, { worldNs: 130e9, wallMs: 0 })).toBe(
      "deadline reached; the node settles it at its next step",
    );
    expect(deadlineWords(sensorWaits, { worldNs: 0, wallMs: 1_700_000_000_000 })).toBe("expires in 45 s");
    expect(deadlineWords(sensorWaits, { worldNs: 0, wallMs: 1_700_000_100_000 })).toBe(
      "deadline reached; the node settles it at its next step",
    );
  });
});
