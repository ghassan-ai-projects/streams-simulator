import { describe, expect, it } from "vitest";
import { failed, success } from "../model/bridge";
import { FakeBridge, failure } from "../test/fakeBridge";
import { PlantMaker } from "./plantMaker";

const base = "compositions/lab.composition.json";
const lab = { id: "lab", title: "Lab", description: "base", equipment: [{ use: "lib/chiller@1", names: ["CH-1"] }] };

function bridge() {
  const fake = new FakeBridge();
  fake.files.set(base, success({ path: base, content: JSON.stringify(lab), etag: "c1" }));
  fake.files.set("domains/lab.domain.json", success({ path: "domains/lab.domain.json", content: "{}", etag: "d1" }));
  fake.actions.set(
    "component.examples",
    success({ examples: [{ id: "pilot", title: "Pilot", description: "small" }, { title: "no id" }] }),
  );
  fake.actions.set("component.example", success({ example: { ...lab, id: "pilot", title: "Pilot" } }));
  fake.actions.set("domain.compose", success({ channels: 1 }));
  return fake;
}

const content = (fake: FakeBridge) => JSON.parse(fake.writes[0]?.content ?? "") as Record<string, unknown>;

describe("plant maker", () => {
  it("copies a workspace plant whole or keeps only its frame", async () => {
    const fake = bridge();
    const maker = new PlantMaker(fake);
    expect(await maker.create({ file: base }, "lab-two", "Lab two", true)).toEqual({
      composition: "compositions/lab-two.composition.json",
      domain: "domains/lab-two.domain.json",
    });
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
    await maker.create({ example: "pilot" }, "lab-three", "Lab three", false);
    expect(fake.calls.at(-1)?.args).toEqual({ id: "lab-three", title: "Lab three", example: "pilot", empty: true });
    expect(JSON.parse(fake.writes[1]?.content ?? "")).not.toHaveProperty("equipment");
  });

  it("refuses an existing plant, an unreadable base and an unknown example", async () => {
    const fake = bridge();
    const maker = new PlantMaker(fake);
    expect(await maker.create({ file: base }, "lab", "Lab", true)).toMatchObject({ code: "plant_exists" });
    fake.files.set("compositions/bad.composition.json", success({ path: "x", content: "[", etag: "e" }));
    expect(await maker.create({ file: "compositions/bad.composition.json" }, "bad-copy", "B", true)).toMatchObject({
      code: "composition_invalid",
    });
    const refusal = {
      ...failure("composition_invalid", 422, "fails"),
      issues: [{ pointer: "/id", message: "bad id" }],
    };
    fake.actions.set("component.new", failed(refusal));
    expect(await maker.create({ example: "pilot" }, "Bad_Id", "B", true)).toMatchObject({ message: "bad id" });
    fake.actions.set("component.new", success({ composition: "not a plant" }));
    expect(await maker.create({ example: "odd" }, "odd-copy", "O", true)).toMatchObject({
      code: "composition_unreadable",
    });
    expect(fake.writes).toEqual([]);
    fake.actions.delete("component.new");
    fake.writeOutcomes.set(
      "compositions/lab-two.composition.json",
      failed(failure("conflict", 409, "changed on disk")),
    );
    expect(await maker.create({ file: base }, "lab-two", "Lab two", true)).toMatchObject({ code: "plant_exists" });
  });

  it("lists the shipped examples and tries one: copied, composed and ready to run", async () => {
    const fake = bridge();
    const maker = new PlantMaker(fake);
    const examples = await maker.examples();
    expect(examples).toEqual([{ id: "pilot", title: "Pilot", description: "small" }]);
    const [pilot] = examples;
    if (pilot === undefined) throw new Error("no example");
    expect(await maker.tryExample(pilot, "pilot")).toEqual({
      composition: "compositions/pilot.composition.json",
      domain: "domains/pilot.domain.json",
    });
    expect(fake.calls.at(-1)).toEqual({
      name: "domain.compose",
      args: { composition: "compositions/pilot.composition.json", out: "domains/pilot.domain.json", etag: "" },
    });
    fake.actions.set("domain.compose", failed(failure("composition_invalid", 422, "bad")));
    expect(await maker.tryExample(pilot, "pilot-2")).toMatchObject({ code: "composition_invalid" });
    fake.actions.set("component.examples", failed(failure("bridge_failed", 500)));
    expect(await maker.examples()).toEqual([]);
  });
});
