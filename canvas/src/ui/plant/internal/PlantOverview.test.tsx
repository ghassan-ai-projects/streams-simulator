import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { campus, domain } from "../../../test/plantFixture";
import type { RunRow } from "../../../model/workspace";
import { Outline } from "../../graph";
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

  it("singularises the one-instance plant summary", () => {
    const single = parsePlant({
      id: "single",
      site: { name: "Site", areas: [{ name: "Site" }] },
      instances: [{ kind: "pump", names: ["P-1"], area: "Site" }],
    });
    render(<PlantOverview plant={single} runs={[]} />);
    expect(screen.getByText("1 instance in 1 area · 0 links")).toBeInTheDocument();
  });

  it("shows one area order on campus and power: the outline rail and the areas list agree", () => {
    const names = (prefix: string, count: number) =>
      Array.from({ length: count }, (_, i) => `${prefix}${String(i + 1)}`);
    const shapes = [
      {
        plant: {
          id: "multi-hall-campus",
          site: {
            name: "Site",
            areas: [
              { name: "Site" },
              { name: "Utilities", parent: "Site" },
              { name: "Hall 1", parent: "Site" },
              { name: "Hall 2", parent: "Site" },
              { name: "Hall 3", parent: "Site" },
            ],
          },
          instances: [
            { kind: "meter", names: ["S-1", "S-2"], area: "Site" },
            { kind: "pump", names: names("U-", 9), area: "Utilities" },
            { kind: "pump", names: names("A-", 5), area: "Hall 1" },
            { kind: "pump", names: names("B-", 5), area: "Hall 2" },
            { kind: "pump", names: names("C-", 5), area: "Hall 3" },
          ],
        },
        declared: ["Site2", "Utilities9", "Hall 15", "Hall 25", "Hall 35"],
      },
      {
        plant: {
          id: "power-distribution-standby",
          site: {
            name: "Site",
            areas: [
              { name: "Site" },
              { name: "Substation", parent: "Site" },
              { name: "Generator yard", parent: "Site" },
              { name: "Line A", parent: "Site" },
              { name: "Line B", parent: "Site" },
            ],
          },
          instances: [
            { kind: "meter", names: ["S-1"], area: "Site" },
            { kind: "pump", names: ["SB-1", "SB-2"], area: "Substation" },
            { kind: "pump", names: ["GY-1", "GY-2"], area: "Generator yard" },
            { kind: "meter", names: names("LA-", 6), area: "Line A" },
            { kind: "meter", names: names("LB-", 6), area: "Line B" },
          ],
        },
        declared: ["Site1", "Substation2", "Generator yard2", "Line A6", "Line B6"],
      },
    ];
    for (const { plant, declared } of shapes) {
      const overview = render(<PlantOverview plant={parsePlant(plant)} runs={[]} />);
      const rail = render(<Outline plant={parsePlant(plant)} hall={null} selected={null} onSelect={vi.fn()} />);
      const areas = within(overview.getByRole("region", { name: "Areas" }))
        .getAllByRole("listitem")
        .map((item) => item.textContent);
      const titles = within(rail.container)
        .getAllByRole("group")
        .map((group) => group.querySelector("summary")?.textContent);
      expect(areas).toEqual(declared);
      expect(titles).toEqual(declared);
      overview.unmount();
      rail.unmount();
    }
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
