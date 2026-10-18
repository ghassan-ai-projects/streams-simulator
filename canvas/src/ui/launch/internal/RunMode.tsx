import { cx, panel } from "../../primitives";
import type { RunKind } from "./useRunMode";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

const choices: readonly (readonly [RunKind, string, string])[] = [
  ["complete", "To completion", "Simulate the plant for a set duration, then save the run for review."],
  [
    "session",
    "Connected session",
    "Link the plant to a plant wire host, and through a gateway to real boards, until you stop it.",
  ],
];

interface RunModeProps {
  readonly kind: RunKind;
  readonly locked: string;
  readonly onKind: (kind: RunKind) => void;
}

export function RunMode({ kind, locked, onKind }: RunModeProps) {
  return (
    <fieldset className={cx(panel.group, styles.kinds)} disabled={locked !== ""} title={locked || undefined}>
      <legend>How this run goes</legend>
      {choices.map(([value, title, words]) => (
        <label key={value} className={launch.mode}>
          <input
            type="radio"
            name="run-kind"
            checked={kind === value}
            onChange={() => {
              onKind(value);
            }}
          />
          <span>
            <strong>{title}</strong>
            <small>{words}</small>
          </span>
        </label>
      ))}
    </fieldset>
  );
}
