import { beforeEach, describe, expect, it, vi } from "vitest";
import { stubFlowDom } from "../../../test/dom";
import { createReviewSession } from "../../../app/review";
import type { ReviewSession } from "../../../app/review";
import { factoryBridge, factoryJournal } from "../../../test/factoryFixture";
import { fakeClock } from "../../../test/fakeClock";
import { mountLivePlant, watchedRun } from "../../../test/livePlantHarness";
import * as graph from "../../../model/graph";
import { watchSix } from "../../../test/watchSix";

beforeEach(() => {
  stubFlowDom();
});

const measure = import.meta.env.VITE_PERF === "1" ? describe : describe.skip;

function ingest(open: (review: ReviewSession) => void, minutes: number) {
  const bridge = factoryBridge();
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  open(review);
  const entries = factoryJournal(0, minutes * 60);
  const started = performance.now();
  bridge.stream(watchedRun, entries, null);
  timers.advance(250);
  const took = performance.now() - started;
  const line = `PERF ingest minutes=${String(minutes)} observations=${String(review.index.size)} held=${String(review.index.held)} ms=${took.toFixed(0)}`;
  return { line, records: review.index.size };
}

measure("the live plant under the factory load (PERF=1 prints numbers)", () => {
  it.each([30, 360, 1440])("prints the ingest cost of %d minutes of the factory in one batch", (minutes) => {
    const { line, records } = ingest((review) => {
      review.watch(watchedRun);
    }, minutes);
    console.info(line);
    expect(records).toBe(minutes * 6 * 82);
  });

  it.each([false, true])("prints ingest and update cost (six watched channels: %s)", async (trends) => {
    const nodes = vi.spyOn(graph, "describeNode");
    const live = await mountLivePlant((review, worldId) => {
      review.watch(worldId);
    });
    if (trends) watchSix(live.review);
    const perFlush: number[] = [];
    const wall: number[] = [];
    for (let second = 0; second < 1800; second += 10) {
      const before = nodes.mock.calls.length;
      const started = performance.now();
      live.feed(second, second + 10);
      wall.push(performance.now() - started);
      perFlush.push(nodes.mock.calls.length - before);
    }
    const total = perFlush.reduce((sum, count) => sum + count, 0);
    const sorted = [...wall].sort((a, b) => a - b);
    console.info(
      `PERF trends=${String(trends)} flushes=${String(perFlush.length)} nodeRenders=${String(total)} perFlushMedian=${String(
        [...perFlush].sort((a, b) => a - b)[perFlush.length >> 1],
      )} commits=${String(live.profile.commits)} phases=${JSON.stringify(live.profile.phases)} actualMs=${live.profile.actualMs.toFixed(0)} baseMs=${live.profile.baseMs.toFixed(0)} flushMsMedian=${(sorted[sorted.length >> 1] ?? 0).toFixed(1)} flushMsP95=${(sorted[Math.floor(sorted.length * 0.95)] ?? 0).toFixed(1)} records=${String(live.review.index.size)}`,
    );
    expect(live.review.index.size).toBeGreaterThan(14000);
  });
});
