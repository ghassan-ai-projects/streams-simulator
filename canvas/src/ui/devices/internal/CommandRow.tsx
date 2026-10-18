import {
  argumentLabel,
  commandFacts,
  noSetpointHint,
  numericArguments,
  pickResponse,
  placeKey,
  responseChoices,
  responseOf,
  setArgument,
  settableChoices,
  twinIssuesOn,
} from "../../../model/devices";
import type { Command, CommandChoice, KindFile, Member } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import tw from "./Behaviour.module.css";
import { DriveRows } from "./DriveRows";
import { Field, ItemIssues, Select } from "./Field";
import { InterfaceNeeds } from "./InterfaceNeeds";
import { choose } from "./twinProps";
import type { Reveal, TwinProps } from "./twinProps";

interface RowProps extends Pick<TwinProps, "twin" | "issues" | "needs"> {
  readonly reveal: Reveal | null;
  readonly kind: KindFile;
  readonly members: readonly Member[];
  readonly member: string;
  readonly command: Command;
  readonly choice: CommandChoice;
}

function SetpointButton({ kind, reveal }: Pick<RowProps, "kind" | "reveal">) {
  const first = kind.metrics.find((metric) => metric.name !== "");
  if (reveal === null) return null;
  return (
    <button
      type="button"
      className={cx(shared.secondary, panel.secondary)}
      onClick={() => {
        reveal(first === undefined ? { part: "kind" } : { part: "metric", key: first.key }, "writable");
      }}
    >
      Set it in the Interface step
    </button>
  );
}

function Sets({ twin, issues, kind, member, command, choice, reveal }: RowProps) {
  const numbers = numericArguments(command);
  const settable = settableChoices(kind);
  const on = (field: string, spot = "") =>
    twinIssuesOn(issues, { scope: "command", kind: member, key: command.key, field, spot });
  if (choice.response === "none") return null;
  if (numbers.length === 0) {
    return <ItemIssues messages={["This command has no number argument, so it cannot set a reading."]} />;
  }
  return (
    <>
      <ItemIssues messages={on("sets")} />
      {numbers.map((argument) => (
        <Field
          key={argument}
          label={argumentLabel(argument)}
          field="sets"
          place={placeKey({ scope: "command", kind: member, key: command.key, field: "sets", spot: argument })}
          hint={settable.length === 1 ? noSetpointHint : undefined}
          issues={on("sets", argument)}
        >
          {(aria) => (
            <Select
              aria={aria}
              value={choice.sets[argument] ?? ""}
              choices={settable}
              onChange={(reading) => {
                choose(twin, (all) => setArgument(all, member, command.key, argument, reading));
              }}
            />
          )}
        </Field>
      ))}
      {settable.length === 1 && <SetpointButton kind={kind} reveal={reveal} />}
    </>
  );
}

export function CommandRow(props: RowProps) {
  const { twin, issues, member, command, choice } = props;
  const on = (field: string) => twinIssuesOn(issues, { scope: "command", kind: member, key: command.key, field });
  return (
    <div role="group" aria-label={`Command ${command.name}`} className={tw.reading}>
      <h4 className={tw.readingHead}>
        <strong>{command.name}</strong>
        <span className={panel.hint}>{commandFacts(command)}</span>
      </h4>
      <ItemIssues messages={[...on(""), ...on("drives")]} />
      <Field
        label="Response"
        field="response"
        place={placeKey({ scope: "command", kind: member, key: command.key, field: "response" })}
        issues={on("response")}
      >
        {(aria) => (
          <Select
            aria={aria}
            value={choice.response}
            choices={responseChoices}
            onChange={(value) => {
              choose(twin, (all) => pickResponse(all, { member, kind: props.kind, command }, responseOf(value)));
            }}
          />
        )}
      </Field>
      <InterfaceNeeds needs={props.needs} part="command" itemKey={command.key} reveal={props.reveal} />
      <Sets {...props} />
      {choice.response === "first_order" && <DriveRows {...props} />}
    </div>
  );
}
