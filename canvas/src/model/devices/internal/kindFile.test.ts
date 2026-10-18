import { describe, expect, it } from "vitest";
import { blankKind, parseKind, serialiseKind } from "./kindFile";

const fan = {
  kind: "fan",
  title: "Cooling fan",
  metrics: [
    {
      name: "duty",
      data_type: "number",
      unit: "%",
      eu_range: { min: 0, max: 100 },
      resolution: 0.5,
      scan_class: "fast",
      writable: true,
      description: "Fan duty",
      instrument_range: { min: 0, max: 120 },
    },
    { name: "mode", data_type: "enum", enum_values: ["auto", "manual"] },
  ],
  commands: [
    {
      name: "set_duty",
      risk_class: "medium",
      safe_stop: false,
      description: "Set the duty",
      args_schema: { type: "object", properties: { duty: { type: "number", minimum: 0, maximum: 100 } } },
    },
    { name: "stop", risk_class: "low", safe_stop: true, args_schema: { type: "object", properties: {} } },
  ],
};

describe("parsing a kind file", () => {
  it("reads every field the form edits and keeps the rest aside", () => {
    const kind = parseKind(JSON.stringify(fan));
    expect(kind?.kind).toBe("fan");
    expect(kind?.extra).toEqual({ title: "Cooling fan" });
    expect(kind?.metrics[0]).toEqual({
      key: "m1",
      name: "duty",
      dataType: "number",
      unit: "%",
      range: { min: 0, max: 100 },
      resolution: 0.5,
      scanClass: "fast",
      writable: true,
      description: "Fan duty",
      extra: { instrument_range: { min: 0, max: 120 } },
    });
    expect(kind?.metrics[1]).toMatchObject({
      key: "m2",
      range: { min: null, max: null },
      extra: { enum_values: ["auto", "manual"] },
    });
    expect(kind?.commands[0]?.args).toEqual({
      kind: "rows",
      rows: [{ name: "duty", type: "number", min: 0, max: 100, required: false }],
    });
    expect(kind?.commands[1]).toMatchObject({ key: "c2", safeStop: true, riskClass: "low" });
  });

  it("treats missing lists as empty and a missing field as blank", () => {
    expect(parseKind("{}")).toEqual({ kind: "", metrics: [], commands: [], extra: {} });
    const sparse = parseKind('{"kind":"k","metrics":[{}],"commands":[{}]}');
    expect(sparse?.metrics[0]).toMatchObject({ name: "", dataType: "", unit: "", writable: false });
    expect(sparse?.commands[0]).toMatchObject({
      name: "",
      riskClass: "",
      safeStop: false,
      args: { kind: "rows", rows: [] },
    });
  });

  it.each([
    ["not JSON", "{"],
    ["a list", "[]"],
    ["metrics that are not a list", '{"metrics":{}}'],
    ["commands that are not a list", '{"commands":"x"}'],
    ["a reading that is not an object", '{"metrics":[1]}'],
    ["a command that is not an object", '{"commands":[null]}'],
  ])("refuses %s", (_name, source) => {
    expect(parseKind(source)).toBeNull();
  });
});

describe("writing a kind file", () => {
  it("reads back to the same file and leaves out what is empty", () => {
    const kind = parseKind(JSON.stringify(fan));
    if (kind === null) throw new Error("unreadable");
    const written = JSON.parse(serialiseKind(kind)) as typeof fan;
    expect(written.metrics[0]).toEqual(fan.metrics[0]);
    expect(written.metrics[1]).toEqual({ ...fan.metrics[1], writable: false });
    expect(written.commands[1]).toEqual(fan.commands[1]);
    expect(written.title).toBe("Cooling fan");
    expect(serialiseKind(kind).endsWith("}\n")).toBe(true);
  });

  it("writes what a blank reading and command carry, so the binary can judge the rest", () => {
    const kind = parseKind('{"kind":"k","metrics":[{"eu_range":{"min":1}}],"commands":[{}]}');
    if (kind === null) throw new Error("unreadable");
    expect(JSON.parse(serialiseKind(kind))).toEqual({
      kind: "k",
      metrics: [{ name: "", eu_range: { min: 1 }, writable: false }],
      commands: [{ name: "", safe_stop: false, args_schema: { type: "object", properties: {} } }],
    });
  });

  it("always states whether a command is a safe stop", () => {
    const kind = parseKind('{"kind":"k","metrics":[],"commands":[{"name":"a"},{"name":"b","safe_stop":true}]}');
    if (kind === null) throw new Error("unreadable");
    const written = JSON.parse(serialiseKind(kind)) as {
      readonly commands: readonly { readonly safe_stop: boolean }[];
    };
    expect(written.commands.map((command) => command.safe_stop)).toEqual([false, true]);
  });

  it("always states whether a reading is writable, because a twin needs the statement", () => {
    const kind = parseKind('{"kind":"k","metrics":[{"name":"a"},{"name":"b","writable":true}],"commands":[]}');
    if (kind === null) throw new Error("unreadable");
    const written = JSON.parse(serialiseKind(kind)) as { readonly metrics: readonly { readonly writable: boolean }[] };
    expect(written.metrics.map((metric) => metric.writable)).toEqual([false, true]);
  });

  it("starts a device with no readings and no commands", () => {
    expect(JSON.parse(serialiseKind(blankKind("fan")))).toEqual({ kind: "fan", metrics: [], commands: [] });
  });
});
