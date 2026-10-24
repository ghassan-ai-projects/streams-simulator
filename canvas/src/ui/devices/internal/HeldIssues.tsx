import type { HeldIssue } from "../../../model/devices";
import { heldSelection } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import prov from "./Provenance.module.css";
import type { Reveal } from "./twinProps";

interface IssuesProps {
  readonly issues: readonly HeldIssue[];
  readonly reveal: Reveal | null;
}

function showWords(issue: HeldIssue): string {
  const where = issue.setting === "" ? issue.field : issue.setting;
  return where === "" ? "Show it in the Interface step" : `Show ${where} in the Interface step`;
}

function Issue({ issue, reveal }: { readonly issue: HeldIssue; readonly reveal: Reveal | null }) {
  return (
    <li className={prov.heldIssue}>
      <span>{issue.sentence}</span>
      {reveal !== null && (
        <button
          type="button"
          className={cx(shared.secondary, panel.secondary)}
          onClick={() => {
            reveal(heldSelection(issue), issue.field);
          }}
        >
          {showWords(issue)}
        </button>
      )}
    </li>
  );
}

export function HeldIssues({ issues, reveal }: IssuesProps) {
  if (issues.length === 0) return null;
  return (
    <>
      <ul className={prov.heldIssues}>
        {issues.map((issue, index) => (
          <Issue key={`${String(index)}:${issue.sentence}`} issue={issue} reveal={reveal} />
        ))}
      </ul>
      <details className={prov.heldDetails}>
        <summary>Details</summary>
        {issues.map((issue) => issue.original).join(" ")}
      </details>
    </>
  );
}
