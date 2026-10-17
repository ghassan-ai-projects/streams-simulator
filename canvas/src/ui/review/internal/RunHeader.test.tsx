import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { navigation } from "../../../test/homeHarness";
import { launchHarness } from "../../../test/launchFixture";
import { reviewState } from "../../../test/reviewFixture";
import { Header } from "./RunHeader";

describe("run header", () => {
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
    expect(nav.location.hash).toContain("run=1");
  });
});
