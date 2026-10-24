import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { flushPromises } from "../../../test/deferred";
import { launchBridge, launchHarness } from "../../../test/launchFixture";
import { deliveryKinds, teachCatalog } from "../../../test/planCatalogFixture";
import { domain } from "../../../test/plantFixture";
import { attributeValues, canary } from "../../../test/runningPlant";
import { LaunchPanel } from "./LaunchPanel";

const secret = {
  ...domain,
  kinds: [{ ...domain.kinds[0], faults: [{ id: "chiller__CANARY_TRUTH_9f3a" }] }, ...domain.kinds.slice(1)],
};

const plan = (explore: boolean) => ({
  name: "Cooling loss",
  purpose: "Observe the chiller with the tower out",
  seed: 1,
  durationS: 3600,
  explore,
  pace: "watch" as const,
  faults: [{ entity: "CH-1", fault: "chiller__CANARY_TRUTH_9f3a", atS: 60, untilS: 300 }],
  perturbations: [{ kind: "CANARY_TRUTH_drop", params: { rate: 0.5 }, atS: 120, untilS: 600 }],
  calls: [],
});

async function openWith(explore: boolean) {
  const fake = launchBridge("w-7");
  const [first] = deliveryKinds;
  if (first !== undefined) teachCatalog(fake, {}, [{ ...first, kind: "CANARY_TRUTH_drop" }]);
  const { launch } = launchHarness(fake);
  launch.propose("mini", plan(explore), [], "w-2");
  render(
    <LaunchPanel
      launch={launch}
      plant={parsePlant(secret)}
      path="domains/mini.domain.json"
      onClose={vi.fn()}
      onFinished={vi.fn()}
    />,
  );
  await act(async () => {
    await flushPromises();
  });
}

async function startIt() {
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    await flushPromises();
  });
  await screen.findByText(/^Running Cooling loss on Mini plant as run w-7/);
}

describe("the plan of an evaluation", () => {
  it("is the operator's own to read while it is being written, in the list, the editors and the review", async () => {
    await openWith(false);
    expect(screen.getByRole("list", { name: "Plan in time order" })).toHaveTextContent(/CANARY.TRUTH.9f3a/i);
    expect(screen.getByRole("list", { name: "Plan in time order" })).toHaveTextContent(/CANARY.TRUTH.drop/i);
    expect(screen.getByLabelText("Fault 1 fault")).toHaveValue("chiller__CANARY_TRUTH_9f3a");
  });

  it("shows the name and purpose but no fault or delivery condition once it is running", async () => {
    await openWith(false);
    await startIt();
    expect(screen.getByText("Observe the chiller with the tower out")).toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(screen.queryByRole("list", { name: "Plan in time order" })).toBeNull();
    expect(screen.queryByRole("region", { name: "Before you start" })).toBeNull();
    expect(screen.queryByLabelText("Fault 1 fault")).toBeNull();
  });

  it("shows the same of an exploration: the plan is not listed once the run is going", async () => {
    await openWith(true);
    await startIt();
    expect(screen.queryByRole("list", { name: "Plan in time order" })).toBeNull();
    expect(screen.getByText("Observe the chiller with the tower out")).toBeInTheDocument();
  });
});
