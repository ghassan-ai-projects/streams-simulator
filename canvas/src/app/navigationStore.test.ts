import { describe, expect, it } from "vitest";
import { NavigationStore } from "./navigationStore";
import type { Location } from "./ports";

function memoryLocation(initial: string): Location & { hash: string; replaced: boolean; fire: () => void } {
  const listeners = new Set<() => void>();
  return {
    hash: initial,
    replaced: false,
    read() {
      return this.hash;
    },
    write(hash, replace = false) {
      this.hash = hash;
      this.replaced = replace;
    },
    subscribe(listener) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    fire() {
      listeners.forEach((listener) => {
        listener();
      });
    },
  };
}

describe("navigation", () => {
  it("starts from the address and writes every move to it", () => {
    const location = memoryLocation("#/run?world=w-2");
    const navigation = new NavigationStore(location);
    expect(navigation.snapshot()).toEqual({ screen: "run", worldId: "w-2" });
    navigation.go({ screen: "plant", domain: "domains/a.domain.json", hall: null });
    expect(location.hash).toBe("#/plant?domain=domains%2Fa.domain.json");
    expect(location.replaced).toBe(false);
  });

  it("replaces the address without a new history entry when asked", () => {
    const location = memoryLocation("#/plant?domain=d&run=1");
    const navigation = new NavigationStore(location);
    navigation.replace({ screen: "plant", domain: "d", hall: null });
    expect([location.hash, location.replaced]).toEqual(["#/plant?domain=d", true]);
    expect(navigation.snapshot()).toEqual({ screen: "plant", domain: "d", hall: null });
  });

  it("follows the address when the user goes back", () => {
    const location = memoryLocation("#/");
    const navigation = new NavigationStore(location);
    location.hash = "#/plant?domain=d&hall=h01";
    location.fire();
    expect(navigation.snapshot()).toEqual({ screen: "plant", domain: "d", hall: "h01" });
  });
});
