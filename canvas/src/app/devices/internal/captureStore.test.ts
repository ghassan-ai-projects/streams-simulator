import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../ports";
import type { Canned } from "../../../test/fakeFailures";
import {
  announced,
  captureResult,
  differs,
  pumpDraft,
  refuseCapture,
  savedPumpPath,
  serveCapture,
} from "../../../test/captureFixture";
import { board, opened, text } from "../../../test/captureSessionHarness";
import { deferred } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import { decisionCounts } from "./captureSelectors";

describe("reading an announcement", () => {
  it("asks the binary and shows what the board announced, nothing decided and nothing written", async () => {
    const { bridge, session } = opened();
    const reading = session.read(text);
    expect(session.snapshot()).toEqual({ phase: "reading", source: text });
    await reading;
    const state = board(session.snapshot());
    expect(state).toMatchObject({ phase: "reviewing", environment: "hardware", bdSeq: 2, source: text });
    expect(state.scan).toMatchObject({ birthsSeen: 1, taken: { line: 1, bdSeq: 2 }, laterBirths: [], issues: [] });
    expect(state.cards.map((card) => [card.kind.kind, card.decision.kind, card.saved])).toEqual([
      ["pump", "open", null],
    ]);
    expect(bridge.calls).toEqual([{ name: "kind.capture", args: text }]);
    expect(bridge.writes).toEqual([]);
    expect(bridge.reads).toEqual([]);
  });

  it("reads a recorded run by its id", async () => {
    const { bridge, session } = opened();
    await session.read({ run: "w-1" });
    expect(bridge.calls[0]).toEqual({ name: "kind.capture", args: { run: "w-1" } });
  });

  it("reads the saved file of a device that differs once, right away, and keeps what it read", async () => {
    const { bridge, session } = opened(captureResult([differs(), announced({ kind: "fan" })]), true);
    await session.read(text);
    expect(bridge.reads).toEqual([savedPumpPath]);
    const [pump, fan] = board(session.snapshot()).cards;
    expect(pump?.saved).toMatchObject({ ok: true, value: { path: savedPumpPath, etag: "k3" } });
    expect(fan?.saved).toBeNull();
  });

  it("keeps the failure when the saved file cannot be read", async () => {
    const { session } = opened(captureResult([differs()]));
    await session.read(text);
    expect(board(session.snapshot()).cards[0]?.saved).toMatchObject({ ok: false, failure: { kind: "missing" } });
  });

  it("is refused with the binary's issues, writes nothing, and can be tried again", async () => {
    const { bridge, session } = opened();
    refuseCapture(bridge);
    await session.read(text);
    expect(session.snapshot()).toMatchObject({
      phase: "refused",
      source: text,
      failure: { kind: "refused", diagnostic: { message: "the birth is refused" } },
    });
    expect(bridge.writes).toEqual([]);
    serveCapture(bridge);
    await session.retry();
    expect(board(session.snapshot()).phase).toBe("reviewing");
    expect(bridge.calls.map((call) => call.name)).toEqual(["kind.capture", "kind.capture"]);
  });

  it("has nothing to retry before anything was read, and nothing after it is cleared", async () => {
    const { bridge, session } = opened();
    await session.retry();
    await session.read(text);
    session.clear();
    await session.retry();
    expect(session.snapshot()).toEqual({ phase: "idle" });
    expect(bridge.calls).toHaveLength(1);
  });

  it("drops an answer that arrives after the review was cleared", async () => {
    const { bridge, session } = opened();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("kind.capture", () => gate.promise);
    const stale = session.read(text);
    session.clear();
    gate.resolve(success(captureResult()));
    await stale;
    expect(session.snapshot()).toEqual({ phase: "idle" });
  });

  it("does not read again while a device is being saved", async () => {
    const { bridge, session } = opened();
    await session.read(text);
    const gates = bridge.holdWrite();
    const saving = session.save(0);
    await session.read({ run: "w-1" });
    expect(board(session.snapshot()).phase).toBe("saving");
    expect(bridge.calls).toHaveLength(1);
    gates[0]?.resolve(success({ path: "kinds/pump.kind.json", etag: "e1" }));
    await saving;
  });

  it("remembers that the screen was left, once", () => {
    const { session } = opened();
    expect(session.comingBack()).toBe(false);
    session.leave();
    expect(session.comingBack()).toBe(true);
    expect(session.comingBack()).toBe(false);
  });
});

describe("saving a new device", () => {
  it("writes the draft exactly as announced to the path the binary named, as a new file", async () => {
    const target = { status: "new", target: "kinds/pump_2.kind.json" };
    const { bridge, session } = opened(captureResult([announced(pumpDraft, { comparison: target })]));
    await session.read(text);
    await session.save(0);
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/pump_2.kind.json", etag: "" });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toEqual(pumpDraft);
    const state = board(session.snapshot());
    expect(state.phase).toBe("saved");
    expect(state.cards[0]?.decision).toEqual({ kind: "saved", path: "kinds/pump_2.kind.json" });
    expect(decisionCounts(state.cards)).toEqual({ saved: 1, skipped: 0 });
  });

  it("says the device is being saved while the file is written, and does not save it twice", async () => {
    const { bridge, session } = opened();
    await session.read(text);
    const gates = bridge.holdWrite();
    const saving = session.save(0);
    void session.save(0);
    expect(board(session.snapshot()).phase).toBe("saving");
    expect(gates).toHaveLength(1);
    gates[0]?.resolve(success({ path: "kinds/pump.kind.json", etag: "e1" }));
    await saving;
    expect(bridge.writes).toHaveLength(1);
  });

  it("keeps a draft the binary finds faults in, as announced, to be repaired in the editor", async () => {
    const issues = [{ pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" }];
    const { bridge, session } = opened(captureResult([announced(pumpDraft, { issues })]));
    await session.read(text);
    await session.save(0);
    expect(bridge.writes).toHaveLength(1);
    expect(board(session.snapshot()).cards[0]?.kind.issues).toEqual(issues);
  });

  it("saves two devices of one announcement independently, even under one name", async () => {
    const at = (path: string) => announced(pumpDraft, { comparison: { status: "new", target: path } });
    const { bridge, session } = opened(captureResult([at("kinds/a.kind.json"), at("kinds/b.kind.json")]));
    await session.read(text);
    await session.save(1);
    expect(board(session.snapshot()).cards.map((card) => card.decision.kind)).toEqual(["open", "saved"]);
    expect(board(session.snapshot()).phase).toBe("reviewing");
    await session.save(0);
    expect(bridge.writes.map((write) => write.path)).toEqual(["kinds/b.kind.json", "kinds/a.kind.json"]);
    expect(board(session.snapshot()).phase).toBe("saved");
  });

  it("says a file of that name exists, writes nothing more, and lets the author try again", async () => {
    const { bridge, session } = opened();
    await session.read(text);
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("conflict", 409, "exists")));
    await session.save(0);
    expect(board(session.snapshot()).cards[0]?.decision).toMatchObject({
      kind: "failed",
      failure: {
        kind: "conflict",
        diagnostic: { message: expect.stringContaining("A device named pump already exists") as string },
      },
    });
    expect(board(session.snapshot()).phase).toBe("reviewing");
    bridge.writeOutcomes.clear();
    await session.save(0);
    expect(board(session.snapshot()).phase).toBe("saved");
  });

  it("keeps a bridge fault as it is", async () => {
    const { bridge, session } = opened();
    await session.read(text);
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("bridge_failed", 500, "down")));
    await session.save(0);
    expect(board(session.snapshot()).cards[0]?.decision).toMatchObject({
      kind: "failed",
      failure: { kind: "unavailable", diagnostic: { message: "down" } },
    });
  });

  it("does nothing for a device it was not shown", async () => {
    const { bridge, session } = opened();
    await session.save(0);
    await session.read(text);
    await session.save(3);
    session.skip(3);
    session.unskip(3);
    expect(bridge.writes).toEqual([]);
  });
});
