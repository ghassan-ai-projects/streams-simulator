import { describe, expect, it } from "vitest";
import { holdsOperation } from "./holdsOperation";

const live = { live: () => true };
const stale = { live: () => false };

describe("holdsOperation", () => {
  it("holds a live operation on the open path", () => {
    expect(holdsOperation(live, { phase: "ready", path: "a.json" }, "a.json")).toBe(true);
  });

  it("lets go when stale, idle or on another path", () => {
    expect(holdsOperation(stale, { phase: "ready", path: "a.json" }, "a.json")).toBe(false);
    expect(holdsOperation(live, { phase: "idle" }, "a.json")).toBe(false);
    expect(holdsOperation(live, { phase: "ready", path: "b.json" }, "a.json")).toBe(false);
  });
});
