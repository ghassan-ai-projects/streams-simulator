import { describe, expect, it, vi } from "vitest";
import { localPreferences, mediaSystemTheme } from "..";

describe("browser adapters", () => {
  it("reads and writes preferences and survives a storage that throws", () => {
    const store = localPreferences(window.localStorage);
    store.write("k", "dark");
    expect(store.read("k")).toBe("dark");
    const broken = {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      },
    } as unknown as Storage;
    expect(localPreferences(broken).read("k")).toBeNull();
    expect(() => {
      localPreferences(broken).write("k", "dark");
    }).not.toThrow();
  });

  it("reports the media query and forwards its changes until unsubscribed", () => {
    const add = vi.fn();
    const remove = vi.fn();
    const media = { matches: true, addEventListener: add, removeEventListener: remove } as unknown as MediaQueryList;
    const theme = mediaSystemTheme(media);
    const listener = vi.fn();
    const unsubscribe = theme.subscribe(listener);
    expect(theme.prefersDark()).toBe(true);
    expect(add).toHaveBeenCalledWith("change", listener);
    unsubscribe();
    expect(remove).toHaveBeenCalledWith("change", listener);
  });
});
