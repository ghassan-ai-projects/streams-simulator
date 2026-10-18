import {
  chooseResponse,
  commandFacts,
  numericArguments,
  responseChoices,
  responseOf,
  setArgument,
  settableChoices,
  twinIssuesOn,
} from "../../../model/devices";
import type { Command, CommandChoice, KindFile } from "../../../model/devices";
import { panel } from "../../primitives";
import tw from "./Behaviour.module.css";
import { DriveRows } from "./DriveRows";
import { Field, ItemIssues, Select } from "./Field";
import { InterfaceNeeds } from "./InterfaceNeeds";
import { choose } from "./twinProps";
import type { Reveal, TwinProps } from "./twinProps";

interface RowProps extends Pick<TwinProps, "twin" | "issues"> {
  readonly reveal: Reveal | null;
  readonly kind: KindFile;
  readonly kinds: readonly KindFile[];
  readonly command: Command;
  readonly choice: CommandChoice;
}

function Sets({ twin, issues, kind, command, choice }: RowProps) {
  const numbers = numericArguments(command);
  const on = (field: string, spot = "") =>
    twinIssuesOn(issues, { scope: "command", kind: kind.kind, key: command.name, field, spot });
  if (choice.response === "none") return null;
  if (numbers.length === 0) {
    return <ItemIssues messages={["This command has no number argument, so it cannot set a reading."]} />;
  }
  return (
    <>
      <ItemIssues messages={on("sets")} />
      {numbers.map((argument) => (
        <Field key={argument} label={`${argument} sets`} issues={on("sets", argument)}>
          {(aria) => (
            <Select
              aria={aria}
              value={choice.sets[argument] ?? ""}
              choices={settableChoices(kind)}
              onChange={(reading) => {
                choose(twin, (all) => setArgument(all, kind.kind, command.name, argument, reading));
              }}
            />
          )}
        </Field>
      ))}
    </>
  );
}

export function CommandRow(props: RowProps) {
  const { twin, issues, kind, command, choice } = props;
  const on = (field: string) => twinIssuesOn(issues, { scope: "command", kind: kind.kind, key: command.name, field });
  return (
    <div role="group" aria-label={`Command ${command.name}`} className={tw.reading}>
      <h4 className={tw.readingHead}>
        <strong>{command.name}</strong>
        <span className={panel.hint}>{commandFacts(command)}</span>
      </h4>
      <ItemIssues messages={[...on(""), ...on("drives")]} />
      <Field label="Response" issues={on("response")}>
        {(aria) => (
          <Select
            aria={aria}
            value={choice.response}
            choices={responseChoices}
            onChange={(value) => {
              choose(twin, (all) => chooseResponse(all, kind.kind, command.name, responseOf(value)));
            }}
          />
        )}
      </Field>
      <InterfaceNeeds
        issues={issues}
        scope="command"
        kind={kind.kind}
        name={command.name}
        selection={{ part: "command", key: command.key }}
        reveal={props.reveal}
      />
      <Sets {...props} />
      {choice.response === "first_order" && <DriveRows {...props} />}
    </div>
  );
}
