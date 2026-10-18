import { describe, expect, it } from "vitest";
import { argumentsOf, schemaOf } from "./args";
import {
  issueSelection,
  issueText,
  issueTitle,
  issuesOn,
  localIssue,
  resolveIssues,
  schemaIssues,
  twinNeeds,
} from "./issues";
import { addArgument, addCommand, addMetric, changeArgument, changeMetric } from "./kindEdits";
import { blankKind, parseKind } from "./kindFile";
import { enumOptions, withEnumOptions } from "./options";
import { deviceNameProblem, devicePath, draftProblems, rangeHint } from "./rules";
import { choicesFor, commandFacts, kindFacts, readingFacts, riskChoices, stemOf, typeChoices } from "./words";
import type { KindFile } from "./types";

function only<T>(value: T | undefined): T {
  if (value === undefined) throw new Error("missing");
  return value;
}

function fan(): KindFile {
  const metric = addMetric(addMetric(blankKind("fan")).kind).kind;
  const named = changeMetric(changeMetric(metric, "m1", { name: "duty" }), "m2", { name: "mode" });
  return addCommand(named).kind;
}

describe("name rules", () => {
  it.each([
    ["", "Give the device a name."],
    ["  ", "Give the device a name."],
    ["Fan", "A device named Fan already exists."],
    ["fan", "A device named fan already exists."],
    ["fan one", ""],
    ["pump", ""],
  ])("judges %j as a new device name: %j, and leaves the pattern to the binary", (name, problem) => {
    expect(deviceNameProblem(name, ["Fan"])).toBe(problem);
  });

  it("names the file after the device", () => {
    expect(devicePath("fan")).toBe("kinds/fan.kind.json");
    expect(stemOf("kinds/fan.kind.json")).toBe("fan");
  });

  it("warns when a range runs backwards", () => {
    expect(rangeHint({ min: 5, max: 1 })).toBe("The minimum is above the maximum.");
    expect(rangeHint({ min: 1, max: 5 })).toBe("");
    expect(rangeHint({ min: null, max: 5 })).toBe("");
  });
});

describe("what the draft already shows is wrong", () => {
  it("lists each blank name and the argument mistakes", () => {
    const empty = addArgument(addArgument(addCommand(addMetric(blankKind(" ")).kind).kind, "c1"), "c1");
    expect(draftProblems(empty).map((problem) => `${problem.part}:${problem.field}`)).toEqual([
      "kind:kind",
      "metric:name",
      "command:name",
      "command:args_schema",
    ]);
    const twice = changeArgument(changeArgument(empty, "c1", 0, { name: "a" }), "c1", 1, { name: "a" });
    expect(draftProblems(twice).map((problem) => problem.message)).toContain("Two arguments share a name.");
  });

  it("is quiet about a device whose names are all given", () => {
    expect(draftProblems(changeMetric(blankKind("fan"), "m1", {}))).toEqual([]);
    const custom = parseKind('{"kind":"k","commands":[{"name":"x","args_schema":{"oneOf":[]}}]}');
    expect(custom === null ? ["unreadable"] : draftProblems(custom)).toEqual([]);
  });
});

describe("the binary's issues placed on the draft", () => {
  const kind = fan();

  it("pins a pointer to the reading or command it names, by key", () => {
    const issues = resolveIssues(kind, [
      { pointer: "/metrics/1/data_type", message: "bad type", keyword: "enum" },
      { pointer: "/commands/0/args_schema/properties", message: "bad args", keyword: "type" },
      { pointer: "/metrics/0", message: "whole reading" },
      { pointer: "/kind", message: "bad name", keyword: "pattern" },
      { pointer: "/metrics", message: "too many", keyword: "maxItems" },
      { pointer: "/metrics/9/name", message: "gone" },
      { pointer: "/metrics/x/name", message: "odd" },
      { pointer: "", message: "root", keyword: "json" },
      { pointer: "/colour", message: "unknown", keyword: "additionalProperties", member: "colour" },
    ]);
    expect(issues.map((issue) => `${issue.part}:${issue.key}:${issue.field}`)).toEqual([
      "metric:m2:data_type",
      "command:c1:args_schema",
      "metric:m1:",
      "kind::kind",
      "kind::metrics",
      "file::",
      "file::",
      "file::",
      "kind::",
    ]);
    expect(issuesOn(issues, "metric", "m2", "data_type")).toEqual(["Choose one of the offered values"]);
    expect(issuesOn(issues, "metric", "m1", "")).toEqual(["The binary does not accept this"]);
    expect(issuesOn(issues, "metric", "m2", "name")).toEqual([]);
  });

  it("places a missing member on the field it names, and says it in plain words", () => {
    const issues = resolveIssues(kind, [
      {
        pointer: "/metrics/0/resolution",
        message: 'needs "resolution"',
        keyword: "twin_member_required",
        member: "resolution",
      },
      { pointer: "/metrics/0", message: "missing name", keyword: "required", member: "name" },
      { pointer: "/metrics/0/unit", message: "metric needs a unit", keyword: "twin_unit_required", member: "unit" },
      { pointer: "/metrics/0/eu_range", message: "min above max", keyword: "twin_range", member: "eu_range" },
      {
        pointer: "/metrics/1/resolution",
        message: 'needs "resolution"',
        keyword: "twin_member_required",
        member: "resolution",
      },
      {
        pointer: "/commands/0/safe_stop",
        message: 'needs "safe_stop"',
        keyword: "twin_member_required",
        member: "safe_stop",
      },
    ]);
    expect(issues.map((issue) => `${issue.part}:${issue.key}:${issue.field}:${issue.text}`)).toEqual([
      "metric:m1:resolution:Needs a resolution",
      "metric:m1:name:Name is required",
      "metric:m1:unit:duty needs a unit",
      "metric:m1:eu_range:The minimum must be below the maximum",
      "metric:m2:resolution:Needs a resolution",
      "command:c1:safe_stop:Needs a safe-stop choice",
    ]);
    expect(issues.map((issue) => issue.twin)).toEqual([true, false, true, true, true, true]);
    expect(schemaIssues(issues)).toHaveLength(1);
    expect(twinNeeds(issues)).toHaveLength(5);
    expect(issues[0]?.message).toBe('needs "resolution"');
    expect(issues[0]?.keyword).toBe("twin_member_required");
  });

  it("words an issue that has no place on its own", () => {
    expect(issueText({ pointer: "/kinds/0/implements", message: "x", keyword: "implements_drift" })).toBe(
      "The twin no longer matches this device",
    );
    expect(issueText({ pointer: "", message: "x", keyword: "required", member: "unit" })).toBe("Unit is required");
    expect(issueText({ pointer: "", message: "x" })).toBe("The binary does not accept this");
    expect(localIssue("kind", "", "kind", "Give the device a name.")).toMatchObject({
      keyword: "",
      twin: false,
      text: "Give the device a name.",
      message: "Give the device a name.",
    });
  });

  it("says which item an issue is about, in words", () => {
    const [first, second, third, fourth] = resolveIssues(kind, [
      { pointer: "/metrics/0/eu_range", message: "a" },
      { pointer: "/commands/0/risk_class", message: "b" },
      { pointer: "/kind", message: "c" },
      { pointer: "/metrics/0", message: "d" },
    ]);
    expect(first && issueTitle(kind, first)).toBe("Reading duty · Range");
    expect(second && issueTitle(kind, second)).toBe("Command · Risk class");
    expect(third && issueTitle(kind, third)).toBe("Device · Device name");
    expect(fourth && issueTitle(kind, fourth)).toBe("Reading duty");
    expect(first && issueSelection(first)).toEqual({ part: "metric", key: "m1" });
    expect(second && issueSelection(second)).toEqual({ part: "command", key: "c1" });
    expect(third && issueSelection(third)).toEqual({ part: "kind" });
  });
});

describe("arguments as rows and as a schema", () => {
  it("reads a simple schema into rows and writes the same schema back", () => {
    const schema = {
      type: "object",
      properties: { duty: { type: "number", minimum: 0, maximum: 100 }, on: { type: "boolean" } },
      required: ["duty"],
    };
    const args = argumentsOf(schema);
    expect(args).toEqual({
      kind: "rows",
      rows: [
        { name: "duty", type: "number", min: 0, max: 100, required: true },
        { name: "on", type: "boolean", min: null, max: null, required: false },
      ],
    });
    expect(schemaOf(args)).toEqual(schema);
  });

  it.each([
    ["a missing type", { properties: {} }],
    ["an open additionalProperties", { type: "object", properties: {}, additionalProperties: true }],
    ["an extra keyword", { type: "object", properties: {}, title: "x" }],
    ["a required name that is not a property", { type: "object", properties: {}, required: ["a"] }],
    ["a property with extra keywords", { type: "object", properties: { a: { type: "string", pattern: "x" } } }],
    ["a property of another type", { type: "object", properties: { a: { type: "array" } } }],
    ["a property that is not an object", { type: "object", properties: { a: true } }],
  ])("keeps a schema with %s as it is", (_name, schema) => {
    expect(argumentsOf(schema)).toEqual({ kind: "custom", schema });
    expect(schemaOf({ kind: "custom", schema })).toBe(schema);
  });

  it("reads a schema that forbids other arguments into rows, and writes the same schema back", () => {
    const schema = {
      type: "object",
      properties: { position: { type: "number", minimum: 0, maximum: 100 } },
      additionalProperties: false,
    };
    const args = argumentsOf(schema);
    expect(args).toEqual({
      kind: "rows",
      rows: [{ name: "position", type: "number", min: 0, max: 100, required: false }],
      closed: true,
    });
    expect(schemaOf(args)).toEqual(schema);
  });

  it("takes an empty schema for no arguments", () => {
    expect(argumentsOf({})).toEqual({ kind: "rows", rows: [] });
    expect(schemaOf({ kind: "rows", rows: [] })).toEqual({ type: "object", properties: {} });
  });
});

describe("choices and words", () => {
  it("offers a placeholder for a missing value and flags a value the binary does not accept", () => {
    expect(choicesFor(riskChoices, "")[0]).toEqual(["", "Choose…"]);
    expect(choicesFor(riskChoices, "low")).toBe(riskChoices);
    expect(choicesFor(riskChoices, "severe").at(-1)).toEqual(["severe", "severe (not accepted)"]);
  });

  it("counts readings and commands, or says the file could not be read", () => {
    expect(kindFacts(1, 2)).toBe("1 reading · 2 commands");
    expect(kindFacts(null, 2)).toBe("Could not be read");
    expect(kindFacts(2, null)).toBe("Could not be read");
  });

  it("describes a reading and a command in a line", () => {
    const kind = fan();
    const reading = only(changeMetric(kind, "m1", { unit: "%" }).metrics[0]);
    expect(readingFacts(reading)).toBe("Number · %");
    expect(readingFacts({ ...reading, dataType: "odd", unit: "" })).toBe("odd");
    const command = only(kind.commands[0]);
    expect(commandFacts({ ...command, riskClass: "low" })).toBe("Low risk");
    expect(commandFacts({ ...command, riskClass: "", safeStop: true })).toBe("safe stop");
    expect(typeChoices.map(([key]) => key)).toContain("enum");
  });

  it("reads and writes the options of a list reading", () => {
    const extra = withEnumOptions({ instrument_range: 1 }, "auto, manual,, ");
    expect(extra).toEqual({ instrument_range: 1, enum_values: ["auto", "manual"] });
    expect(withEnumOptions(extra, " ")).toEqual({ instrument_range: 1 });
    const metric = { ...only(fan().metrics[0]), extra: { enum_values: ["a", 3, "b"] } };
    expect(enumOptions(metric)).toBe("a, b");
  });
});
