import { describe, expect, it, vi } from "vitest";
import type { JobStatus } from "../../../model/launch";
import { success } from "../../ports";
import { deferred, flushPromises } from "../../../test/deferred";
import { fakeClock } from "../../../test/fakeClock";
import { StatusPoll } from "./statusPoll";

const status: JobStatus = {
  worldId: "w-1",
  runId: "r-1",
  domain: "d",
  seed: 1,
  state: "running",
  startNs: 0,
  reachedNs: 0,
  requestedEndNs: 60e9,
  rate: 0,
  explore: true,
  calls: [],
  failure: "",
  name: "",
  purpose: "",
  faults: [],
  perturbations: [],
};

describe("status poll", () => {
  it("reads twice a second, one reading at a time, for as long as it is told to go on", async () => {
    const timers = fakeClock();
    const read = vi.fn(() => Promise.resolve(success(status)));
    const take = vi.fn(() => true);
    new StatusPoll(timers, read, take).begin();
    timers.advance(499);
    expect(read).not.toHaveBeenCalled();
    timers.advance(1);
    await flushPromises();
    expect(take).toHaveBeenCalledTimes(1);
    timers.advance(500);
    await flushPromises();
    expect(read).toHaveBeenCalledTimes(2);
  });

  it("stops when the reading says to stop", async () => {
    const timers = fakeClock();
    new StatusPoll(
      timers,
      () => Promise.resolve(success(status)),
      () => false,
    ).begin();
    timers.advance(500);
    await flushPromises();
    expect(timers.pending()).toBe(0);
  });

  it("cancels the wait when ended, and drops a reading still on its way", async () => {
    const timers = fakeClock();
    const waiting = new StatusPoll(
      timers,
      () => Promise.resolve(success(status)),
      () => true,
    );
    waiting.begin();
    waiting.end();
    expect(timers.pending()).toBe(0);
    const gate = deferred<ReturnType<typeof success<JobStatus>>>();
    const take = vi.fn(() => true);
    const reading = new StatusPoll(timers, () => gate.promise, take);
    reading.begin();
    timers.advance(500);
    reading.end();
    gate.resolve(success(status));
    await flushPromises();
    expect(take).not.toHaveBeenCalled();
    expect(timers.pending()).toBe(0);
  });

  it("goes on without taking a reading that was set aside", async () => {
    const timers = fakeClock();
    const take = vi.fn(() => true);
    new StatusPoll(timers, () => Promise.resolve(null), take).begin();
    timers.advance(500);
    await flushPromises();
    expect(take).not.toHaveBeenCalled();
    expect(timers.pending()).toBe(1);
  });
});
