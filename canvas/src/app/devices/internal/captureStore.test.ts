import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../ports";
import type { Canned } from "../../../test/fakeFailures";
import {
  announced,
  captureResult,
  differs,
  pumpDraft,
  refuseCapture,
  same,
  serveCapture,
} from "../../../test/captureFixture";
import { deferred } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { createCaptureSession } from "./captureStore";
import type { CaptureState } from "./captureState";
import { decisionCounts, reviewedOf } from "./captureSelectors";

const entry = (path: string, kind: string) => ({ path, kind, title: "" });
const text = { content: '{"type":"NBIRTH"}' };

function opened(result: object = captureResult()) {
  const bridge = new FakeBridge();
  serveCapture(bridge, result);
  return { bridge, session: createCaptureSession(bridge.devices) };
}

function board(state: CaptureState) {
  const found = reviewedOf(state);
  if (found === null) throw new Error(state.phase);
  return found;
}

describe("reading an announcement", () => {
  it("asks the binary and shows what the board announced, nothing decided and nothing written", async () => {
    const { bridge, session } = opened();
    const reading = session.read(text, []);
    expect(session.snapshot()).toEqual({ phase: "reading", source: text });
    await reading;
    const state = board(session.snapshot());
    expect(state).toMatchObject({ phase: "reviewing", environment: "hardware", bdSeq: 2, source: text });
    expect(state.cards.map((card) => [card.kind.kind, card.decision.kind])).toEqual([["pump", "open"]]);
    expect(bridge.calls).toEqual([{ name: "kind.capture", args: text }]);
    expect(bridge.writes).toEqual([]);
  });

  it("reads a recorded run by its id", async () => {
    const { bridge, session } = opened();
    await session.read({ run: "w-1" }, []);
    expect(bridge.calls[0]).toEqual({ name: "kind.capture", args: { run: "w-1" } });
  });

  it("is refused with the binary's issues, writes nothing, and can be tried again", async () => {
    const { bridge, session } = opened();
    refuseCapture(bridge);
    await session.read(text, []);
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

  it("has nothing to retry before anything was read, and nothing after it is closed", async () => {
    const { bridge, session } = opened();
    await session.retry();
    await session.read(text, []);
    session.close();
    await session.retry();
    expect(session.snapshot()).toEqual({ phase: "idle" });
    expect(bridge.calls).toHaveLength(1);
  });

  it("drops an answer that arrives after the screen was closed or another text was read", async () => {
    const { bridge, session } = opened();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("kind.capture", () => gate.promise);
    const stale = session.read(text, []);
    session.close();
    gate.resolve(success(captureResult()));
    await stale;
    expect(session.snapshot()).toEqual({ phase: "idle" });
  });
});

describe("saving a new device", () => {
  it("writes the draft exactly as announced to the device's own file, as a new file", async () => {
    const { bridge, session } = opened();
    await session.read(text, []);
    await session.save("pump");
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/pump.kind.json", etag: "" });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toEqual(pumpDraft);
    const state = board(session.snapshot());
    expect(state.phase).toBe("saved");
    expect(state.cards[0]?.decision).toEqual({ kind: "saved", path: "kinds/pump.kind.json" });
    expect(decisionCounts(state.cards)).toEqual({ saved: 1, skipped: 0 });
  });

  it("says the device is being saved while the file is written, and does not save it twice", async () => {
    const { bridge, session } = opened();
    await session.read(text, []);
    const gates = bridge.holdWrite();
    const saving = session.save("pump");
    void session.save("pump");
    expect(board(session.snapshot()).phase).toBe("saving");
    expect(gates).toHaveLength(1);
    gates[0]?.resolve(success({ path: "kinds/pump.kind.json", etag: "e1" }));
    await saving;
    expect(bridge.writes).toHaveLength(1);
  });

  it("keeps a draft the binary finds faults in, to be repaired in the editor", async () => {
    const issues = [{ pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" }];
    const { bridge, session } = opened(captureResult([announced(pumpDraft, { issues })]));
    await session.read(text, []);
    await session.save("pump");
    expect(bridge.writes).toHaveLength(1);
    expect(board(session.snapshot()).cards[0]?.kind.issues).toEqual(issues);
  });

  it("says a file of that name exists, writes nothing more, and lets the author try again", async () => {
    const { bridge, session } = opened();
    await session.read(text, []);
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("conflict", 409, "exists")));
    await session.save("pump");
    const card = board(session.snapshot()).cards[0];
    expect(card?.decision).toMatchObject({
      kind: "failed",
      failure: {
        kind: "conflict",
        diagnostic: { message: expect.stringContaining("A device named pump already exists") as string },
      },
    });
    expect(board(session.snapshot()).phase).toBe("reviewing");
    bridge.writeOutcomes.clear();
    await session.save("pump");
    expect(board(session.snapshot()).phase).toBe("saved");
  });

  it("keeps a bridge fault as it is", async () => {
    const { bridge, session } = opened();
    await session.read(text, []);
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("bridge_failed", 500, "down")));
    await session.save("pump");
    expect(board(session.snapshot()).cards[0]?.decision).toMatchObject({
      kind: "failed",
      failure: { kind: "unavailable", diagnostic: { message: "down" } },
    });
  });

  it.each([["a/b"], ["..\\x"], [".hidden"]])("does not write a device named %j into a file", async (name) => {
    const { bridge, session } = opened(captureResult([announced({ kind: name })]));
    await session.read(text, []);
    await session.save(name);
    expect(bridge.writes).toEqual([]);
  });

  it("does nothing for a device it was not shown", async () => {
    const { bridge, session } = opened();
    await session.save("pump");
    await session.read(text, []);
    await session.save("fan");
    session.skip("fan");
    expect(bridge.writes).toEqual([]);
  });
});

describe("replacing a saved device", () => {
  it("never overwrites silently: saving a device that differs from its file does nothing", async () => {
    const { bridge, session } = opened(captureResult([differs()]));
    await session.read(text, [entry("kinds/pump.kind.json", "pump")]);
    await session.save("pump");
    expect(bridge.writes).toEqual([]);
  });

  it("writes over the saved file with the etag it has now", async () => {
    const { bridge, session } = opened(captureResult([differs()]));
    bridge.addFile("kinds/old-pump.kind.json", { kind: "pump" }, "k7");
    await session.read(text, [entry("kinds/old-pump.kind.json", "pump")]);
    await session.replace("pump");
    expect(bridge.writes).toEqual([
      { path: "kinds/old-pump.kind.json", etag: "k7", content: expect.stringContaining('"kind": "pump"') as string },
    ]);
    expect(board(session.snapshot()).phase).toBe("saved");
  });

  it("says the saved device changed since it was compared, and writes nothing over it", async () => {
    const { bridge, session } = opened(captureResult([differs()]));
    bridge.addFile("kinds/pump.kind.json", { kind: "pump" }, "k7");
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("conflict", 409, "changed")));
    await session.read(text, [entry("kinds/pump.kind.json", "pump")]);
    await session.replace("pump");
    expect(board(session.snapshot()).cards[0]?.decision).toMatchObject({
      kind: "failed",
      failure: {
        kind: "conflict",
        diagnostic: { message: expect.stringContaining("changed since it was compared") as string },
      },
    });
  });

  it("creates the file again when it has vanished, and stops when it cannot be read", async () => {
    const { bridge, session } = opened(captureResult([differs()]));
    await session.read(text, []);
    await session.replace("pump");
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/pump.kind.json", etag: "" });
    const broken = opened(captureResult([differs()]));
    broken.bridge.files.set("kinds/pump.kind.json", failed(cannedFailure("bridge_failed", 500, "down")));
    await broken.session.read(text, []);
    await broken.session.replace("pump");
    expect(broken.bridge.writes).toEqual([]);
    expect(board(broken.session.snapshot()).cards[0]?.decision).toMatchObject({ kind: "failed" });
  });

  it("only replaces a device that differs", async () => {
    const { bridge, session } = opened();
    await session.read(text, []);
    await session.replace("pump");
    expect(bridge.writes).toEqual([]);
  });
});

describe("a device the same as its file, and a device skipped", () => {
  it("offers nothing to save when the board announced what is already saved", async () => {
    const { bridge, session } = opened(captureResult([same()]));
    await session.read(text, [entry("kinds/pump.kind.json", "pump")]);
    await session.save("pump");
    await session.replace("pump");
    expect(bridge.writes).toEqual([]);
  });

  it("skips a device, and the screen is done when every device is saved or skipped", async () => {
    const fan = { kind: "fan", metrics: [], commands: [] };
    const { session } = opened(captureResult([announced(), announced(fan)]));
    await session.read(text, []);
    session.skip("fan");
    expect(board(session.snapshot()).phase).toBe("reviewing");
    await session.save("pump");
    const state = board(session.snapshot());
    expect(state.phase).toBe("saved");
    expect(decisionCounts(state.cards)).toEqual({ saved: 1, skipped: 1 });
    session.skip("pump");
    expect(decisionCounts(board(session.snapshot()).cards)).toEqual({ saved: 1, skipped: 1 });
  });

  it("is not done when the board announced nothing", async () => {
    const { session } = opened(captureResult([]));
    await session.read(text, []);
    expect(board(session.snapshot())).toMatchObject({ phase: "reviewing", cards: [] });
  });
});
