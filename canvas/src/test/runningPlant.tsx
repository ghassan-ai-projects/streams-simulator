import { act, fireEvent, screen } from "@testing-library/react";
import { mountApp } from "./appHarness";
import type { FakeBridge } from "./fakeBridge";
import { elapse } from "./launchFixture";
import { miniBridge, planHash } from "./screenFixture";

export const canary = /canary[ _-]*truth/i;
export const roles = [
  "group",
  "button",
  "status",
  "application",
  "region",
  "img",
  "listitem",
  "complementary",
  "row",
  "cell",
  "table",
] as const;

const origin = { domain: "mini", title: "Mini plant", path: "domains/mini.domain.json" };

function plan(explore: boolean) {
  return {
    seed: 1,
    durationS: 3600,
    explore,
    pace: "watch" as const,
    faults: [{ entity: "CH-1", fault: "chiller__CANARY_TRUTH_9f3a", atS: 0 }],
    calls: [{ entity: "CT-1", effector: "start", atS: 0, args: {} }],
  };
}

export async function runningPlant(explore: boolean, prepare: (bridge: FakeBridge) => void = () => undefined) {
  const bridge = miniBridge();
  prepare(bridge);
  const app = mountApp(bridge, planHash());
  await screen.findByRole("region", { name: /^Plant / });
  act(() => {
    void app.stores.launch.start(origin, plan(explore));
  });
  await screen.findByRole("button", { name: /^Running · / });
  await act(async () => {
    await elapse(app.timers, 500);
  });
  fireEvent.click(screen.getByRole("button", { name: /^Running · / }));
  await screen.findByLabelText("Simulated time");
  return { ...app, bridge };
}

export function attributeValues(): string[] {
  return [...document.body.querySelectorAll("*")].flatMap((element) =>
    [...element.attributes].map((attribute) => attribute.value),
  );
}

export function namedBy(role: (typeof roles)[number]): number {
  return screen.queryAllByRole(role, { name: canary }).length;
}
