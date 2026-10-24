import type { Clipboard, Clock, ReviewGateway, Scheduler } from "../../ports";
import { createOperationIdentity } from "../../shared";
import type { ReviewSession } from "./api";
import type { ControlSession } from "./controlStore";
import { ControlStore } from "./controlStore";
import { ReviewStore } from "./reviewStore";
import type { TwinSession } from "./twinStore";
import { TwinStore } from "./twinStore";
import type { WatchSession } from "./watchStore";
import { WatchStore } from "./watchStore";

const noClipboard: Clipboard = { copy: () => Promise.resolve(false) };

class ControlledReview extends ReviewStore implements ReviewSession {
  readonly controls: ControlSession;
  readonly twin: TwinSession;
  readonly watchList: WatchSession = new WatchStore();

  constructor(
    gateway: ReviewGateway,
    clock: Clock,
    scheduler: Scheduler,
    private readonly clipboard: Clipboard,
  ) {
    super(gateway, clock, scheduler, createOperationIdentity());
    this.controls = new ControlStore(gateway, {
      run: () => this.snapshot().worldId,
      newest: () => this.snapshot().liveHistory?.entries.at(-1) ?? null,
    });
    this.subscribe(() => {
      this.controls.bind(this.snapshot().worldId);
    });
    this.twin = new TwinStore(gateway);
  }

  copyText(text: string): Promise<boolean> {
    return this.clipboard.copy(text);
  }
}

export function createReviewSession(
  gateway: ReviewGateway,
  clock: Clock,
  scheduler: Scheduler,
  clipboard: Clipboard = noClipboard,
): ReviewSession {
  return new ControlledReview(gateway, clock, scheduler, clipboard);
}
