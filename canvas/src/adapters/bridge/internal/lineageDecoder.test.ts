import { describe, expect, it } from "vitest";
import { lineageWire } from "../../../test/provenanceFixture";
import { lineage } from "./lineageDecoder";

function withPlant(change: Record<string, unknown>) {
  return { ...lineageWire, plants: [{ ...lineageWire.plants[0], ...change }] };
}

function withDevice(change: Record<string, unknown>) {
  return {
    ...lineageWire,
    device_kinds: [{ kind: "x", file: "f", has_own_twin: false, twins_including: [], ...change }],
  };
}

const editable = (change: Record<string, unknown>, route: Record<string, unknown> = {}) => ({
  editable: {
    verdict: "editable",
    reason: "composition",
    route: { code: "recompose", sentence: "s", ...route },
    ...change,
  },
});

describe("the lineage of the plants", () => {
  it("reads the plants with their sources, runs and route, ignoring members it does not know", () => {
    expect(lineage(lineageWire)).toMatchObject({
      plants: [
        {
          id: "factory-utilities",
          composition: { file: "compositions/factory-utilities.composition.json", status: "in_sync", reason: "" },
          twin: null,
          runs: [
            { id: "w-1", mode: "evaluate", recorded: true, domainVersion: "0.1.0", sameAsCurrent: true },
            { id: "w-2", mode: "explore", sameAsCurrent: false, domainDigest: `sha256:${"b".repeat(64)}` },
          ],
          editable: { verdict: "editable", reason: "composition", route: { code: "edit_composition" } },
        },
        {
          id: "arduino-thermal-twin",
          editable: { verdict: "read_only", reason: "twin" },
          kinds: [{ kind: "fan", found: true }, { kind: "thermal_sensor" }],
        },
        { id: "chiller-pilot", editable: { reason: "example" } },
        { id: "alarm-playback", editable: { reason: "hand_authored" } },
        { id: "campus-3", composition: { status: "out_of_sync", reason: "version differs" } },
      ],
    });
  });

  it("reads the device kinds and the files that could not be read", () => {
    expect(lineage(lineageWire)).toMatchObject({
      devices: [
        {
          kind: "fan",
          hasOwnTwin: false,
          twinsIncluding: [{ twin: "arduino-thermal-twin", file: "twins/fan.twin.json", kinds: 2 }],
        },
        { kind: "thermal_sensor" },
        { kind: "heater", hasOwnTwin: true },
        { kind: "lonely", twinsIncluding: [], implementedBy: [] },
      ],
      unreadable: [{ file: "kinds/broken.kind.json", code: "kind_file_unreadable", message: "not JSON" }],
    });
  });

  it("marks a device file that is not found", () => {
    const plant = {
      ...withPlant({}).plants[0],
      device_kinds: [{ kind: "gone", status: "missing" }, { status: "found" }],
    };
    expect(lineage({ ...lineageWire, plants: [plant] })).toBeNull();
    const missing = { ...plant, device_kinds: [{ kind: "gone", status: "missing" }] };
    expect(lineage({ ...lineageWire, plants: [missing] })?.plants[0]?.kinds).toEqual([
      { kind: "gone", found: false, file: "" },
    ]);
  });

  it.each([
    ["not a record", null],
    ["another schema version", { ...lineageWire, schema_version: "0.2" }],
    ["no plants", { ...lineageWire, plants: undefined }],
    ["plants that are not rows", { ...lineageWire, plants: [3] }],
    ["a plant without its domain", withPlant({ domain: { id: "x" } })],
    ["a plant whose domain is not a record", withPlant({ domain: 4 })],
    [
      "a composition with a status it does not know",
      withPlant({ composition: { file: "f", id: "i", version: "1", status: "odd" } }),
    ],
    ["a composition that is not a record", withPlant({ composition: 3 })],
    ["a twin without its file", withPlant({ twin: { id: "i", version: "1", status: "in_sync" } })],
    ["device kinds that are not rows", withPlant({ device_kinds: [null] })],
    ["a device kind without a name", withPlant({ device_kinds: [{ status: "found" }] })],
    ["a run without its version", withPlant({ runs: [{ id: "w" }] })],
    [
      "a run without the same-as-current flag",
      withPlant({ runs: [{ id: "w", domain_version: "1", domain_digest: "d" }] }),
    ],
    ["no editable verdict", withPlant({ editable: undefined })],
    ["a verdict it does not know", withPlant(editable({ verdict: "maybe" }))],
    ["a reason it does not know", withPlant(editable({ reason: "odd" }))],
    ["a route it does not know", withPlant(editable({}, { code: "odd" }))],
    ["device kinds without a file", withDevice({ file: undefined })],
    ["a device kind without its twin flag", withDevice({ has_own_twin: undefined })],
    ["twins that include a kind without a file", withDevice({ twins_including: [{ twin: "t" }] })],
    ["an unreadable file without a code", { ...lineageWire, unreadable: [{ file: "f" }] }],
  ] as readonly (readonly [string, unknown])[])("refuses %s", (_name, wire) => {
    expect(lineage(wire)).toBeNull();
  });
});
