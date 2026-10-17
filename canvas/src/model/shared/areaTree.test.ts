import { describe, expect, it } from "vitest";
import { areaTree } from "./areaTree";

describe("area tree", () => {
  it("orders areas under their parent, with the path and depth", () => {
    const nodes = areaTree([
      { name: "Floors", parent: "Site" },
      { name: "Site", parent: null },
      { name: "Floors 1-3", parent: "Floors" },
      { name: "Plant room", parent: "Site" },
    ]);
    expect(nodes.map((node) => [node.name, node.depth, node.path])).toEqual([
      ["Site", 0, "Site"],
      ["Floors", 1, "Site › Floors"],
      ["Floors 1-3", 2, "Site › Floors › Floors 1-3"],
      ["Plant room", 1, "Site › Plant room"],
    ]);
  });

  it("keeps an area whose parent is missing or that sits in a cycle", () => {
    const nodes = areaTree([
      { name: "A", parent: "Nowhere" },
      { name: "B", parent: "C" },
      { name: "C", parent: "B" },
    ]);
    expect(nodes.map((node) => node.name).sort()).toEqual(["A", "B", "C"]);
  });
});
