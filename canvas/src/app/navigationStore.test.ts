import { describe, expect, it } from "vitest";
import { NavigationStore } from "./navigationStore";
import type { Location } from "./ports";

function memoryLocation(initial: string): Location & { hash: string; fire: () => void } {
  const listeners = new Set<() => void>();
  return {
    hash: initial,
    read() {
      return this.hash;
    },
    write(hash) {
      this.hash = hash;
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
  });

  it("follows the address when the user goes back", () => {
    const location = memoryLocation("#/");
    const navigation = new NavigationStore(location);
    location.hash = "#/plant?domain=d&hall=h01";
    location.fire();
    expect(navigation.snapshot()).toEqual({ screen: "plant", domain: "d", hall: "h01" });
  });
});
