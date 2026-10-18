import { TriangleAlert } from "lucide-react";

interface IssueProps {
  readonly title: string;
  readonly text: string;
  readonly message: string;
}

export function CaptureIssue({ title, text, message }: IssueProps) {
  return (
    <li>
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
