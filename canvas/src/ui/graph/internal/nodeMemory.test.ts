import { describe, expect, it } from "vitest";
import type { FlowNode } from "./flow";
import { NodeMemory, reshaped, shapesOf } from "./nodeMemory";

function frame(id: string, count: number): FlowNode {
  return { id, type: "frame", position: { x: 0, y: 0 }, data: { title: id, count } };
}

function instance(id: string, rows: number, hidden: number, value = "1"): FlowNode {
  const readout = { rows: Array.from({ length: rows }, (_, at) => ({ key: String(at), value })), hidden };
  const view = { id, pins: { ins: [], outs: [] } };
  return {
    id,
    type: "instance",
    position: { x: 0, y: 0 },
    width: 200,
    height: 80,
    data: { view, live: { readout } },
  } as unknown as FlowNode;
}

describe("nodes kept between renders", () => {
  it("hands back the very same node when nothing in it changed", () => {
    const memory = new NodeMemory();
    const first = memory.stable([instance("a", 2, 0), instance("b", 2, 0)]);
    const second = memory.stable([instance("a", 2, 0), instance("b", 2, 0)]);
    expect(second[0]).toBe(first[0]);
    expect(second[1]).toBe(first[1]);
  });

  it("replaces only the node whose content changed", () => {
    const memory = new NodeMemory();
    const first = memory.stable([instance("a", 2, 0), instance("b", 2, 0)]);
    const second = memory.stable([instance("a", 2, 0), instance("b", 2, 0, "9")]);
    expect(second[0]).toBe(first[0]);
    expect(second[1]).not.toBe(first[1]);
  });
});

describe("nodes whose shape changed", () => {
  it("does not call a node reshaped because its values changed", () => {
    const before = shapesOf([instance("a", 2, 0, "1"), frame("f", 3)]);
    const now = shapesOf([instance("a", 2, 0, "2"), frame("f", 3)]);
    expect(reshaped(before, now)).toEqual([]);
  });

  it("names a node whose number of rows, hidden channels or frame count changed, and any new node", () => {
    const before = shapesOf([instance("a", 2, 0), instance("b", 2, 0), frame("f", 3)]);
    const now = shapesOf([instance("a", 3, 0), instance("b", 2, 4), frame("f", 4), instance("c", 1, 0)]);
    expect(reshaped(before, now)).toEqual(["a", "b", "f", "c"]);
  });

  it("names a node that went from drawn ports to a live readout", () => {
    const ported = { ...instance("a", 2, 0), data: { view: { id: "a" }, live: null } } as unknown as FlowNode;
    expect(reshaped(shapesOf([ported]), shapesOf([instance("a", 2, 0)]))).toEqual(["a"]);
  });
});
