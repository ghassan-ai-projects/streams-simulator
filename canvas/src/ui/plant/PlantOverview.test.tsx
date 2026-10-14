import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { parsePlant } from "../../model/plant";
import { campus, domain } from "../../test/plantFixture";
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
    expect(screen.getByRole("region", { name: "Areas" })).toBeInTheDocument();
  });
});
