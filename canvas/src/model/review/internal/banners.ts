import { grouped } from "../../shared/units";
import type { Replay, Report } from "./report";
import { replayLine } from "./report";

export interface BannerWarning {
  readonly text: string;
  readonly details: string;
}

export type VerifyLoad =
  | { readonly phase: "loading" }
  | { readonly phase: "failed"; readonly failure: { readonly diagnostic: { readonly message: string } } }
  | { readonly phase: "ready"; readonly value: Replay };

export interface PlaybackFacts {
  readonly phase: "loading" | "ready" | "failed";
  readonly records: number;
}

function verifyWarning(verified: VerifyLoad): BannerWarning | null {
  if (verified.phase !== "failed") return null;
  return { text: replayLine(verified), details: verified.failure.diagnostic.message };
}

function differWarning(report: Report): BannerWarning | null {
  if (report.verification !== "files_differ") return null;
  return {
    text: "The run's files differ from what the run wrote: playback shows the files as they are now.",
    details: "",
  };
}

function countWarning(report: Report, playback: PlaybackFacts): BannerWarning | null {
  const delivered = report.counts.delivered ?? 0;
  const records = grouped(playback.records);
  if (playback.phase === "failed")
    return {
      text: `Playback is incomplete: ${records} of ${grouped(delivered)} delivered observations arrived.`,
      details: "",
    };
  if (playback.phase === "ready" && report.ended && playback.records !== delivered)
    return {
      text: `Playback has ${records} observations; the report counts ${grouped(delivered)} delivered.`,
      details: "",
    };
  return null;
}

export function bannerWarnings(
  report: Report | null,
  verified: VerifyLoad,
  playback: PlaybackFacts,
): readonly BannerWarning[] {
  if (report === null) return [];
  return [verifyWarning(verified), differWarning(report), countWarning(report, playback)].filter(
    (warning): warning is BannerWarning => warning !== null,
  );
}
