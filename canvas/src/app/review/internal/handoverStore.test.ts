import { describe, expect, it } from "vitest";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { deferred, flushPromises } from "../../../test/deferred";
import { HandoverStore, handoverFor } from "./handoverStore";

function setup(saved = true) {
  const bridge = new FakeBridge();
  const asked: Record<string, unknown>[] = [];
  const files: { name: string; text: string }[] = [];
  bridge.handlers.set("run.report", (args) => {
    asked.push({ ...args });
    return success({ format: "markdown", markdown: "# Run w-2\n\n## Open questions\n" });
  });
  const saver = {
    save: (name: string, text: string) => {
      files.push({ name, text });
      return saved;
    },
  };
  return { bridge, asked, files, store: new HandoverStore(bridge.review, saver) };
}

describe("the handover of a run", () => {
  it("is idle until it is asked for", () => {
    expect(setup().store.snapshot()).toEqual({ phase: "idle" });
  });

  it("downloads the binary's markdown as <run id>-handover.md and assembles nothing itself", async () => {
    const { store, files, asked } = setup();
    await store.save("w-2", { replay: false, fileStem: "r-2" });
    expect(asked).toEqual([{ run: "w-2", format: "markdown" }]);
    expect(files).toEqual([{ name: "r-2-handover.md", text: "# Run w-2\n\n## Open questions\n" }]);
    expect(store.snapshot()).toEqual({ phase: "saved", run: "w-2", name: "r-2-handover.md" });
  });

  it("asks for the replay when it was already checked in this view", async () => {
    const { store, asked } = setup();
    await store.save("w-2", { replay: true, fileStem: "r-2" });
    expect(asked).toEqual([{ run: "w-2", format: "markdown", replay: true }]);
  });

  it("shows the binary's refusal word for word and saves no file", async () => {
    const { bridge, store, files } = setup();
    const message = "run w-2 has not ended";
    bridge.handlers.set("run.report", () => failure(cannedFailure("run_not_ended", 422, message)));
    await store.save("w-2", { replay: false, fileStem: "r-2" });
    expect(store.snapshot()).toEqual({ phase: "failed", run: "w-2", message });
    expect(files).toEqual([]);
  });

  it("says so when the browser could not save the file", async () => {
    const { store } = setup(false);
    await store.save("w-2", { replay: false, fileStem: "r-2" });
    expect(store.snapshot()).toMatchObject({ phase: "failed", message: "The browser could not save the file." });
  });

  it("is working while the binary builds it, and ignores a second request meanwhile", async () => {
    const { bridge, store, asked } = setup();
    const held = deferred<Canned<unknown>>();
    bridge.handlers.set("run.report", (args) => {
      asked.push({ ...args });
      return held.promise;
    });
    const first = store.save("w-2", { replay: false, fileStem: "r-2" });
    await flushPromises();
    expect(store.snapshot()).toEqual({ phase: "working", run: "w-2" });
    await store.save("w-2", { replay: false, fileStem: "r-2" });
    expect(asked).toHaveLength(1);
    held.resolve(success({ format: "markdown", markdown: "x" }));
    await first;
    expect(store.snapshot().phase).toBe("saved");
  });

  it("belongs to the run it was made for", async () => {
    const { store } = setup();
    await store.save("w-2", { replay: false, fileStem: "r-2" });
    expect(handoverFor(store.snapshot(), "w-2").phase).toBe("saved");
    expect(handoverFor(store.snapshot(), "w-3")).toEqual({ phase: "idle" });
  });
});
