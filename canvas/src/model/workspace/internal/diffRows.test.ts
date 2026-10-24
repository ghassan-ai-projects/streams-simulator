import { describe, expect, it } from "vitest";
import { diffModel } from "../../../test/provenanceModels";
import { diffView, valueWords, versionWarning } from "./diffRows";

describe("the model diff as a view", () => {
  const view = diffView(diffModel);

  it("warns that the model changed while its version did not", () => {
    expect(view.warning).toBe("The model changed but its version did not.");
    expect(versionWarning).toBe(view.warning);
  });

  it("carries no warning when the version moved", () => {
    const moved = diffView({ ...diffModel, identity: { ...diffModel.identity, versionKeptDigestChanged: false } });
    expect(moved.warning).toBe("");
  });

  it("counts the changes and names both sides with their source", () => {
    expect(view.headline).toBe("1 change in 1 device kind");
    expect(view.sides[0]).toEqual(["Model A", "factory-utilities 0.1.0 · sha256:aaaaaaaa…aaaaaaaa · recorded in w-1"]);
    expect(view.sides[1]?.[1]).toContain("domains/factory-utilities.domain.json");
  });

  it("groups the changes by device kind with the instances and the link they run over", () => {
    expect(view.groups).toEqual([
      {
        kind: "thermal_sensor",
        status: "changed",
        affects: "Affects zone-01",
        relations: ["fan_drives_thermal_sensor: fan to thermal_sensor (fan-01>zone-01)"],
        changes: [
          {
            key: "0:thermal_sensor/dynamics/temperature_value/inputs/duty@fan_drives_thermal_sensor:gain_changed",
            label: "gain changed",
            subject: "thermal_sensor/dynamics/temperature_value/inputs/duty@fan_drives_thermal_sensor",
            sentence: "thermal_sensor/dynamics/temperature_value: coef changed from -0.02 to -0.05.",
            before: "-0.02",
            after: "-0.05",
          },
        ],
      },
    ]);
  });

  it("lists what did not change", () => {
    expect(view.unchanged).toBe("Unchanged: 2 kinds, 5 readings, 4 commands, 14 faults.");
  });

  it("says when two models are identical", () => {
    const same = diffView({
      ...diffModel,
      identity: { ...diffModel.identity, sameDigest: true },
      kinds: [],
      changes: 0,
      unchanged: [],
    });
    expect(same.headline).toBe("The two models are identical.");
    expect(same.unchanged).toBe("");
  });

  it("describes instances from an id template and a removed kind by its earlier instances", () => {
    const base = diffModel.kinds[0] ?? ({} as never);
    const template = { names: [], count: 12, idTemplates: ["M-{n}"] };
    const view2 = diffView({
      ...diffModel,
      kinds: [
        { ...base, status: "changed", after: template, relations: [] },
        {
          ...base,
          kind: "gone",
          status: "removed",
          before: { names: ["a", "b"], count: 2, idTemplates: [] },
          after: { names: [], count: 0, idTemplates: [] },
        },
        { ...base, kind: "counted", after: { names: [], count: 3, idTemplates: [] } },
        { ...base, kind: "empty", after: { names: [], count: 0, idTemplates: [] } },
      ],
    });
    expect(view2.groups.map((group) => group.affects)).toEqual([
      "Affects 12 instances from M-{n}",
      "Affects a, b",
      "Affects 3 instances",
      "",
    ]);
  });

  it("shows domain-wide changes", () => {
    const change = {
      code: "link_added",
      subject: "links/a",
      section: "links",
      field: "link",
      before: null,
      after: { from: "a" },
      sentence: "link a added.",
    };
    const rows = diffView({ ...diffModel, domainChanges: [change] }).domain;
    expect(rows[0]).toMatchObject({ label: "link added", before: "none", after: '{"from":"a"}' });
  });
});

describe("the words of a changed value", () => {
  it.each([
    [null, "", "none"],
    [undefined, "", "none"],
    ["fast", "", "fast"],
    [12, "", "12"],
    [180, "s", "180 s"],
    [{ min: 0, max: 50 }, "", "0 to 50"],
    [{ min: 0, max: 50 }, "kPa", "0 to 50 kPa"],
    [{ a: 1 }, "", '{"a":1}'],
    [true, "", "true"],
  ])("writes %j with unit %j as %j", (value, unit, words) => {
    expect(valueWords(value, unit)).toBe(words);
  });

  it("writes seconds and milliseconds after the values of fields that are named in them", () => {
    const change = (field: string, code = "time_constant_changed") => ({
      ...(diffModel.kinds[0]?.changes[0] ?? ({} as never)),
      code,
      field,
      before: 10,
      after: 20,
    });
    const rows = (field: string, code?: string) =>
      diffView({ ...diffModel, domainChanges: [change(field, code)] }).domain[0];
    expect(rows("time_constant_s")).toMatchObject({ before: "10 s", after: "20 s" });
    expect(rows("effect.latency_ms")).toMatchObject({ before: "10 ms", after: "20 ms" });
    expect(rows("gain")).toMatchObject({ before: "10", after: "20" });
  });

  it("writes unit codes as their symbols", () => {
    const change = {
      ...(diffModel.kinds[0]?.changes[0] ?? ({} as never)),
      code: "unit_changed",
      field: "unit",
      before: "degC",
      after: "kPa",
    };
    expect(diffView({ ...diffModel, domainChanges: [change] }).domain[0]).toMatchObject({ before: "°C", after: "kPa" });
  });
});
