import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { FakeBridge } from "../../../test/fakeBridge";
import { domain } from "../../../test/plantFixture";
import { launchHarness } from "../../../test/launchFixture";
import { LaunchPanel } from "./LaunchPanel";

const path = "domains/mini.domain.json";

describe("prefilled run plan", () => {
  it("clears a prefilled fault the plant no longer has, says so, and does not start with it", () => {
    const { launch } = launchHarness(new FakeBridge());
    launch.propose(
      "mini",
      {
        seed: 5,
        durationS: 3600,
        explore: true,
        faults: [{ entity: "CH-1", fault: "chiller__gone", atS: 600 }],
        perturbations: [],
        calls: [],
      },
      [],
      "w-6",
    );
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={vi.fn()} />,
    );
    expect(screen.getByLabelText("Fault 1 fault")).toHaveValue("");
    expect(screen.getByRole("note")).toHaveTextContent("Not copied: gone on CH-1 (no longer in this plant).");
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(screen.getByLabelText("Fault 1 fault")).toHaveAccessibleDescription(/Fault 1 needs a fault\./);
  });

  it("says the plan of a run that did not finish could not be recovered, and offers fresh settings in its mode", () => {
    const { launch } = launchHarness(new FakeBridge());
    launch.proposeFresh(
      "mini",
      { seed: 1, durationS: 1800, explore: false, pace: "watch", faults: [], perturbations: [], calls: [] },
      "w-9",
    );
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={vi.fn()} />,
    );
    expect(screen.getByRole("note")).toHaveTextContent(
      "The plan of this run could not be recovered because it did not finish; these are fresh settings. The mode is the one the run opened with: Evaluate.",
    );
    expect(screen.getByRole("note")).not.toHaveTextContent("Prefilled from");
    expect(screen.getByRole("radio", { name: /^Evaluate/ })).toBeChecked();
    expect(screen.getByLabelText("Duration")).toHaveValue("30");
    expect(screen.queryByLabelText("Fault 1 instance")).toBeNull();
  });

  it("puts the note inside the plan's scrolling content, before the Run group, not above the form", () => {
    const { launch } = launchHarness(new FakeBridge());
    launch.propose(
      "mini",
      { seed: 5, durationS: 3600, explore: true, faults: [], perturbations: [], calls: [] },
      [],
      "w-6",
    );
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={vi.fn()} />,
    );
    const scroller = screen.getByRole("form", { name: "Run plan" }).firstElementChild;
    expect(scroller).toContainElement(screen.getByRole("note"));
    expect(scroller?.firstElementChild).toBe(screen.getByRole("note"));
  });
});
