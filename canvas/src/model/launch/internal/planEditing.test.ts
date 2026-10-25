import { describe, expect, it } from "vitest";
import { parsePlant } from "../../plant";
import { draftOf, firstDraft, planOf } from "./draft";
import type { Draft } from "./draft";
import type { LaunchPlan } from "./launch";
import type { PerturbationKind } from "./perturbation";
import { planProblems } from "./planProblems";

const plant = parsePlant({ id: "mini", instances: [{ kind: "chiller", names: ["CH-1"] }], kinds: [] });
const anySpec = (_entity: string, effector: string) => ({ name: effector, entityArg: null, args: [] });

const kinds: readonly PerturbationKind[] = [
  {
    kind: "drop",
    category: "delivery",
    description: "Loses records in transport.",
    requiresEnd: false,
    params: [{ name: "rate", type: "number", minimum: 0, maximum: 1 }],
  },
  {
    kind: "producer_flap",
    category: "delivery",
    description: "Silences the producer during its window.",
    requiresEnd: true,
    params: [
      { name: "burst", type: "integer", minimum: 1, maximum: 9 },
      { name: "mode", type: "enum", values: ["hold", "lose"] },
      { name: "fields", type: "string_list" },
    ],
  },
];

const plan = (overrides: Partial<LaunchPlan>): LaunchPlan => ({
  seed: 7,
  durationS: 1800,
  explore: true,
  faults: [],
  perturbations: [],
  calls: [],
  ...overrides,
});

const draft = (overrides: Partial<Draft>): Draft => ({ ...firstDraft, ...overrides });

describe("name and purpose in a plan", () => {
  it("carries them trimmed, and leaves an unset one out", () => {
    const set = planOf(draft({ name: "  Fan failure  ", purpose: " How does cooling respond? " }), plant);
    expect(set).toMatchObject({ name: "Fan failure", purpose: "How does cooling respond?" });
    const unset = planOf(draft({ name: " ", purpose: "" }), plant);
    expect("name" in unset).toBe(false);
    expect("purpose" in unset).toBe(false);
  });

  it("round-trips through the draft", () => {
    const given = plan({ name: "Fan failure", purpose: "Observe recovery" });
    expect(draftOf(given)).toMatchObject({ name: "Fan failure", purpose: "Observe recovery" });
    expect(planOf(draftOf(given), plant)).toEqual({ ...given, pace: "watch" });
  });
});

describe("faults that recover", () => {
  it("turns an end minute into whole seconds and leaves a blank one out", () => {
    const faults = [
      { key: 1, entity: "CH-1", fault: "fouling", atMinutes: "2", untilMinutes: "5.5" },
      { key: 2, entity: "CH-1", fault: "fouling", atMinutes: "3", untilMinutes: "" },
    ];
    const planned = planOf(draft({ faults }), plant).faults;
    expect(planned).toEqual([
      { entity: "CH-1", fault: "fouling", atS: 120, untilS: 330 },
      { entity: "CH-1", fault: "fouling", atS: 180 },
    ]);
    expect(draftOf(plan({ faults: planned })).faults.map((fault) => fault.untilMinutes)).toEqual(["5.5", ""]);
  });

  it("names an end that does not follow its start, runs past the run, or is not a number", () => {
    const texts = (untilS: number, atS = 120) =>
      planProblems(plan({ faults: [{ entity: "CH-1", fault: "f", atS, untilS }] }), anySpec).map((p) => p.text);
    expect(texts(300)).toEqual([]);
    expect(texts(120)).toEqual(["Fault 1 ends after it starts."]);
    expect(texts(60)).toEqual(["Fault 1 ends after it starts."]);
    expect(texts(1801)).toEqual(["Fault 1 ends by the end of the run."]);
    expect(texts(1800)).toEqual([]);
    expect(texts(Number.NaN)).toEqual(["Fault 1 ends at a number of minutes."]);
    expect(planProblems(plan({ faults: [{ entity: "CH-1", fault: "f", atS: 120, untilS: 60 }] }), anySpec)).toEqual([
      { field: "faultEnd", index: 0, text: "Fault 1 ends after it starts." },
    ]);
  });
});

describe("delivery conditions in a plan", () => {
  const drop = { kind: "drop", params: { rate: 0.5 }, atS: 60 };

  it("types each parameter from the catalog and omits a blank one", () => {
    const perturbations = [
      {
        key: 1,
        kind: "producer_flap",
        atMinutes: "1",
        untilMinutes: "4",
        values: { burst: "3", mode: "hold", fields: "a, b ,,c", rate: "" },
      },
    ];
    expect(planOf(draft({ perturbations }), plant, kinds).perturbations).toEqual([
      { kind: "producer_flap", params: { burst: 3, mode: "hold", fields: ["a", "b", "c"] }, atS: 60, untilS: 240 },
    ]);
  });

  it("reads a parameter as a number when the catalog is not there to type it", () => {
    const perturbations = [
      { key: 1, kind: "drop", atMinutes: "1", untilMinutes: "", values: { rate: "0.5", mode: "x" } },
    ];
    expect(planOf(draft({ perturbations }), plant).perturbations[0]?.params).toEqual({ rate: 0.5, mode: "x" });
  });

  it("round-trips through the draft, parameters as text", () => {
    const given = plan({
      perturbations: [{ kind: "producer_flap", params: { fields: ["a", "b"], burst: 2 }, atS: 90, untilS: 150 }],
    });
    const back = draftOf(given).perturbations[0];
    expect(back).toEqual({
      key: 1,
      kind: "producer_flap",
      atMinutes: "1.5",
      untilMinutes: "2.5",
      values: { fields: "a, b", burst: "2" },
    });
    expect(planOf(draftOf(given), plant, kinds).perturbations).toEqual(given.perturbations);
  });

  it("is clean for a complete condition and names what is wrong otherwise", () => {
    const texts = (...perturbations: LaunchPlan["perturbations"][number][]) =>
      planProblems(plan({ perturbations }), anySpec, kinds).map((problem) => problem.text);
    expect(texts(drop)).toEqual([]);
    expect(texts({ ...drop, kind: "" })).toEqual(["Delivery condition 1 needs a kind."]);
    expect(texts({ ...drop, kind: "teleport" })).toEqual([
      "Delivery condition 1: the simulator offers no delivery condition named teleport.",
    ]);
    expect(texts({ ...drop, atS: 1800 })).toEqual(["Delivery condition 1 starts inside the run, before its end."]);
    expect(texts({ ...drop, untilS: 60 })).toEqual(["Delivery condition 1 ends after it starts."]);
    expect(texts({ ...drop, untilS: 1900 })).toEqual(["Delivery condition 1 ends by the end of the run."]);
    expect(texts({ ...drop, untilS: Number.NaN })).toEqual(["Delivery condition 1 ends at a number of minutes."]);
    expect(texts({ kind: "producer_flap", params: {}, atS: 60 })).toEqual([
      "Delivery condition 1 needs an end: Silences the producer during its window.",
    ]);
  });

  it("holds each parameter to the range the catalog declares", () => {
    const texts = (params: LaunchPlan["perturbations"][number]["params"], kind = "drop") =>
      planProblems(plan({ perturbations: [{ kind, params, atS: 60, untilS: 120 }] }), anySpec, kinds).map(
        (p) => p.text,
      );
    expect(texts({ rate: 1.5 })).toEqual(["Delivery condition 1: rate is a number from 0 to 1."]);
    expect(texts({ rate: Number.NaN })).toEqual(["Delivery condition 1: rate is a number from 0 to 1."]);
    expect(texts({ burst: 2.5 }, "producer_flap")).toEqual([
      "Delivery condition 1: burst is a whole number from 1 to 9.",
    ]);
    expect(texts({ mode: "other" }, "producer_flap")).toEqual(["Delivery condition 1: mode is one of hold, lose."]);
    expect(texts({ burst: 2, mode: "hold", fields: ["a"] }, "producer_flap")).toEqual([]);
  });

  it("flags the part of the condition that is wrong", () => {
    const found = planProblems(plan({ perturbations: [{ kind: "drop", params: {}, atS: 1800 }] }), anySpec, kinds);
    expect(found).toEqual([
      {
        field: "delivery",
        index: 0,
        part: "start",
        text: "Delivery condition 1 starts inside the run, before its end.",
      },
    ]);
  });
});
