import { describe, expect, it } from "vitest";
import type { PreferenceStore, SystemTheme } from "../../ports";
import { ThemeStore } from "./themeStore";

function memory(initial: string | null): PreferenceStore & { value: string | null } {
  return {
    value: initial,
    read() {
      return this.value;
    },
    write(_key, value) {
      this.value = value;
    },
  };
}

function system(dark: boolean): SystemTheme & { set: (dark: boolean) => void } {
  let current = dark;
  const listeners = new Set<() => void>();
  return {
    prefersDark: () => current,
    subscribe(listener) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    set(next) {
      current = next;
      listeners.forEach((listener) => {
        listener();
      });
    },
  };
}

describe("ThemeStore", () => {
  it("starts from the stored preference and persists each change", () => {
    const stored = memory("dark");
    const store = new ThemeStore(stored, system(false));
    expect(store.snapshot()).toEqual({ preference: "dark", theme: "dark" });
    store.cycle();
    expect(store.snapshot()).toEqual({ preference: "system", theme: "light" });
    expect(stored.value).toBe("system");
  });

  it("follows the system theme while the preference is system and tells its listeners", () => {
    const os = system(false);
    const store = new ThemeStore(memory(null), os);
    let calls = 0;
    const unsubscribe = store.subscribe(() => calls++);
    os.set(true);
    expect(store.snapshot().theme).toBe("dark");
    unsubscribe();
    os.set(false);
    expect(calls).toBe(1);
  });
});
