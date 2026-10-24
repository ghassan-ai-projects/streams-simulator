import { act, fireEvent, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { elapse } from "../../../test/launchFixture";
import { miniBridge, planHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

const canary = /canary[ _-]*truth/i;
const origin = { domain: "mini", title: "Mini plant", path: "domains/mini.domain.json" };
const roles = ["group", "button", "status", "application", "region", "img", "listitem", "complementary"] as const;

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

async function running(explore: boolean) {
  const bridge = miniBridge();
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
  return app;
}

function attributeValues(): string[] {
  return [...document.body.querySelectorAll("*")].flatMap((element) =>
    [...element.attributes].map((attribute) => attribute.value),
  );
}

function namedBy(role: (typeof roles)[number]): number {
  return screen.queryAllByRole(role, { name: canary }).length;
}

describe("the plant while a run is going", () => {
  it("shows no planned fault anywhere on the page in Evaluate", async () => {
    await running(false);
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(roles.map(namedBy)).toEqual(roles.map(() => 0));
    expect(screen.getByRole("group", { name: /^CH-1, chiller/ })).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("Fault ·");
  });

  it("keeps the operator's own effector call visible in Evaluate", async () => {
    await running(false);
    expect(await screen.findByRole("group", { name: /^CT-1, .*effector call: start · accepted/ })).toBeInTheDocument();
  });

  it("shows the planned fault in Explore", async () => {
    await running(true);
    expect(document.body.textContent).toMatch(canary);
    expect(attributeValues().some((value) => canary.test(value))).toBe(true);
    expect(namedBy("group")).toBe(1);
  });
});
