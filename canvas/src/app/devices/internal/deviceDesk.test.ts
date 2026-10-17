import { describe, expect, it } from "vitest";
import { failure as failed } from "../../ports";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { createDeviceDesk } from "./deviceDesk";

const entry = (path: string, kind: string) => ({ path, kind, title: "" });

describe("device desk", () => {
  it("counts the readings and commands of each device file, by name", async () => {
    const fake = new FakeBridge();
    fake.addFile("kinds/pump.kind.json", { kind: "pump", metrics: [{ name: "flow" }], commands: [] });
    fake.addFile("kinds/fan.kind.json", { kind: "fan", metrics: [], commands: [{ name: "stop" }, { name: "go" }] });
    fake.addFile("kinds/odd.kind.json", "not json");
    const rows = await createDeviceDesk(fake.devices).rows([
      entry("kinds/pump.kind.json", "pump"),
      entry("kinds/fan.kind.json", "fan"),
      entry("kinds/odd.kind.json", ""),
      entry("kinds/gone.kind.json", "gone"),
    ]);
    expect(rows).toEqual([
      { path: "kinds/fan.kind.json", name: "fan", metrics: 0, commands: 2 },
      { path: "kinds/gone.kind.json", name: "gone", metrics: null, commands: null },
      { path: "kinds/odd.kind.json", name: "odd", metrics: null, commands: null },
      { path: "kinds/pump.kind.json", name: "pump", metrics: 1, commands: 0 },
    ]);
  });

  it("writes a new empty device that does not exist yet", async () => {
    const fake = new FakeBridge();
    const created = await createDeviceDesk(fake.devices).create(" fan ");
    expect(created).toEqual({ ok: true, value: "kinds/fan.kind.json" });
    expect(fake.writes).toHaveLength(1);
    expect(fake.writes[0]).toMatchObject({ path: "kinds/fan.kind.json", etag: "" });
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toEqual({ kind: "fan", metrics: [], commands: [] });
  });

  it("says a device exists when the file does, and passes any other failure on", async () => {
    const fake = new FakeBridge();
    fake.writeOutcomes.set("kinds/fan.kind.json", failed(cannedFailure("conflict", 409, "exists")));
    fake.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("bridge_failed", 500, "disk full")));
    const desk = createDeviceDesk(fake.devices);
    expect(await desk.create("fan")).toMatchObject({
      ok: false,
      failure: { kind: "conflict", diagnostic: { message: "A device named fan already exists." } },
    });
    expect(await desk.create("pump")).toMatchObject({ ok: false, failure: { kind: "unavailable" } });
  });
});
