interface Area {
  readonly name: string;
  readonly parent: string | null;
}

interface Channel {
  readonly name: string;
  readonly unit: string;
}

interface Kind {
  readonly name: string;
  readonly channels: readonly Channel[];
  readonly states: readonly string[];
  readonly faults: readonly string[];
  readonly effectors: readonly string[];
}

export interface Instance {
  readonly id: string;
  readonly kind: string;
  readonly area: string;
  readonly component: string | null;
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
  }));
  return {
    name: text(entry.name),
    channels,
    states: names(entry.state, "name"),
    faults: names(entry.faults, "id"),
    effectors: names(entry.effectors, "name"),
  };
}

function instances(value: unknown): readonly Instance[] {
  return list(value).flatMap((raw) => {
    const entry = record(raw);
    const component = text(record(entry.provenance).component) || null;
    const ids = list(entry.names).map(text);
    const single = text(entry.id);
    return (ids.length > 0 ? ids : [single])
      .filter((id) => id !== "")
      .map((id) => ({ id, kind: text(entry.kind), area: text(entry.area), component }));
  });
}

function areas(site: Json): readonly Area[] {
  return list(site.areas).map((raw) => {
    const entry = record(raw);
    return { name: text(entry.name), parent: text(entry.parent) || null };
  });
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
    return { from, to, type: text(entry.type), fromPort: ports?.from ?? null, toPort: ports?.to ?? null };
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
