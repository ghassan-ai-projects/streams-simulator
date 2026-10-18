import { instrumentOf, parseKind, sortInstruments } from "../model/devices";
import type { Instrument, KindFile, Metric } from "../model/devices";
import { FakeBridge, cannedFailure } from "./fakeBridge";
import { failure, success } from "./outcome";
import { sampleWorkspace } from "./workspaceFixture";

export const coolerPath = "kinds/cooler.kind.json";
export const sensorPath = "kinds/sensor.kind.json";
export const choicesPath = "twins/cooler.twin.json";
export const domainPath = "domains/cooler-twin.domain.json";

function argumentsOf(properties: Record<string, unknown>) {
  return { type: "object", properties, required: ["duty"] };
}

export const coolerDocument = {
  kind: "cooler",
  metrics: [
    { name: "temperature", data_type: "number", unit: "Cel", eu_range: { min: 0, max: 100 } },
    { name: "duty", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, writable: true },
    { name: "mode", data_type: "enum", enum_values: ["auto", "manual"] },
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
  metrics: [{ name: "humidity", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 } }],
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

const library = [
  {
    id: "streamsim/thermistor",
    title: "Thermistor",
    family: "temperature",
    unit: "degC",
    range: { min: -40, max: 125 },
  },
  {
    id: "streamsim/humidity",
    title: "Capacitive humidity sensor",
    family: "humidity",
    unit: "%RH",
    range: { min: 0, max: 100 },
  },
  { id: "streamsim/tachometer", title: "Tachometer", family: "speed", unit: "rpm", range: null },
];

function instrumentFrom(entry: (typeof library)[number]): Instrument {
  const { id, title, family, unit, range } = entry;
  return instrumentOf({ ref: `${id}@1`, title }, { family, unit, min: range?.min ?? null, max: range?.max ?? null });
}

function missing(): never {
  throw new Error("missing fixture");
}

export const thermistor = instrumentFrom(library[0] ?? missing());
export const humidity = instrumentFrom(library[1] ?? missing());
export const tachometer = instrumentFrom(library[2] ?? missing());
export const instruments: readonly Instrument[] = sortInstruments([thermistor, humidity, tachometer]);

const libraryWire = {
  components: [
    ...library.map(({ id, title }) => ({ id, version: 1, kind: "instrument", title, class: "x", tags: [] })),
    { id: "streamsim/pump", version: 1, kind: "equipment", title: "Pump", class: "process/pump", tags: [] },
  ],
};

function specOf(ref: unknown) {
  const found = library.find(({ id }) => `${id}@1` === ref);
  if (found === undefined) return failure(cannedFailure("not_found", 404, "no such component"));
  const document = {
    unit_family: found.family,
    unit: found.unit,
    ...(found.range === null ? {} : { range: found.range }),
  };
  return success({ entry: { id: found.id }, document });
}

export function refusal(pointer: string, message: string) {
  return failure({ ...cannedFailure("kind_failed", 422, "the choices are refused"), issues: [{ pointer, message }] });
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
  fake.actions.set("component.list", success(libraryWire));
  fake.handlers.set("component.show", (args) => specOf(args.ref));
  fake.handlers.set("kind.twin", (args) => {
    fake.addFile(String(args.out), { id: "cooler-twin" }, "d1");
    return success({
      id: "cooler-twin",
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
    success({ id: "cooler-twin", digest: "sha256:t", channels: 3, effectors: 1, faults: 2, valid: true }),
  );
  fake.actions.set("domain.lint", success({ domain: "cooler-twin", findings: [], errors: 0, warnings: 0 }));
  return fake;
}
