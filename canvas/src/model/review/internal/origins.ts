import type { InstanceOrigin } from "../../graph";
import { plannedSimulated, unknownOrigin } from "../../graph";
import type { BindingNote } from "../../launch";
import { bindingWords } from "../../launch";

export type Connection = "none" | "running" | "saved";

export function withOrigin(
  origins: ReadonlyMap<string, InstanceOrigin>,
  origin: InstanceOrigin,
): ReadonlyMap<string, InstanceOrigin> {
  if (origin.instance === "") return origins;
  return new Map([...origins, [origin.instance, origin]]);
}

export function shownOrigins(
  origins: ReadonlyMap<string, InstanceOrigin>,
  instances: readonly string[],
  connection: Connection,
  simulated: readonly string[] = [],
): ReadonlyMap<string, InstanceOrigin> {
  if (connection === "none") return origins;
  const absent = (id: string) => (simulated.includes(id) ? plannedSimulated(id) : unknownOrigin(id));
  return new Map(instances.map((id) => [id, origins.get(id) ?? absent(id)]));
}

export function bindingLine(binding: BindingNote): string {
  return bindingWords(binding);
}

export function connectionWords(connection: "running" | "saved"): string {
  return connection === "running" ? "Connected session · running" : "Connected session";
}
