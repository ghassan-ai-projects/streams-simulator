import type { CommandChoice, Drive, KindChoice, TwinChoices, TwinIssue } from "./twinTypes";

export function twinPath(stem: string): string {
  return `twins/${stem}.twin.json`;
}

export function twinDomainPath(name: string): string {
  return `domains/${name}.domain.json`;
}

function problem(scope: TwinIssue["scope"], kind: string, key: string, field: string, text: string): TwinIssue {
  return { scope, kind, key, field, spot: "", pointer: "", keyword: "", message: text, text };
}

function twinProblems(choices: TwinChoices): readonly TwinIssue[] {
  return [
    ...(choices.id.trim() === "" ? [problem("twin", "", "", "id", "Give the twin a name.")] : []),
    ...(choices.title.trim() === "" ? [problem("twin", "", "", "title", "Give the twin a title.")] : []),
  ];
}

function incomplete(drive: Drive): boolean {
  const missing = drive.from === "" || drive.kind === "" || drive.metric === "";
  return missing || drive.gain === null || drive.timeConstantS === null;
}

function slowness(kind: string, key: string, drives: readonly Drive[]): readonly TwinIssue[] {
  return drives.flatMap((drive, index) =>
    drive.timeConstantS !== null && drive.timeConstantS <= 0
      ? [{ ...problem("command", kind, key, "time_constant_s", "Must be above zero"), spot: String(index) }]
      : [],
  );
}

function commandProblems(kind: string, key: string, choice: CommandChoice): readonly TwinIssue[] {
  const sets = Object.keys(choice.sets).length > 0;
  const unset = choice.response !== "none" && !sets;
  const unfinished =
    choice.response === "first_order" && (choice.drives.length === 0 || choice.drives.some(incomplete));
  return [
    ...(unset ? [problem("command", kind, key, "sets", "Choose the reading this command sets.")] : []),
    ...(unfinished ? [problem("command", kind, key, "drives", "Fill in every part of the effect.")] : []),
    ...slowness(kind, key, choice.drives),
  ];
}

function kindProblems(choice: KindChoice): readonly TwinIssue[] {
  const named =
    choice.instance.trim() === ""
      ? [problem("kind", choice.kind, "", "instance", "Give the device a name in the twin.")]
      : [];
  const commands = Object.entries(choice.commands).flatMap(([key, value]) => commandProblems(choice.kind, key, value));
  return [...named, ...commands];
}

export function choiceProblems(choices: TwinChoices): readonly TwinIssue[] {
  return [...twinProblems(choices), ...choices.kinds.flatMap(kindProblems)];
}
