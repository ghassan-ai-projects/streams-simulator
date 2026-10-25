import type { ReplayCheck } from "../../../app/review";
import type { OpenQuestion } from "../../../model/review";
import { shared } from "../../primitives";
import styles from "./ThisRun.module.css";

export interface ReplayControl {
  readonly check: ReplayCheck;
  readonly onCheck: () => void;
}

export const noReplayControl: ReplayControl = { check: { phase: "idle" }, onCheck: () => undefined };

export function ReplayButton({ control }: { readonly control: ReplayControl }) {
  const checking = control.check.phase === "checking";
  return (
    <>
      <button type="button" className={shared.secondary} disabled={checking} onClick={control.onCheck}>
        Check replay now
      </button>
      {checking && (
        <span role="status" className={styles.progress}>
          Replaying the run…
        </span>
      )}
      {control.check.phase === "failed" && (
        <span role="alert" className={styles.failure}>
          {control.check.message}
        </span>
      )}
    </>
  );
}

function Question({ question, replay }: { readonly question: OpenQuestion; readonly replay: ReplayControl }) {
  return (
    <li className={styles.question} data-code={question.code}>
      <span>{question.message}</span>
      {question.code === "replay_not_checked" && <ReplayButton control={replay} />}
    </li>
  );
}

export function OpenQuestions({
  questions,
  replay,
}: {
  readonly questions: readonly OpenQuestion[];
  readonly replay: ReplayControl;
}) {
  return (
    <section className={styles.questions} aria-label="Open questions">
      <h3 className={styles.subtitle}>Open questions</h3>
      {questions.length === 0 ? (
        <p className={styles.empty}>Nothing open.</p>
      ) : (
        <ul className={styles.list}>
          {questions.map((question) => (
            <Question key={question.code} question={question} replay={replay} />
          ))}
        </ul>
      )}
    </section>
  );
}
