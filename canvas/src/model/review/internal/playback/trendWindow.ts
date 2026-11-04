import type { PlaybackIndex } from "./playback";
import type { Span } from "./trend";
import type { Watched } from "./watchList";

export type WindowChoice = "10" | "30" | "all";

export const windowChoices: readonly { readonly id: WindowChoice; readonly label: string }[] = [
  { id: "10", label: "10 min" },
  { id: "30", label: "30 min" },
  { id: "all", label: "Whole run" },
];

const shortestNs = 60e9;
const minuteNs = 60e9;

export interface WindowFacts {
  readonly choice: WindowChoice | null;
  readonly running: boolean;
  readonly reachedNs: number;
  readonly cursorNs: number;
  readonly runStartNs: number;
  readonly runEndNs: number;
}

export interface TrendSpan {
  readonly span: Span;
  readonly choice: WindowChoice;
  readonly limitedFromNs: number | null;
}

export function choiceOf(choice: WindowChoice | null, run: { readonly running: boolean }): WindowChoice {
  return choice ?? (run.running ? "10" : "all");
}

function requestedSpan(facts: WindowFacts, choice: WindowChoice): Span {
  const endNs = facts.running ? facts.reachedNs : choice === "all" ? facts.runEndNs : facts.cursorNs;
  const startNs = choice === "all" ? facts.runStartNs : endNs - Number(choice) * minuteNs;
  return { startNs: Math.max(facts.runStartNs, startNs), endNs: Math.max(endNs, facts.runStartNs + 1) };
}

function widened(span: Span): Span {
  return span.endNs - span.startNs >= shortestNs ? span : { startNs: span.endNs - shortestNs, endNs: span.endNs };
}

function trimmedIn(index: PlaybackIndex, watched: readonly Watched[], span: Span): number | null {
  const trims = watched
    .map((entry) => index.trimmedBeforeNs(entry.entity, entry.channel))
    .filter((at): at is number => at !== null && at >= span.startNs);
  return trims.length === 0 ? null : Math.max(...trims);
}

export function trendSpanOf(index: PlaybackIndex, watched: readonly Watched[], facts: WindowFacts): TrendSpan {
  const choice = choiceOf(facts.choice, facts);
  const requested = widened(requestedSpan(facts, choice));
  const limitedFromNs = trimmedIn(index, watched, requested);
  const startNs = limitedFromNs === null ? requested.startNs : Math.min(limitedFromNs, requested.endNs - 1);
  return { span: { startNs, endNs: requested.endNs }, choice, limitedFromNs };
}

export function windowWords(choice: WindowChoice, run: { readonly running: boolean }): string {
  const { running } = run;
  if (choice === "all") return running ? "Everything the run has produced so far" : "The whole run";
  return `${choice} min of simulated time ${running ? "up to the reached run clock" : "ending at the cursor"}`;
}
