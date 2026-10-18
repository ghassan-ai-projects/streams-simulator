import { describe, expect, it } from "vitest";
import { captureContent, captureLines, differenceWords, draftIssues } from "./captureDraft";

const pump = {
  commands: [
    { args_schema: { type: "object" }, name: "start", risk_class: "high", safe_stop: false },
    {
      args_schema: {
        type: "object",
        required: ["pct"],
        properties: { pct: { type: "number" }, lease_ms: { type: "integer" } },
      },
      name: "set_speed",
      risk_class: "medium",
      safe_stop: false,
    },
    { args_schema: { type: "object", anyOf: [] }, name: "tune", risk_class: "low", safe_stop: true },
  ],
  kind: "pump",
  metrics: [
    { data_type: "number", eu_range: { max: 120, min: 0 }, name: "flow", unit: "m3/h", writable: false },
    { data_type: "enum", enum_values: ["auto", "hand"], name: "mode" },
    { data_type: "number", name: "level", unit: "m", writable: true, eu_range: { min: 1 } },
  ],
};

describe("the content of a captured kind file", () => {
  it("keeps every member the board announced and invents none", () => {
    expect(JSON.parse(captureContent(pump))).toEqual(pump);
  });

  it("writes it the way the interface editor does: two spaces, a final newline, the editor's order", () => {
    const content = captureContent(pump);
    expect(content.endsWith("}\n")).toBe(true);
    expect(content).toContain('\n  "metrics": [');
    const written = JSON.parse(content) as typeof pump;
    expect(Object.keys(written)).toEqual(["kind", "metrics", "commands"]);
    expect(Object.keys(written.metrics[0] ?? {})).toEqual(["name", "data_type", "unit", "eu_range", "writable"]);
    expect(Object.keys(written.commands[0] ?? {})).toEqual(["name", "risk_class", "safe_stop", "args_schema"]);
  });

  it("puts members the editor does not know after the ones it does, in the order they came", () => {
    const content = captureContent({ title: "Pump", kind: "pump", metrics: [{ extra: 1, name: "flow" }] });
    expect(Object.keys(JSON.parse(content) as object)).toEqual(["kind", "metrics", "title"]);
    expect(content).toContain('"name": "flow",\n      "extra": 1');
  });

  it("leaves a draft without readings or commands without them, and entries it cannot order as they are", () => {
    expect(JSON.parse(captureContent({ kind: "bare" }))).toEqual({ kind: "bare" });
    expect(JSON.parse(captureContent({ kind: "odd", metrics: "none", commands: [1, null] }))).toEqual({
      kind: "odd",
      metrics: "none",
      commands: [1, null],
    });
  });
});

describe("reviewing a captured draft", () => {
  it("says what each reading and command is in the editor's words", () => {
    expect(captureLines(pump)).toEqual({
      readings: [
        { name: "flow", facts: "Number · m3/h · 0 to 120" },
        { name: "mode", facts: "One of a list · options auto, hand" },
        { name: "level", facts: "Number · m · from 1 · setpoint" },
      ],
      commands: [
        { name: "start", facts: "no arguments · High risk" },
        { name: "set_speed", facts: "takes pct, lease_ms · Medium risk" },
        { name: "tune", facts: "custom arguments · Low risk · safe stop" },
      ],
    });
  });

  it("names a reading or command the board left unnamed, and cannot summarise a draft it cannot read", () => {
    expect(captureLines({ kind: "x", metrics: [{}], commands: [{}] })).toEqual({
      readings: [{ name: "Unnamed reading", facts: "" }],
      commands: [{ name: "Unnamed command", facts: "no arguments" }],
    });
    expect(captureLines({ kind: "x", metrics: "none" })).toBeNull();
  });

  it("places an issue on the reading it names, with the words and the binary's message", () => {
    const issue = { pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" };
    expect(draftIssues(pump, [issue])).toEqual([
      { title: "Reading flow · Unit", text: "Unit is required", message: "needs a unit" },
    ]);
  });

  it("still words an issue of a draft that cannot be read", () => {
    const issue = { pointer: "", message: "broken", keyword: "kind_draft" };
    expect(draftIssues({ kind: "x", metrics: "none" }, [issue])).toEqual([
      { title: "", text: "The announced device does not meet the rules for a device file", message: "broken" },
    ]);
  });
});

describe("where the saved file first differs", () => {
  it.each([
    ["/metrics/0/unit", "The first difference is in Reading flow · Unit."],
    ["/commands/1/risk_class", "The first difference is in Command set_speed · Risk class."],
    ["/kind", "The first difference is in Device · Device name."],
    ["/metrics/9/unit", "The first difference is somewhere in the file."],
    ["", "The first difference is somewhere in the file."],
  ])("says %j as %j", (pointer, words) => {
    expect(differenceWords(pump, pointer)).toBe(words);
  });

  it("does not guess at a draft it cannot read", () => {
    expect(differenceWords({ kind: "x", metrics: "none" }, "/metrics/0")).toBe(
      "The first difference is somewhere in the file.",
    );
  });
});
