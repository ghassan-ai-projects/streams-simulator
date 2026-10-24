import { describe, expect, it } from "vitest";
import { assumptionsWire } from "../../../test/provenanceFixture";
import { assumptions } from "./assumptionsDecoder";

const first = assumptionsWire.kinds[0];

function withKind(change: Record<string, unknown>) {
  return { ...assumptionsWire, kinds: [{ ...first, ...change }] };
}

describe("the assumptions of a model", () => {
  it("reads the origin, readings, dynamics, commands, faults and what is not modelled", () => {
    expect(assumptions(assumptionsWire)).toMatchObject({
      origin: { kind: "twin", twin: { id: "arduino-thermal-twin", version: "0.1.0", status: "in_sync" } },
      kinds: [
        {
          instances: ["zone-01 in thermal-bench"],
          readings: [
            {
              channel: "temperature",
              unit: "Cel",
              range: { min: 0, max: 50 },
              noise: { model: "gaussian", sigma: 0.1 },
              cadence: { mode: "periodic", periodS: 10 },
              instrument: "streamsim/thermistor@1",
              recordedIn: "twin_choices",
              suggestion: false,
            },
            { instrument: "", suggestion: true },
          ],
          dynamics: [{ target: "temperature_value", gain: 1, timeConstantS: 180, clamp: { min: 0, max: 50 } }],
          commands: [
            {
              name: "set_led",
              response: "first_order",
              timeConstantS: 180,
              suggestion: true,
              drives: [
                { from: "fan-01", kind: "thermal_sensor", metric: "temperature", gain: -0.02, timeConstantS: 180 },
              ],
            },
          ],
          faults: [{ id: "thermal_sensor__temperature__offset", onsetShape: "step" }],
        },
      ],
      notModeled: [
        {
          code: "no_calibration_evidence",
          scope: "domain",
          sentence: "No calibration or hardware validation is attached.",
        },
        {},
      ],
    });
  });

  it("describes what a model leaves unsaid: no suggestion, range, period, instrument, gain or risk", () => {
    const bare = {
      ...assumptionsWire,
      origin: { kind: "domain" },
      kinds: [
        {
          kind: "k",
          instances: [{ id_template: "M-{n}" }, {}],
          readings: [{ channel: "c", noise: {}, cadence: { mode: "event" } }],
          dynamics: [{ target: "t" }],
          commands: [{ name: "n" }],
          faults: [{ id: "f" }],
        },
      ],
    };
    expect(assumptions(bare)).toMatchObject({
      origin: { kind: "domain", twin: null },
      kinds: [
        {
          instances: ["M-{n}", ""],
          readings: [{ range: null, suggestion: null, cadence: { mode: "event", periodS: null } }],
          dynamics: [{ gain: null, timeConstantS: null, clamp: null }],
          commands: [{ riskClass: "", suggestion: null, drives: [] }],
        },
      ],
    });
  });

  it.each([
    ["not a record", 3],
    ["another schema version", { ...assumptionsWire, schema_version: "9" }],
    ["no domain", { ...assumptionsWire, domain: undefined }],
    ["a domain without its digest", { ...assumptionsWire, domain: { id: "a", version: "1" } }],
    ["an origin it does not know", { ...assumptionsWire, origin: { kind: "odd" } }],
    ["no origin", { ...assumptionsWire, origin: undefined }],
    ["kinds that are not rows", { ...assumptionsWire, kinds: [1] }],
    ["a kind without its name", withKind({ kind: "" })],
    ["a kind without its readings", withKind({ readings: undefined })],
    ["a kind without its dynamics", withKind({ dynamics: undefined })],
    ["a kind without its commands", withKind({ commands: undefined })],
    ["a kind without its faults", withKind({ faults: undefined })],
    ["a kind without its instances", withKind({ instances: undefined })],
    ["a reading without its channel", withKind({ readings: [{}] })],
    ["a dynamics without its target", withKind({ dynamics: [{}] })],
    ["a command without its name", withKind({ commands: [{}] })],
    ["a fault without its id", withKind({ faults: [{}] })],
    ["a limit without its sentence", { ...assumptionsWire, not_modeled: [{ code: "c" }] }],
  ] as readonly (readonly [string, unknown])[])("refuses %s", (_name, wire) => {
    expect(assumptions(wire)).toBeNull();
  });
});
