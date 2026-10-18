import { serialiseNamed } from "./twinFile";
import type { CommandChoice, Drive, KindChoice, TwinChoices } from "./twinTypes";
import type { KindFile, Member } from "./types";

function metricKey(kind: KindFile, name: string): string {
  return kind.metrics.find((metric) => metric.name === name)?.key ?? "";
}

function metricName(kind: KindFile, key: string): string {
  return kind.metrics.find((metric) => metric.key === key)?.name ?? "";
}

function commandKey(kind: KindFile, name: string): string {
  return kind.commands.find((command) => command.name === name)?.key ?? "";
}

function commandName(kind: KindFile, key: string): string {
  return kind.commands.find((command) => command.key === key)?.name ?? "";
}

function byName(members: readonly Member[], name: string): Member | undefined {
  return members.find((member) => member.kind.kind === name);
}

function byPath(members: readonly Member[], path: string): Member | undefined {
  return members.find((member) => member.path === path);
}

function remap<T>(
  source: Readonly<Record<string, T>>,
  convert: (name: string) => string,
  change: (value: T) => T = (value) => value,
): Readonly<Record<string, T>> {
  const entries = Object.entries(source).map(([name, value]) => [convert(name), change(value)] as const);
  return Object.fromEntries(entries.filter(([name]) => name !== ""));
}

interface Translation {
  readonly find: (members: readonly Member[], ref: string) => Member | undefined;
  readonly refOf: (member: Member) => string;
  readonly metric: (kind: KindFile, ref: string) => string;
  readonly command: (kind: KindFile, ref: string) => string;
}

const toKeys: Translation = {
  find: byName,
  refOf: (member) => member.path,
  metric: metricKey,
  command: commandKey,
};

const toNames: Translation = {
  find: byPath,
  refOf: (member) => member.kind.kind,
  metric: metricName,
  command: commandName,
};

function driveOf(drive: Drive, kind: KindFile, members: readonly Member[], way: Translation): Drive {
  const target = way.find(members, drive.kind);
  return {
    ...drive,
    from: way.metric(kind, drive.from),
    kind: target === undefined ? "" : way.refOf(target),
    metric: target === undefined ? "" : way.metric(target.kind, drive.metric),
  };
}

function setsOf(sets: CommandChoice["sets"], kind: KindFile, way: Translation): CommandChoice["sets"] {
  const mapped = Object.entries(sets).map(([argument, reading]) => [argument, way.metric(kind, reading)] as const);
  return Object.fromEntries(mapped.filter(([, reading]) => reading !== ""));
}

function commandOf(choice: CommandChoice, kind: KindFile, members: readonly Member[], way: Translation) {
  return {
    ...choice,
    sets: setsOf(choice.sets, kind, way),
    drives: choice.drives.map((drive) => driveOf(drive, kind, members, way)),
  };
}

function kindOf(choice: KindChoice, member: Member, members: readonly Member[], way: Translation): KindChoice {
  const { kind } = member;
  const command = (value: CommandChoice) => commandOf(value, kind, members, way);
  return {
    kind: way.refOf(member),
    instance: choice.instance,
    metrics: remap(choice.metrics, (ref) => way.metric(kind, ref)),
    commands: remap(choice.commands, (ref) => way.command(kind, ref), command),
  };
}

function translate(choices: TwinChoices, members: readonly Member[], way: Translation): TwinChoices {
  const kinds = choices.kinds.flatMap((choice) => {
    const member = way.find(members, choice.kind);
    return member === undefined ? [] : [kindOf(choice, member, members, way)];
  });
  return { ...choices, kinds };
}

export function keyChoices(named: TwinChoices, members: readonly Member[]): TwinChoices {
  return translate(named, members, toKeys);
}

export function nameChoices(choices: TwinChoices, members: readonly Member[]): TwinChoices {
  return translate(choices, members, toNames);
}

export function serialiseChoices(choices: TwinChoices, members: readonly Member[]): string {
  return serialiseNamed(nameChoices(choices, members));
}
