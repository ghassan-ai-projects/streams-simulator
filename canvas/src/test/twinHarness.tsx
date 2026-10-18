import { fireEvent, screen, within } from "@testing-library/react";
import { expect, vi } from "vitest";
import { mountApp } from "./appHarness";
import type { FakeBridge } from "./fakeBridge";
import { coolerPath, twinBridge } from "./twinFixture";

export const hash = `#/device?device=${encodeURIComponent(coolerPath)}`;

export async function behaviourOf(fake: FakeBridge = twinBridge()) {
  const app = mountApp(fake, hash);
  await screen.findByRole("region", { name: "Editing cooler" });
  await vi.waitFor(() => {
    const state = app.stores.devices.snapshot();
    expect(state.phase === "ready" && state.check.kind !== "checking").toBe(true);
  });
  fireEvent.click(screen.getByRole("button", { name: /^Behaviour/ }));
  await screen.findByRole("region", { name: "Behaviour of cooler" });
  return { ...app, fake };
}

export function inReading(name: string) {
  return within(screen.getByRole("group", { name: `Reading ${name}` }));
}

export function inCommand(name: string) {
  return within(screen.getByRole("group", { name: `Command ${name}` }));
}

export function pickStep(name: RegExp) {
  fireEvent.click(within(screen.getByRole("navigation", { name: "Device steps" })).getByRole("button", { name }));
}

export function createTwin() {
  fireEvent.click(screen.getByRole("button", { name: "Create twin" }));
}
