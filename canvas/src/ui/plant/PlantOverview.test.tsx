import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { parsePlant } from "../../model/plant";
import { campus, domain } from "../../test/plantFixture";
import type { RunRow } from "../../model/workspace";
import { PlantOverview } from "./PlantOverview";

describe("plant overview", () => {
  it("counts a campus by hall and keeps its areas apart", () => {
    render(<PlantOverview plant={parsePlant(campus(3))} runs={[]} />);
    expect(screen.getByText("7 instances in 3 halls and the site · 9 links")).toBeInTheDocument();
    const halls = screen.getByRole("region", { name: "Halls" });
    expect(
      within(halls)
        .getAllByRole("listitem")
        .map((item) => item.textContent),
    ).toEqual(["h012", "h022", "h032"]);
    expect(screen.getByRole("region", { name: "Areas across the halls" })).toBeInTheDocument();
  });

  it("counts a plant without halls by area", () => {
    render(<PlantOverview plant={parsePlant(domain)} runs={[]} />);
    expect(screen.queryByRole("region", { name: "Halls" })).toBeNull();
    const areas = screen.getByRole("region", { name: "Areas" });
    expect(
      within(areas)
        .getAllByRole("listitem")
        .map((item) => item.textContent),
    ).toEqual(["Cooling2", "Power1"]);
  });

  it("shows five recent runs and offers the rest on request", () => {
    const runs: RunRow[] = Array.from({ length: 8 }, (_, i) => ({
      plantTitle: "Campus",
      run: {
        worldId: `w-${String(i)}`,
        ended: true,
        complete: true,
        incompleteReason: "",
        explore: false,
        seed: null,
        plant: "campus",
        durationS: 60,
        recordedAt: "2026-10-14T10:00:00Z",
      },
    }));
    render(<PlantOverview plant={parsePlant(domain)} runs={runs} />);
    const recent = screen.getByRole("region", { name: "Recent runs" });
    expect(within(recent).getAllByRole("listitem")).toHaveLength(5);
    expect(within(recent).getAllByRole("listitem")[0]?.textContent).toMatch(/14 Oct 2026, \d\d:00/);
    fireEvent.click(within(recent).getByRole("button", { name: "Show all 8 runs" }));
    expect(within(recent).getAllByRole("listitem")).toHaveLength(8);
  });
});
