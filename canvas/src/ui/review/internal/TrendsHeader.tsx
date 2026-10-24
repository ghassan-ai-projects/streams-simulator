import { ChevronDown, ChevronUp, X } from "lucide-react";
import type { WatchSession } from "../../../app/review";
import { windowChoices, windowWords } from "../../../model/review";
import type { TrendSize, WindowChoice } from "../../../model/review";
import { resizedTrend, trendSizes } from "../../../model/review";
import styles from "./Trends.module.css";

interface Props {
  readonly watch: WatchSession;
  readonly plantId: string;
  readonly count: number;
  readonly choice: WindowChoice;
  readonly running: boolean;
  readonly cursorWords: string;
  readonly size: TrendSize;
}

function Sizing({ watch, size }: Pick<Props, "watch" | "size">) {
  const last = trendSizes.length - 1;
  const grow = (direction: "larger" | "smaller") => {
    watch.setSize(resizedTrend(size, direction));
  };
  return (
    <div className={styles.sizing} role="group" aria-label="Panel height">
      <button
        type="button"
        className={styles.action}
        aria-label="Make the trends panel shorter"
        title="Shorter"
        aria-disabled={size === trendSizes[0]}
        onClick={() => {
          grow("smaller");
        }}
      >
        <ChevronDown size={14} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={styles.action}
        aria-label="Make the trends panel taller"
        title="Taller"
        aria-disabled={size === trendSizes[last]}
        onClick={() => {
          grow("larger");
        }}
      >
        <ChevronUp size={14} aria-hidden="true" />
      </button>
    </div>
  );
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

export function TrendsHeader({ watch, plantId, count, choice, running, cursorWords, size }: Props) {
  return (
    <>
      <div className={styles.head}>
        <h2 className={styles.title}>{`Trends · ${String(count)} watched`}</h2>
        <span className={styles.when} data-part="cursor-words">
          {cursorWords}
        </span>
        <WindowChoices watch={watch} choice={choice} />
        <Sizing watch={watch} size={size} />
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
