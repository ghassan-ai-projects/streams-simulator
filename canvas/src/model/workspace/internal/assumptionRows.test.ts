import { describe, expect, it } from "vitest";
import { assumptionsModel } from "../../../test/provenanceModels";
import { kindAssumptions, originWords } from "./assumptionRows";

const kind = kindAssumptions(assumptionsModel.kinds[0] ?? ({} as never));

describe("the assumptions of a device kind as rows", () => {
  it("counts what the kind holds and where its instances are", () => {
    expect(kind.summary).toBe("zone-01 in thermal-bench · 2 readings · 1 command · 1 fault");
  });

  it("states what a reading is: instrument, range and resolution with units, noise, cadence and where it was recorded", () => {
    expect(kind.readings[0]).toEqual({
      key: "temperature",
      title: "temperature",
      statement: "",
      facts: [
        { label: "Instrument", value: "streamsim/thermistor@1" },
        { label: "Range", value: "0 to 50 °C" },
        { label: "Resolution", value: "1 °C" },
        { label: "Noise", value: "gaussian, sigma 0.1 °C" },
        { label: "Cadence", value: "periodic, every 10 s" },
        { label: "Recorded in", value: "the twin choices" },
      ],
      suggestion: "Differs from the binary's suggestion",
    });
  });

  it("says when a reading is a plain channel and when it equals the binary's suggestion", () => {
    expect(kind.readings[1]?.facts[0]).toEqual({ label: "Instrument", value: "plain channel, no library instrument" });
    expect(kind.readings[1]?.suggestion).toBe("Suggested by the binary");
  });

  it("says nothing about a suggestion the binary did not compare", () => {
    const bare = kindAssumptions({
      ...(assumptionsModel.kinds[0] ?? ({} as never)),
      readings: [
        {
          ...(assumptionsModel.kinds[0]?.readings[0] ?? ({} as never)),
          suggestion: null,
          range: null,
          unit: "",
          instrument: "",
          basis: "domain_declared",
          cadence: { mode: "on_change", periodS: null },
        },
      ],
    });
    expect(bare.readings[0]).toMatchObject({ suggestion: "" });
    expect(bare.readings[0]?.facts).toContainEqual({ label: "Instrument", value: "declared in the domain file" });
    expect(bare.readings[0]?.facts).toContainEqual({ label: "Range", value: "not stated" });
    expect(bare.readings[0]?.facts).toContainEqual({ label: "Cadence", value: "on change" });
  });

  it("states what a command does, its gain and time constant, risk and safe stop, in the binary's own statement", () => {
    expect(kind.commands[0]).toEqual({
      key: "set_led",
      title: "set_led",
      statement: "Moves temperature toward the commanded value.",
      facts: [
        { label: "Response", value: "moves the readings it drives through a first-order lag" },
        { label: "Drives fan-01", value: "thermal_sensor.temperature: gain -0.02, time constant 180 s" },
        { label: "Dead time / time constant", value: "0 s / 180 s" },
        { label: "Risk", value: "low" },
        { label: "Safe stop", value: "no" },
        { label: "Recorded in", value: "the twin choices" },
      ],
      suggestion: "Suggested by the binary",
    });
  });

  it("keeps a response and a recording place it does not know, and says when no risk is stated", () => {
    const base = assumptionsModel.kinds[0]?.commands[0] ?? ({} as never);
    const odd = kindAssumptions({
      ...(assumptionsModel.kinds[0] ?? ({} as never)),
      commands: [{ ...base, response: "new_kind", recordedIn: "elsewhere", riskClass: "", safeStop: true, drives: [] }],
    });
    expect(odd.commands[0]?.facts).toEqual([
      { label: "Response", value: "new_kind" },
      { label: "Dead time / time constant", value: "0 s / 180 s" },
      { label: "Risk", value: "not stated" },
      { label: "Safe stop", value: "yes" },
      { label: "Recorded in", value: "elsewhere" },
    ]);
  });

  it("states the dynamics and the faults", () => {
    expect(kind.dynamics[0]?.facts).toEqual([
      { label: "Form", value: "first order lag" },
      { label: "Gain", value: "1" },
      { label: "Time constant", value: "180 s" },
      { label: "Clamp", value: "0 to 50" },
    ]);
    expect(kind.faults[0]).toMatchObject({
      title: "thermal_sensor__temperature__offset",
      facts: [
        { label: "Onset", value: "step" },
        { label: "Affects", value: "temperature__bias" },
      ],
    });
  });

  it("states a dynamics and a fault that say little", () => {
    const base = assumptionsModel.kinds[0] ?? ({} as never);
    const small = kindAssumptions({
      ...base,
      dynamics: [{ target: "t", form: "", gain: null, timeConstantS: null, clamp: null }],
      faults: [{ id: "f", description: "d", onsetShape: "step", affects: [] }],
    });
    expect(small.dynamics[0]?.facts).toEqual([{ label: "Form", value: "not stated" }]);
    expect(small.faults[0]).toMatchObject({ statement: "d", facts: [{ label: "Onset", value: "step" }] });
  });
});

describe("where the assumptions were read from", () => {
  it.each([
    [{ kind: "composition", twin: null }, "Read from the plant composed from a composition."],
    [{ kind: "domain", twin: null }, "Read from the domain file alone."],
    [
      { kind: "twin", twin: { id: "t", version: "1", status: "in_sync" } },
      "Read from a plant generated from twin choices that are in sync with it.",
    ],
    [
      { kind: "twin", twin: { id: "t", version: "1", status: "out_of_sync" } },
      "The twin choices no longer produce this plant: read from the domain file alone.",
    ],
    [{ kind: "twin", twin: null }, "The twin choices no longer produce this plant: read from the domain file alone."],
  ] as const)("says %j", (origin, words) => {
    expect(originWords(origin)).toBe(words);
  });
});
