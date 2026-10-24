import { humanise, spaced } from "../../shared/words";

interface CallLike {
  readonly entity: string;
  readonly effector: string;
  readonly args: Readonly<Record<string, unknown>>;
}

function unitWords(symbol: string): string {
  return symbol === "" ? "" : ` ${symbol}`;
}

export type UnitOf = (entity: string, effector: string, arg: string) => string;

export const noUnits: UnitOf = () => "";

export function callWords(call: CallLike, unitOf: UnitOf = noUnits): string {
  if (call.entity === "" || call.effector === "") return "an effector call to choose";
  const given = Object.entries(call.args)
    .filter(([, value]) => value !== call.entity)
    .map(([name, value]) => `${spaced(name)} ${String(value)}${unitWords(unitOf(call.entity, call.effector, name))}`);
  const args = given.length === 0 ? "" : ` (${given.join(", ")})`;
  return `${call.entity} ${humanise(call.effector)}${args}`;
}
