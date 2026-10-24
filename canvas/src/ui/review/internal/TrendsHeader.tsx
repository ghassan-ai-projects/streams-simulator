import { X } from "lucide-react";
import type { WatchSession } from "../../../app/review";
import { windowChoices, windowWords } from "../../../model/review";
import type { WindowChoice } from "../../../model/review";
import styles from "./Trends.module.css";

interface Props {
  readonly watch: WatchSession;
  readonly plantId: string;
  readonly count: number;
  readonly choice: WindowChoice;
  readonly running: boolean;
  readonly cursorWords: string;
}

function WindowChoices({ watch, choice }: Pick<Props, "watch" | "choice">) {
  return (
    <div className={styles.choices} role="group" aria-label="Time window">
      {windowChoices.map((entry) => (
        <button
          key={entry.id}
          type="button"
          className={styles.choice}
          aria-pressed={choice === entry.id}
          onClick={() => {
            watch.setChoice(entry.id);
          }}
        >
          {entry.label}
        </button>
      ))}
    </div>
  );
}

export function TrendsHeader({ watch, plantId, count, choice, running, cursorWords }: Props) {
  return (
    <>
      <div className={styles.head}>
        <h2 className={styles.title}>{`Trends · ${String(count)} watched`}</h2>
        <span className={styles.when} data-part="cursor-words">
          {cursorWords}
        </span>
        <WindowChoices watch={watch} choice={choice} />
        {count > 0 && (
          <button
            type="button"
            className={styles.action}
            onClick={() => {
              watch.clear(plantId);
            }}
          >
            Clear
          </button>
        )}
        <button
          type="button"
          className={styles.action}
          aria-label="Hide the trends"
          onClick={() => {
            watch.hide();
          }}
        >
          <X size={14} aria-hidden="true" />
        </button>
      </div>
      <p className={styles.note}>{windowWords(choice, { running })}</p>
    </>
  );
}
