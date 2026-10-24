import { describe, expect, it } from "vitest";
import { comparisonWire } from "../../../test/compareFixture";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { CompareStore, compareFor } from "./compareStore";

function setup() {
  const bridge = new FakeBridge();
  const sent: Record<string, unknown>[] = [];
  bridge.handlers.set("run.compare", (args) => {
    sent.push({ ...args });
    return success(comparisonWire);
  });
  return { bridge, sent, store: new CompareStore(bridge.review) };
}

describe("the comparison of this run with a reference", () => {
  it("asks for nothing until a reference is chosen", () => {
    const { store, sent } = setup();
    expect(store.snapshot()).toEqual({ phase: "idle" });
    expect(sent).toEqual([]);
  });

  it("sends the reference as a and this run as b, once, and keeps what the binary answered", async () => {
    const { store, sent } = setup();
    await store.choose("w-2", "w-1");
    expect(sent).toEqual([{ a: "w-1", b: "w-2" }]);
    expect(store.snapshot()).toMatchObject({
      phase: "ready",
      run: "w-2",
      reference: "w-1",
      comparison: { a: { runId: "w-1" }, b: { runId: "w-2" } },
    });
  });

  it("shows loading while the binary reads both runs", async () => {
    const { bridge, store } = setup();
    const held = deferred<Canned<unknown>>();
    bridge.handlers.set("run.compare", () => held.promise);
    const choosing = store.choose("w-2", "w-1");
    await flushPromises();
    expect(store.snapshot()).toEqual({ phase: "loading", run: "w-2", reference: "w-1" });
    held.resolve(success(comparisonWire));
    await choosing;
    expect(store.snapshot().phase).toBe("ready");
  });

  it("keeps the binary's refusal word for word", async () => {
    const { bridge, store } = setup();
    const message = "run B has not ended: its truth is still sealed";
    bridge.handlers.set("run.compare", () => failure(cannedFailure("run_not_ended", 422, message)));
    await store.choose("w-2", "w-1");
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      reference: "w-1",
      failure: { kind: "refused", diagnostic: { message, displayCode: "run_not_ended" } },
    });
  });

  it("drops an answer that arrives after another reference was chosen", async () => {
    const { bridge, store } = setup();
    const first = deferred<Canned<unknown>>();
    bridge.handlers.set("run.compare", () => first.promise);
    const slow = store.choose("w-2", "w-1");
    await flushPromises();
    bridge.handlers.set("run.compare", () => success(comparisonWire));
    await store.choose("w-2", "w-3");
    first.resolve(success(comparisonWire));
    await slow;
    expect(store.snapshot()).toMatchObject({ phase: "ready", reference: "w-3" });
  });

  it("goes back to choosing when cleared", async () => {
    const { store } = setup();
    await store.choose("w-2", "w-1");
    store.clear();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });

  it("belongs to the run it was made for, and is idle on any other", async () => {
    const { store } = setup();
    await store.choose("w-2", "w-1");
    expect(compareFor(store.snapshot(), "w-2").phase).toBe("ready");
    expect(compareFor(store.snapshot(), "w-9")).toEqual({ phase: "idle" });
    expect(compareFor(store.snapshot(), null)).toEqual({ phase: "idle" });
  });
});
