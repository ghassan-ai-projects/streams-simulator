import { act, fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";
import type { Plant } from "../model/plant";
import { LaunchPanel } from "../ui/launch";
import type { FakeBridge } from "./fakeBridge";
import { flushPromises } from "./deferred";
import { elapse, launchBridge, launchHarness } from "./launchFixture";
import { setpointPlant } from "./setpointPlant";

type Mode = "Explore" | "Evaluate";

interface Options {
  readonly mode?: Mode;
  readonly bridge?: FakeBridge;
  readonly plant?: Plant;
}

export async function runningPanel({
  mode = "Explore",
  bridge = launchBridge("w-7"),
  plant = setpointPlant,
}: Options = {}) {
  const { launch, timers } = launchHarness(bridge);
  const view = render(
    <LaunchPanel launch={launch} plant={plant} path="p.json" onClose={vi.fn()} onFinished={vi.fn()} />,
  );
  fireEvent.click(screen.getByRole("radio", { name: new RegExp(`^${mode}`) }));
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    await flushPromises();
  });
  await screen.findByText(/^Running Mini plant as run w-7/);
  const tick = (ms: number) =>
    act(async () => {
      await elapse(timers, ms);
    });
  const press = (name: string) =>
    act(async () => {
      fireEvent.click(screen.getByRole("button", { name }));
      await flushPromises();
    });
  const choose = (label: string, value: string) =>
    act(async () => {
      fireEvent.change(screen.getByLabelText(label), { target: { value } });
      await flushPromises();
    });
  return { bridge, launch, timers, view, tick, press, choose };
}
