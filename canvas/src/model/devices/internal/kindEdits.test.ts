import { describe, expect, it } from "vitest";
import {
  addArgument,
  addCommand,
  addMetric,
  changeArgument,
  changeCommand,
  changeMetric,
  hasItem,
  removeArgument,
  removeItem,
  renameKind,
} from "./kindEdits";
import { blankKind, parseKind, serialiseKind } from "./kindFile";
import type { KindFile } from "./types";

function read(source: string): KindFile {
  const kind = parseKind(source);
  if (kind === null) throw new Error("unreadable");
  return kind;
}

describe("editing a device", () => {
  it("adds a blank reading and a blank command, each with a key of its own", () => {
    const first = addMetric(blankKind("fan"));
    const second = addMetric(first.kind);
    expect(first.selected).toEqual({ part: "metric", key: "m1" });
    expect(second.selected).toEqual({ part: "metric", key: "m2" });
    expect(second.kind.metrics[1]).toMatchObject({ name: "", dataType: "number", writable: false });
    const command = addCommand(second.kind);
    expect(command.selected).toEqual({ part: "command", key: "c1" });
    expect(command.kind.commands[0]).toMatchObject({ name: "", riskClass: "low", safeStop: false });
  });

  it("never reuses the key of a removed item", () => {
    const two = addMetric(addMetric(blankKind("fan")).kind).kind;
    const left = removeItem(two, { part: "metric", key: "m1" });
    expect(addMetric(left).selected).toEqual({ part: "metric", key: "m3" });
  });

  it("changes one reading or command and leaves the others", () => {
    const kind = addCommand(addMetric(addMetric(renameKind(blankKind("fan"), "pump")).kind).kind).kind;
    const renamed = changeMetric(kind, "m2", { name: "speed", writable: true });
    expect(renamed.kind).toBe("pump");
    expect(renamed.metrics.map((metric) => metric.name)).toEqual(["", "speed"]);
    expect(changeCommand(kind, "c1", { safeStop: true }).commands[0]?.safeStop).toBe(true);
  });

  it("removes only the item selected, and not the device itself", () => {
    const kind = addCommand(addMetric(blankKind("fan")).kind).kind;
    expect(removeItem(kind, { part: "metric", key: "m1" }).metrics).toEqual([]);
    expect(removeItem(kind, { part: "command", key: "c1" }).commands).toEqual([]);
    expect(removeItem(kind, { part: "kind" })).toBe(kind);
  });

  it("knows whether a selection is still there", () => {
    const kind = addCommand(addMetric(blankKind("fan")).kind).kind;
    expect(hasItem(kind, { part: "kind" })).toBe(true);
    expect(hasItem(kind, { part: "metric", key: "m1" })).toBe(true);
    expect(hasItem(kind, { part: "metric", key: "m9" })).toBe(false);
    expect(hasItem(kind, { part: "command", key: "c1" })).toBe(true);
    expect(hasItem(kind, { part: "command", key: "c9" })).toBe(false);
  });
});

describe("editing the arguments of a command", () => {
  const base = addCommand(blankKind("fan")).kind;

  it("adds, changes and removes an argument row", () => {
    const added = addArgument(addArgument(base, "c1"), "c1");
    const changed = changeArgument(added, "c1", 1, { name: "duty", type: "integer", required: true });
    expect(changed.commands[0]?.args).toMatchObject({
      kind: "rows",
      rows: [{ name: "" }, { name: "duty", type: "integer", required: true }],
    });
    const removed = removeArgument(changed, "c1", 0);
    expect(removed.commands[0]?.args).toMatchObject({ rows: [{ name: "duty" }] });
  });

  it("keeps a schema that forbids other arguments closed through an edit", () => {
    const closed = read(
      '{"kind":"k","commands":[{"name":"x","args_schema":{"type":"object","properties":{"position":{"type":"number"}},"additionalProperties":false}}]}',
    );
    const edited = addArgument(closed, "c1");
    expect(edited.commands[0]?.args).toMatchObject({ kind: "rows", closed: true });
    const written = JSON.parse(serialiseKind(edited)) as { commands: { args_schema: Record<string, unknown> }[] };
    expect(written.commands[0]?.args_schema.additionalProperties).toBe(false);
  });

  it("leaves a command whose arguments it cannot show as it is", () => {
    const custom = read('{"kind":"k","commands":[{"name":"x","args_schema":{"oneOf":[]}}]}');
    expect(addArgument(custom, "c1")).toBe(custom);
    expect(addArgument(base, "c9")).toBe(base);
  });
});
