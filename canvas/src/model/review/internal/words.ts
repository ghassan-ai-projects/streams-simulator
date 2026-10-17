import { evaluateRule, spaced } from "../../shared/words";

export const modeWords = {
  explore: "Explore: the truth is visible; not for scoring.",
  evaluate: `Evaluate: ${evaluateRule}`,
  evaluateRule,
} as const;

const alarmWords: Readonly<Record<string, readonly [string, string]>> = {
  activations: ["Alarms raised", "Alarms that came on during the run."],
  peak_per_10_min: ["Most alarms in 10 minutes", "The busiest 10-minute spell: how many alarms came on in it."],
  flood_intervals: ["Overload spells", "10-minute spells where alarms came too fast for an operator to follow."],
  chattering: ["Repeating alarms", "Alarms that kept switching on and off in a short time."],
  stale: ["Stale alarms", "Alarms still active long after they came on, with no operator action."],
  standing_at_end: ["Still active at the end", "Alarms still on when the run ended."],
};

export function alarmRow(key: string, value: string): readonly [string, string, string] {
  const [label, meaning] = alarmWords[key] ?? [spaced(key), ""];
  return [label, value, meaning];
}
