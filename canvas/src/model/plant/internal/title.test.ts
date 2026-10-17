import { describe, expect, it } from "vitest";
import { plantTitle } from "./title";

describe("plantTitle", () => {
  it("prefers the title and falls back to the id", () => {
    expect(plantTitle({ title: "Boiler hall", id: "boiler" })).toBe("Boiler hall");
    expect(plantTitle({ title: "", id: "boiler" })).toBe("boiler");
  });
});
