import { describe, expect, it, vi } from "vitest";
import { failure as failed, success } from "../../ports";
import { flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { createEditorSession } from "./editorSession";

const path = "compositions/lab.composition.json";
const doc = {
  id: "lab",
  title: "Lab",
  areas: [{ name: "Site" }],
  equipment: [{ use: "streamsim/chiller@1", names: ["CH-1"], area: "Site" }],
  connections: [],
};

function bridge() {
  const fake = new FakeBridge();
  fake.addFile(path, doc, "c1");
  fake.addFile("domains/lab.domain.json", "{}", "d1");
  fake.actions.set("component.list", success({ components: [] }));
  fake.actions.set("component.show", success({ document: {} }));
  fake.actions.set("domain.compose", success({ channels: 3, faults: 2, effectors: 1 }));
  return fake;
}

async function opened(fake = bridge()) {
  const store = createEditorSession(fake.editor);
  await store.open(path);
  return { store, fake };
}

describe("the plausibility check after a save", () => {
  const report = { errors: 0, warnings: 1, findings: [{ severity: "warning", entity: "CH-1", detail: "flat" }] };

  function composedState(store: ReturnType<typeof createEditorSession>) {
    const state = store.snapshot();
    if (state.phase !== "ready" || state.save.kind !== "composed") throw new Error("not composed");
    return state.save;
  }

  it("keeps the digest of the compose and runs domain.lint on the composed file, showing it while it runs", async () => {
    const fake = bridge();
    fake.actions.set("domain.compose", success({ channels: 3, faults: 2, effectors: 1, digest: "sha256:ab" }));
    const gate = fake.holdAction("domain.lint");
    fake.actions.set("domain.lint", success(report));
    const { store } = await opened(fake);
    await store.save();
    expect(composedState(store)).toMatchObject({ digest: "sha256:ab", plausibility: { kind: "running" } });
    gate.release();
    await vi.waitFor(() => {
      expect(composedState(store).plausibility).toEqual({ kind: "done", report });
    });
    expect(fake.calls.find((call) => call.name === "domain.lint")?.args).toEqual({ path: "domains/lab.domain.json" });
  });

  it("says why the check could not be made", async () => {
    const fake = bridge();
    fake.actions.set("domain.lint", failed(cannedFailure("domain_failed", 500, "lint broke")));
    const { store } = await opened(fake);
    await store.save();
    await vi.waitFor(() => {
      expect(composedState(store).plausibility).toEqual({ kind: "failed", message: "lint broke (domain_failed)" });
    });
  });

  it("drops the answer when the editor moved on to another save before it arrived", async () => {
    const fake = bridge();
    const slow = fake.holdAction("domain.lint");
    const { store } = await opened(fake);
    await store.save();
    store.close();
    slow.release();
    await flushPromises();
    expect(store.snapshot().phase).toBe("idle");
  });

  it("is not asked for when the composition is refused", async () => {
    const fake = bridge();
    fake.actions.set("domain.compose", failed(cannedFailure("composition_invalid", 422)));
    const { store } = await opened(fake);
    await store.save();
    expect(fake.calls.some((call) => call.name === "domain.lint")).toBe(false);
  });
});
