import { TriangleAlert } from "lucide-react";
import type { Ref } from "react";

interface IssueProps {
  readonly title: string;
  readonly text: string;
  readonly message: string;
  readonly focus?: Ref<HTMLLIElement> | undefined;
}

export function CaptureIssue({ title, text, message, focus }: IssueProps) {
  return (
    <li ref={focus} tabIndex={focus === undefined ? undefined : -1}>
      <TriangleAlert size={14} aria-hidden="true" />
      <span>
        {title !== "" && <strong>{title} </strong>}
        {text}
        {message !== text && (
          <details>
            <summary>Details</summary>
            {message}
          </details>
        )}
      </span>
    </li>
  );
}
