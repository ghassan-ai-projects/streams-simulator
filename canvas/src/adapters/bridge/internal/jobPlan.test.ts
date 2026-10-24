import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { jobStatus } from "./jobs";

const wire = {
  world_id: "w-9",
  run_id: "r-9",
  domain: "factory",
  seed: 3,
  state: "running",
  start_ns: 100e9,
  reached_ns: 160e9,
  requested_end_ns: 1900e9,
  rate: 60,
  explore: true,
  calls: [],
};

function answering(result: unknown) {
  return bridgeTransport(() => respond(200, { result }));
}

function lastBody(t: ReturnType<typeof bridgeTransport>): unknown {
  const [, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined;
}

describe("starting a job with a plan", () => {
  it("starts a named job with a fault that recovers and a delivery condition", async () => {
    const t = answering(wire);
    await bridgeAdapter(t).launch.startJob({
      name: "Fan failure",
      purpose: "Observe recovery",
      domain: "factory",
      seed: 3,
      explore: true,
      durationS: 1800,
      rate: 0,
      faults: [{ entity: "CT-1", fault: "fan", atS: 10, untilS: 90 }],
      perturbations: [
        { kind: "drop", params: { rate: 0.5, fields: ["a", "b"] }, atS: 30, untilS: 120 },
        { kind: "reorder", params: {}, atS: 60 },
      ],
      calls: [],
    });
    expect(lastBody(t)).toEqual({
      args: {
        name: "Fan failure",
        purpose: "Observe recovery",
        domain: "factory",
        seed: 3,
        explore: true,
        duration_s: 1800,
        rate: 0,
        faults: [{ entity_id: "CT-1", fault: "fan", at_s: 10, until_s: 90 }],
        perturbations: [
          { kind: "drop", params: { rate: 0.5, fields: ["a", "b"] }, at_s: 30, until_s: 120 },
          { kind: "reorder", at_s: 60 },
        ],
        calls: [],
      },
    });
  });
});

describe("a job's plan in its status", () => {
  const planned = {
    ...wire,
    name: "Fan failure",
    purpose: "Observe recovery",
    faults: [
      { entity_id: "CT-1", fault: "fan", at_s: 10, until_s: 90, fault_id: "f-0" },
      { entity_id: "CT-2", fault: "fan", at_s: 20 },
    ],
    perturbations: [
      { kind: "drop", params: { rate: 0.5, mode: "hold", fields: ["a"] }, at_s: 30, until_s: 120 },
      { kind: "reorder", at_s: 60 },
    ],
  };

  it("reads the name, the purpose, the faults with their clear times and the delivery conditions", () => {
    expect(jobStatus(planned)).toMatchObject({
      name: "Fan failure",
      purpose: "Observe recovery",
      faults: [
        { entity: "CT-1", fault: "fan", atS: 10, untilS: 90 },
        { entity: "CT-2", fault: "fan", atS: 20 },
      ],
      perturbations: [
        { kind: "drop", params: { rate: 0.5, mode: "hold", fields: ["a"] }, atS: 30, untilS: 120 },
        { kind: "reorder", params: {}, atS: 60 },
      ],
    });
    expect(jobStatus(planned)?.faults[1]).not.toHaveProperty("untilS");
  });

  it("reads a status from a binary that predates the plan fields as an unnamed job with no plan", () => {
    expect(jobStatus(wire)).toMatchObject({ name: "", purpose: "", faults: [], perturbations: [] });
  });

  it("ignores fields it does not know", () => {
    const extra = { ...planned, faults: [{ ...planned.faults[1], severity: 3 }], future: true };
    expect(jobStatus(extra)?.faults).toEqual([{ entity: "CT-2", fault: "fan", atS: 20 }]);
  });

  it.each([
    ["a numeric name", { name: 4 }],
    ["a numeric purpose", { purpose: 4 }],
    ["faults that are not a list", { faults: {} }],
    ["a fault that is not an object", { faults: [3] }],
    ["a fault without an instance", { faults: [{ fault: "f", at_s: 1 }] }],
    ["a fault without a name", { faults: [{ entity_id: "e", at_s: 1 }] }],
    ["a fault with a text time", { faults: [{ entity_id: "e", fault: "f", at_s: "1" }] }],
    ["a fault with a text end", { faults: [{ entity_id: "e", fault: "f", at_s: 1, until_s: "2" }] }],
    ["perturbations that are not a list", { perturbations: "drop" }],
    ["a perturbation that is not an object", { perturbations: [3] }],
    ["a perturbation without a kind", { perturbations: [{ at_s: 1 }] }],
    ["a perturbation with a text time", { perturbations: [{ kind: "drop", at_s: "1" }] }],
    ["a perturbation with a text end", { perturbations: [{ kind: "drop", at_s: 1, until_s: "2" }] }],
    ["perturbation params that are not an object", { perturbations: [{ kind: "drop", at_s: 1, params: 3 }] }],
    ["a perturbation param that is a boolean", { perturbations: [{ kind: "drop", at_s: 1, params: { a: true } }] }],
    ["a perturbation list param with a number", { perturbations: [{ kind: "drop", at_s: 1, params: { a: [1] } }] }],
  ])("refuses a job with %s", (_name, change) => {
    expect(jobStatus({ ...wire, ...change })).toBeNull();
  });
});
