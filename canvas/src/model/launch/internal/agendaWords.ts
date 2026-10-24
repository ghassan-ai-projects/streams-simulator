import { humanise, spaced } from "../../shared/words";

interface CallLike {
  readonly entity: string;
  readonly effector: string;
  readonly args: Readonly<Record<string, unknown>>;
}

export function callWords(call: CallLike): string {
  if (call.entity === "" || call.effector === "") return "an effector call to choose";
  const given = Object.entries(call.args)
    .filter(([, value]) => value !== call.entity)
    .map(([name, value]) => `${spaced(name)} ${String(value)}`);
  const args = given.length === 0 ? "" : ` (${given.join(", ")})`;
  return `${call.entity} ${humanise(call.effector)}${args}`;
}
