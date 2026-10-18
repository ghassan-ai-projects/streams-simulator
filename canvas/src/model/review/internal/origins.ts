import type { InstanceOrigin } from "../../graph";
import { unknownOrigin } from "../../graph";
import type { BindingNote } from "../../launch";
import { bindingWords } from "../../launch";

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
  connection: "none" | "running" | "saved",
): ReadonlyMap<string, InstanceOrigin> {
  if (connection === "none") return origins;
  return new Map(instances.map((id) => [id, origins.get(id) ?? unknownOrigin(id)]));
}

export function bindingLine(binding: BindingNote): string {
  return bindingWords(binding);
}

export function connectionWords(connection: "running" | "saved"): string {
  return connection === "running" ? "Connected session · running" : "Connected session";
}
