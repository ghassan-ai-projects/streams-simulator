import { describe, expect, it } from "vitest";
import { fakeClock } from "./fakeClock";

describe("the fake clock", () => {
  it("runs a scheduled task only when its time arrives, and cancels it", () => {
    const clock = fakeClock(1000);
    const fired: string[] = [];
    clock.schedule(250, () => {
      fired.push("kept");
    });
    const cancel = clock.schedule(500, () => {
      fired.push("cancelled");
    });
    clock.advance(250);
    expect(fired).toEqual(["kept"]);
    cancel();
    clock.advance(500);
    expect(fired).toEqual(["kept"]);
  });

  it("resolves a delay when the clock reaches it and reports what is pending", async () => {
    const clock = fakeClock();
    const order: string[] = [];
    const waiting = clock.delay(100).then(() => {
      order.push("delayed");
    });
    expect(clock.pending()).toBe(1);
    clock.advance(99);
    await Promise.resolve();
    expect(order).toEqual([]);
    clock.advance(1);
    await waiting;
    expect(order).toEqual(["delayed"]);
    expect(clock.pending()).toBe(0);
  });

  it("rejects a delay whose signal is already aborted, as an AbortError", async () => {
    const clock = fakeClock();
    const controller = new AbortController();
    controller.abort();
    await expect(clock.delay(10, controller.signal)).rejects.toMatchObject({ name: "AbortError" });
    expect(clock.pending()).toBe(0);
  });

  it("rejects a delay aborted while it waits, and runs the delay that was not aborted", async () => {
    const clock = fakeClock();
    const controller = new AbortController();
    const aborted = clock.delay(50, controller.signal);
    const kept = clock.delay(50);
    expect(clock.pending()).toBe(2);
    clock.advance(10);
    controller.abort();
    await expect(aborted).rejects.toMatchObject({ name: "AbortError" });
    expect(clock.pending()).toBe(1);
    clock.advance(40);
    await kept;
    expect(clock.pending()).toBe(0);
  });

  it("does not run a scheduled task that was cancelled after its time", () => {
    const clock = fakeClock();
    const fired: string[] = [];
    const cancel = clock.schedule(10, () => {
      fired.push("late");
    });
    clock.advance(10);
    cancel();
    expect(fired).toEqual(["late"]);
    expect(clock.pending()).toBe(0);
  });

  it("runs tasks in time order even when one schedules another", () => {
    const clock = fakeClock();
    const order: string[] = [];
    clock.schedule(20, () => {
      order.push("second");
    });
    clock.schedule(10, () => {
      order.push("first");
      clock.schedule(5, () => {
        order.push("nested");
      });
    });
    clock.advance(30);
    expect(order).toEqual(["first", "nested", "second"]);
    expect(clock.nowMs()).toBe(30);
  });
});
