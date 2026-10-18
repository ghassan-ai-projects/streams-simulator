import { TriangleAlert } from "lucide-react";
import type { PlanIssue, PlanMember, SessionDraft } from "../../../model/launch";
import { issueWords, memberIssues } from "../../../model/launch";
import { cx, panel } from "../../primitives";
import launch from "./LaunchPanel.module.css";

export interface DraftEdit {
  readonly draft: SessionDraft;
  readonly change: (next: SessionDraft) => void;
  readonly issues: readonly PlanIssue[];
}

export function IssueNote({ id, issues }: { readonly id: string; readonly issues: readonly PlanIssue[] }) {
  return (
    <small id={id} className={cx(launch.problem, launch.inField)}>
      {issues.length > 0 && <TriangleAlert size={13} aria-hidden="true" />}
      {issues.map((issue) => issueWords(issue)).join(" ")}
    </small>
  );
}

export function flagged(issues: readonly PlanIssue[], id: string) {
  return issues.length === 0 ? {} : { "aria-invalid": true, "aria-describedby": id };
}

interface TextFieldProps {
  readonly id: string;
  readonly label: string;
  readonly value: string;
  readonly issues: readonly PlanIssue[];
  readonly placeholder: string;
  readonly onText: (text: string) => void;
}

function TextField({ id, label, value, issues, placeholder, onText }: TextFieldProps) {
  return (
    <div className={panel.field}>
      <label htmlFor={id}>{label}</label>
      <input
        id={id}
        value={value}
        placeholder={placeholder}
        {...flagged(issues, `${id}-issue`)}
        onChange={(event) => {
          onText(event.target.value);
        }}
      />
      <IssueNote id={`${id}-issue`} issues={issues} />
    </div>
  );
}

type TextKey = "seed" | "hostSocket" | "gatewaySocket";

const edits: Readonly<Record<TextKey, (draft: SessionDraft, text: string) => SessionDraft>> = {
  seed: (draft, seed) => ({ ...draft, seed }),
  hostSocket: (draft, hostSocket) => ({ ...draft, hostSocket }),
  gatewaySocket: (draft, gatewaySocket) => ({ ...draft, gatewaySocket }),
};

const fields: readonly (readonly [PlanMember, string, string, TextKey, string])[] = [
  ["seed", "Seed", "a whole number, such as 1", "seed", "session-seed"],
  ["host_socket", "Host socket", "/full/path/to/host.sock", "hostSocket", "session-host"],
  ["gateway_socket", "Gateway socket", "empty: every instance simulated", "gatewaySocket", "session-gateway"],
];

export function SessionFields({ draft, change, issues }: DraftEdit) {
  return (
    <fieldset className={panel.group}>
      <legend>Session</legend>
      {fields.map(([member, label, placeholder, key, id]) => (
        <TextField
          key={member}
          id={id}
          label={label}
          value={draft[key]}
          issues={memberIssues(issues, member)}
          placeholder={placeholder}
          onText={(text) => {
            change(edits[key](draft, text));
          }}
        />
      ))}
      <p className={panel.hint}>
        The host socket is the one the plant wire host listens on. The gateway socket is the one this session opens for
        a gateway with real boards; leave it empty to run every instance simulated.
      </p>
    </fieldset>
  );
}
