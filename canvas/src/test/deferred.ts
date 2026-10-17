export interface Deferred<T> {
  readonly promise: Promise<T>;
  resolve(value: T): void;
  reject(reason: unknown): void;
}

export function deferred<T>(): Deferred<T> {
  let fulfil: ((value: T) => void) | null = null;
  let fail: ((reason: unknown) => void) | null = null;
  const promise = new Promise<T>((resolve, reject) => {
    fulfil = resolve;
    fail = reject;
  });
  return {
    promise,
    resolve: (value) => {
      fulfil?.(value);
    },
    reject: (reason) => {
      fail?.(reason);
    },
  };
}

export async function flushPromises(): Promise<void> {
  for (let tick = 0; tick < 30; tick += 1) await Promise.resolve();
}
