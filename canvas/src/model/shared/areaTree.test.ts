import { describe, expect, it } from "vitest";
import { areaTree } from "./areaTree";
import { areaNodesOf } from "../../test/modelBridge";
import { addArea } from "../../test/modelBridge";

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

  it("adds an area under a parent once, to a plant that had none", () => {
    const doc = addArea({ id: "x" }, "Site", "");
    const nested = addArea(doc, "Plant room", "Site");
    expect(addArea(nested, "Plant room", "Site")).toBe(nested);
    expect(nested.areas).toEqual([{ name: "Site" }, { name: "Plant room", parent: "Site" }]);
    expect(areaNodesOf(nested).map((node) => node.path)).toEqual(["Site", "Site › Plant room"]);
  });
});
