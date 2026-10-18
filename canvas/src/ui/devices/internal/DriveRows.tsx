import { Plus } from "lucide-react";
import {
  addDrive,
  changeDrive,
  driveFollows,
  effectWords,
  followChoices,
  followTarget,
  gainHint,
  removeDrive,
  timeConstantHint,
  twinIssuesOn,
} from "../../../model/devices";
import type { Command, CommandChoice, Drive, KindFile } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import tw from "./Behaviour.module.css";
import { Field, ItemIssues, NumberInput, RemoveButton, Select } from "./Field";
import { choose, commit, persist } from "./twinProps";
import type { TwinProps } from "./twinProps";

interface DriveProps extends Pick<TwinProps, "twin" | "issues"> {
  readonly kind: KindFile;
  readonly kinds: readonly KindFile[];
  readonly command: Command;
  readonly choice: CommandChoice;
}

interface EffectFields {
  readonly twin: TwinProps["twin"];
  readonly drive: Drive;
  readonly choice: CommandChoice;
  readonly kinds: readonly KindFile[];
  readonly on: (field: string) => readonly string[];
  readonly change: (edit: Partial<Drive>) => void;
}

function Timing({ twin, drive, on, change }: EffectFields) {
  return (
    <>
      <Field label="Gain" hint={gainHint} issues={on("gain")}>
        {(aria) => (
          <NumberInput
            aria={aria}
            name="Gain"
            value={drive.gain}
            onChange={(gain) => {
              change({ gain });
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
      <Field label="Time constant" hint={timeConstantHint} issues={on("time_constant_s")}>
        {(aria) => (
          <NumberInput
            aria={aria}
            name="Time constant in seconds"
            value={drive.timeConstantS}
            onChange={(timeConstantS) => {
              change({ timeConstantS });
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
    </>
  );
}

function Targets({ twin, drive, choice, kinds, on, change }: EffectFields) {
  const sources = [...new Set(Object.values(choice.sets))].map((name) => [name, name] as const);
  return (
    <>
      <Field label="Reading it sets" issues={on("from")}>
        {(aria) => (
          <Select
            aria={aria}
            value={drive.from}
            choices={sources}
            onChange={(from) => {
              change({ from });
              persist(twin);
            }}
          />
        )}
      </Field>
      <Field label="Reading that follows" issues={on("follows")}>
        {(aria) => (
          <Select
            aria={aria}
            value={driveFollows(drive)}
            choices={followChoices(kinds)}
            onChange={(value) => {
              change(followTarget(value));
              persist(twin);
            }}
          />
        )}
      </Field>
    </>
  );
}

function Effect(props: DriveProps & { readonly drive: Drive; readonly index: number }) {
  const { twin, issues, kind, command, drive, index } = props;
  const on = (field: string) =>
    twinIssuesOn(issues, { scope: "command", kind: kind.kind, key: command.name, field, spot: String(index) });
  const at = { kind: kind.kind, command: command.name, index };
  const change = (edit: Partial<Drive>) => {
    twin.edit((all) => changeDrive(all, at, edit));
  };
  const fields = { ...props, on, change };
  const words = effectWords(drive);
  return (
    <div role="group" aria-label={`Effect ${String(index + 1)} of ${command.name}`} className={tw.drive}>
      <ItemIssues messages={on("drives")} />
      <Targets {...fields} />
      <Timing {...fields} />
      {words !== "" && <p className={panel.hint}>{words}</p>}
      <RemoveButton
        words="Remove this effect"
        onRemove={() => {
          choose(twin, (all) => removeDrive(all, kind.kind, command.name, index));
        }}
      />
    </div>
  );
}

export function DriveRows(props: DriveProps) {
  const { twin, kind, command, choice } = props;
  return (
    <section className={panel.group} aria-label={`Effects of ${command.name}`}>
      {choice.drives.map((drive, index) => (
        <Effect key={String(index)} {...props} drive={drive} index={index} />
      ))}
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          choose(twin, (all) => addDrive(all, kind.kind, command.name));
        }}
      >
        <Plus size={14} aria-hidden="true" />
        Add another effect
      </button>
    </section>
  );
}
