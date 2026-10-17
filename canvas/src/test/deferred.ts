export interface Deferred<T> {
  readonly promise: Promise<T>;
  resolve(value: T): void;
  reject(reason: unknown): void;
}

export function deferred<T>(): Deferred<T> {
  let settle: ((value: T) => void) | null = null;
  let fail: ((reason: unknown) => void) | null = null;
  const promise = new Promise<T>((resolve, reject) => {
    settle = resolve;
    fail = reject;
  });
  return {
    promise,
    resolve: (value) => {
      settle?.(value);
    },
    reject: (reason) => {
      fail?.(reason);
    },
  };
}

/** Drains the microtask queue: a store's await chain plus the double's projection hop per gateway call. */
export async function settle(): Promise<void> {
  for (let tick = 0; tick < 5; tick += 1) await Promise.resolve();
}
