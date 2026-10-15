export const alarmWords: Readonly<Record<string, readonly [string, string]>> = {
  activations: ["Alarms raised", "Alarms that came on during the run."],
  peak_per_10_min: ["Most alarms in 10 minutes", "The busiest 10-minute spell: how many alarms came on in it."],
  flood_intervals: [
    "Overload spells",
    "10-minute spells where alarms came too fast for an operator to follow (ISA-18.2: floods).",
  ],
  chattering: ["Repeating alarms", "Alarms that kept switching on and off in a short time (ISA-18.2: chattering)."],
  stale: ["Stale alarms", "Alarms still active long after they came on, with no operator action (ISA-18.2: stale)."],
  standing_at_end: ["Still active at the end", "Alarms still on when the run ended (ISA-18.2: standing)."],
};

export function alarmRow(key: string, value: string): readonly [string, string, string] {
  const [label, meaning] = alarmWords[key] ?? [key.replaceAll("_", " "), ""];
  return [label, value, meaning];
}
