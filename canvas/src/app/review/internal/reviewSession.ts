import type { Clipboard, Clock, FileSaver, PreferenceStore, ReviewGateway, Scheduler } from "../../ports";
import { createOperationIdentity } from "../../shared";
import type { ReviewSession } from "./api";
import type { CompareSession } from "./compareStore";
import { CompareStore } from "./compareStore";
import type { ControlSession } from "./controlStore";
import { ControlStore } from "./controlStore";
import type { HandoverSession } from "./handoverStore";
import { HandoverStore } from "./handoverStore";
import { ReviewStore } from "./reviewStore";
import type { TwinSession } from "./twinStore";
import { TwinStore } from "./twinStore";
import type { WatchSession } from "./watchStore";
import { WatchStore } from "./watchStore";

const noClipboard: Clipboard = { copy: () => Promise.resolve(false) };
const noSaver: FileSaver = { save: () => false };
const noViewer: PreferenceStore = { read: () => null, write: () => undefined };

class ControlledReview extends ReviewStore implements ReviewSession {
  readonly controls: ControlSession;
  readonly twin: TwinSession;
  readonly watchList: WatchSession;
  readonly compare: CompareSession;
  readonly handover: HandoverSession;

  constructor(
    gateway: ReviewGateway,
    clock: Clock,
    scheduler: Scheduler,
    private readonly clipboard: Clipboard,
    saver: FileSaver,
    viewer: PreferenceStore,
  ) {
    super(gateway, clock, scheduler, createOperationIdentity());
    this.watchList = new WatchStore(viewer);
    this.controls = new ControlStore(gateway, {
      run: () => this.snapshot().worldId,
      newest: () => this.snapshot().liveHistory?.entries.at(-1) ?? null,
    });
    this.subscribe(() => {
      this.controls.bind(this.snapshot().worldId);
    });
    this.twin = new TwinStore(gateway);
    this.compare = new CompareStore(gateway);
    this.handover = new HandoverStore(gateway, saver);
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
  saver: FileSaver = noSaver,
  viewer: PreferenceStore = noViewer,
): ReviewSession {
  return new ControlledReview(gateway, clock, scheduler, clipboard, saver, viewer);
}
