import { describe, expect, it } from "vitest";
import { baseChoices, newPlantReadiness, startingMode } from "./baseChoice";

const bases = [
  { path: "compositions/lab.composition.json", id: "lab", title: "Lab", instances: 3 },
  { path: "compositions/pump.composition.json", id: "pump", title: "", instances: 1 },
];
const examples = [{ id: "pilot", title: "Pilot", description: "small" }];

describe("what a new plant starts from", () => {
  it("lists the workspace plants first, then the shipped examples, each with the source the binary reads", () => {
    expect(baseChoices(bases, examples)).toEqual([
      {
        key: "file:compositions/lab.composition.json",
        label: "Lab",
        group: "plants",
        source: { kind: "file", path: bases[0]?.path },
      },
      {
        key: "file:compositions/pump.composition.json",
        label: "pump",
        group: "plants",
        source: { kind: "file", path: bases[1]?.path },
      },
      { key: "example:pilot", label: "Pilot", group: "examples", source: { kind: "example", id: "pilot" } },
    ]);
    expect(baseChoices([], examples)[0]?.source).toEqual({ kind: "example", id: "pilot" });
    expect(baseChoices([], [])).toEqual([]);
  });

  it("starts a blank plant empty and any other copied", () => {
    expect(startingMode({ blank: true })).toBe("empty");
    expect(startingMode({ blank: false })).toBe("copy");
  });

  it("says what is missing before a plant can be created, in the order the form asks", () => {
    const [first] = baseChoices(bases, examples);
    expect(newPlantReadiness(" ", "", null)).toBe("Name the plant first.");
    expect(newPlantReadiness("Lab", " ", null)).toBe("Give it an id.");
    expect(newPlantReadiness("Lab", "lab-2", null)).toBe("Choose what to start from.");
    expect(newPlantReadiness("Lab", "lab-2", first ?? null)).toBe("");
  });
});
