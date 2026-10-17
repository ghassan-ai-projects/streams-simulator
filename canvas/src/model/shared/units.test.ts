import { describe, expect, it } from "vitest";
import { grouped } from "./units";

describe("units", () => {
  it("groups thousands with a thin space", () => {
    expect([grouped(7), grouped(14_483), grouped(1_234_567)]).toEqual(["7", "14 483", "1 234 567"]);
  });
});
