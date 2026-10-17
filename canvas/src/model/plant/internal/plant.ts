import type { AreaRef } from "../../shared/areaTree";
import { groupBy } from "../../shared/collect";
import { list, record, text } from "../../shared/json";
import type { EffectorSpec } from "./effectors";
import { parseEffectors } from "./effectors";

export interface ChannelState {
  readonly value: number;
  readonly name: string;
  readonly attention: "alarm" | "warn" | null;
}

export interface ChannelRange {
  readonly min: number;
  readonly max: number;
}

interface Channel {
  readonly name: string;
  readonly unit: string;
  readonly states: readonly ChannelState[];
  readonly range: ChannelRange | null;
}

function rangeFrom(value: unknown): ChannelRange | null {
  const range = record(value);
  return typeof range.min === "number" && typeof range.max === "number" ? { min: range.min, max: range.max } : null;
}

function attention(value: unknown): ChannelState["attention"] {
  return value === "alarm" || value === "warn" ? value : null;
}

function channelStates(value: unknown): readonly ChannelState[] {
  return list(value).flatMap((raw) => {
    const entry = record(raw);
    return typeof entry.value === "number"
      ? [{ value: entry.value, name: text(entry.name), attention: attention(entry.attention) }]
      : [];
  });
}

interface Kind {
  readonly name: string;
  readonly channels: readonly Channel[];
  readonly states: readonly string[];
  readonly units: ReadonlyMap<string, string>;
  readonly faults: readonly string[];
  readonly effectors: readonly string[];
  readonly effectorSpecs: readonly EffectorSpec[];
}

export interface Instance {
  readonly id: string;
  readonly kind: string;
  readonly area: string;
  readonly component: string | null;
  readonly instruments: ReadonlyMap<string, string>;
}

export interface Link {
  readonly from: string;
  readonly to: string;
  readonly type: string;
  readonly fromPort: string | null;
  readonly toPort: string | null;
}

export interface Plant {
  readonly id: string;
  readonly title: string;
  readonly site: string;
  readonly areas: readonly AreaRef[];
  readonly instances: readonly Instance[];
  readonly links: readonly Link[];
  readonly kinds: ReadonlyMap<string, Kind>;
}

function names(value: unknown, key: string): readonly string[] {
  return list(value)
    .map((entry) => text(record(entry)[key]))
    .filter((name) => name !== "");
}

function kind(value: unknown): Kind {
  const entry = record(value);
  const channels = list(entry.channels).map((channel) => ({
    name: text(record(channel).name),
    unit: text(record(channel).unit),
    states: channelStates(record(channel).states),
    range: rangeFrom(record(channel).range),
  }));
  return {
    name: text(entry.name),
    channels,
    states: names(entry.state, "name"),
    units: new Map(list(entry.state).map((state) => [text(record(state).name), text(record(state).unit)] as const)),
    faults: names(entry.faults, "id"),
    effectors: names(entry.effectors, "name"),
    effectorSpecs: parseEffectors(entry.effectors),
  };
}

function instances(value: unknown): readonly Instance[] {
  return list(value).flatMap((raw) => {
    const entry = record(raw);
    const provenance = record(entry.provenance);
    const component = text(provenance.component) || null;
    const instruments = new Map(
      Object.entries(record(provenance.instruments)).map(([point, ref]) => [point, text(ref)] as const),
    );
    const ids = list(entry.names).map(text);
    const single = text(entry.id);
    return (ids.length > 0 ? ids : [single])
      .filter((id) => id !== "")
      .map((id) => ({ id, kind: text(entry.kind), area: text(entry.area), component, instruments }));
  });
}

function areas(site: Readonly<Record<string, unknown>>): readonly AreaRef[] {
  return list(site.areas).map((raw) => {
    const entry = record(raw);
    return { name: text(entry.name), parent: text(entry.parent) || null };
  });
}

function present(value: string | undefined): string | undefined {
  return value === "" ? undefined : value;
}

export function endpointOf(value: unknown): { readonly id: string; readonly port: string } {
  const raw = text(value);
  const dot = raw.lastIndexOf(".");
  return dot < 0 ? { id: raw, port: "" } : { id: raw.slice(0, dot), port: raw.slice(dot + 1) };
}

interface Wire {
  readonly from: ReturnType<typeof endpointOf>;
  readonly to: ReturnType<typeof endpointOf>;
}

function portQueues(composition: unknown): Map<string, Wire[]> {
  const wires = list(record(composition).connections).map((raw) => ({
    from: endpointOf(record(raw).from),
    to: endpointOf(record(raw).to),
  }));
  return groupBy(wires, (wire) => `${wire.from.id}\u0000${wire.to.id}`);
}

function links(value: unknown, composition: unknown): readonly Link[] {
  const queues = portQueues(composition);
  return list(value).map((raw) => {
    const entry = record(raw);
    const from = text(entry.from);
    const to = text(entry.to);
    const ports = queues.get(`${from}\u0000${to}`)?.shift();
    const fromPort = present(text(entry.from_port)) ?? present(ports?.from.port) ?? null;
    const toPort = present(text(entry.to_port)) ?? present(ports?.to.port) ?? null;
    return { from, to, type: text(entry.type), fromPort, toPort };
  });
}

export function parsePlant(domain: unknown, composition: unknown = null): Plant {
  const body = record(domain);
  const site = record(body.site);
  return {
    id: text(body.id),
    title: text(body.title),
    site: text(site.name),
    areas: areas(site),
    instances: instances(body.instances),
    links: links(body.links, composition),
    kinds: new Map(
      list(body.kinds)
        .map(kind)
        .map((entry) => [entry.name, entry] as const),
    ),
  };
}

export function areaOf(plant: Plant, instance: Instance): string {
  return instance.area || plant.site || "Plant";
}

export function instanceIdTaken(plant: Plant, id: string): boolean {
  return plant.instances.some((instance) => instance.id === id);
}

export function linksOf(
  plant: Plant,
  id: string,
): { readonly incoming: readonly Link[]; readonly outgoing: readonly Link[] } {
  return {
    incoming: plant.links.filter((link) => link.to === id),
    outgoing: plant.links.filter((link) => link.from === id),
  };
}

function kindOfEntity(plant: Plant, entity: string): Kind | undefined {
  const kind = plant.instances.find((instance) => instance.id === entity)?.kind;
  return kind === undefined ? undefined : plant.kinds.get(kind);
}

export function stateUnit(plant: Plant, entity: string, state: string): string {
  return kindOfEntity(plant, entity)?.units.get(state) ?? "";
}

export function channelOf(plant: Plant, entity: string, channel: string): Channel | null {
  return kindOfEntity(plant, entity)?.channels.find((entry) => entry.name === channel) ?? null;
}

export function faultsOf(plant: Plant, entity: string): readonly string[] | null {
  return instanceIdTaken(plant, entity) ? (kindOfEntity(plant, entity)?.faults ?? []) : null;
}

export function effectorOf(plant: Plant, entity: string, effector: string): EffectorSpec | undefined {
  return kindOfEntity(plant, entity)?.effectorSpecs.find((spec) => spec.name === effector);
}

export function effectorNamesOf(plant: Plant, entity: string): readonly string[] {
  return kindOfEntity(plant, entity)?.effectorSpecs.map((spec) => spec.name) ?? [];
}

export function hasEffectors(plant: Plant): boolean {
  return [...plant.kinds.values()].some((kind) => kind.effectorSpecs.length > 0);
}
