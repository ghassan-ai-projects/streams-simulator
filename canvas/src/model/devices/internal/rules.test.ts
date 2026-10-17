import { describe, expect, it } from "vitest";
import { argumentsOf, schemaOf } from "./args";
import { issueSelection, issueTitle, issuesOn, resolveIssues } from "./issues";
import { addArgument, addCommand, addMetric, changeArgument, changeMetric } from "./kindEdits";
import { blankKind, parseKind } from "./kindFile";
import { enumOptions, withEnumOptions } from "./options";
import { deviceNameProblem, devicePath, draftProblems, looksLikeId, rangeHint } from "./rules";
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
    ["fan", true],
    ["fan-1.v2:a_b", true],
    ["", false],
    ["-fan", false],
    ["fan one", false],
    ["a".repeat(129), false],
  ])("judges %j as a kind name: %s", (name, fits) => {
    expect(looksLikeId(name)).toBe(fits);
  });

  it.each([
    ["", "Give the device a name."],
    ["fan one", "Use letters, digits and . _ - only, starting with a letter or digit."],
    ["fan:1", "Use letters, digits and . _ - only, starting with a letter or digit."],
    ["Fan", "A device named Fan already exists."],
    ["fan", "A device named fan already exists."],
    ["pump", ""],
  ])("judges %j as a new device name: %j", (name, problem) => {
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
      { pointer: "/metrics/1/data_type", message: "bad type" },
      { pointer: "/commands/0/args_schema/properties", message: "bad args" },
      { pointer: "/metrics/0", message: "whole reading" },
      { pointer: "/kind", message: "bad name" },
      { pointer: "/metrics", message: "too many" },
      { pointer: "/metrics/9/name", message: "gone" },
      { pointer: "/metrics/x/name", message: "odd" },
      { pointer: "", message: "root" },
    ]);
    expect(issues.map((issue) => `${issue.part}:${issue.key}:${issue.field}`)).toEqual([
      "metric:m2:data_type",
      "command:c1:args_schema",
      "metric:m1:",
      "kind::kind",
      "file::",
      "file::",
      "file::",
      "file::",
    ]);
    expect(issuesOn(issues, "metric", "m2", "data_type")).toEqual(["bad type"]);
    expect(issuesOn(issues, "metric", "m1", "")).toEqual(["whole reading"]);
    expect(issuesOn(issues, "metric", "m2", "name")).toEqual([]);
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
    ["an extra keyword", { type: "object", properties: {}, additionalProperties: false }],
    ["a required name that is not a property", { type: "object", properties: {}, required: ["a"] }],
    ["a property with extra keywords", { type: "object", properties: { a: { type: "string", pattern: "x" } } }],
    ["a property of another type", { type: "object", properties: { a: { type: "array" } } }],
    ["a property that is not an object", { type: "object", properties: { a: true } }],
  ])("keeps a schema with %s as it is", (_name, schema) => {
    expect(argumentsOf(schema)).toEqual({ kind: "custom", schema });
    expect(schemaOf({ kind: "custom", schema })).toBe(schema);
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
