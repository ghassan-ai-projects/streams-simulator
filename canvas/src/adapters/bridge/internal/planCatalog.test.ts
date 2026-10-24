import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { faultCategories, perturbationCatalog } from "./planCatalog";

const wireKinds = {
  perturbations: [
    {
      kind: "drop",
      category: "delivery",
      description: "Loses records in transport.",
      requires_end: false,
      params: [{ name: "rate", type: "number", minimum: 0, maximum: 1 }],
    },
    {
      kind: "producer_flap",
      category: "delivery",
      description: "Silences the producer.",
      requires_end: true,
      params: [
        { name: "mode", type: "enum", values: ["hold", "lose"] },
        { name: "fields", type: "string_list" },
        { name: "burst", type: "integer", minimum: 1, maximum: 9, future: true },
      ],
    },
  ],
};

const kinds = [
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
    description: "Silences the producer.",
    requiresEnd: true,
    params: [
      { name: "mode", type: "enum", values: ["hold", "lose"] },
      { name: "fields", type: "string_list" },
      { name: "burst", type: "integer", minimum: 1, maximum: 9 },
    ],
  },
];

function lastCall(t: ReturnType<typeof bridgeTransport>): { url: unknown; body: unknown } {
  const [url, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return { url, body: typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined };
}

describe("the delivery conditions the simulator offers", () => {
  it("asks the catalog and reads each kind with the contract of its parameters", async () => {
    const t = bridgeTransport(() => respond(200, { result: wireKinds }));
    expect(await bridgeAdapter(t).launch.perturbationCatalog()).toEqual({ ok: true, value: kinds });
    expect(lastCall(t)).toEqual({ url: "/api/actions/catalog.perturbations", body: { args: {} } });
  });

  it("reads a catalog that is empty, and a kind whose requires_end the binary left out", () => {
    expect(perturbationCatalog({ perturbations: [] })).toEqual([]);
    const bare = { perturbations: [{ kind: "drop", description: "d" }] };
    expect(perturbationCatalog(bare)).toEqual([
      { kind: "drop", category: "", description: "d", requiresEnd: false, params: [] },
    ]);
  });

  it.each([
    ["not an object", null],
    ["no list", {}],
    ["an entry that is not an object", { perturbations: [3] }],
    ["a kind without a name", { perturbations: [{ description: "d" }] }],
    ["a text requires_end", { perturbations: [{ kind: "drop", requires_end: "yes" }] }],
    ["params that are not a list", { perturbations: [{ kind: "drop", params: {} }] }],
    ["a param that is not an object", { perturbations: [{ kind: "drop", params: [3] }] }],
    ["a param without a name", { perturbations: [{ kind: "drop", params: [{ type: "number" }] }] }],
    ["a param of an unknown type", { perturbations: [{ kind: "drop", params: [{ name: "a", type: "blob" }] }] }],
    ["a text minimum", { perturbations: [{ kind: "drop", params: [{ name: "a", type: "number", minimum: "0" }] }] }],
    [
      "enum values that are not text",
      { perturbations: [{ kind: "drop", params: [{ name: "a", type: "enum", values: [1] }] }] },
    ],
  ])("refuses a catalog with %s", (_name, bad) => {
    expect(perturbationCatalog(bad)).toBeNull();
  });

  it("says that the catalog could not be read", async () => {
    const down = bridgeTransport(() => respond(500, { error: { code: "boom", message: "down" } }));
    expect(await bridgeAdapter(down).launch.perturbationCatalog()).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    const refused = bridgeTransport(() => respond(404, { error: { code: "nope", message: "no" } }));
    expect(await bridgeAdapter(refused).launch.perturbationCatalog()).toMatchObject({
      ok: false,
      failure: { kind: "rejected", step: "catalog" },
    });
  });
});

describe("which layer each fault of a plant acts on", () => {
  const described = {
    spec: { id: "factory" },
    digest: "sha256:x",
    fault_categories: {
      tower: { fan_failure: "world", sensor_drift: "measurement", stuck: "world_and_measurement", odd: "other" },
      chiller: {},
    },
  };

  it("asks the catalog about the plant and reads the category of each fault by kind", async () => {
    const t = bridgeTransport(() => respond(200, { result: described }));
    expect(await bridgeAdapter(t).launch.faultCategories("factory")).toEqual({
      ok: true,
      value: {
        tower: { fan_failure: "world", sensor_drift: "measurement", stuck: "world_and_measurement" },
        chiller: {},
      },
    });
    expect(lastCall(t)).toEqual({ url: "/api/actions/catalog.describe", body: { args: { domain: "factory" } } });
  });

  it("reads a description that names no categories as none", () => {
    expect(faultCategories({ spec: {} })).toEqual({});
    expect(faultCategories({ fault_categories: { tower: 3 } })).toEqual({ tower: {} });
    expect(faultCategories(null)).toBeNull();
  });

  it("says that the description could not be read", async () => {
    const refused = bridgeTransport(() => respond(404, { error: { code: "domain_invalid", message: "no such" } }));
    expect(await bridgeAdapter(refused).launch.faultCategories("x")).toMatchObject({
      ok: false,
      failure: { kind: "rejected", step: "catalog", diagnostic: { message: "no such" } },
    });
  });
});
