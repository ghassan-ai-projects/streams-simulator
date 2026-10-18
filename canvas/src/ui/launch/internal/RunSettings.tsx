import { Dices, TriangleAlert } from "lucide-react";
import type { Draft, PlanProblem } from "../../../model/launch";
import { modeWords } from "../../../model/launch";
import { cx, panel } from "../../primitives";
import type { Change } from "./FaultRows";
import styles from "./LaunchPanel.module.css";

type RunField = "seed" | "duration";

interface DraftProps {
  readonly draft: Draft;
  readonly change: (change: Change) => void;
  readonly problems: readonly PlanProblem[];
}

function FieldProblem({
  id,
  problems,
  field,
}: { readonly id: string; readonly field: RunField } & Pick<DraftProps, "problems">) {
  const text = problems.find((problem) => problem.field === field)?.text ?? "";
  return (
    <small id={id} className={cx(styles.problem, styles.inField)}>
      {text !== "" && <TriangleAlert size={13} aria-hidden="true" />}
      {text}
    </small>
  );
}

function described(problems: readonly PlanProblem[], field: RunField, id: string, hint = "") {
  const wrong = problems.some((problem) => problem.field === field);
  const ids = [hint, wrong ? id : ""].filter((part) => part !== "").join(" ");
  return { "aria-invalid": wrong || undefined, "aria-describedby": ids === "" ? undefined : ids };
}

export function Duration({ draft, change, problems }: DraftProps) {
  return (
    <div className={panel.field}>
      <label htmlFor="launch-duration">Duration</label>
      <span className={cx(panel.pair, styles.pair)}>
        <input
          id="launch-duration"
          inputMode="decimal"
          value={draft.duration}
          {...described(problems, "duration", "launch-duration-problem")}
          onChange={(event) => {
            change((d) => ({ ...d, duration: event.target.value }));
          }}
        />
        <select
          aria-label="Duration unit"
          value={draft.unit}
          onChange={(event) => {
            change((d) => ({ ...d, unit: event.target.value === "hours" ? "hours" : "minutes" }));
          }}
        >
          <option value="minutes">minutes</option>
          <option value="hours">hours</option>
        </select>
      </span>
      <FieldProblem id="launch-duration-problem" field="duration" problems={problems} />
    </div>
  );
}

function ModeChoice({
  explore,
  checked,
  onChoose,
}: {
  readonly explore: boolean;
  readonly checked: boolean;
  readonly onChoose: () => void;
}) {
  return (
    <label className={styles.mode}>
      <input type="radio" name="launch-mode" checked={checked} onChange={onChoose} />
      <span>
        <strong>{explore ? "Explore" : "Evaluate"}</strong>
        <small>{modeWords(explore)}</small>
      </span>
    </label>
  );
}

export function Mode({ draft, change }: Omit<DraftProps, "problems">) {
  return (
    <fieldset className={panel.group}>
      <legend>Mode</legend>
      {[true, false].map((explore) => (
        <ModeChoice
          key={String(explore)}
          explore={explore}
          checked={draft.explore === explore}
          onChoose={() => {
            change((d) => ({ ...d, explore }));
          }}
        />
      ))}
    </fieldset>
  );
}

export function Seed({ draft, change, problems }: DraftProps) {
  return (
    <div className={panel.field}>
      <label htmlFor="launch-seed">Seed</label>
      <span className={styles.seed}>
        <input
          id="launch-seed"
          inputMode="numeric"
          value={draft.seed}
          {...described(problems, "seed", "launch-seed-problem", "launch-seed-hint")}
          onChange={(event) => {
            change((d) => ({ ...d, seed: event.target.value }));
          }}
        />
        <button
          type="button"
          className={panel.icon}
          aria-label="Pick a new seed"
          title="Pick a new seed"
          onClick={() => {
            change((d) => ({ ...d, seed: String(Math.floor(Math.random() * 1_000_000)) }));
          }}
        >
          <Dices size={16} aria-hidden="true" />
        </button>
      </span>
      <FieldProblem id="launch-seed-problem" field="seed" problems={problems} />
      <small id="launch-seed-hint" className={styles.note}>
        The same seed, plant, faults and simulator version give the same run.
      </small>
    </div>
  );
}
