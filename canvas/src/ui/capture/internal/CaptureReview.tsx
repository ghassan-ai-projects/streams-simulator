import { Cpu, FlaskConical } from "lucide-react";
import { useRef } from "react";
import type { CaptureSession, Reviewed } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { birthWords, claimWords, originWords } from "../../../model/devices";
import { cx, useFocusOnMount } from "../../primitives";
import styles from "./Capture.module.css";
import { CaptureCard } from "./CaptureCard";
import cardStyles from "./CaptureCard.module.css";
import { CaptureNotices } from "./CaptureNotices";
import { useCardFocus } from "./useCardFocus";

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
  const list = useRef<HTMLUListElement>(null);
  const focus = useCardFocus(list, heading);
  const restarts = birthWords(board.bdSeq);
  return (
    <section className={styles.panel} aria-labelledby="capture-review">
      <h2 id="capture-review" ref={heading} tabIndex={-1} className={styles.panelTitle}>
        Review what the board announced
      </h2>
      <div className={styles.header}>
        <Origin environment={board.environment} />
        {restarts !== "" && <span className={styles.facts}>{restarts}</span>}
      </div>
      <p className={styles.claim}>{claimWords}</p>
      <CaptureNotices scan={board.scan} />
      {board.cards.length === 0 ? (
        <p className={styles.note}>The board announced no devices.</p>
      ) : (
        <ul className={cardStyles.cards} ref={list}>
          {board.cards.map((card, index) => (
            <CaptureCard
              key={`${String(index)}:${card.kind.kind}`}
              card={card}
              index={index}
              session={session}
              navigation={navigation}
              focus={focus}
            />
          ))}
        </ul>
      )}
    </section>
  );
}
