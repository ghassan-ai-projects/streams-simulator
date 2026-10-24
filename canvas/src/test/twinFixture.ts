import { choicesOfDocument, instrumentOf, keyChoices, parseKind, reconcileChoices } from "../model/devices";
import { list, record, text } from "../model/shared/json";
import type { Instrument, KindFile, Member, Metric, TwinChoices } from "../model/devices";
import { FakeBridge, cannedFailure } from "./fakeBridge";
import { kindCheck } from "./kindCheckFake";
import { failure, success } from "./outcome";
import { sampleWorkspace } from "./workspaceFixture";

export const coolerPath = "kinds/cooler.kind.json";
export const sensorPath = "kinds/sensor.kind.json";
export const choicesPath = "twins/cooler.twin.json";
export const domainPath = "domains/cooler-sensor-twin.domain.json";

function argumentsOf(properties: Record<string, unknown>) {
  return { type: "object", properties, required: ["duty"] };
}

export const coolerDocument = {
  kind: "cooler",
  metrics: [
    { name: "temperature", data_type: "number", unit: "Cel", eu_range: { min: 0, max: 100 }, resolution: 0.1 },
    { name: "duty", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, resolution: 1, writable: true },
    { name: "mode", data_type: "enum", enum_values: ["auto", "manual"], resolution: 1 },
  ],
  commands: [
    {
      name: "set_duty",
      risk_class: "medium",
      args_schema: argumentsOf({ duty: { type: "number", minimum: 0, maximum: 100 }, lease_ms: { type: "integer" } }),
    },
    { name: "stop", risk_class: "low", safe_stop: true, args_schema: { type: "object", properties: {} } },
  ],
};

export const sensorDocument = {
  kind: "sensor",
  metrics: [{ name: "humidity", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, resolution: 1 }],
  commands: [],
};

export function kindOf(document: unknown): KindFile {
  const kind = parseKind(JSON.stringify(document));
  if (kind === null) throw new Error("not a kind file");
  return kind;
}

export function metricOf(extra: Record<string, unknown> = {}): Metric {
  const kind = kindOf({
    kind: "k",
    metrics: [{ name: "m", data_type: "number", unit: "Cel", eu_range: { min: 0, max: 50 }, ...extra }],
  });
  const [metric] = kind.metrics;
  if (metric === undefined) throw new Error("no metric");
  return metric;
}

export const cooler: Member = { path: coolerPath, kind: kindOf(coolerDocument) };
export const sensor: Member = { path: sensorPath, kind: kindOf(sensorDocument) };

export const suggestion = {
  schema_version: "0.1",
  id: "cooler-sensor-twin",
  version: "0.1.0",
  title: "Twin of cooler, sensor",
  kinds: [
    {
      kind: "cooler",
      instance: "cooler-01",
      metrics: { temperature: { instrument: "none" }, duty: { instrument: "none" }, mode: { instrument: "none" } },
      commands: { set_duty: { response: "direct", sets: { duty: "duty" } } },
    },
    { kind: "sensor", instance: "sensor-01", metrics: { humidity: { instrument: "streamsim/humidity@1" } } },
  ],
};

export function startingChoices(members: readonly Member[] = [cooler]): TwinChoices {
  const suggested = keyChoices(choicesOfDocument(suggestion), members);
  return reconcileChoices(suggested, members, suggested);
}

interface Entry {
  readonly id: string;
  readonly title: string;
  readonly unit: string;
  readonly range: { readonly min: number; readonly max: number } | null;
}

const library: readonly Entry[] = [
  { id: "streamsim/thermistor", title: "Thermistor", unit: "degC", range: { min: -40, max: 125 } },
  { id: "streamsim/humidity", title: "Capacitive humidity sensor", unit: "%RH", range: { min: 0, max: 100 } },
  { id: "streamsim/tachometer", title: "Tachometer", unit: "rpm", range: null },
];

function instrumentFrom(entry: Entry, match: Instrument["match"] = "", fit: Instrument["fit"] = null): Instrument {
  return instrumentOf({
    ref: `${entry.id}@1`,
    title: entry.title,
    unit: entry.unit,
    min: entry.range?.min ?? null,
    max: entry.range?.max ?? null,
    match,
    fit,
  });
}

function missing(): never {
  throw new Error("missing fixture");
}

export const thermistor = instrumentFrom(library[0] ?? missing());
export const humidity = instrumentFrom(library[1] ?? missing());
const tachometer = instrumentFrom(library[2] ?? missing());
export const instruments: readonly Instrument[] = [thermistor, humidity, tachometer];
export const convertibleThermistor = instrumentFrom(library[0] ?? missing(), "convertible");

const matches: Readonly<Record<string, Readonly<Record<string, "exact" | "convertible">>>> = {
  Cel: { "streamsim/thermistor": "convertible" },
  degC: { "streamsim/thermistor": "exact" },
  "%": { "streamsim/humidity": "convertible" },
  rpm: { "streamsim/tachometer": "exact" },
};

function fitFor(entry: Entry, args: Readonly<Record<string, unknown>>) {
  const { min, max } = args;
  if (entry.range === null) return { ok: true };
  if (typeof min !== "number" || typeof max !== "number") {
    return { ok: false, code: "range_unstated", member: "eu_range" };
  }
  const covers = entry.range.min <= min && entry.range.max >= max;
  return covers ? { ok: true } : { ok: false, code: "range_narrower", member: "eu_range" };
}

function wire(entry: Entry, match?: string, args?: Readonly<Record<string, unknown>>) {
  const { id, title, unit, range } = entry;
  const base = { id, version: 1, kind: "instrument", title, class: "x", tags: [], unit };
  const fit = args === undefined ? {} : { fit: fitFor(entry, args) };
  return { ...base, ...(range === null ? {} : { range }), ...(match === undefined ? {} : { match }), ...fit };
}

function listed(args: Readonly<Record<string, unknown>>) {
  const equipment = {
    id: "streamsim/pump",
    version: 1,
    kind: "equipment",
    title: "Pump",
    class: "process/pump",
    tags: [],
  };
  const unit = typeof args.unit === "string" ? args.unit : "";
  if (unit === "") return success({ components: [...library.map((entry) => wire(entry)), equipment] });
  const found = matches[unit] ?? {};
  const entries = library.flatMap((entry) => {
    const match = found[entry.id];
    return match === undefined ? [] : [wire(entry, match, args)];
  });
  return success({ components: entries });
}

interface RawIssue {
  readonly pointer?: string;
  readonly message: string;
  readonly keyword?: string;
  readonly member?: string;
}

export function refusal(...issues: readonly RawIssue[]) {
  return failure({ ...cannedFailure("kind_failed", 422, "the choices are refused"), issues });
}

export function withPlant(fake: FakeBridge): FakeBridge {
  const listing = fake.workspaceOutcome;
  if (!listing.ok) throw new Error("no listing");
  const domains = [{ path: domainPath, id: "cooler-sensor-twin", title: "Cooler", instances: 1 }];
  fake.workspaceOutcome = success({ ...listing.value, domains });
  fake.addFile(domainPath, { id: "cooler-sensor-twin" }, "d5");
  return fake;
}

function lacksTwinMembers(content: string): boolean {
  try {
    const doc = record(JSON.parse(content) as unknown);
    return list(doc.metrics)
      .map(record)
      .some((metric) => {
        const numeric = metric.data_type === "number" || metric.data_type === "integer";
        return (
          metric.resolution === undefined || metric.scan_class !== undefined || (numeric && text(metric.unit) === "")
        );
      });
  } catch {
    return false;
  }
}

function suggested(fake: FakeBridge, args: Readonly<Record<string, unknown>>) {
  const stems = list(args.kinds).map(text);
  const files = stems.map((stem) => fake.files.get(`kinds/${stem}.kind.json`));
  if (files.some((file) => !file?.ok)) return failure(cannedFailure("not_found", 404, "no kind"));
  const lacking = files.some((file) => file !== undefined && file.ok && lacksTwinMembers(file.value.content));
  const issue = {
    pointer: "/kinds/0/kind",
    message: 'needs "resolution"',
    keyword: "twin_member_required",
    member: "resolution",
  };
  return lacking ? refusal(issue) : success({ choices: suggestion });
}

export function twinBridge() {
  const fake = new FakeBridge();
  fake.sessionOutcome = success(true);
  fake.persistWrites();
  fake.addFile(coolerPath, coolerDocument, "k1");
  fake.addFile(sensorPath, sensorDocument, "k2");
  const kinds = [
    { path: coolerPath, kind: "cooler", title: "" },
    { path: sensorPath, kind: "sensor", title: "" },
  ];
  fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
  fake.actions.set(
    "kind.validate",
    success({ valid: true, kinds: [{ kind: "cooler", kind_digest: "sha256:abc", metrics: 3, commands: 2 }] }),
  );
  fake.handlers.set("kind.check", kindCheck);
  fake.handlers.set("kind.suggest", (args) => suggested(fake, args));
  fake.handlers.set("component.list", listed);
  fake.handlers.set("kind.twin", (args) => {
    fake.addFile(String(args.out), { id: "cooler-sensor-twin" }, "d1");
    return success({
      id: "cooler-sensor-twin",
      digest: "sha256:t",
      channels: 3,
      faults: 2,
      effectors: 1,
      implements: ["cooler"],
      composed: args.out,
    });
  });
  fake.actions.set(
    "domain.validate",
    success({ id: "cooler-sensor-twin", digest: "sha256:t", channels: 3, effectors: 1, faults: 2, valid: true }),
  );
  fake.actions.set("domain.lint", success({ domain: "cooler-sensor-twin", findings: [], errors: 0, warnings: 0 }));
  return fake;
}
