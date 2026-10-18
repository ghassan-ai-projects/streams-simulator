import { describe, expect, it } from "vitest";
import { changeMetric, renameTwin } from "../../../model/devices";
import { deviceBridge, editingDevice } from "../../../test/deviceHarness";
import { coolerPath, twinBridge } from "../../../test/twinFixture";
import { createDeviceSession } from "./deviceStore";
import { hasUnsavedWork } from "./unsaved";

describe("work that would be lost by leaving", () => {
  it("is none for a device nobody changed", async () => {
    const { stores } = await editingDevice(deviceBridge());
    expect(hasUnsavedWork(stores.devices.snapshot(), stores.devices.twin.snapshot())).toBe(false);
    expect(hasUnsavedWork({ phase: "idle" }, { phase: "idle" })).toBe(false);
  });

  it("is the unsaved interface, or choices the user changed and that are not yet written", async () => {
    const fake = twinBridge();
    const session = createDeviceSession(fake.devices);
    await session.open(coolerPath);
    expect(hasUnsavedWork(session.snapshot(), session.twin.snapshot())).toBe(false);
    session.twin.edit((choices) => renameTwin(choices, "plant-a"));
    expect(hasUnsavedWork(session.snapshot(), session.twin.snapshot())).toBe(true);
    await session.twin.persist();
    expect(hasUnsavedWork(session.snapshot(), session.twin.snapshot())).toBe(false);
    session.edit((kind) => changeMetric(kind, "m1", { unit: "Pa" }));
    expect(hasUnsavedWork(session.snapshot(), session.twin.snapshot())).toBe(true);
  });
});
