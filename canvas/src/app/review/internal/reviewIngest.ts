import { PlaybackIndex } from "../../../model/review";
import type { JournalEntry } from "../../../model/review";
import type { OpenMode, SessionEntry } from "./reviewState";

const watchedReadings = 2048;

export function indexFor(mode: OpenMode): PlaybackIndex {
  return mode === "watch" ? new PlaybackIndex({ capacity: watchedReadings }) : new PlaybackIndex();
}

export function ingest(index: PlaybackIndex, entry: JournalEntry): void {
  if (entry.kind === "observation") index.add(entry.observation);
  else if (entry.kind === "call") index.addCall(entry.call);
  else if (entry.kind === "clock") index.advanceClock(entry.atNs);
  else if (entry.kind === "opened") index.markOpened(entry.atNs);
}

export function isSessionEntry(entry: JournalEntry): entry is SessionEntry {
  return entry.kind === "origin" || entry.kind === "binding" || entry.kind === "closed" || entry.kind === "session";
}
