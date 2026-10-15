import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { LaunchStore } from "../../app/launchStore";
import { parsePlant } from "../../model/plant";
import { FakeBridge } from "../../test/fakeBridge";
import { domain } from "../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

const path = "domains/mini.domain.json";

describe("prefilled run plan", () => {
  it("clears a prefilled fault the plant no longer has, says so, and does not start with it", () => {
    const launch = new LaunchStore(new FakeBridge());
    launch.propose(
      "mini",
      {
        seed: 5,
        durationS: 3600,
        explore: true,
        faults: [{ entity: "CH-1", fault: "chiller__gone", atS: 600 }],
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
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    expect(screen.getByLabelText("Fault 1 fault")).toHaveAccessibleDescription(/Fault 1 needs a fault\./);
  });
});
