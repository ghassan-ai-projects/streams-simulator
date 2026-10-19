import { TriangleAlert } from "lucide-react";
import { useRef } from "react";
import type { PolicyDefaultState } from "../../../app/launch";
import type { PlanIssue, PolicyNote, SessionDraft } from "../../../model/launch";
import {
  policyHint,
  policyNotes,
  policyNoteWords,
  prefilledWords,
  unreadableDefaultWords,
  withPolicyText,
} from "../../../model/launch";
import { cx, panel, shared } from "../../primitives";
import launch from "./LaunchPanel.module.css";
import styles from "./Session.module.css";

interface PolicyFieldProps {
  readonly draft: SessionDraft;
  readonly change: (next: SessionDraft) => void;
  readonly issues: readonly PlanIssue[];
  readonly defaults: PolicyDefaultState;
}

function sourceNote(draft: SessionDraft, defaults: PolicyDefaultState): string {
  if (draft.policiesFrom === "default" && draft.policies !== "") return prefilledWords;
  return defaults.file === "unreadable" ? unreadableDefaultWords : "";
}

function describedBy(note: string, notes: readonly PolicyNote[]): string {
  const ids = ["session-policies-hint"];
  if (note !== "") ids.push("session-policies-note");
  if (notes.length > 0) ids.push("session-policies-issue");
  return ids.join(" ");
}

function PolicyNotes({ note, notes }: { readonly note: string; readonly notes: readonly PolicyNote[] }) {
  return (
    <>
      <small id="session-policies-hint" className={panel.hint}>
        {policyHint}
      </small>
      {note !== "" && (
        <small id="session-policies-note" className={panel.hint}>
          {note}
        </small>
      )}
      {notes.length > 0 && (
        <small id="session-policies-issue" className={cx(launch.problem, styles.policyIssues)}>
          <TriangleAlert size={13} aria-hidden="true" />
          {notes.map((item) => policyNoteWords(item)).join(" ")}
        </small>
      )}
    </>
  );
}

export function PolicyField({ draft, change, issues, defaults }: PolicyFieldProps) {
  const field = useRef<HTMLTextAreaElement>(null);
  const notes = policyNotes(issues);
  const note = sourceNote(draft, defaults);
  return (
    <div className={styles.policyField}>
      <label htmlFor="session-policies">Allowed policy digests</label>
      <textarea
        id="session-policies"
        ref={field}
        rows={3}
        spellCheck={false}
        className={cx(shared.sunkenField, styles.policies)}
        value={draft.policies}
        placeholder="sha256:… (one per line)"
        aria-describedby={describedBy(note, notes)}
        {...(notes.length > 0 ? { "aria-invalid": true } : {})}
        onChange={(event) => {
          change(withPolicyText(draft, event.target.value));
        }}
      />
      <PolicyNotes note={note} notes={notes} />
      {draft.policies !== "" && (
        <button
          type="button"
          className={cx(shared.secondary, panel.secondary)}
          onClick={() => {
            change(withPolicyText(draft, ""));
            field.current?.focus();
          }}
        >
          Clear the list
        </button>
      )}
    </div>
  );
}
