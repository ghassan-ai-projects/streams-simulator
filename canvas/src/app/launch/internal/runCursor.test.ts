import { describe, expect, it } from "vitest";
import { RunCursor } from "./runCursor";

const plan = { seed: 1, durationS: 600, explore: true, faults: [], calls: [] };

describe("run cursor", () => {
  it("reports no speed until a second of wall time has been sampled, then simulated seconds per second", () => {
    const cursor = new RunCursor({ domain: "d", title: "D", path: "p" }, plan, 0);
    expect(cursor.sampleRate(0, 0)).toBeNull();
    expect(cursor.sampleRate(500, 30)).toBeNull();
    expect(cursor.sampleRate(2000, 120)).toBe(60);
  });

  it("forgets samples older than five seconds", () => {
    const cursor = new RunCursor({ domain: "d", title: "D", path: "p" }, plan, 0);
    cursor.sampleRate(0, 0);
    cursor.sampleRate(5000, 500);
    expect(cursor.sampleRate(7000, 620)).toBe(60);
  });
});
