import { render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { parsePlant } from "../../model/plant";
import { composition, domain } from "../../test/plantFixture";
import { Inspector } from "./Inspector";

describe("inspector", () => {
  it("asks for a selection when nothing is selected", () => {
    render(<Inspector plant={parsePlant(domain)} selected={null} />);
    expect(screen.getByRole("complementary", { name: "Inspector" })).toHaveTextContent("Select an instance");
  });

  it("shows the selected instance's kind and links with no accessibility violation", async () => {
    const view = render(<Inspector plant={parsePlant(domain, composition)} selected="CH-1" />);
    expect(screen.getByRole("heading", { name: "CH-1" })).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Channels" })).getByText("chiller.power · kW"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Links out" })).getByText("heat_out → CT-1.heat_in"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Links in" })).getByText("CT-1.basin_out → cond_in"),
    ).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Faults" })).getByText("chiller__fouling")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says None for an instance whose kind has nothing to list", () => {
    render(<Inspector plant={parsePlant(domain)} selected="MB-1" />);
    expect(within(screen.getByRole("region", { name: "Faults" })).getByText("None")).toBeInTheDocument();
  });
});
