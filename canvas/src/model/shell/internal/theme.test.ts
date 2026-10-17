import { describe, expect, it } from "vitest";
import { nextPreference, parsePreference, preferenceLabel, resolveTheme } from "./theme";

describe("theme", () => {
  it("follows the system only when the preference says so", () => {
    expect(resolveTheme("system", true)).toBe("dark");
    expect(resolveTheme("system", false)).toBe("light");
    expect(resolveTheme("light", true)).toBe("light");
    expect(resolveTheme("dark", false)).toBe("dark");
  });

  it("cycles system, light, dark and back", () => {
    expect(nextPreference("system")).toBe("light");
    expect(nextPreference("light")).toBe("dark");
    expect(nextPreference("dark")).toBe("system");
  });

  it("reads an unknown or absent stored value as system", () => {
    expect(parsePreference(null)).toBe("system");
    expect(parsePreference("purple")).toBe("system");
    expect(parsePreference("dark")).toBe("dark");
  });

  it("labels every preference", () => {
    expect(["system", "light", "dark"].map((p) => preferenceLabel(parsePreference(p)))).toEqual([
      "Theme: system",
      "Theme: light",
      "Theme: dark",
    ]);
  });
});
