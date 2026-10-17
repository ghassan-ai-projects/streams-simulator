import { describe, expect, it } from "vitest";
import { addProblem, libraryRows, libraryWords, suggestedWords } from "./addWords";
import { compositionPlant, shortKind } from "./composition";

const plant = compositionPlant(
  { id: "lab", equipment: [{ use: "lib/chiller@1", names: ["CH-1"], area: "Site" }] },
  shortKind,
);

describe("add equipment words", () => {
  it("names the first thing missing, in the order the form is filled", () => {
    expect(addProblem("", "", "", plant)).toBe("Choose equipment from the library.");
    expect(addProblem("lib/pump@1", "", "Site", plant)).toBe("Give the instance a name.");
    expect(addProblem("lib/pump@1", "bad name", "Site", plant)).toMatch(
      /^"bad name" is not a valid name\. An instance name/,
    );
    expect(addProblem("lib/pump@1", "CH-1", "Site", plant)).toBe("CH-1 is already used by another instance.");
    expect(addProblem("lib/pump@1", "P-1", "", plant)).toBe("Choose an area.");
    expect(addProblem("lib/pump@1", "P-1", "Site", plant)).toBeNull();
  });

  it("shows two to six library rows", () => {
    expect([0, 1, 2, 4, 6, 40].map(libraryRows)).toEqual([2, 2, 2, 4, 6, 6]);
  });

  it("counts the matches and says when the list scrolls or is empty", () => {
    expect(libraryWords(1, 2, "")).toBe("1 component");
    expect(libraryWords(6, 6, "")).toBe("6 components");
    expect(libraryWords(9, 6, "")).toBe("9 components · scroll the list for the rest");
    expect(libraryWords(0, 2, " zz ")).toBe('Nothing in the library matches "zz".');
  });

  it("offers the suggested name only when there is one", () => {
    expect(suggestedWords("")).toBe("Leave empty to use the suggested name.");
    expect(suggestedWords("P-1")).toBe("Leave empty to use the suggested name, P-1.");
  });
});
