import { describe, expect, it } from "vitest";
import { Observable } from "./observable";

class Counter extends Observable<{ readonly n: number }> {
  put(next: { readonly n: number }): void {
    this.set(next);
  }
}

describe("observable", () => {
  it("notifies on a new state and stays silent when the same state is set again", () => {
    const counter = new Counter({ n: 0 });
    let heard = 0;
    const stop = counter.subscribe(() => {
      heard += 1;
    });
    const next = { n: 1 };
    counter.put(next);
    counter.put(next);
    expect(heard).toBe(1);
    stop();
    counter.put({ n: 2 });
    expect(heard).toBe(1);
    expect(counter.snapshot()).toEqual({ n: 2 });
  });
});
