import { act, fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";
import { parsePlant } from "../model/plant";
import { LaunchPanel } from "../ui/launch";
import type { FakeBridge } from "./fakeBridge";
import { flushPromises } from "./deferred";
import { launchBridge, launchHarness } from "./launchFixture";
import { domain } from "./plantFixture";

const richDomain = {
  ...domain,
  kinds: [
    {
      ...domain.kinds[0],
      faults: [{ id: "chiller__fouling" }, { id: "chiller__sensor_drift" }, { id: "chiller__stuck_valve" }],
      effectors: [
        {
          name: "chiller__set_setpoint",
          args_schema: {
            required: ["chiller_id", "setpoint_c"],
            properties: { chiller_id: { type: "string" }, setpoint_c: { type: "number", minimum: 4, maximum: 12 } },
          },
        },
      ],
    },
    ...domain.kinds.slice(1),
  ],
};

export async function renderPanel(fake: FakeBridge = launchBridge()) {
  const { launch, timers } = launchHarness(fake);
  const view = render(
    <LaunchPanel
      launch={launch}
      plant={parsePlant(richDomain)}
      path="domains/mini.domain.json"
      onClose={vi.fn()}
      onFinished={vi.fn()}
    />,
  );
  await act(async () => {
    await flushPromises();
  });
  return { fake, launch, timers, view };
}

export function change(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

export function press(name: string | RegExp) {
  fireEvent.click(screen.getByRole("button", { name }));
}

export async function start(fake: FakeBridge) {
  fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
  await act(async () => {
    press("Start the run");
    await flushPromises();
  });
  return fake.jobCalls.find((call) => call.name === "job.start")?.args;
}
