import {
  chooseResponse,
  commandFacts,
  numericArguments,
  responseChoices,
  responseOf,
  setArgument,
  settableChoices,
  spaced,
  twinIssuesOn,
} from "../../../model/devices";
import type { Command, CommandChoice, KindFile, Member } from "../../../model/devices";
import { panel } from "../../primitives";
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

function Sets({ twin, issues, kind, member, command, choice }: RowProps) {
  const numbers = numericArguments(command);
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
        <Field key={argument} label={`Reading set by ${spaced(argument)}`} field="sets" issues={on("sets", argument)}>
          {(aria) => (
            <Select
              aria={aria}
              value={choice.sets[argument] ?? ""}
              choices={settableChoices(kind)}
              onChange={(reading) => {
                choose(twin, (all) => setArgument(all, member, command.key, argument, reading));
              }}
            />
          )}
        </Field>
      ))}
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
      <Field label="Response" field="response" issues={on("response")}>
        {(aria) => (
          <Select
            aria={aria}
            value={choice.response}
            choices={responseChoices}
            onChange={(value) => {
              choose(twin, (all) => chooseResponse(all, member, command.key, responseOf(value)));
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
