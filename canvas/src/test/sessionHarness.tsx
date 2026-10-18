import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";
import { parsePlant } from "../model/plant";
import { LaunchPanel } from "../ui/launch";
import type { FakeBridge } from "./fakeBridge";
import { launchBridge, launchHarness } from "./launchFixture";
import { success } from "./outcome";
import { domain } from "./plantFixture";

export const check = {
  domain: "mini",
  environment: "mixed",
  instances: [
    { instance: "CH-1", origin: "real", kind: "chiller", board: "uno-07", device: "ch-1" },
    { instance: "CT-1", origin: "simulated", kind: "tower" },
  ],
};

export function sessionBridge(): FakeBridge {
  const fake = launchBridge();
  fake.actions.set("session.status", success({ state: "none" }));
  fake.actions.set("session.check", success(check));
  fake.actions.set("session.start", success({ run: "c-1", state: "running", domain: "mini", plan: check }));
  fake.actions.set("session.stop", success({ run: "c-1", state: "ended", domain: "mini" }));
  return fake;
}

export function renderSession(fake: FakeBridge, onFinished = vi.fn<(run: string) => void>()) {
  const { launch } = launchHarness(fake);
  const view = render(
    <LaunchPanel
      launch={launch}
      plant={parsePlant(domain)}
      path="domains/mini.domain.json"
      onClose={vi.fn()}
      onFinished={onFinished}
    />,
  );
  fireEvent.click(screen.getByRole("radio", { name: /^Connected session/ }));
  return { launch, view, onFinished };
}

export function fill(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

export function planWithARealBoard() {
  fill("Host socket", "/tmp/h.sock");
  fill("Gateway socket", "/tmp/g.sock");
  fireEvent.click(screen.getByRole("button", { name: "Add a real board" }));
  fill("Real board 1 board", "uno-07");
  fill("Real board 1 device", "ch-1");
}
