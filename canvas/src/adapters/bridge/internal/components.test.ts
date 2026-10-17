import { describe, expect, it } from "vitest";
import { component, examples, library, presets } from "./decoders";

describe("library payload", () => {
  const wire = {
    components: [
      { kind: "equipment", id: "pump", title: "Pump", version: 2, class: "process/fluid", tags: ["Fluids", "Wet"] },
      { kind: "equipment", id: "chiller", class: "hvac", tags: "none" },
      { kind: "instrument", id: "probe", title: "Probe" },
      { kind: "equipment", id: "", title: "Nameless" },
      { kind: "equipment", id: "aaa", title: "Alpha", version: "3" },
      "stray",
    ],
  };

  it("keeps equipment, builds refs and search words, and sorts by title", () => {
    expect(library(wire)).toEqual([
      { component: "aaa", ref: "aaa@1", title: "Alpha", words: "alpha aaa " },
      { component: "chiller", ref: "chiller@1", title: "chiller", words: " chiller hvac" },
      { component: "pump", ref: "pump@2", title: "Pump", words: "pump pump process/fluid fluids wet" },
    ]);
  });

  it("is an empty library when there are no components", () => {
    expect(library({ components: [] })).toEqual([]);
  });

  it.each([[null], [{}], [{ components: "all" }], [[]]])("refuses %j", (bad) => {
    expect(library(bad)).toBeNull();
  });
});

describe("presets payload", () => {
  it("reads the options with title fallback, the default and the description", () => {
    expect(
      presets({
        document: {
          presets: { eco: { title: "Economy" }, peak: {}, odd: 4 },
          default_preset: "eco",
          description: "Load shapes",
        },
      }),
    ).toEqual({
      options: [
        { name: "eco", title: "Economy" },
        { name: "peak", title: "peak" },
        { name: "odd", title: "odd" },
      ],
      preferred: "eco",
      description: "Load shapes",
    });
  });

  it("allows a document with no presets", () => {
    expect(presets({ document: {} })).toEqual({ options: [], preferred: "", description: "" });
  });

  it.each([[null], [{}], [{ document: [] }], [{ document: "doc" }]])("refuses %j", (bad) => {
    expect(presets(bad)).toBeNull();
  });
});

describe("examples payload", () => {
  it("lists examples that have an id, falling back to the id for the title", () => {
    expect(
      examples({
        examples: [{ id: "a", title: "Alpha", description: "first" }, { id: "b" }, { title: "No id" }, null],
      }),
    ).toEqual([
      { id: "a", title: "Alpha", description: "first" },
      { id: "b", title: "b", description: "" },
    ]);
  });

  it.each([[null], [{}], [{ examples: {} }]])("refuses %j", (bad) => {
    expect(examples(bad)).toBeNull();
  });
});

describe("component payload", () => {
  const document = {
    title: "Document title",
    description: "Makes cold water",
    ports: [
      { name: "supply", type: "water", direction: "provides", state: "flow" },
      { name: "power", type: "electric", direction: "requires" },
      { name: "loose", type: "x", state: "  " },
      "stray",
    ],
  };

  it("reads ports with direction and channel, and the entry's title over the document's", () => {
    expect(component({ entry: { title: "Entry title", class: "electrical/motor" }, document })).toEqual({
      ports: [
        { name: "supply", type: "water", direction: "out", channel: "flow" },
        { name: "power", type: "electric", direction: "in" },
        { name: "loose", type: "x", direction: "in" },
        { name: "", type: "", direction: "in" },
      ],
      category: "elec",
      title: "Entry title",
      description: "Makes cold water",
    });
  });

  it("falls back to the document title and omits a blank title and description", () => {
    expect(component({ entry: { title: "  " }, document: { title: "Doc", ports: [] } })).toEqual({
      ports: [],
      category: "env",
      title: "Doc",
    });
    expect(component({ entry: {}, document: { description: " " } })).toEqual({ ports: [], category: "env" });
  });

  it.each([
    ["electrical/motor", [], "elec"],
    ["process/line", ["hvac"], "prod"],
    ["environment/room", ["drives"], "env"],
    ["", ["compressed-air"], "air"],
    ["", ["hvac"], "cool"],
    ["", ["fluids"], "cool"],
    ["", ["heat", "drives"], "cool"],
    ["", ["drives"], "prod"],
    ["", ["material-handling"], "prod"],
    ["", ["unrelated"], "env"],
    ["", [], "env"],
  ])("categorises class %j with tags %j as %s, class before tags", (componentClass, tags, expected) => {
    const spec = component({ entry: { class: componentClass, tags }, document: {} });
    expect(spec?.category).toBe(expected);
  });

  it("treats a document without a ports list as having none", () => {
    expect(component({ entry: {}, document: { ports: "none" } })?.ports).toEqual([]);
  });

  it.each([
    [null],
    [{}],
    [{ entry: {} }],
    [{ document: {} }],
    [{ entry: [], document: {} }],
    [{ entry: {}, document: 3 }],
  ])("refuses %j", (bad) => {
    expect(component(bad)).toBeNull();
  });
});
