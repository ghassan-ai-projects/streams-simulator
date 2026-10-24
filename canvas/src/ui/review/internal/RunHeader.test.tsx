import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { navigation } from "../../../test/workspaceHarness";
import { launchHarness } from "../../../test/launchFixture";
import { reviewState } from "../../../test/reviewFixture";
import { report } from "../../../test/runFixture";
import { Header } from "./RunHeader";

describe("run header", () => {
  const header = (state: ReturnType<typeof reviewState>) => {
    const { launch } = launchHarness(new FakeBridge());
    return render(
      <Header
        state={state}
        worldId="w-4"
        hall={null}
        onHall={vi.fn()}
        plant={{ title: "Mini plant", path: "domains/mini.domain.json" }}
        navigation={navigation().store}
        launch={launch}
      />,
    );
  };

  it("puts the name of a named run above its id, which stays the heading", () => {
    header(reviewState({ report: { phase: "ready", value: { ...report, name: "Cooling loss" } } }));
    expect(screen.getByText("Cooling loss")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1, name: "Run w-4" })).toHaveAccessibleDescription("Cooling loss");
  });

  it("shows no name for a run that has none", () => {
    header(reviewState({}));
    expect(screen.queryByText("Cooling loss")).toBeNull();
  });

  it("sets up another run from this run's plan and names the run it is prefilled from", () => {
    const nav = navigation();
    const { launch } = launchHarness(new FakeBridge());
    render(
      <Header
        state={reviewState({})}
        worldId="w-4"
        hall={null}
        onHall={vi.fn()}
        plant={{ title: "Mini plant", path: "domains/mini.domain.json" }}
        navigation={nav.store}
        launch={launch}
      />,
    );
    expect(screen.queryByRole("button", { name: "Run again" })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Set up another run" }));
    expect(launch.proposal("mini")).toMatchObject({ from: "w-4" });
    expect(nav.location.hash).toContain("launch=1");
  });
});
