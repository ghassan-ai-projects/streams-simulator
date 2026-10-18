import { looksLikeFileStem } from "./rules";
import type { CommandChoice, Drive, KindChoice, TwinChoices, TwinIssue } from "./twinTypes";

const idPattern = /^[a-z][a-z0-9-]{2,63}$/;

export const twinNameRule =
  "3 to 64 lower-case letters, digits and hyphens, starting with a letter. The plant made from the twin is saved under this name.";

export function twinPath(stem: string): string {
  return `twins/${stem}.twin.json`;
}

export function twinDomainPath(name: string): string {
  return `domains/${name}.domain.json`;
}

export function twinNameLooksRight(name: string): boolean {
  return idPattern.test(name);
}

function problem(scope: TwinIssue["scope"], kind: string, key: string, field: string, message: string): TwinIssue {
  return { scope, kind, key, field, spot: "", message };
}

function twinProblems(choices: TwinChoices): readonly TwinIssue[] {
  const named = looksLikeFileStem(choices.id)
    ? []
    : [problem("twin", "", "", "id", "Give the twin a name: letters, digits and - only.")];
  const titled = choices.title.trim() === "" ? [problem("twin", "", "", "title", "Give the twin a title.")] : [];
  return [...named, ...titled];
}

function incomplete(drive: Drive): boolean {
  const missing = drive.from === "" || drive.kind === "" || drive.metric === "";
  return missing || drive.gain === null || drive.timeConstantS === null;
}

function commandProblems(kind: string, name: string, choice: CommandChoice): readonly TwinIssue[] {
  const sets = Object.keys(choice.sets).length > 0;
  const unset = choice.response !== "none" && !sets;
  const unfinished =
    choice.response === "first_order" && (choice.drives.length === 0 || choice.drives.some(incomplete));
  return [
    ...(unset ? [problem("command", kind, name, "sets", "Choose the reading this command sets.")] : []),
    ...(unfinished ? [problem("command", kind, name, "drives", "Fill in every part of the effect.")] : []),
  ];
}

function kindProblems(kind: KindChoice): readonly TwinIssue[] {
  const named =
    kind.instance.trim() === ""
      ? [problem("kind", kind.kind, "", "instance", "Give the device a name in the twin.")]
      : [];
  const commands = Object.entries(kind.commands).flatMap(([name, choice]) => commandProblems(kind.kind, name, choice));
  return [...named, ...commands];
}

export function choiceProblems(choices: TwinChoices): readonly TwinIssue[] {
  return [...twinProblems(choices), ...choices.kinds.flatMap(kindProblems)];
}
