export interface Clock {
  nowMs(): number;
}

export interface Scheduler {
  schedule(delayMs: number, task: () => void): () => void;
  delay(delayMs: number, signal?: AbortSignal): Promise<void>;
}
