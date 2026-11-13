import { useStoreApi } from "@xyflow/react";
import { useSyncExternalStore } from "react";

type Listener = () => void;

class HiddenFamilies {
  private value: ReadonlySet<string> = new Set();
  private readonly listeners = new Set<Listener>();

  readonly get = (): ReadonlySet<string> => this.value;

  readonly subscribe = (listener: Listener): (() => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };

  readonly toggle = (family: string): void => {
    const next = new Set(this.value);
    if (!next.delete(family)) next.add(family);
    this.value = next;
    this.listeners.forEach((listener) => {
      listener();
    });
  };
}

const byGraph = new WeakMap<object, HiddenFamilies>();

function filterOf(graph: object): HiddenFamilies {
  const known = byGraph.get(graph);
  if (known !== undefined) return known;
  const made = new HiddenFamilies();
  byGraph.set(graph, made);
  return made;
}

export function useHiddenFamilies(): {
  readonly hidden: ReadonlySet<string>;
  readonly toggle: (family: string) => void;
} {
  const filter = filterOf(useStoreApi().getState);
  return { hidden: useSyncExternalStore(filter.subscribe, filter.get), toggle: filter.toggle };
}
