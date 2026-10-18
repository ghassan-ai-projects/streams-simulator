import {
  askedCommands,
  neutralCommand,
  neutralMetric,
  numericArguments,
  settleCommand,
  writableReadings,
} from "./twinMembers";
import type { CommandChoice, Drive, KindChoice, TwinChoices } from "./twinTypes";
import type { Command, KindFile, Member } from "./types";

function neutralKind(member: Member): KindChoice {
  return { kind: member.path, instance: member.kind.kind, metrics: {}, commands: {} };
}

function keepSets(kind: KindFile, command: Command, choice: CommandChoice): CommandChoice {
  const arguments_ = numericArguments(command);
  const writable = writableReadings(kind).map((metric) => metric.key);
  const kept = Object.entries(choice.sets).filter(
    ([argument, reading]) => arguments_.includes(argument) && writable.includes(reading),
  );
  return { ...choice, sets: Object.fromEntries(kept) };
}

function reconcileCommands(member: Member, base: KindChoice, existing: KindChoice | undefined) {
  const commands = askedCommands(member.kind).map((command) => {
    const previous = existing?.commands[command.key];
    const choice =
      previous === undefined
        ? (base.commands[command.key] ?? neutralCommand)
        : keepSets(member.kind, command, previous);
    return [command.key, settleCommand(choice)] as const;
  });
  return Object.fromEntries(commands);
}

function reconcileKind(member: Member, existing: KindChoice | undefined, fresh: KindChoice | undefined): KindChoice {
  const base = fresh ?? neutralKind(member);
  const named = member.kind.metrics.filter((metric) => metric.name !== "");
  const metrics = named.map(
    (metric) => [metric.key, existing?.metrics[metric.key] ?? base.metrics[metric.key] ?? neutralMetric] as const,
  );
  return {
    kind: member.path,
    instance: existing?.instance ?? base.instance,
    metrics: Object.fromEntries(metrics),
    commands: reconcileCommands(member, base, existing),
  };
}

function targetKnown(drive: Drive, kinds: readonly KindChoice[]): boolean {
  return kinds.some((kind) => kind.kind === drive.kind && drive.metric in kind.metrics);
}

function clearDangling(choice: CommandChoice, kinds: readonly KindChoice[]): CommandChoice {
  const clear = (drive: Drive): Drive =>
    drive.metric === "" || targetKnown(drive, kinds) ? drive : { ...drive, kind: "", metric: "" };
  return { ...choice, drives: choice.drives.map(clear) };
}

function withoutDangling(kind: KindChoice, kinds: readonly KindChoice[]): KindChoice {
  const commands = Object.entries(kind.commands).map(([key, choice]) => [key, clearDangling(choice, kinds)] as const);
  return { ...kind, commands: Object.fromEntries(commands) };
}

export function reconcileChoices(
  choices: TwinChoices,
  members: readonly Member[],
  suggested: TwinChoices | null,
): TwinChoices {
  const find = (list: readonly KindChoice[] | undefined, member: Member) =>
    list?.find((kind) => kind.kind === member.path);
  const rebuilt = members.map((member) =>
    reconcileKind(member, find(choices.kinds, member), find(suggested?.kinds, member)),
  );
  return { ...choices, kinds: rebuilt.map((kind) => withoutDangling(kind, rebuilt)) };
}
