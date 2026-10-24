import type { JobStatus } from "../../../model/launch";
import type { LaunchFailure, Outcome, Scheduler } from "../../ports";

type Reading = Outcome<JobStatus, LaunchFailure>;

const pollMs = 500;

export class StatusPoll {
  private cancel: (() => void) | null = null;
  private live = true;

  constructor(
    private readonly scheduler: Scheduler,
    private readonly read: () => Promise<Reading>,
    private readonly take: (reading: Reading) => boolean,
  ) {}

  begin(): void {
    this.arm();
  }

  end(): void {
    this.live = false;
    this.cancel?.();
    this.cancel = null;
  }

  private arm(): void {
    if (!this.live) return;
    this.cancel = this.scheduler.schedule(pollMs, () => {
      void this.poll();
    });
  }

  private async poll(): Promise<void> {
    this.cancel = null;
    const reading = await this.read();
    if (this.live && this.take(reading)) this.arm();
  }
}
