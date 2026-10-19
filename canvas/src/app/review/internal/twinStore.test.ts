import { describe, expect, it } from "vitest";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { cannotJudgeWire, twinReportWire } from "../../../test/twinReportFixture";
import { TwinStore, twinFor } from "./twinStore";

const request = { run: "c-1", domain: "mini", tolerance: "" };

function setup() {
  const bridge = new FakeBridge();
  const sent: Record<string, unknown>[] = [];
  bridge.handlers.set("run.twin_report", (args) => {
    sent.push({ ...args });
    return success(twinReportWire());
  });
  return { bridge, sent, store: new TwinStore(bridge.review) };
}

describe("the twin comparison of a run", () => {
  it("asks for nothing until it is told to, and is idle", () => {
    const { store, sent } = setup();
    expect(store.snapshot()).toEqual({ phase: "idle" });
    expect(sent).toEqual([]);
  });

  it("asks once for the run and keeps the report it got with the request that made it", async () => {
    const { store, sent } = setup();
    await store.compare({ ...request, tolerance: "sensor.temperature=2" });
    expect(sent).toEqual([{ run: "c-1", domain: "mini", tolerance: "sensor.temperature=2" }]);
    expect(store.snapshot()).toMatchObject({
      phase: "ready",
      request: { run: "c-1", tolerance: "sensor.temperature=2" },
      report: { verdict: "diverges" },
    });
  });

  it("marks a new answer fresh once, so a screen focuses its result on arrival and never again", async () => {
    const { store } = setup();
    store.markSeen();
    expect(store.snapshot()).toEqual({ phase: "idle" });
    await store.compare(request);
    expect(store.snapshot()).toMatchObject({ phase: "ready", fresh: true });
    store.markSeen();
    expect(store.snapshot()).toMatchObject({ phase: "ready", fresh: false });
    store.markSeen();
    expect(store.snapshot()).toMatchObject({ phase: "ready", fresh: false });
  });

  it("marks a refusal fresh the same way", async () => {
    const { bridge, store } = setup();
    bridge.handlers.set("run.twin_report", () => failure(cannedFailure("plantwire_failed", 422, "invalid tolerance")));
    await store.compare(request);
    expect(store.snapshot()).toMatchObject({ phase: "failed", fresh: true });
    store.markSeen();
    expect(store.snapshot()).toMatchObject({ phase: "failed", fresh: false });
  });

  it("keeps a report that cannot judge as a result", async () => {
    const { bridge, store } = setup();
    bridge.handlers.set("run.twin_report", () => success(cannotJudgeWire()));
    await store.compare(request);
    expect(store.snapshot()).toMatchObject({ phase: "ready", report: { verdict: "cannot_judge" } });
  });

  it("shows the binary's refusal as it came", async () => {
    const { bridge, store } = setup();
    bridge.handlers.set("run.twin_report", () => failure(cannedFailure("plantwire_failed", 422, "invalid tolerance")));
    await store.compare(request);
    expect(store.snapshot()).toMatchObject({
      phase: "failed",
      request,
      failure: { kind: "refused", diagnostic: { message: "invalid tolerance", displayCode: "plantwire_failed" } },
    });
  });

  it("is loading while it waits, and a second press for the same run asks nothing more", async () => {
    const { bridge, store, sent } = setup();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("run.twin_report", (args) => {
      sent.push({ ...args });
      return gate.promise;
    });
    const first = store.compare(request);
    await flushPromises();
    expect(store.snapshot()).toEqual({ phase: "loading", request });
    await store.compare(request);
    expect(sent).toHaveLength(1);
    gate.resolve(success(twinReportWire()));
    await first;
    expect(store.snapshot().phase).toBe("ready");
  });

  it("forgets an answer that arrives after another run was asked for", async () => {
    const { bridge, store } = setup();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("run.twin_report", () => gate.promise);
    const slow = store.compare(request);
    await flushPromises();
    bridge.handlers.set("run.twin_report", () => success(cannotJudgeWire()));
    await store.compare({ ...request, run: "c-2" });
    gate.resolve(success(twinReportWire()));
    await slow;
    expect(store.snapshot()).toMatchObject({
      phase: "ready",
      request: { run: "c-2" },
      report: { verdict: "cannot_judge" },
    });
  });

  it("shows only what was asked for the run on screen", async () => {
    const { store } = setup();
    await store.compare(request);
    expect(twinFor(store.snapshot(), "c-1").phase).toBe("ready");
    expect(twinFor(store.snapshot(), "c-2")).toEqual({ phase: "idle" });
    expect(twinFor(store.snapshot(), null)).toEqual({ phase: "idle" });
    expect(twinFor({ phase: "idle" }, "c-1")).toEqual({ phase: "idle" });
  });
});
