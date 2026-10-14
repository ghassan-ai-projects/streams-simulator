import type { Instance, Plant } from "./plant";

export type Family = "air" | "pressure" | "water" | "cw" | "temp" | "heat" | "power" | "signal" | "env";
export type Shape = "circle" | "square" | "diamond";

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

const shapes: Readonly<Record<Family, Shape>> = {
  air: "circle",
  water: "circle",
  cw: "circle",
  heat: "square",
  power: "square",
  pressure: "diamond",
  temp: "diamond",
  signal: "diamond",
  env: "diamond",
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
}

export type Category = "air" | "cool" | "prod" | "elec" | "env";

export interface ComponentSpec {
  readonly ports: readonly PortSpec[];
  readonly category: Category;
}

export function portsMatching(spec: ComponentSpec | undefined, direction: "in" | "out", type = ""): readonly string[] {
  return (spec?.ports ?? [])
    .filter((port) => port.direction === direction && (type === "" || port.type === "" || port.type === type))
    .map((port) => port.name);
}

export function portType(spec: ComponentSpec | undefined, name: string): string {
  return spec?.ports.find((port) => port.name === name)?.type ?? "";
}

const classCategories: readonly (readonly [string, Category])[] = [
  ["electrical/", "elec"],
  ["process/", "prod"],
  ["environment/", "env"],
];

const tagCategories: readonly (readonly [string, Category])[] = [
  ["compressed-air", "air"],
  ["hvac", "cool"],
  ["fluids", "cool"],
  ["heat", "cool"],
  ["drives", "prod"],
  ["material-handling", "prod"],
];

function categoryOf(componentClass: string, tags: readonly string[]): Category {
  const byClass = classCategories.find(([prefix]) => componentClass.startsWith(prefix));
  if (byClass) return byClass[1];
  return tagCategories.find(([tag]) => tags.includes(tag))?.[1] ?? "env";
}

export function parseComponent(shown: unknown): ComponentSpec {
  const entry = field(shown, "entry");
  const tags = field(entry, "tags");
  const tagList = Array.isArray(tags) ? tags.filter((tag): tag is string => typeof tag === "string") : [];
  return { ports: parsePorts(field(shown, "document")), category: categoryOf(text(field(entry, "class")), tagList) };
}

type Connection = "connected" | "open" | "unknown";

export interface Pin {
  readonly name: string;
  readonly type: string;
  readonly family: Family | null;
  readonly connection: Connection;
}

export interface Pins {
  readonly ins: readonly Pin[];
  readonly outs: readonly Pin[];
}

function field(value: unknown, key: string): unknown {
  return typeof value === "object" && value !== null ? (value as Record<string, unknown>)[key] : undefined;
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

export function parsePorts(document: unknown): readonly PortSpec[] {
  const ports = field(document, "ports");
  return (Array.isArray(ports) ? ports : []).map((port: unknown) => ({
    name: text(field(port, "name")),
    type: text(field(port, "type")),
    direction: field(port, "direction") === "provides" ? ("out" as const) : ("in" as const),
  }));
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
  return [...names].sort().map((name) => ({ name, type: "", family: null, connection: "connected" as const }));
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
  });
  return {
    ins: specs.filter((spec) => spec.direction === "in").map((spec) => pin(spec, connected.ins, connected.unknownIn)),
    outs: specs
      .filter((spec) => spec.direction === "out")
      .map((spec) => pin(spec, connected.outs, connected.unknownOut)),
  };
}
