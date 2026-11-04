import { Check, CircleAlert, GitCompare, Plus, TriangleAlert } from "lucide-react";
import { useId } from "react";
import type { Card, CaptureSession } from "../../../app/devices";
import type { NavigationStore } from "../../../app/shell";
import { captureLines, comparisonWords, draftIssues, kindFacts } from "../../../model/devices";
import type { DraftLines } from "../../../model/devices";
import { cx } from "../../primitives";
import { CaptureActions } from "./CaptureActions";
import cardStyles from "./CaptureCard.module.css";
import { Claims, Comparing, Fingerprint, Issues, RawDraft, Summary } from "./CaptureDetail";
import type { CardFocus } from "./useCardFocus";

interface CardProps {
  readonly card: Card;
  readonly index: number;
  readonly session: CaptureSession;
  readonly navigation: NavigationStore;
  readonly focus: CardFocus;
}

type Comparison = Card["kind"]["comparison"];

const looks = {
  new: { Icon: Plus, style: cardStyles.chipNew },
  same: { Icon: Check, style: cardStyles.chipSame },
  differs: { Icon: GitCompare, style: cardStyles.chipDiffers },
  unreadable: { Icon: TriangleAlert, style: cardStyles.chipDiffers },
} as const;

function Status({ comparison, saved }: { readonly comparison: Comparison; readonly saved: boolean }) {
  const { Icon, style } = saved ? looks.same : looks[comparison.status];
  const title = comparison.status === "new" ? undefined : comparison.file;
  return (
    <span className={cx(cardStyles.chip, style)} title={title}>
      <Icon size={12} aria-hidden="true" />
      {saved ? "Saved" : comparisonWords(comparison)}
    </span>
  );
}

function Head({
  card,
  id,
  lines,
  invalid,
}: {
  readonly card: Card;
  readonly id: string;
  readonly lines: DraftLines | null;
  readonly invalid: boolean;
}) {
  const facts = lines === null ? kindFacts(null, null) : kindFacts(lines.readings.length, lines.commands.length);
  return (
    <header className={cardStyles.cardHead}>
      <h3 id={id} className={cardStyles.cardName}>
        {card.kind.kind}
      </h3>
      <span className={cardStyles.reading}>{facts}</span>
      <Status comparison={card.kind.comparison} saved={card.decision.kind === "saved"} />
      {invalid && (
        <span className={cx(cardStyles.chip, cardStyles.chipInvalid)}>
          <CircleAlert size={12} aria-hidden="true" />
          Not yet valid
        </span>
      )}
    </header>
  );
}

export function CaptureCard(props: CardProps) {
  const id = useId();
  const { kind } = props.card;
  const lines = captureLines(kind.draft);
  const issues = draftIssues(kind.draft, kind.issues);
  return (
    <li>
      <article className={cardStyles.card} aria-labelledby={id}>
        <Head card={props.card} id={id} lines={lines} invalid={issues.length > 0} />
        <Comparing card={props.card} />
        <Issues name={kind.kind} issues={issues} />
        <Summary lines={lines} />
        <Claims claims={kind.claims} />
        <Fingerprint card={props.card} />
        <RawDraft draft={kind.draft} />
        <CaptureActions {...props} />
      </article>
    </li>
  );
}
