import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { failed, success } from "../../model/bridge";
import { parsePlant } from "../../model/plant";
import { FakeBridge, failure } from "../../test/fakeBridge";
import { domain } from "../../test/plantFixture";
import { launchHarness } from "../../test/launchFixture";
import { LaunchPanel } from "./LaunchPanel";

describe("a run that stops on a fault", () => {
  it("names the fault and the instance in one sentence and keeps the binary message in a disclosure", async () => {
    const fake = new FakeBridge();
    fake.actions.set("world.create", success({ world_id: "w-8" }));
    fake.actions.set("fault.inject", failed(failure("unknown_fault", 422, 'unknown fault "chiller__fouling"')));
    fake.actions.set("world.destroy", success({}));
    const { launch, scheduler } = launchHarness(fake);
    launch.propose("mini", {
      seed: 1,
      durationS: 3600,
      explore: true,
      faults: [{ entity: "CH-1", fault: "chiller__fouling", atS: 0 }],
      calls: [],
    });
    render(
      <LaunchPanel
        launch={launch}
        scheduler={scheduler}
        plant={parsePlant(domain)}
        path="domains/mini.domain.json"
        onClose={vi.fn()}
        onFinished={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent(
      'The run did not finish. The fault "fouling" could not be added to CH-1. Check that this plant still has it, choose a fault, and start again.',
    );
    expect(alert.querySelector("details")).toHaveTextContent('unknown fault "chiller__fouling"');
  });
});
