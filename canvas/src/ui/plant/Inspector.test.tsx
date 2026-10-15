import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../model/plant";
import type { ComponentSpec } from "../../model/ports";
import { campus, composition, domain } from "../../test/plantFixture";
import { Inspector } from "./Inspector";

describe("inspector", () => {
  it("asks for a selection when nothing is selected", () => {
    render(<Inspector plant={parsePlant(domain)} selected={null} />);
    expect(screen.getByRole("complementary", { name: "Inspector" })).toHaveTextContent("Select an instance");
  });

  it("shows the selected instance's kind and links with no accessibility violation", async () => {
    const view = render(<Inspector plant={parsePlant(domain, composition)} selected="CH-1" />);
    expect(screen.getByRole("heading", { name: "CH-1" })).toBeInTheDocument();
    expect(screen.getByText("chiller · Cooling")).toBeInTheDocument();
    expect(screen.getByText("lib/chiller")).toBeInTheDocument();
    expect(screen.queryByText("What this component models")).not.toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Channels" })).getByText("power · kW")).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Links out" })).getByText("heat out → CT-1 heat in"),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "Links in" })).getByText("CT-1 basin out → cond in"),
    ).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Faults" })).getByText("fouling")).toHaveAttribute(
      "title",
      "chiller__fouling",
    );
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("shows a declared component title and purpose while keeping the raw id as the only repeat", async () => {
    const spec: ComponentSpec = {
      category: "cool",
      ports: [],
      title: "Chiller",
      description: "Moves heat from chilled water to condenser water.",
    };
    const view = render(
      <Inspector plant={parsePlant(domain, composition)} selected="CH-1" specs={new Map([["lib/chiller", spec]])} />,
    );
    expect(screen.getByText("Chiller · Cooling")).toBeInTheDocument();
    expect(screen.getByText("lib/chiller")).toBeInTheDocument();
    expect(screen.queryByText(/Kind: /)).not.toBeInTheDocument();
    fireEvent.click(screen.getByText("What this component models"));
    expect(screen.getByText("Moves heat from chilled water to condenser water.")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says None for an instance whose kind has nothing to list", () => {
    render(<Inspector plant={parsePlant(domain)} selected="MB-1" />);
    expect(within(screen.getByRole("region", { name: "Faults" })).getByText("None")).toBeInTheDocument();
  });

  it("lists cross-hall receivers in order and lets the operator select one", () => {
    const onSelect = vi.fn();
    render(<Inspector plant={parsePlant(campus(3))} selected="CAL-1" onSelect={onSelect} />);
    const rows = screen.getByRole("region", { name: "Links out" });
    const hall = within(rows).getByRole("region", { name: "h02 linked instances" });
    expect(
      within(hall)
        .getAllByRole("button")
        .map((button) => button.textContent),
    ).toEqual([
      "CAL-1 · port not recorded → h02/M-1 · port not recorded",
      "CAL-1 · port not recorded → h02/M-2 · port not recorded",
    ]);
    fireEvent.click(within(hall).getByRole("button", { name: /Select receiver h02\/M-1/ }));
    expect(onSelect).toHaveBeenCalledWith("h02/M-1");
  });
});
