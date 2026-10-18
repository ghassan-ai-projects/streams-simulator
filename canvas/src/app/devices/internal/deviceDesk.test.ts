import { describe, expect, it } from "vitest";
import { failure as failed } from "../../ports";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { kindCheck } from "../../../test/kindCheckFake";
import { createDeviceDesk } from "./deviceDesk";

const entry = (path: string, kind: string) => ({ path, kind, title: "" });

function bridge() {
  const fake = new FakeBridge();
  fake.handlers.set("kind.check", kindCheck);
  return fake;
}

describe("device desk", () => {
  it("counts the readings and commands of each device file and says whether it has a twin", async () => {
    const fake = bridge();
    fake.addFile("kinds/pump.kind.json", { kind: "pump", metrics: [{ name: "flow" }], commands: [] });
    fake.addFile("twins/pump.twin.json", { id: "pump-twin", kinds: [] });
    fake.addFile("kinds/fan.kind.json", { kind: "fan", metrics: [], commands: [{ name: "stop" }, { name: "go" }] });
    fake.addFile("kinds/odd.kind.json", "not json");
    fake.addFile("kinds/lost.kind.json", { kind: "lost", metrics: [], commands: [] });
    fake.addFile("twins/lost.twin.json", { id: "lost-twin", kinds: [] });
    fake.addFile("kinds/half.kind.json", { kind: "half", metrics: [], commands: [] });
    fake.addFile("twins/half.twin.json", "not json");
    const rows = await createDeviceDesk(fake.devices).rows(
      [
        entry("kinds/pump.kind.json", "pump"),
        entry("kinds/fan.kind.json", "fan"),
        entry("kinds/odd.kind.json", ""),
        entry("kinds/gone.kind.json", "gone"),
        entry("kinds/lost.kind.json", "lost"),
        entry("kinds/half.kind.json", "half"),
      ],
      ["domains/pump-twin.domain.json"],
    );
    expect(rows).toEqual([
      { path: "kinds/fan.kind.json", name: "fan", metrics: 0, commands: 2, status: "needs-twin" },
      { path: "kinds/gone.kind.json", name: "gone", metrics: null, commands: null, status: "invalid" },
      { path: "kinds/half.kind.json", name: "half", metrics: 0, commands: 0, status: "needs-twin" },
      { path: "kinds/lost.kind.json", name: "lost", metrics: 0, commands: 0, status: "needs-twin" },
      { path: "kinds/odd.kind.json", name: "odd", metrics: null, commands: null, status: "invalid" },
      { path: "kinds/pump.kind.json", name: "pump", metrics: 1, commands: 0, status: "twin-ready" },
    ]);
  });

  it("checks a new empty device with the binary, then writes it, when it does not exist yet", async () => {
    const fake = bridge();
    const created = await createDeviceDesk(fake.devices).create(" fan ");
    expect(created).toEqual({ ok: true, value: "kinds/fan.kind.json" });
    expect(fake.calls[0]).toMatchObject({ name: "kind.check", args: { twin: false } });
    expect(fake.writes).toHaveLength(1);
    expect(fake.writes[0]).toMatchObject({ path: "kinds/fan.kind.json", etag: "" });
    expect(JSON.parse(fake.writes[0]?.content ?? "")).toEqual({ kind: "fan", metrics: [], commands: [] });
  });

  it("writes nothing when the binary refuses the name, and says so in plain words", async () => {
    const fake = bridge();
    const created = await createDeviceDesk(fake.devices).create("bad name");
    expect(created).toMatchObject({
      ok: false,
      failure: { kind: "refused", diagnostic: { message: "Use letters, digits and . _ : - only" } },
    });
    expect(fake.writes).toEqual([]);
    fake.handlers.set("kind.check", () => failed(cannedFailure("bridge_failed", 500, "down")));
    expect(await createDeviceDesk(fake.devices).create("fan")).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    expect(fake.writes).toEqual([]);
  });

  it("says a device exists when the file does, and passes any other failure on", async () => {
    const fake = bridge();
    fake.writeOutcomes.set("kinds/fan.kind.json", failed(cannedFailure("conflict", 409, "exists")));
    fake.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("bridge_failed", 500, "disk full")));
    const desk = createDeviceDesk(fake.devices);
    expect(await desk.create("fan")).toMatchObject({
      ok: false,
      failure: { kind: "conflict", diagnostic: { message: "A device named fan already exists." } },
    });
    expect(await desk.create("pump")).toMatchObject({ ok: false, failure: { kind: "unavailable" } });
  });

  it("keeps one capture session, so a review survives the screen that showed it", () => {
    const desk = createDeviceDesk(bridge().devices);
    expect(desk.capture()).toBe(desk.capture());
    expect(createDeviceDesk(bridge().devices).capture()).not.toBe(desk.capture());
  });
});
