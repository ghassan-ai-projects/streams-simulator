import { describe, expect, it } from "vitest";
import type { SessionLive } from "../../../model/review";
import { failure, success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { emptyState, fanWaits, liveOf } from "../../../test/interruptFixture";
import { controlReplyWire, interruptsWire } from "../../../test/liveFixture";
import { ControlStore } from "./controlStore";

const board = { kind: "board", id: "mega-01" } as const;
const everything = { kind: "all", id: "" } as const;

function setup(options: { readonly key?: string; readonly run?: string | null } = {}) {
  const bridge = new FakeBridge();
  const requests: Record<string, unknown>[] = [];
  const source = {
    key: options.key ?? "key-1",
    run: options.run === undefined ? "run-1" : options.run,
    live: null as SessionLive | null,
  };
  const store = new ControlStore(bridge.review, {
    run: () => source.run,
    newest: () => source.live ?? liveOf({ key: source.key, interrupts: emptyState }),
  });
  const answer = (...replies: Canned<unknown>[]) => {
    bridge.handlers.set("session.control", (args) => {
      requests.push({ ...args });
      return replies.shift() ?? success(controlReplyWire());
    });
  };
  return { bridge, store, answer, requests, source };
}

describe("controls on a live session", () => {
  it("sends each control as one request to the node and keeps its answer as the state seen", async () => {
    const { store, answer, requests } = setup();
    answer();
    await store.hold(board);
    await store.release(board);
    await store.approveEach(board);
    await store.stopApproving(board);
    await store.emergencyStop(everything);
    await store.approve(fanWaits);
    await store.deny({ commandId: "cmd-8", contentDigest: "sha256:x" });
    expect(requests).toEqual([
      { op: "hold", scope: { kind: "board", id: "mega-01" } },
      { op: "release", scope: { kind: "board", id: "mega-01" } },
      { op: "approve_mode", scope: { kind: "board", id: "mega-01" }, enabled: true },
      { op: "approve_mode", scope: { kind: "board", id: "mega-01" }, enabled: false },
      { op: "estop", scope: { kind: "all" } },
      { op: "approve", command_id: "cmd-7", content_digest: fanWaits.contentDigest },
      { op: "deny", command_id: "cmd-8", content_digest: "sha256:x" },
    ]);
    const { seen, last, sending } = store.snapshot();
    expect(sending).toEqual([]);
    expect(seen).toMatchObject({ run: "run-1", seenKey: "key-1" });
    expect(seen?.state).toMatchObject({ holds: [{ scope: { id: "uno-07" } }] });
    expect(last).toMatchObject({ request: { op: "deny" }, result: { kind: "reply" }, check: "none" });
  });

  it("remembers the state shown before an answer, so a snapshot older than the answer can be told", async () => {
    const { store, answer } = setup();
    answer();
    await store.hold(board);
    expect(store.snapshot().seen?.older).toEqual([emptyState]);
    await store.release(board);
    expect(store.snapshot().seen?.older).toHaveLength(2);
  });

  it("shows a control in flight and sends no second ordinary one, saying so, until the node answered", async () => {
    const { bridge, store, requests } = setup();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("session.control", (args) => {
      requests.push({ ...args });
      return gate.promise;
    });
    const first = store.hold(board);
    expect(store.snapshot().sending.map((entry) => entry.request)).toEqual([{ op: "hold", scope: board }]);
    expect(await store.release(board)).toBe("not-sent");
    expect(await store.approve(fanWaits)).toBe("not-sent");
    expect(requests).toHaveLength(1);
    expect(store.snapshot().blocked).toMatchObject({ op: "approve" });
    gate.resolve(success(controlReplyWire()));
    expect(await first).toBe("sent");
    expect(store.snapshot()).toMatchObject({ sending: [], blocked: null });
  });

  it("never drops an emergency stop, even while another control waits for the node", async () => {
    const { bridge, store, requests } = setup();
    const gates = [deferred<Canned<unknown>>(), deferred<Canned<unknown>>()];
    bridge.handlers.set("session.control", (args) => {
      requests.push({ ...args });
      return gates[requests.length - 1]?.promise ?? success(controlReplyWire());
    });
    const approving = store.approveEach(everything);
    const stopping = store.emergencyStop(everything);
    await flushPromises();
    expect(requests).toEqual([
      { op: "approve_mode", scope: { kind: "all" }, enabled: true },
      { op: "estop", scope: { kind: "all" } },
    ]);
    expect(store.snapshot().sending).toHaveLength(2);
    gates[1]?.resolve(success(controlReplyWire({ op: "estop", stopped: ["fan"] })));
    expect(await stopping).toBe("sent");
    gates[0]?.resolve(success(controlReplyWire({ op: "approve_mode" })));
    expect(await approving).toBe("sent");
    expect(store.snapshot().sending).toEqual([]);
    expect(store.snapshot().last).toMatchObject({ request: { op: "estop" }, result: { kind: "reply" } });
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

  it("does not read the state again when there is no session, no socket or the bridge refused the request", async () => {
    for (const [code, status] of [
      ["no_session", 422],
      ["control_unavailable", 422],
      ["bad_arguments", 400],
    ] as const) {
      const { store, answer, requests } = setup();
      answer(failure(cannedFailure(code, status, "words")));
      await store.deny(fanWaits);
      await flushPromises();
      expect(requests).toHaveLength(1);
      expect(store.snapshot().last).toMatchObject({ result: { kind: "failed" }, check: "none" });
    }
  });

  it("keeps the bridge's words with a request it refused", async () => {
    const { store, answer } = setup();
    answer(failure(cannedFailure("bad_arguments", 400, "unknown member")));
    await store.hold(board);
    expect(store.snapshot().last?.result).toEqual({ kind: "failed", failure: "rejected", detail: "unknown member" });
  });

  it("leaves a newer control's outcome and state alone when an older state read finishes", async () => {
    const { bridge, store, requests } = setup();
    const gate = deferred<Canned<unknown>>();
    const replies = [
      success(controlReplyWire({ result: "refused", reason: "node_busy" })),
      gate.promise,
      success(controlReplyWire({ op: "release", state: interruptsWire({ holds: [] }) })),
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
    expect(store.snapshot().seen?.state.holds).toEqual([]);
  });
});

describe("controls and the run they belong to", () => {
  it("forgets the answers, the outcome and the scope picked when another run is bound", async () => {
    const { store, answer } = setup();
    answer();
    store.pick(board);
    await store.hold(board);
    expect(store.snapshot()).toMatchObject({ run: "run-1", picked: board });
    expect(store.snapshot().seen).not.toBeNull();
    store.bind("run-2");
    expect(store.snapshot()).toMatchObject({
      run: "run-2",
      seen: null,
      last: null,
      sending: [],
      blocked: null,
      picked: everything,
    });
  });

  it("keeps the scope picked while the same run stays bound", () => {
    const { store } = setup();
    store.pick(board);
    store.bind("run-1");
    expect(store.snapshot().picked).toEqual(board);
  });

  it("drops the answer to a control that was sent for a run that is no longer bound", async () => {
    const { bridge, store } = setup();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("session.control", () => gate.promise);
    const sending = store.hold(board);
    store.bind("run-2");
    gate.resolve(success(controlReplyWire()));
    await sending;
    expect(store.snapshot()).toMatchObject({ run: "run-2", seen: null, last: null, sending: [] });
  });

  it("starts bound to the run the source names", () => {
    expect(setup({ run: "run-9" }).store.snapshot().run).toBe("run-9");
    expect(setup({ run: null }).store.snapshot().run).toBeNull();
  });
});
