import { describe, expect, it } from "vitest";
import { status } from "./status";

describe("status", () => {
  it("defaults the detail to null", () => {
    expect(status("Saved", "ok")).toEqual({ text: "Saved", detail: null, tone: "ok" });
  });

  it("keeps a given detail", () => {
    expect(status("Not saved", "warn", "why")).toEqual({ text: "Not saved", detail: "why", tone: "warn" });
  });
});
