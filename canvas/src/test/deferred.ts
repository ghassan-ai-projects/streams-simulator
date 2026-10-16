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

/** Waits for the microtask queue to drain twice, which is what a store's await chain needs before it publishes. */
export async function settle(): Promise<void> {
  await Promise.resolve();
  await Promise.resolve();
}
