import type { Clock, Scheduler } from "../../ports";

const frameMs = 100;

export class FrameLoop {
  private cancel: (() => void) | null = null;
  private running = false;

  constructor(
    private readonly clock: Clock,
    private readonly scheduler: Scheduler,
  ) {}

  start(frame: (elapsedMs: number) => void): void {
    if (this.running) return;
    this.running = true;
    this.arm(this.clock.nowMs(), frame);
  }

  stop(): void {
    this.running = false;
    this.cancel?.();
    this.cancel = null;
  }

  private arm(last: number, frame: (elapsedMs: number) => void): void {
    this.cancel = this.scheduler.schedule(frameMs, () => {
      const now = this.clock.nowMs();
      frame(now - last);
      if (this.running) this.arm(now, frame);
    });
  }
}
