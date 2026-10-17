import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../ports";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { PlantMaker } from "./plantMaker";

const base = "compositions/lab.composition.json";
const lab = { id: "lab", title: "Lab", description: "base", equipment: [{ use: "lib/chiller@1", names: ["CH-1"] }] };

function bridge() {
  const fake = new FakeBridge();
  fake.addFile(base, lab, "c1");
  fake.addFile("domains/lab.domain.json", "{}", "d1");
  fake.actions.set(
    "component.examples",
    success({ examples: [{ id: "pilot", title: "Pilot", description: "small" }, { title: "no id" }] }),
  );
  fake.actions.set(
    "component.example",
    success({
      example: { ...lab, id: "pilot", title: "Pilot", site: { name: "Pilot plant" }, areas: [{ name: "Utilities" }] },
    }),
  );
  fake.actions.set("domain.compose", success({ channels: 1, faults: 0, effectors: 0 }));
  return fake;
}

const content = (fake: FakeBridge) => JSON.parse(fake.writes[0]?.content ?? "") as Record<string, unknown>;

describe("plant maker", () => {
  it("copies a workspace plant whole or keeps only its frame", async () => {
    const fake = bridge();
    const maker = new PlantMaker(fake.plant);
    expect(await maker.create({ kind: "file", path: base }, "lab-two", "Lab two", "copy")).toEqual(
      success({
        composition: "compositions/lab-two.composition.json",
        domain: "domains/lab-two.domain.json",
      }),
    );
    expect(fake.writes[0]).toMatchObject({ path: "compositions/lab-two.composition.json", etag: "" });
    expect(content(fake)).toMatchObject({
      id: "lab-two",
      title: "Lab two",
      equipment: lab.equipment,
      description: "base",
    });
    expect(fake.calls.find((call) => call.name === "component.new")?.args).toEqual({
      id: "lab-two",
      title: "Lab two",
      from: base,
    });
    expect(fake.calls.find((call) => call.name === "domain.compose")?.args).toMatchObject({
      composition: "compositions/lab-two.composition.json",
      out: "domains/lab-two.domain.json",
    });
    await maker.create({ kind: "example", id: "pilot" }, "lab-three", "Lab three", "empty");
    expect(fake.calls.at(-1)?.args).toEqual({ id: "lab-three", title: "Lab three", example: "pilot", empty: true });
    expect(fake.calls.filter((call) => call.name === "domain.compose")).toHaveLength(1);
    const made = JSON.parse(fake.writes[1]?.content ?? "") as Record<string, unknown>;
    expect(made).toMatchObject({ site: { name: "Pilot plant" }, areas: [{ name: "Pilot plant" }] });
    expect(made).not.toHaveProperty("equipment");
  });

  it("refuses an existing plant, an unreadable base and an unknown example", async () => {
    const fake = bridge();
    const maker = new PlantMaker(fake.plant);
    expect(await maker.create({ kind: "file", path: base }, "lab", "Lab", "copy")).toMatchObject({
      ok: false,
      failure: { kind: "conflict", diagnostic: { displayCode: "plant_exists" } },
    });
    fake.addFile("compositions/bad.composition.json", "[");
    expect(
      await maker.create({ kind: "file", path: "compositions/bad.composition.json" }, "bad-copy", "B", "copy"),
    ).toMatchObject({ ok: false, failure: { kind: "unreadable", diagnostic: { displayCode: "composition_invalid" } } });
    const refusal = {
      ...cannedFailure("composition_invalid", 422, "fails"),
      issues: [{ pointer: "/id", message: "bad id" }],
    };
    fake.actions.set("component.new", failed(refusal));
    expect(await maker.create({ kind: "example", id: "pilot" }, "Bad_Id", "B", "copy")).toMatchObject({
      ok: false,
      failure: {
        kind: "invalid-composition",
        diagnostic: { message: "" },
        issues: [{ pointer: "/id", message: "bad id" }],
      },
    });
    expect(await maker.idProblem({ kind: "example", id: "pilot" }, "Bad_Id", "B")).toBe("bad id");
    fake.actions.set(
      "component.new",
      failed({
        ...cannedFailure("composition_invalid", 422, "fails"),
        issues: [
          { pointer: "/equipment", message: "missing required property" },
          { pointer: "/id", message: "bad id" },
        ],
      }),
    );
    expect(await maker.create({ kind: "example", id: "pilot" }, "Bad_Id", "B", "copy")).toMatchObject({
      ok: false,
      failure: { kind: "invalid-composition", diagnostic: { message: "missing required property" } },
    });
    expect(await maker.idProblem({ kind: "example", id: "pilot" }, "Bad_Id", "B")).toBe("bad id");
    fake.actions.set("component.new", success({ composition: "not a plant" }));
    expect(await maker.create({ kind: "example", id: "odd" }, "odd-copy", "O", "copy")).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
    expect(fake.writes).toEqual([]);
    fake.actions.delete("component.new");
    fake.writeOutcomes.set(
      "compositions/lab-two.composition.json",
      failed(cannedFailure("conflict", 409, "changed on disk")),
    );
    expect(await maker.create({ kind: "file", path: base }, "lab-two", "Lab two", "copy")).toMatchObject({
      ok: false,
      failure: { kind: "conflict", diagnostic: { displayCode: "plant_exists" } },
    });
  });

  it("lists the shipped examples and tries one: copied, composed and ready to run", async () => {
    const fake = bridge();
    const maker = new PlantMaker(fake.plant);
    const examples = await maker.examples();
    expect(examples).toEqual([{ id: "pilot", title: "Pilot", description: "small" }]);
    const [pilot] = examples;
    if (pilot === undefined) throw new Error("no example");
    expect(await maker.tryExample(pilot, "pilot")).toEqual(
      success({
        composition: "compositions/pilot.composition.json",
        domain: "domains/pilot.domain.json",
      }),
    );
    expect(fake.calls.at(-1)).toEqual({
      name: "domain.compose",
      args: { composition: "compositions/pilot.composition.json", out: "domains/pilot.domain.json", etag: "" },
    });
    fake.actions.set("domain.compose", failed(cannedFailure("composition_invalid", 422, "bad")));
    expect(await maker.tryExample(pilot, "pilot-2")).toMatchObject({
      ok: false,
      failure: { kind: "unreadable", diagnostic: { displayCode: "composition_invalid" } },
    });
    fake.actions.set("component.examples", failed(cannedFailure("bridge_failed", 500)));
    expect(await maker.examples()).toEqual([]);
  });

  it("keeps a created copy when composing it fails, because the editor opens the composition and composes on save", async () => {
    const fake = bridge();
    fake.actions.set("domain.compose", failed(cannedFailure("composition_invalid", 422, "bad")));
    const maker = new PlantMaker(fake.plant);
    expect(await maker.create({ kind: "file", path: base }, "lab-two", "Lab two", "copy")).toMatchObject({
      ok: true,
      value: { composition: "compositions/lab-two.composition.json" },
    });
    expect(fake.calls.filter((call) => call.name === "domain.compose")).toHaveLength(1);
  });
});
