import { Plus } from "lucide-react";
import {
  addDrive,
  changeDrive,
  driveFollows,
  effectWords,
  followChoices,
  followTarget,
  gainHint,
  noFollowerHint,
  removeDrive,
  timeConstantHint,
  twinIssuesOn,
} from "../../../model/devices";
import type { Command, CommandChoice, Drive, KindFile, Member } from "../../../model/devices";
import { cx, panel, shared } from "../../primitives";
import tw from "./Behaviour.module.css";
import { Field, ItemIssues, NumberInput, RemoveButton, Select } from "./Field";
import { choose, commit, persist } from "./twinProps";
import type { TwinProps } from "./twinProps";

interface DriveProps extends Pick<TwinProps, "twin" | "issues"> {
  readonly kind: KindFile;
  readonly members: readonly Member[];
  readonly member: string;
  readonly command: Command;
  readonly choice: CommandChoice;
}

interface EffectFields {
  readonly twin: TwinProps["twin"];
  readonly drive: Drive;
  readonly choice: CommandChoice;
  readonly kind: KindFile;
  readonly members: readonly Member[];
  readonly on: (field: string) => readonly string[];
  readonly change: (edit: Partial<Drive>) => void;
}

function Timing({ twin, drive, on, change }: EffectFields) {
  return (
    <>
      <Field label="Gain" field="gain" hint={gainHint} issues={on("gain")}>
        {(aria) => (
          <NumberInput
            aria={aria}
            value={drive.gain}
            onChange={(gain) => {
              change({ gain });
            }}
            onCommit={commit(twin)}
          />
        )}
      </Field>
      <Field label="Time constant" field="time_constant_s" hint={timeConstantHint} issues={on("time_constant_s")}>
        {(aria) => (
          <NumberInput
            aria={aria}
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

function Source({ twin, drive, choice, kind, on, change }: EffectFields) {
  const sources = [...new Set(Object.values(choice.sets))].flatMap((key) => {
    const name = kind.metrics.find((metric) => metric.key === key)?.name;
    return name === undefined ? [] : [[key, name] as const];
  });
  if (sources.length < 2) return null;
  return (
    <Field label="Reading it sets" field="from" issues={on("from")}>
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
  );
}

function Follower({ twin, drive, members, on, change }: EffectFields) {
  const choices = followChoices(members);
  return (
    <Field
      label="Reading that follows"
      field="follows"
      hint={choices.length === 0 ? noFollowerHint : undefined}
      issues={on("follows")}
    >
      {(aria) => (
        <Select
          aria={aria}
          value={driveFollows(drive)}
          choices={choices}
          onChange={(value) => {
            change(followTarget(value));
            persist(twin);
          }}
        />
      )}
    </Field>
  );
}

function Effect(props: DriveProps & { readonly drive: Drive; readonly index: number }) {
  const { twin, issues, member, command, drive, index } = props;
  const on = (field: string) =>
    twinIssuesOn(issues, { scope: "command", kind: member, key: command.key, field, spot: String(index) });
  const at = { kind: member, command: command.key, index };
  const change = (edit: Partial<Drive>) => {
    twin.edit((all) => changeDrive(all, at, edit));
  };
  const fields = { ...props, on, change };
  const words = effectWords(drive, member, props.members);
  return (
    <div role="group" aria-label={`Effect ${String(index + 1)} of ${command.name}`} className={tw.drive}>
      <ItemIssues messages={on("drives")} />
      <Source {...fields} />
      <Follower {...fields} />
      <Timing {...fields} />
      {words !== "" && <p className={panel.hint}>{words}</p>}
      <RemoveButton
        words="Remove this effect"
        onRemove={() => {
          choose(twin, (all) => removeDrive(all, member, command.key, index));
        }}
      />
    </div>
  );
}

export function DriveRows(props: DriveProps) {
  const { twin, member, command, choice } = props;
  return (
    <section className={panel.group} aria-label={`Effects of ${command.name}`}>
      {choice.drives.map((drive, index) => (
        <Effect key={String(index)} {...props} drive={drive} index={index} />
      ))}
      <button
        type="button"
        className={cx(shared.secondary, panel.secondary)}
        onClick={() => {
          choose(twin, (all) => addDrive(all, member, command.key));
        }}
      >
        <Plus size={14} aria-hidden="true" />
        Add another effect
      </button>
    </section>
  );
}
