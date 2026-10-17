import { screen } from "@testing-library/react";
import { expect, vi } from "vitest";
import { mountApp } from "./appHarness";
import { FakeBridge, cannedFailure } from "./fakeBridge";
import { failure, success } from "./outcome";

export const path = "kinds/fan.kind.json";
export const hash = `#/device?device=${encodeURIComponent(path)}`;

export const fan = {
  kind: "fan",
  metrics: [
    { name: "duty", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, writable: true },
    { name: "mode", data_type: "enum", enum_values: ["auto", "manual"] },
  ],
  commands: [
    {
      name: "set_duty",
      risk_class: "medium",
      args_schema: {
        type: "object",
        properties: { duty: { type: "number", minimum: 0, maximum: 100 } },
        required: ["duty"],
      },
    },
    { name: "stop", risk_class: "low", safe_stop: true, args_schema: { type: "object", properties: {} } },
  ],
};

export const valid = success({
  valid: true,
  kinds: [{ path, kind: "fan", kind_digest: "sha256:abc", metrics: 2, commands: 2 }],
});

export function refusal(pointer: string, message: string) {
  return failure({
    ...cannedFailure("command_failed", 422, "kind file fails its schema"),
    issues: [{ pointer, message }],
  });
}

export function deviceBridge(file: unknown = fan) {
  const fake = new FakeBridge();
  fake.sessionOutcome = success(true);
  fake.addFile(path, file, "k1");
  fake.actions.set("kind.validate", valid);
  return fake;
}

export async function editingDevice(fake = deviceBridge()) {
  const app = mountApp(fake, hash);
  await screen.findByRole("region", { name: "Editing fan" });
  await vi.waitFor(() => {
    const state = app.stores.devices.snapshot();
    expect(state.phase === "ready" && state.check.kind !== "checking").toBe(true);
  });
  return { ...app, fake };
}
