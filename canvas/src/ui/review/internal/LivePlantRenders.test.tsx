import type { MockInstance } from "vitest";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as graph from "../../../model/graph";
import { clockOnly, factoryJournal, singleReadings } from "../../../test/factoryFixture";
import { stubFlowDom } from "../../../test/dom";
import { mountLivePlant } from "../../../test/livePlantHarness";

beforeEach(() => {
  stubFlowDom();
});

async function settled() {
  const renders = vi.spyOn(graph, "describeNode");
  const live = await mountLivePlant((review, worldId) => {
    review.watch(worldId);
  });
  live.feed(0, 600);
  return { live, renders, since: () => renders.mock.calls.length };
}

function rendered(renders: MockInstance<typeof graph.describeNode>, before: number): string[] {
  return renders.mock.calls.slice(before).map((call) => call[0].id);
}

describe("what a flush redraws on the factory plant", () => {
  it("leaves every node alone when the flush brings the clock and nothing else", async () => {
    const { live, renders, since } = await settled();
    const before = since();
    live.send([clockOnly(660)]);
    live.send([clockOnly(720)]);
    expect(rendered(renders, before)).toEqual([]);
  });

  it("redraws only the instance that reported, not the other twenty-two", async () => {
    const { live, renders, since } = await settled();
    const before = since();
    live.send(singleReadings("CT-1", 600, 610));
    const drawn = new Set(rendered(renders, before));
    expect(drawn).toEqual(new Set(["CT-1"]));
    expect(rendered(renders, before).length).toBeLessThanOrEqual(2);
  });

  it("redraws each instance once, and only those whose readings changed, when nearly all of them reported", async () => {
    const { live, renders, since } = await settled();
    const before = since();
    live.send(factoryJournal(600, 610));
    const drawn = rendered(renders, before);
    expect(new Set(drawn).size).toBeGreaterThanOrEqual(20);
    expect(drawn).toHaveLength(new Set(drawn).size);
  });
});
