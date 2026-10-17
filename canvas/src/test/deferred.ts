export interface Deferred<T> {
  readonly promise: Promise<T>;
  resolve(value: T): void;
}

export function deferred<T>(): Deferred<T> {
  let fulfil: ((value: T) => void) | null = null;
  const promise = new Promise<T>((resolve) => {
    fulfil = resolve;
  });
  return {
    promise,
    resolve: (value) => {
      fulfil?.(value);
    },
  };
}

export async function flushPromises(): Promise<void> {
  for (let tick = 0; tick < 30; tick += 1) await Promise.resolve();
}
