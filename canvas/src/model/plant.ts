interface Area {
  readonly name: string;
  readonly parent: string | null;
}

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
  readonly areas: readonly Area[];
  readonly instances: readonly Instance[];
  readonly links: readonly Link[];
  readonly kinds: ReadonlyMap<string, Kind>;
}

type Json = Record<string, unknown>;

function record(value: unknown): Json {
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : {};
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function list(value: unknown): readonly unknown[] {
  return Array.isArray(value) ? value : [];
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

function areas(site: Json): readonly Area[] {
  return list(site.areas).map((raw) => {
    const entry = record(raw);
    return { name: text(entry.name), parent: text(entry.parent) || null };
  });
}

function present(value: string | undefined): string | undefined {
  return value === "" ? undefined : value;
}

function endpoint(value: unknown): { readonly id: string; readonly port: string } {
  const raw = text(value);
  const dot = raw.lastIndexOf(".");
  return dot < 0 ? { id: raw, port: "" } : { id: raw.slice(0, dot), port: raw.slice(dot + 1) };
}

function portQueues(composition: unknown): Map<string, { from: string; to: string }[]> {
  const queues = new Map<string, { from: string; to: string }[]>();
  for (const raw of list(record(composition).connections)) {
    const from = endpoint(record(raw).from);
    const to = endpoint(record(raw).to);
    const key = `${from.id}\u0000${to.id}`;
    queues.set(key, [...(queues.get(key) ?? []), { from: from.port, to: to.port }]);
  }
  return queues;
}

function links(value: unknown, composition: unknown): readonly Link[] {
  const queues = portQueues(composition);
  return list(value).map((raw) => {
    const entry = record(raw);
    const from = text(entry.from);
    const to = text(entry.to);
    const ports = queues.get(`${from}\u0000${to}`)?.shift();
    const fromPort = present(text(entry.from_port)) ?? present(ports?.from) ?? null;
    const toPort = present(text(entry.to_port)) ?? present(ports?.to) ?? null;
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

export function linksOf(
  plant: Plant,
  id: string,
): { readonly incoming: readonly Link[]; readonly outgoing: readonly Link[] } {
  return {
    incoming: plant.links.filter((link) => link.to === id),
    outgoing: plant.links.filter((link) => link.from === id),
  };
}

const listedLinks = 12;

function linkEnd(id: string, port: string | null): string {
  return port === null ? id : `${id}.${port}`;
}

function counted(n: number, noun: string): string {
  return `${String(n)} ${noun}${n === 1 ? "" : "s"}`;
}

export function linkLines(links: readonly Link[], self: string, side: "in" | "out"): readonly string[] {
  const own = (link: Link) => (side === "in" ? link.toPort : link.fromPort) ?? self;
  const other = (link: Link) => (side === "in" ? linkEnd(link.from, link.fromPort) : linkEnd(link.to, link.toPort));
  if (links.length <= listedLinks)
    return links.map((link) => (side === "in" ? `${other(link)} → ${own(link)}` : `${own(link)} → ${other(link)}`));
  const groups = new Map<string, Link[]>();
  for (const link of links) groups.set(own(link), [...(groups.get(own(link)) ?? []), link]);
  return [...groups].map(([port, group]) => {
    const ends = new Set(group.map((link) => (side === "in" ? link.from : link.to))).size;
    const count = `${counted(group.length, "link")}, ${counted(ends, "instance")}`;
    return side === "in" ? `${count} → ${port}` : `${port} → ${count}`;
  });
}

export function stateUnit(plant: Plant, entity: string, state: string): string {
  const kind = plant.instances.find((instance) => instance.id === entity)?.kind;
  return (kind === undefined ? undefined : plant.kinds.get(kind)?.units.get(state)) ?? "";
}

export function channelOf(plant: Plant, entity: string, channel: string): Channel | null {
  const kindName = plant.instances.find((instance) => instance.id === entity)?.kind ?? "";
  return plant.kinds.get(kindName)?.channels.find((entry) => entry.name === channel) ?? null;
}
