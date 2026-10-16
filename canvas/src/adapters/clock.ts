import type { Clock, Scheduler } from "../app/clock";

function abortError(): Error {
  return new DOMException("the delay was aborted", "AbortError");
}

function delayOnBrowser(delayMs: number, signal?: AbortSignal): Promise<void> {
  if (signal?.aborted === true) return Promise.reject(abortError());
  return new Promise<void>((resolve, reject) => {
    const timer = window.setTimeout(() => {
      signal?.removeEventListener("abort", onAbort);
      resolve();
    }, delayMs);
    const onAbort = (): void => {
      window.clearTimeout(timer);
      reject(abortError());
    };
    signal?.addEventListener("abort", onAbort, { once: true });
  });
}

export function browserClock(): Clock & Scheduler {
  return {
    nowMs: () => Date.now(),
    schedule(delayMs, task) {
      const timer = window.setTimeout(task, delayMs);
      return () => window.clearTimeout(timer);
    },
    delay: delayOnBrowser,
  };
}
