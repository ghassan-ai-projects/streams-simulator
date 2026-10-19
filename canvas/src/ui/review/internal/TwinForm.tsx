import { useId, useState } from "react";
import type { ReviewSession } from "../../../app/review";
import { shared } from "../../primitives";
import styles from "./Twin.module.css";

interface TwinFormProps {
  readonly review: ReviewSession;
  readonly run: string;
  readonly domain: string;
  readonly busy: boolean;
}

const helpWords =
  "Optional. Absolute per metric, written device.metric=value or kind.metric=value, separated by commas. " +
  "A metric without one uses a share of its declared range. The binary checks what you type.";

export function TwinForm({ review, run, domain, busy }: TwinFormProps) {
  const [tolerance, setTolerance] = useState("");
  const help = useId();
  return (
    <form
      className={styles.form}
      onSubmit={(event) => {
        event.preventDefault();
        if (!busy) void review.twin.compare({ run, domain, tolerance });
      }}
    >
      <label className={styles.field}>
        Tolerances (optional)
        <input
          type="text"
          className={styles.input}
          value={tolerance}
          spellCheck={false}
          autoComplete="off"
          aria-describedby={help}
          onChange={(event) => {
            setTolerance(event.target.value);
          }}
        />
      </label>
      <p id={help} className={styles.help}>
        {helpWords}
      </p>
      <div>
        <button type="submit" className={shared.primary} aria-disabled={busy}>
          Compare with twin
        </button>
      </div>
    </form>
  );
}
