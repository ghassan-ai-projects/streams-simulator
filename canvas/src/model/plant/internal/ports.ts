import { spaced } from "../../shared/words";
import type { Instance, Plant } from "./plant";

export type Family = "air" | "pressure" | "water" | "cw" | "temp" | "heat" | "power" | "signal" | "env";
type Shape = "circle" | "square" | "diamond" | "hexagon" | "bar";

const families: Readonly<Record<string, Family>> = {
  compressed_air: "air",
  pressure: "pressure",
  chilled_water: "water",
  hot_water: "water",
  liquid_flow: "water",
  level: "water",
  steam: "water",
  condenser_water: "cw",
  temperature: "temp",
  cooling_water_temp: "temp",
  dew_point: "temp",
  thermal_load: "heat",
  heat_rejection: "heat",
  electrical: "power",
  load_torque: "power",
  shift: "signal",
  tariff: "signal",
  dry_bulb_temp: "env",
  wet_bulb_temp: "env",
};

// Look-alike hues never share a mark: the greys (env, signal) and the warm pair (temp, heat) split, so hue and mark both carry the family.
const shapes: Readonly<Record<Family, Shape>> = {
  air: "circle",
  water: "circle",
  cw: "circle",
  heat: "bar",
  power: "square",
  pressure: "diamond",
  temp: "hexagon",
  signal: "hexagon",
  env: "bar",
};

export const familyNames: Readonly<Record<Family, string>> = {
  air: "Compressed air",
  pressure: "Pressure",
  water: "Water",
  cw: "Condenser water",
  temp: "Temperature",
  heat: "Heat",
  power: "Power",
  signal: "Signal",
  env: "Weather",
};

export function familyOf(type: string): Family | null {
  return families[type] ?? null;
}

export function shapeOf(family: Family): Shape {
  return shapes[family];
}

export interface PortSpec {
  readonly name: string;
  readonly type: string;
  readonly direction: "in" | "out";
  readonly channel?: string;
}

export type Category = "air" | "cool" | "prod" | "elec" | "env";

export interface ComponentSpec {
  readonly ports: readonly PortSpec[];
  readonly category: Category;
  readonly title?: string;
  readonly description?: string;
}

export function portsMatching(spec: ComponentSpec | undefined, direction: "in" | "out", type = ""): readonly string[] {
  return (spec?.ports ?? [])
    .filter((port) => port.direction === direction && (type === "" || port.type === "" || port.type === type))
    .map((port) => port.name);
}

export function portType(spec: ComponentSpec | undefined, name: string): string {
  return spec?.ports.find((port) => port.name === name)?.type ?? "";
}

export function specOf(
  specs: ReadonlyMap<string, ComponentSpec>,
  instance: Instance | undefined,
): ComponentSpec | undefined {
  return instance?.component ? specs.get(instance.component) : undefined;
}

function listed(items: readonly string[]): string {
  if (items.length < 2) return items.join("");
  return `${items.slice(0, -1).join(", ")} and ${items.at(-1) ?? ""}`;
}

function inputTypesElsewhere(
  plant: Plant,
  specs: ReadonlyMap<string, ComponentSpec>,
  from: Instance,
): readonly string[] {
  const types = plant.instances
    .filter((instance) => instance.id !== from.id)
    .flatMap((instance) => specOf(specs, instance)?.ports ?? [])
    .filter((port) => port.direction === "in")
    .map((port) => port.type);
  return [...new Set(types)];
}

export function noTargetReason(
  plant: Plant,
  specs: ReadonlyMap<string, ComponentSpec>,
  from: Instance,
  output: string,
): string {
  const type = portType(specOf(specs, from), output);
  if (type === "") return "No instance in this plant has an input.";
  const taken = inputTypesElsewhere(plant, specs, from);
  const inputs =
    taken.length === 0
      ? "The other instances here have no inputs."
      : `The inputs in this plant take ${listed(taken.map(spaced))}.`;
  return `No instance takes ${spaced(type)}: ${from.id}'s ${spaced(output)} is ${spaced(type)}. ${inputs}`;
}

type Connection = "connected" | "open" | "unknown";

export interface Pin {
  readonly name: string;
  readonly type: string;
  readonly family: Family | null;
  readonly connection: Connection;
  readonly channel: string | null;
}

export interface Pins {
  readonly ins: readonly Pin[];
  readonly outs: readonly Pin[];
}

interface Used {
  readonly ins: ReadonlySet<string>;
  readonly outs: ReadonlySet<string>;
  readonly unknownIn: boolean;
  readonly unknownOut: boolean;
}

function used(plant: Plant, id: string): Used {
  const ins = new Set<string>();
  const outs = new Set<string>();
  let unknownIn = false;
  let unknownOut = false;
  for (const link of plant.links) {
    if (link.to === id && link.toPort === null) unknownIn = true;
    if (link.to === id && link.toPort !== null) ins.add(link.toPort);
    if (link.from === id && link.fromPort === null) unknownOut = true;
    if (link.from === id && link.fromPort !== null) outs.add(link.fromPort);
  }
  return { ins, outs, unknownIn, unknownOut };
}

function fromLinks(names: ReadonlySet<string>): readonly Pin[] {
  return [...names]
    .sort()
    .map((name) => ({ name, type: "", family: null, connection: "connected" as const, channel: null }));
}

export function pinsOf(plant: Plant, instance: Instance, specs: readonly PortSpec[] | undefined): Pins {
  const connected = used(plant, instance.id);
  if (specs === undefined || specs.length === 0) {
    return { ins: fromLinks(connected.ins), outs: fromLinks(connected.outs) };
  }
  const pin = (spec: PortSpec, names: ReadonlySet<string>, unknown: boolean): Pin => ({
    name: spec.name,
    type: spec.type,
    family: familyOf(spec.type),
    connection: names.has(spec.name) ? "connected" : unknown ? "unknown" : "open",
    channel: spec.channel ?? null,
  });
  return {
    ins: specs.filter((spec) => spec.direction === "in").map((spec) => pin(spec, connected.ins, connected.unknownIn)),
    outs: specs
      .filter((spec) => spec.direction === "out")
      .map((spec) => pin(spec, connected.outs, connected.unknownOut)),
  };
}
