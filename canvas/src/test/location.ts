import type { Location } from "../app/ports";

export interface MemoryLocation extends Location {
  hash: string;
  replaced: boolean;
  fire(): void;
}

export function memoryLocation(initial = "#/"): MemoryLocation {
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
