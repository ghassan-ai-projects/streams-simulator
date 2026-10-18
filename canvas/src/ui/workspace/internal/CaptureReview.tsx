import { Cpu, FlaskConical } from "lucide-react";
import type { CaptureSession, Reviewed } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { birthWords, claimWords, originWords } from "../../../model/devices";
import { cx, useFocusOnMount } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureCard } from "./CaptureCard";

interface ReviewProps {
  readonly board: Reviewed;
  readonly session: CaptureSession;
  readonly navigation: NavigationStore;
}

function Origin({ environment }: Pick<Reviewed, "environment">) {
  const real = environment === "hardware";
  const Icon = real ? Cpu : FlaskConical;
  return (
    <span className={cx(styles.origin, real ? styles.originReal : styles.originSimulated)}>
      <Icon size={13} aria-hidden="true" />
      {originWords(environment)}
    </span>
  );
}

export function CaptureReview({ board, session, navigation }: ReviewProps) {
  const heading = useFocusOnMount<HTMLHeadingElement>();
  return (
    <section className={styles.panel} aria-labelledby="capture-review">
      <h2 id="capture-review" ref={heading} tabIndex={-1} className={styles.panelTitle}>
        Review what the board announced
      </h2>
      <div className={styles.header}>
        <Origin environment={board.environment} />
        <span className={styles.facts}>{birthWords(board.bdSeq)}</span>
      </div>
      <p className={styles.claim}>{claimWords}</p>
      {board.cards.length === 0 ? (
        <p className={styles.note}>The board announced no devices.</p>
      ) : (
        <ul className={styles.cards}>
          {board.cards.map((card) => (
            <CaptureCard key={card.kind.kind} card={card} session={session} navigation={navigation} />
          ))}
        </ul>
      )}
    </section>
  );
}
