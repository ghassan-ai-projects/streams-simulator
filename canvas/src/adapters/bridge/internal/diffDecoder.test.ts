import { describe, expect, it } from "vitest";
import { diffWire } from "../../../test/provenanceFixture";
import { modelDiff } from "./diffDecoder";

const first = diffWire.kinds[0];

function withKind(change: Record<string, unknown>) {
  return { ...diffWire, kinds: [{ ...first, ...change }] };
}

describe("the diff of two models", () => {
  it("reads both sides, the identity, the changes per kind with instances and relations, and the unchanged counts", () => {
    expect(modelDiff(diffWire)).toMatchObject({
      a: { id: "factory-utilities", source: { kind: "run", label: "w-1" } },
      identity: { sameId: true, sameVersion: true, sameDigest: false, versionKeptDigestChanged: true },
      kinds: [
        {
          kind: "thermal_sensor",
          status: "changed",
          after: { count: 1, names: ["zone-01"], idTemplates: [] },
          relations: [
            { type: "fan_drives_thermal_sensor", fromKind: "fan", toKind: "thermal_sensor", links: ["fan-01>zone-01"] },
          ],
          changes: [{ code: "gain_changed", field: "coef", before: -0.02, after: -0.05 }],
        },
      ],
      changes: 1,
      unchanged: [
        ["kinds", 2],
        ["readings", 5],
        ["dynamics", 0],
        ["commands", 4],
        ["faults", 14],
      ],
    });
  });

  it("reads a change that carries no value, a kind without relations and no domain-wide changes", () => {
    const bare = {
      ...diffWire,
      kinds: [{ kind: "k", status: "added", changes: [{ code: "kind_added", subject: "k", sentence: "s" }] }],
      domain_changes: undefined,
      unchanged_summary: { kinds: "many" },
    };
    expect(modelDiff(bare)).toMatchObject({
      kinds: [
        { relations: [], before: { count: 0, names: [] }, changes: [{ before: null, after: null, section: "" }] },
      ],
      domainChanges: [],
      unchanged: [],
    });
  });

  it.each([
    ["not a record", null],
    ["another schema version", { ...diffWire, schema_version: "2" }],
    ["a side without its source", { ...diffWire, a: { id: "i", version: "1", digest: "d" } }],
    ["a side without its digest", { ...diffWire, a: { ...diffWire.a, digest: "" } }],
    ["a source it does not know", { ...diffWire, a: { ...diffWire.a, source: { kind: "web" } } }],
    ["an identity without its flags", { ...diffWire, identity: { same_id: true } }],
    ["kinds that are not rows", { ...diffWire, kinds: [4] }],
    ["a kind with a status it does not know", withKind({ status: "odd" })],
    ["a kind without a name", withKind({ kind: "" })],
    ["a kind without its changes", withKind({ changes: undefined })],
    ["a kind with relations that are not rows", withKind({ relations: [1] })],
    ["a relation without its type", withKind({ relations: [{}] })],
    ["a change without its sentence", withKind({ changes: [{ code: "c", subject: "s" }] })],
    ["domain changes that are not rows", { ...diffWire, domain_changes: [1] }],
  ] as readonly (readonly [string, unknown])[])("refuses %s", (_name, wire) => {
    expect(modelDiff(wire)).toBeNull();
  });
});
