/** Implements the Clock and Scheduler signatures of the accepted architecture (A5) with a controlled queue. */
export interface FakeClock {
  nowMs(): number;
  schedule(delayMs: number, task: () => void): () => void;
  delay(delayMs: number, signal?: AbortSignal): Promise<void>;
  advance(ms: number): void;
  pending(): number;
}

interface Task {
  readonly atMs: number;
  readonly run: () => void;
}

interface Queue {
  now: number;
  tasks: Task[];
}

function enqueue(queue: Queue, delayMs: number, run: () => void): Task {
  const entry = { atMs: queue.now + delayMs, run };
  queue.tasks = [...queue.tasks, entry];
  return entry;
}

function drop(queue: Queue, entry: Task): void {
  queue.tasks = queue.tasks.filter((task) => task !== entry);
}

function advanceTo(queue: Queue, target: number): void {
  for (;;) {
    const next = queue.tasks.filter((task) => task.atMs <= target).sort((a, b) => a.atMs - b.atMs)[0];
    if (next === undefined) break;
    drop(queue, next);
    queue.now = Math.max(queue.now, next.atMs);
    next.run();
  }
  queue.now = target;
}

function abortError(): Error {
  return new DOMException("the delay was aborted", "AbortError");
}

function delayOn(queue: Queue, delayMs: number, signal?: AbortSignal): Promise<void> {
  if (signal?.aborted === true) return Promise.reject(abortError());
  return new Promise<void>((resolve, reject) => {
    const entry = enqueue(queue, delayMs, () => {
      signal?.removeEventListener("abort", onAbort);
      resolve();
    });
    const onAbort = (): void => {
      drop(queue, entry);
      reject(abortError());
    };
    signal?.addEventListener("abort", onAbort, { once: true });
  });
}

export function fakeClock(startMs = 0): FakeClock {
  const queue: Queue = { now: startMs, tasks: [] };
  const enqueueIn = (delayMs: number, run: () => void): Task => enqueue(queue, delayMs, run);
  const dropFrom = (entry: Task): void => {
    drop(queue, entry);
  };
  return {
    nowMs: () => queue.now,
    schedule(delayMs, task) {
      const entry = enqueueIn(delayMs, task);
      return () => {
        dropFrom(entry);
      };
    },
    delay: (delayMs, signal) => delayOn(queue, delayMs, signal),
    advance: (ms) => {
      advanceTo(queue, queue.now + ms);
    },
    pending: () => queue.tasks.length,
  };
}
