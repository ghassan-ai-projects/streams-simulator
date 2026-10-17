import { useSyncExternalStore } from "react";

interface Readable<S> {
  readonly snapshot: () => S;
  readonly subscribe: (listener: () => void) => () => void;
}

export function useStore<S>(store: Readable<S>): S {
  return useSyncExternalStore(store.subscribe, store.snapshot);
}
