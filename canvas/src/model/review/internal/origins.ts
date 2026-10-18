import type { InstanceOrigin } from "../../graph";

export function withOrigin(
  origins: ReadonlyMap<string, InstanceOrigin>,
  origin: InstanceOrigin,
): ReadonlyMap<string, InstanceOrigin> {
  if (origin.instance === "") return origins;
  return new Map([...origins, [origin.instance, origin]]);
}
