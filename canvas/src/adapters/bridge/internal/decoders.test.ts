import { describe, expect, it } from "vitest";
import {
  acknowledgement,
  advanceResult,
  composeSummary,
  composition,
  destroyed,
  faultInjected,
  fileContent,
  paired,
  runDomain,
  runEnded,
  workspace,
  written,
} from "./decoders";

const notRecords: readonly unknown[] = [null, undefined, "text", 7, ["a"]];

describe("simple payloads", () => {
  it.each([
    ["paired", paired, { paired: true }, true],
    ["paired false", paired, { paired: false }, false],
    ["file", fileContent, { path: "a", content: "c", etag: "e", extra: 1 }, { path: "a", content: "c", etag: "e" }],
    ["written", written, { path: "a", etag: "e", size: 3 }, { path: "a", etag: "e" }],
    ["summary", composeSummary, { channels: 3, faults: 2, effectors: 1 }, { channels: 3, faults: 2, effectors: 1 }],
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

  it.each([paired, fileContent, written, composeSummary, acknowledgement, workspace, composition, runDomain])(
    "refuses a payload that is not an object",
    (decode) => {
      for (const value of notRecords) expect(decode(value)).toBeNull();
    },
  );
});

describe("acknowledgement", () => {
  it.each([
    [
      { accepted: true, reason: "ok", effect_eta_ns: 5e9 },
      { accepted: true, reason: "ok", effectEtaNs: 5e9 },
    ],
    [
      { accepted: false, reason: "busy" },
      { accepted: false, reason: "busy", effectEtaNs: null },
    ],
    [{ accepted: true }, { accepted: true, reason: "", effectEtaNs: null }],
    [
      { accepted: true, effect_eta_ns: 0 },
      { accepted: true, reason: "", effectEtaNs: 0 },
    ],
  ])("reads %j", (wire, expected) => {
    expect(acknowledgement(wire)).toEqual(expected);
  });

  it.each([
    [{ reason: "no flag" }],
    [{ accepted: "true" }],
    [{ accepted: true, reason: 3 }],
    [{ accepted: true, effect_eta_ns: "soon" }],
  ])("refuses %j", (wire) => {
    expect(acknowledgement(wire)).toBeNull();
  });
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
      runs: [
        {
          worldId: "w-1",
          ended: true,
          complete: true,
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

describe("confirmations", () => {
  it.each([
    [advanceResult, { clock: "c", emitted: 3 }, { clock: "c" }],
    [advanceResult, { clock: "c", emitted: 0 }, { clock: "c", emitted: Infinity }],
    [faultInjected, { fault_id: "f1", simulated: true }, { fault_id: "f1", simulated: false }],
    [faultInjected, { fault_id: "f1", simulated: true }, { simulated: true }],
    [runEnded, { run_artifact_path: "runs/w/run.json" }, { run_artifact_path: "" }],
    [destroyed, { destroyed: true }, { destroyed: "true" }],
  ])("accepts the confirmation and refuses its malformed twin", (decode, good, bad) => {
    expect(decode(good)).toBe(true);
    expect(decode(bad)).toBeNull();
    expect(decode(null)).toBeNull();
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
