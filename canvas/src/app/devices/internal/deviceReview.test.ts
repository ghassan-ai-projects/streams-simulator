import { describe, expect, it, vi } from "vitest";
import { changeMetric } from "../../../model/devices";
import { failure as failed } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import { bridge, checks, fan, opened, ready, refusal } from "../../../test/deviceSessionHarness";
import { kindCheck } from "../../../test/kindCheckFake";
import { interfaceIssues, twinNeedsOf } from "./deviceSelectors";

describe("what a twin would need, asked while the device is edited", () => {
  it("lists every need at once, placed on the reading it names, and none once they are met", async () => {
    const { store, fake } = await opened(bridge({ ...fan, metrics: [{ name: "duty", data_type: "number" }] }));
    expect(twinNeedsOf(ready(store)).map((issue) => `${issue.key}:${issue.field}:${issue.text}`)).toEqual([
      "m1:unit:duty needs a unit",
      "m1:resolution:Needs a resolution",
    ]);
    expect(ready(store).review.schema).toEqual([]);
    store.edit((kind) => changeMetric(kind, "m1", { unit: "%", resolution: 1 }));
    await vi.waitFor(() => {
      expect(twinNeedsOf(ready(store))).toEqual([]);
      expect(ready(store).review.state).toBe("done");
    });
    expect(checks(fake, true)).toHaveLength(2);
  });

  it("separates what the schema refuses from what only a twin needs", async () => {
    const { store } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { name: "bad name", unit: "" }));
    await vi.waitFor(() => {
      expect(ready(store).review.schema).toHaveLength(1);
    });
    const draft = ready(store);
    expect(draft.review.schema.map((issue) => issue.text)).toEqual(["Use letters, digits and . _ : - only"]);
    expect(draft.review.needs.map((issue) => issue.text)).toEqual(["bad name needs a unit"]);
    expect(interfaceIssues(draft).map((issue) => issue.field)).toEqual(["name"]);
  });

  it("asks only once at a time and again for the latest draft when edits arrive meanwhile", async () => {
    const { store, fake } = await opened();
    const slow = deferred<ReturnType<typeof kindCheck>>();
    fake.handlers.set("kind.check", () => slow.promise);
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    await vi.waitFor(() => {
      expect(checks(fake, true)).toHaveLength(2);
    });
    expect(ready(store).review.state).toBe("checking");
    store.edit((kind) => changeMetric(kind, "m1", { unit: "kPa" }));
    store.edit((kind) => changeMetric(kind, "m1", { unit: "bar" }));
    fake.handlers.set("kind.check", kindCheck);
    slow.resolve(kindCheck({ content: JSON.stringify(fan), twin: true }));
    await vi.waitFor(() => {
      expect(checks(fake, true)).toHaveLength(3);
      expect(ready(store).review.state).toBe("done");
    });
    const asked = checks(fake, true).map(
      (call) => JSON.parse(String(call.args.content)) as { metrics: { unit: string }[] },
    );
    expect(asked.map((document) => document.metrics[0]?.unit)).toEqual(["%", "Pa", "bar"]);
    expect(ready(store).review.state).toBe("done");
  });

  it("does not ask while a name is blank, and says nothing until the draft is whole again", async () => {
    const { store, fake } = await opened();
    store.edit((kind) => changeMetric(kind, "m1", { name: " " }));
    await vi.waitFor(() => {
      expect(ready(store).review).toEqual({ state: "idle", schema: [], needs: [] });
    });
    expect(checks(fake, true)).toHaveLength(1);
    store.revert();
    await vi.waitFor(() => {
      expect(ready(store).review.state).toBe("done");
    });
    expect(checks(fake, true)).toHaveLength(2);
  });

  it("says it could not ask when the binary is unreachable or refuses without naming an issue", async () => {
    const { store, fake } = await opened();
    fake.handlers.set("kind.check", () => failed(cannedFailure("bridge_failed", 500, "down")));
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    await vi.waitFor(() => {
      expect(ready(store).review).toEqual({ state: "failed", schema: [], needs: [] });
    });
    fake.handlers.set("kind.check", () => failed(cannedFailure("kind_failed", 422, "refused")));
    store.edit((kind) => changeMetric(kind, "m1", { unit: "bar" }));
    await vi.waitFor(() => {
      expect(checks(fake, true)).toHaveLength(3);
      expect(ready(store).review.state).toBe("failed");
    });
  });

  it("drops an answer that arrives after the session closed", async () => {
    const { store, fake } = await opened();
    const slow = deferred<ReturnType<typeof kindCheck>>();
    fake.handlers.set("kind.check", () => slow.promise);
    store.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    await vi.waitFor(() => {
      expect(checks(fake, true)).toHaveLength(2);
    });
    store.close();
    slow.resolve(refusal("/kind", "late", "pattern"));
    await flushPromises();
    expect(store.snapshot()).toEqual({ phase: "idle" });
  });
});
