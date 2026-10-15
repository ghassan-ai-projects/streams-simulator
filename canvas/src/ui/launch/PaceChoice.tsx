import type { Draft } from "./draft";
import type { Change } from "./FaultRows";
import { watchWords } from "../../model/launch";
import panel from "../panel.module.css";
import styles from "./LaunchPanel.module.css";

const paceChoices = [
  { value: "watch", name: "Watch" },
  { value: "fast", name: "As fast as possible" },
] as const;

function paceNote(pace: "watch" | "fast", durationS: number): string {
  if (pace === "fast") return "The run is made in one go; the review is ready as soon as it ends.";
  const length = Number.isFinite(durationS) && durationS > 0 ? watchWords(durationS) : "30 s";
  return `The run plays out in ${length}, so you can follow its time, faults and calls while it runs.`;
}

export function Pace({
  draft,
  change,
  durationS,
}: {
  readonly draft: Draft;
  readonly change: (change: Change) => void;
  readonly durationS: number;
}) {
  return (
    <fieldset className={panel.group}>
      <legend>Run speed</legend>
      {paceChoices.map((choice) => (
        <label key={choice.value} className={styles.mode}>
          <input
            type="radio"
            name="launch-pace"
            checked={draft.pace === choice.value}
            onChange={() => {
              change((d) => ({ ...d, pace: choice.value }));
            }}
          />
          <span>
            <strong>{choice.value === "watch" ? `Watch: ${watchWords(durationS)}` : choice.name}</strong>
            <small>{paceNote(choice.value, durationS)}</small>
          </span>
        </label>
      ))}
    </fieldset>
  );
}
