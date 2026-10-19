import { describe, expect, it } from "vitest";
import { failure, success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { controlReplyWire, interruptsWire } from "../../../test/liveFixture";
import { ControlStore } from "./controlStore";

const board = { kind: "board", id: "mega-01" } as const;

function setup(newestKey = "key-1") {
  const bridge = new FakeBridge();
  const requests: Record<string, unknown>[] = [];
  const store = new ControlStore(bridge.review, () => newestKey);
  const answer = (...replies: Canned<unknown>[]) => {
    bridge.handlers.set("session.control", (args) => {
      requests.push({ ...args });
      return replies.shift() ?? success(controlReplyWire());
    });
  };
  return { bridge, store, answer, requests };
}

describe("controls on a live session", () => {
  it("sends each control as one request to the node and keeps its answer as the state seen", async () => {
    const { store, answer, requests } = setup();
    answer();
    await store.hold(board);
    await store.release(board);
    await store.approveEach(board);
    await store.stopApproving(board);
    await store.emergencyStop({ kind: "all", id: "" });
    await store.approve("cmd-7");
    await store.deny("cmd-8");
    expect(requests).toEqual([
      { op: "hold", scope: { kind: "board", id: "mega-01" } },
      { op: "release", scope: { kind: "board", id: "mega-01" } },
      { op: "approve_mode", scope: { kind: "board", id: "mega-01" }, enabled: true },
      { op: "approve_mode", scope: { kind: "board", id: "mega-01" }, enabled: false },
      { op: "estop", scope: { kind: "all" } },
      { op: "approve", command_id: "cmd-7" },
      { op: "deny", command_id: "cmd-8" },
    ]);
    const { seen, last, sending } = store.snapshot();
    expect(sending).toBeNull();
    expect(seen?.seenKey).toBe("key-1");
    expect(seen?.state).toMatchObject({ holds: [{ scope: { id: "uno-07" } }] });
    expect(last).toMatchObject({ request: { op: "deny" }, result: { kind: "reply" }, check: "none" });
  });

  it("shows a control in flight and takes no second one until the node answered", async () => {
    const { bridge, store, requests } = setup();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("session.control", (args) => {
      requests.push({ ...args });
      return gate.promise;
    });
    const first = store.hold(board);
    expect(store.snapshot().sending).toEqual({ op: "hold", scope: board });
    await store.emergencyStop(board);
    expect(requests).toHaveLength(1);
    gate.resolve(success(controlReplyWire()));
    await first;
    expect(store.snapshot().sending).toBeNull();
  });

  it("keeps a refusal of the node as its answer and does not read the state again", async () => {
    const { store, answer, requests } = setup();
    answer(
      success(controlReplyWire({ result: "refused", reason: "unknown_board", state: interruptsWire({ holds: [] }) })),
    );
    await store.hold(board);
    expect(requests).toHaveLength(1);
    expect(store.snapshot().last).toMatchObject({
      result: { kind: "reply", reply: { result: "refused" } },
      check: "none",
    });
    expect(store.snapshot().seen?.state.holds).toEqual([]);
  });

  it("answers a busy node plainly and reads the node's state again", async () => {
    const { store, answer, requests } = setup();
    const busy = success(
      controlReplyWire({ result: "refused", reason: "node_busy", state: interruptsWire({ holds: [] }) }),
    );
    answer(busy, success(controlReplyWire({ op: "state", state: interruptsWire() })));
    await store.hold(board);
    await flushPromises();
    expect(requests).toEqual([{ op: "hold", scope: { kind: "board", id: "mega-01" } }, { op: "state" }]);
    expect(store.snapshot().last).toMatchObject({ check: "read" });
    expect(store.snapshot().seen?.state.holds).toHaveLength(1);
  });

  it.each([
    ["timeout", "control_timeout"],
    ["gone", "control_gone"],
    ["unavailable", "bridge_failed"],
  ])("shows a %s plainly and reads the state again, saying when that failed too", async (kind, code) => {
    const { store, answer, requests } = setup();
    const status = code === "bridge_failed" ? 500 : 422;
    answer(failure(cannedFailure(code, status, "words")), failure(cannedFailure(code, status, "words")));
    await store.hold(board);
    await flushPromises();
    expect(requests).toHaveLength(2);
    expect(store.snapshot().last).toMatchObject({ result: { kind: "failed", failure: kind }, check: "failed" });
    expect(store.snapshot().seen).toBeNull();
  });

  it("does not read the state again when there is no session or no control socket to read it from", async () => {
    for (const code of ["no_session", "control_unavailable"]) {
      const { store, answer, requests } = setup();
      answer(failure(cannedFailure(code, 422, "words")));
      await store.deny("cmd-1");
      await flushPromises();
      expect(requests).toHaveLength(1);
      expect(store.snapshot().last).toMatchObject({ result: { kind: "failed" }, check: "none" });
    }
  });

  it("leaves a newer control's outcome alone when an older state read finishes", async () => {
    const { bridge, store, requests } = setup();
    const gate = deferred<Canned<unknown>>();
    const replies = [
      success(controlReplyWire({ result: "refused", reason: "node_busy" })),
      gate.promise,
      success(controlReplyWire({ op: "release" })),
    ];
    bridge.handlers.set("session.control", (args) => {
      requests.push({ ...args });
      return replies.shift() ?? success(controlReplyWire());
    });
    await store.hold(board);
    await flushPromises();
    expect(store.snapshot().last).toMatchObject({ request: { op: "hold" }, check: "pending" });
    await store.release(board);
    gate.resolve(success(controlReplyWire({ op: "state", state: interruptsWire() })));
    await flushPromises();
    expect(store.snapshot().last).toMatchObject({ request: { op: "release" }, check: "none" });
  });
});
