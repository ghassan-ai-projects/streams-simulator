import { describe, expect, it } from "vitest";
import { failure, success } from "../../ports";
import { cannedFailure } from "../../../test/fakeFailures";
import { deferred, flushPromises } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import { deliveryKinds, miniCategories, teachCatalog } from "../../../test/planCatalogFixture";
import { PlanCatalog } from "./planCatalog";

describe("the plan catalog", () => {
  it("loads the delivery conditions and the fault categories of a plant, once", async () => {
    const fake = new FakeBridge();
    teachCatalog(fake);
    const catalog = new PlanCatalog(fake.launch);
    expect(catalog.snapshot()).toEqual({
      perturbations: { status: "loading" },
      faultCategories: { status: "loading" },
    });
    catalog.load("mini");
    await flushPromises();
    expect(catalog.snapshot()).toEqual({
      perturbations: { status: "ready", value: deliveryKinds },
      faultCategories: { status: "ready", value: miniCategories },
    });
    catalog.load("mini");
    catalog.load("other");
    await flushPromises();
    const asked = fake.calls.map((call) => call.name);
    expect(asked.filter((name) => name === "catalog.perturbations")).toHaveLength(1);
    expect(fake.calls.filter((call) => call.name === "catalog.describe").map((call) => call.args.domain)).toEqual([
      "mini",
      "other",
    ]);
  });

  it("shows the categories of the plant asked for last, whichever answer comes back first", async () => {
    const fake = new FakeBridge();
    teachCatalog(fake);
    const first = deferred<ReturnType<typeof success<object>>>();
    fake.handlers.set("catalog.describe", (args) =>
      args.domain === "slow" ? first.promise : success({ fault_categories: { b: { f: "world" } } }),
    );
    const catalog = new PlanCatalog(fake.launch);
    catalog.load("slow");
    catalog.load("quick");
    await flushPromises();
    first.resolve(success({ fault_categories: { a: { f: "measurement" } } }));
    await flushPromises();
    expect(catalog.snapshot().faultCategories).toEqual({ status: "ready", value: { b: { f: "world" } } });
  });

  it("says what went wrong and asks again on the next load", async () => {
    const fake = new FakeBridge();
    fake.actions.set("catalog.perturbations", failure(cannedFailure("boom", 500, "the director is away")));
    fake.actions.set("catalog.describe", failure(cannedFailure("domain_invalid", 422, "no such plant")));
    const catalog = new PlanCatalog(fake.launch);
    catalog.load("mini");
    await flushPromises();
    expect(catalog.snapshot()).toEqual({
      perturbations: { status: "unavailable", words: "the director is away" },
      faultCategories: { status: "unavailable", words: "no such plant" },
    });
    teachCatalog(fake);
    catalog.load("mini");
    await flushPromises();
    expect(catalog.snapshot().perturbations.status).toBe("ready");
    expect(catalog.snapshot().faultCategories.status).toBe("ready");
  });
});
