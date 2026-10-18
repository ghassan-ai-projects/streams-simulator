import { TriangleAlert } from "lucide-react";
import type { PlanIssue, PlanMember, SessionDraft } from "../../../model/launch";
import { memberIssue } from "../../../model/launch";
import { cx, panel, useFocusOnMount } from "../../primitives";
import launch from "./LaunchPanel.module.css";

export interface DraftEdit {
  readonly draft: SessionDraft;
  readonly change: (next: SessionDraft) => void;
  readonly issues: readonly PlanIssue[];
}

export function IssueNote({ id, issue }: { readonly id: string; readonly issue: PlanIssue | undefined }) {
  return (
    <small id={id} className={cx(launch.problem, launch.inField)}>
      {issue !== undefined && <TriangleAlert size={13} aria-hidden="true" />}
      {issue?.message ?? ""}
    </small>
  );
}

export function flagged(issue: PlanIssue | undefined, id: string) {
  return issue === undefined ? {} : { "aria-invalid": true, "aria-describedby": id };
}

interface TextFieldProps {
  readonly id: string;
  readonly label: string;
  readonly value: string;
  readonly issue: PlanIssue | undefined;
  readonly placeholder: string;
  readonly onText: (text: string) => void;
}

function TextField({ id, label, value, issue, placeholder, onText }: TextFieldProps) {
  const first = useFocusOnMount<HTMLInputElement>({ when: id === "session-seed" });
  return (
    <div className={panel.field}>
      <label htmlFor={id}>{label}</label>
      <input
        ref={first}
        id={id}
        value={value}
        placeholder={placeholder}
        {...flagged(issue, `${id}-issue`)}
        onChange={(event) => {
          onText(event.target.value);
        }}
      />
      <IssueNote id={`${id}-issue`} issue={issue} />
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
  ["seed", "Seed", "1", "seed", "session-seed"],
  ["host_socket", "Host socket", "/path/to/host.sock", "hostSocket", "session-host"],
  ["gateway_socket", "Gateway socket", "optional: /path/to/gateway.sock", "gatewaySocket", "session-gateway"],
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
          issue={memberIssue(issues, member)}
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
