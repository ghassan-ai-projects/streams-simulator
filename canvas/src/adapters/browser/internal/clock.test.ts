import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { browserClock } from "./clock";

describe("browser clock", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("reads the current time in milliseconds", () => {
    vi.setSystemTime(new Date("2026-10-17T10:00:00Z"));
    expect(browserClock().nowMs()).toBe(Date.parse("2026-10-17T10:00:00Z"));
    vi.advanceTimersByTime(1500);
    expect(browserClock().nowMs()).toBe(Date.parse("2026-10-17T10:00:01.500Z"));
  });

  it("runs a scheduled task once after its delay", () => {
    const task = vi.fn();
    browserClock().schedule(100, task);
    vi.advanceTimersByTime(99);
    expect(task).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    vi.advanceTimersByTime(1000);
    expect(task).toHaveBeenCalledOnce();
  });

  it("never runs a task that was cancelled", () => {
    const task = vi.fn();
    const cancel = browserClock().schedule(100, task);
    cancel();
    vi.advanceTimersByTime(1000);
    expect(task).not.toHaveBeenCalled();
  });

  it("resolves a delay once its time has passed", async () => {
    let done = false;
    const pending = browserClock()
      .delay(50)
      .then(() => {
        done = true;
      });
    await vi.advanceTimersByTimeAsync(49);
    expect(done).toBe(false);
    await vi.advanceTimersByTimeAsync(1);
    await pending;
    expect(done).toBe(true);
  });

  it("rejects an aborted delay with an AbortError and stops its timer", async () => {
    const controller = new AbortController();
    const pending = browserClock().delay(50, controller.signal);
    controller.abort();
    await expect(pending).rejects.toMatchObject({ name: "AbortError" });
    expect(vi.getTimerCount()).toBe(0);
  });

  it("rejects at once when the signal was already aborted", async () => {
    const controller = new AbortController();
    controller.abort();
    await expect(browserClock().delay(50, controller.signal)).rejects.toMatchObject({ name: "AbortError" });
    expect(vi.getTimerCount()).toBe(0);
  });

  it("detaches its abort listener when the delay finishes", async () => {
    const controller = new AbortController();
    const pending = browserClock().delay(10, controller.signal);
    await vi.advanceTimersByTimeAsync(10);
    await pending;
    controller.abort();
    await expect(pending).resolves.toBeUndefined();
  });
});
