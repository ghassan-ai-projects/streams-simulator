import type { RunFollower } from "../../../app/ports";
import type { Transport } from "./http";
import { journalEntry } from "./journalDecoder";

const streamClosed = 2;

export function followRun(transport: Transport, runId: string, follower: RunFollower): () => void {
  const source = transport.events(`/api/runs/${encodeURIComponent(runId)}/events`);
  source.addEventListener("record", (event: Event) => {
    const data = event instanceof MessageEvent && typeof event.data === "string" ? event.data : "";
    const entry = journalEntry(data);
    if (entry === null) {
      source.close();
      follower.failed();
      return;
    }
    follower.entry(entry);
  });
  source.addEventListener("caught_up", () => {
    follower.caughtUp?.();
  });
  source.addEventListener("closed", () => {
    source.close();
    follower.closed();
  });
  source.addEventListener("error", () => {
    if (source.readyState === streamClosed) follower.failed();
  });
  return () => {
    source.close();
  };
}
