import { describe, expect, it } from "vitest";
import { layout } from "./layout";

const spacing = { pad: 24, header: 40, column: 96, group: 48, lane: 24 } as const;

const node = (id: string, group = "A", height = 60) => ({ id, group, width: 100, height });

describe("layout", () => {
  it("ranks along links and breaks cycles in a stable order", () => {
    const placed = layout(
      [node("a"), node("b"), node("c")],
      [
        { from: "a", to: "b" },
        { from: "b", to: "c" },
        { from: "c", to: "a" },
        { from: "a", to: "a" },
      ],
      ["A"],
    );
    const step = 100 + spacing.column;
    const xs = ["a", "b", "c"].map((id) => placed.nodes.get(id)?.x);
    expect(xs).toEqual([spacing.pad, spacing.pad + step, spacing.pad + 2 * step]);
  });

  it("places each rank in its own column inside a framed group", () => {
    const placed = layout([node("a"), node("b")], [{ from: "a", to: "b" }], ["A"]);
    const a = placed.nodes.get("a");
    const b = placed.nodes.get("b");
    expect(a).toEqual({ x: spacing.pad, y: spacing.header, width: 100, height: 60 });
    expect(b?.x).toBe(spacing.pad + 100 + spacing.column);
    expect(placed.groups.get("A")).toEqual({
      x: 0,
      y: 0,
      width: 2 * spacing.pad + 200 + spacing.column,
      height: spacing.header + 60 + spacing.pad,
    });
  });

  it("wraps a tall column and orders rows by their predecessors", () => {
    const nodes = ["s1", "s2", "t1", "t2", "t3", "t4"].map((id) => node(id));
    const edges = [
      { from: "s2", to: "t1" },
      { from: "s1", to: "t2" },
      { from: "s1", to: "t3" },
      { from: "s2", to: "t4" },
    ];
    const placed = layout(nodes, edges, ["A"]);
    const columns = new Set(["t1", "t2", "t3", "t4"].map((id) => placed.nodes.get(id)?.x));
    expect(columns.size).toBe(2);
    expect((placed.nodes.get("t2")?.y ?? 0) < (placed.nodes.get("t1")?.y ?? 0)).toBe(true);
  });

  it("puts groups in the given order and wraps them to a new row past the width", () => {
    const wide = Array.from({ length: 40 }, (_, i) => node(`n${String(i)}`, "B"));
    const chain = wide.slice(1).map((n, i) => ({ from: `n${String(i)}`, to: n.id }));
    const placed = layout([node("x", "A"), ...wide, node("z", "C")], chain, ["C", "A"]);
    expect(placed.groups.get("C")?.x).toBe(0);
    expect((placed.groups.get("A")?.x ?? 0) > 0).toBe(true);
    expect((placed.groups.get("B")?.y ?? 0) > 0).toBe(true);
  });

  it("packs areas into rows that suit the canvas shape", () => {
    const nodes = ["a", "b", "c", "d", "e", "f"].map((group) => ({ id: group, group, width: 200, height: 120 }));
    const extent = (aspect?: number) => {
      const boxes = [...layout(nodes, [], [], undefined, aspect).groups.values()];
      return {
        width: Math.max(...boxes.map((box) => box.x + box.width)),
        height: Math.max(...boxes.map((box) => box.y + box.height)),
      };
    };
    const wide = extent(4);
    const square = extent(1);
    expect(wide.width).toBeGreaterThan(square.width);
    expect(square.height).toBeGreaterThan(wide.height);
    expect(extent()).toEqual(extent(10));
  });
});

describe("layout header reserve", () => {
  it("opens the gap between area rows by the tallest header, and leaves the gap alone without one", () => {
    const nodes = [
      { id: "a", group: "One", width: 200, height: 100 },
      { id: "b", group: "Two", width: 200, height: 100 },
    ];
    const plain = layout(nodes, [], ["One", "Two"], () => 3, 0.2);
    const reserved = layout(
      nodes,
      [],
      ["One", "Two"],
      () => 3,
      0.2,
      () => 70,
    );
    const gap = (placed: ReturnType<typeof layout>) => {
      const one = placed.groups.get("One");
      const two = placed.groups.get("Two");
      if (one === undefined || two === undefined) throw new Error("no groups");
      return Math.abs(two.y - (one.y + one.height));
    };
    expect(gap(plain)).toBe(spacing.group);
    expect(gap(reserved)).toBe(70 + spacing.lane);
  });
});
