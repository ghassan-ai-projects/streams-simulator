import { settleCommand } from "./twinMembers";
import type { CommandChoice, Drive, KindChoice, MetricChoice, Response, TwinChoices } from "./twinTypes";

const blankDrive: Drive = { from: "", kind: "", metric: "", gain: null, timeConstantS: null };

function changeKind(choices: TwinChoices, kind: string, change: (kind: KindChoice) => KindChoice): TwinChoices {
  return { ...choices, kinds: choices.kinds.map((entry) => (entry.kind === kind ? change(entry) : entry)) };
}

function changeMetricChoice(
  choices: TwinChoices,
  kind: string,
  metric: string,
  change: (choice: MetricChoice) => MetricChoice,
): TwinChoices {
  return changeKind(choices, kind, (entry) => {
    const current = entry.metrics[metric];
    return current === undefined ? entry : { ...entry, metrics: { ...entry.metrics, [metric]: change(current) } };
  });
}

function changeCommandChoice(
  choices: TwinChoices,
  kind: string,
  command: string,
  change: (choice: CommandChoice) => CommandChoice,
): TwinChoices {
  return changeKind(choices, kind, (entry) => {
    const current = entry.commands[command];
    return current === undefined ? entry : { ...entry, commands: { ...entry.commands, [command]: change(current) } };
  });
}

export function renameTwin(choices: TwinChoices, id: string): TwinChoices {
  return { ...choices, id };
}

export function retitleTwin(choices: TwinChoices, title: string): TwinChoices {
  return { ...choices, title };
}

export function renameInstance(choices: TwinChoices, kind: string, instance: string): TwinChoices {
  return changeKind(choices, kind, (entry) => ({ ...entry, instance }));
}

export function chooseInstrument(choices: TwinChoices, kind: string, metric: string, instrument: string): TwinChoices {
  return changeMetricChoice(choices, kind, metric, (choice) => ({ ...choice, instrument }));
}

export function startAt(choices: TwinChoices, kind: string, metric: string, initial: number | null): TwinChoices {
  return changeMetricChoice(choices, kind, metric, (choice) => ({ ...choice, initial }));
}

function withDrives(current: CommandChoice, response: Response): readonly Drive[] {
  if (response !== "first_order") return [];
  return current.drives.length > 0 ? current.drives : [blankDrive];
}

export function chooseResponse(choices: TwinChoices, kind: string, command: string, response: Response): TwinChoices {
  return changeCommandChoice(choices, kind, command, (current) =>
    settleCommand({
      response,
      sets: response === "none" ? {} : current.sets,
      drives: withDrives(current, response),
    }),
  );
}

export function setArgument(
  choices: TwinChoices,
  kind: string,
  command: string,
  argument: string,
  reading: string,
): TwinChoices {
  return changeCommandChoice(choices, kind, command, (current) => {
    const rest = Object.entries(current.sets).filter(([name]) => name !== argument);
    const next = reading === "" ? rest : [...rest, [argument, reading] as const];
    return settleCommand({ ...current, sets: Object.fromEntries(next) });
  });
}

export function addDrive(choices: TwinChoices, kind: string, command: string): TwinChoices {
  return changeCommandChoice(choices, kind, command, (current) =>
    settleCommand({ ...current, drives: [...current.drives, blankDrive] }),
  );
}

export function changeDrive(
  choices: TwinChoices,
  at: { readonly kind: string; readonly command: string; readonly index: number },
  change: Partial<Drive>,
): TwinChoices {
  return changeCommandChoice(choices, at.kind, at.command, (current) => ({
    ...current,
    drives: current.drives.map((drive, index) => (index === at.index ? { ...drive, ...change } : drive)),
  }));
}

export function removeDrive(choices: TwinChoices, kind: string, command: string, index: number): TwinChoices {
  return changeCommandChoice(choices, kind, command, (current) => ({
    ...current,
    drives: current.drives.filter((_drive, at) => at !== index),
  }));
}
