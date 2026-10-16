import { createEvent, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../model/plant";
import { launchBridge } from "../../test/launchFixture";
import { launchHarness } from "../../test/launchFixture";
import { domain } from "../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

describe("launch panel keys", () => {
  it("keeps Enter in a select from starting the run", () => {
    const fake = launchBridge();
    const { launch, scheduler } = launchHarness(fake);
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
    fireEvent.click(screen.getByRole("button", { name: "Add fault" }));
    fireEvent.change(screen.getByLabelText("Fault 1 instance"), { target: { value: "CH-1" } });
    const select = screen.getByLabelText("Fault 1 fault");
    const enter = createEvent.keyDown(select, { key: "Enter" });
    fireEvent(select, enter);
    expect(enter.defaultPrevented).toBe(true);
    expect(fake.calls).toEqual([]);
  });
});
