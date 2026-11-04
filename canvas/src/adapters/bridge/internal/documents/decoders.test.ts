import { describe, expect, it } from "vitest";
import { composeSummary, composition, fileContent, paired, runDomain, workspace, written } from "./decoders";

const notRecords: readonly unknown[] = [null, undefined, "text", 7, ["a"]];

describe("simple payloads", () => {
  it.each([
    ["paired", paired, { paired: true }, true],
    ["paired false", paired, { paired: false }, false],
    ["file", fileContent, { path: "a", content: "c", etag: "e", extra: 1 }, { path: "a", content: "c", etag: "e" }],
    ["written", written, { path: "a", etag: "e", size: 3 }, { path: "a", etag: "e" }],
    [
      "summary",
      composeSummary,
      { channels: 3, faults: 2, effectors: 1, digest: "sha256:ab" },
      { channels: 3, faults: 2, effectors: 1, digest: "sha256:ab" },
    ],
    [
      "summary without a digest",
      composeSummary,
      { channels: 3, faults: 2, effectors: 1 },
      { channels: 3, faults: 2, effectors: 1, digest: "" },
    ],
  ])("decodes %s", (_name, decode, wire, expected) => {
    expect(decode(wire)).toEqual(expected);
  });

  it.each([
    ["paired", paired, { paired: "yes" }],
    ["file without etag", fileContent, { path: "a", content: "c" }],
    ["file with numeric content", fileContent, { path: "a", content: 1, etag: "e" }],
    ["written without path", written, { etag: "e" }],
    ["written with numeric etag", written, { path: "a", etag: 4 }],
    ["summary missing effectors", composeSummary, { channels: 3, faults: 2 }],
    ["summary with infinite channels", composeSummary, { channels: Infinity, faults: 2, effectors: 1 }],
    ["summary with a string", composeSummary, { channels: "3", faults: 2, effectors: 1 }],
  ])("refuses %s", (_name, decode, wire) => {
    expect(decode(wire)).toBeNull();
  });

  it.each([paired, fileContent, written, composeSummary, workspace, composition, runDomain])(
    "refuses a payload that is not an object",
    (decode) => {
      for (const value of notRecords) expect(decode(value)).toBeNull();
    },
  );
});

describe("workspace payload", () => {
  const wire = {
    name: "lab",
    domains: [{ path: "d.json", id: "mini", title: "Mini", instances: 4 }],
    compositions: [{ path: "c.json", id: "comp", title: "Comp" }],
    scenarios: [{ path: "s.json" }],
    runs: [
      {
        world_id: "w-1",
        ended: true,
        complete: true,
        incomplete_reason: "",
        explore: false,
        seed: 42,
        plant: "mini",
        duration_ns: 90.4e9,
        recorded_at: "2026-10-17T10:00:00Z",
        faults: 2,
      },
    ],
  };

  it("reads each listing with its fields", () => {
    expect(workspace(wire)).toEqual({
      name: "lab",
      domains: [{ path: "d.json", id: "mini", title: "Mini", instances: 4 }],
      compositions: [{ path: "c.json", id: "comp", title: "Comp", instances: 0 }],
      scenarios: [{ path: "s.json", id: "", title: "", instances: 0 }],
      kinds: [],
      components: [],
      runs: [
        {
          worldId: "w-1",
          ended: true,
          complete: true,
          running: false,
          incompleteReason: "",
          explore: false,
          seed: 42,
          plant: "mini",
          durationS: 90,
          recordedAt: "2026-10-17T10:00:00Z",
          faults: 2,
        },
      ],
    });
  });

  it("reads the length a run was asked for and how it ended, and leaves them out when absent or not numbers and text", () => {
    const stopped = { ...wire.runs[0], requested_duration_ns: 1800e9, termination_reason: "stopped_by_operator" };
    const decoded = workspace({
      ...wire,
      runs: [stopped, { ...stopped, world_id: "w-2", requested_duration_ns: "x", termination_reason: 3 }],
    });
    expect(decoded?.runs[0]).toMatchObject({ requestedS: 1800, terminationReason: "stopped_by_operator" });
    expect(decoded?.runs[1]).not.toHaveProperty("requestedS");
    expect(decoded?.runs[1]).not.toHaveProperty("terminationReason");
    expect(workspace(wire)?.runs[0]).not.toHaveProperty("requestedS");
  });

  it("reads the name of a run, and leaves it out when the run has none or the name is not text", () => {
    const named = { ...wire.runs[0], name: "Fan failure" };
    const decoded = workspace({
      ...wire,
      runs: [named, { ...named, world_id: "w-2", name: "" }, { ...named, world_id: "w-3", name: 4 }],
    });
    expect(decoded?.runs[0]).toMatchObject({ name: "Fan failure" });
    expect(decoded?.runs[1]).not.toHaveProperty("name");
    expect(decoded?.runs[2]).not.toHaveProperty("name");
    expect(workspace(wire)?.runs[0]).not.toHaveProperty("name");
  });

  it("reads the kind files and components the workspace lists", () => {
    const decoded = workspace({
      ...wire,
      kinds: [{ path: "kinds/fan.kind.json", kind: "fan", title: "Fan" }, { path: "kinds/bad.kind.json" }],
      components: [{ path: "components/equipment/pump.component.json", id: "pump", title: "Pump", class: "equipment" }],
    });
    expect(decoded?.kinds).toEqual([
      { path: "kinds/fan.kind.json", kind: "fan", title: "Fan" },
      { path: "kinds/bad.kind.json", kind: "", title: "" },
    ]);
    expect(decoded?.components).toEqual([
      { path: "components/equipment/pump.component.json", id: "pump", title: "Pump", class: "equipment" },
    ]);
  });

  it("reads a run the bridge started as a connected session, running or saved, and no other run as one", () => {
    const decoded = workspace({
      ...wire,
      runs: [
        {
          world_id: "c-1",
          connected: true,
          running: true,
          plan: {
            instances: [
              { instance: "zone-01", origin: "real" },
              { instance: "fan-01", origin: "simulated" },
            ],
          },
        },
        { world_id: "c-2", connected: true },
        { world_id: "w-3", connected: "yes", running: true },
      ],
    });
    expect(decoded?.runs.map((run) => run.connected)).toEqual(["running", "saved", undefined]);
    expect(decoded?.runs[2]).not.toHaveProperty("connected");
    expect(decoded?.runs.map((run) => run.simulated)).toEqual([["fan-01"], [], undefined]);
  });

  it("reads whether a run is running for every run, not only for connected sessions", () => {
    const decoded = workspace({
      ...wire,
      runs: [
        { world_id: "w-1", running: true },
        { world_id: "w-2", running: false },
        { world_id: "w-3" },
        { world_id: "w-4", running: "yes" },
      ],
    });
    expect(decoded?.runs.map((run) => run.running)).toEqual([true, false, false, false]);
    expect(decoded?.runs.map((run) => run.connected)).toEqual([undefined, undefined, undefined, undefined]);
  });

  it("defaults a run's missing or mistyped fields", () => {
    const decoded = workspace({
      ...wire,
      runs: [{ world_id: "w-2", seed: "7", duration_ns: "long", faults: null }],
    });
    expect(decoded?.runs).toEqual([
      {
        worldId: "w-2",
        ended: false,
        complete: false,
        running: false,
        incompleteReason: "",
        explore: false,
        seed: null,
        plant: "",
        durationS: 0,
        recordedAt: "",
        faults: 0,
      },
    ]);
  });

  it.each([
    ["no name", { ...wire, name: undefined }],
    ["numeric name", { ...wire, name: 3 }],
    ["domains missing", { ...wire, domains: undefined }],
    ["compositions not a list", { ...wire, compositions: {} }],
    ["scenarios null", { ...wire, scenarios: null }],
    ["runs a string", { ...wire, runs: "none" }],
    ["a run that is not an object", { ...wire, runs: [5] }],
    ["a run without a world id", { ...wire, runs: [{ ended: true }] }],
    ["a run with an empty world id", { ...wire, runs: [{ world_id: "" }] }],
    ["a run with a numeric world id", { ...wire, runs: [{ world_id: 3 }] }],
    ["a domain without a path", { ...wire, domains: [{ id: "mini" }] }],
    ["a composition with an empty path", { ...wire, compositions: [{ path: "", id: "comp" }] }],
    ["a scenario that is not an object", { ...wire, scenarios: ["s.json"] }],
    ["kinds a string", { ...wire, kinds: "none" }],
    ["a kind without a path", { ...wire, kinds: [{ kind: "fan" }] }],
    ["components null", { ...wire, components: null }],
    ["a component without a path", { ...wire, components: [{ id: "pump" }] }],
  ])("refuses a workspace with %s", (_name, bad) => {
    expect(workspace(bad)).toBeNull();
  });
});

describe("run domain", () => {
  it("extracts the domain the run used", () => {
    expect(runDomain({ content: JSON.stringify({ domain_spec: { id: "mini" }, other: 1 }) })).toEqual({ id: "mini" });
  });

  it.each([
    ["content missing", {}],
    ["content not a string", { content: 4 }],
    ["content not JSON", { content: "{oops" }],
    ["artifact an array", { content: "[]" }],
    ["no domain_spec", { content: "{}" }],
    ["domain_spec a list", { content: JSON.stringify({ domain_spec: [] }) }],
  ])("refuses a run whose %s", (_name, wire) => {
    expect(runDomain(wire)).toBeNull();
  });
});

describe("composition payload", () => {
  const document = {
    schema_version: "0.1",
    id: "plant",
    version: "1",
    title: "Plant",
    stresses: "load",
    axes: {},
    profiles: [],
    ground_truth: {},
  };

  it("returns the composition document untouched", () => {
    expect(composition({ composition: { ...document, equipment: [] } })).toEqual({ ...document, equipment: [] });
    expect(composition({ composition: document })).toEqual(document);
  });

  it.each([
    ["no composition", {}],
    ["wrong schema version", { composition: { ...document, schema_version: "0.2" } }],
    ["blank title", { composition: { ...document, title: "" } }],
    ["missing stresses", { composition: { ...document, stresses: undefined } }],
    ["axes a list", { composition: { ...document, axes: [] } }],
    ["profiles an object", { composition: { ...document, profiles: {} } }],
    ["no ground truth", { composition: { ...document, ground_truth: undefined } }],
    ["equipment an object", { composition: { ...document, equipment: {} } }],
  ])("refuses a composition with %s", (_name, wire) => {
    expect(composition(wire)).toBeNull();
  });
});
