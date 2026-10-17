import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { Link } from "../../../model/plant";
import { InspectorLinks } from "./InspectorLinks";

const links: readonly Link[] = [
  { from: "CAL-1", to: "h10/M-1", type: "line", fromPort: "out", toPort: "in" },
  { from: "CAL-1", to: "h2/M-10", type: "line", fromPort: "zeta_out", toPort: "shared_in" },
  { from: "CAL-1", to: "h2/M-2", type: "line", fromPort: "out", toPort: "in" },
  { from: "CAL-1", to: "h2/M-2", type: "line", fromPort: "alpha_out", toPort: "in" },
];

describe("inspector link destinations", () => {
  it("orders hall and receiver numerically and preserves both declared ports", () => {
    const onSelect = vi.fn();
    render(<InspectorLinks links={links} side="out" onSelect={onSelect} />);
    const hall = screen.getByRole("region", { name: "h2 linked instances" });
    expect(
      within(hall)
        .getAllByRole("button")
        .map((button) => button.textContent),
    ).toEqual(["CAL-1 alpha out → h2/M-2 in", "CAL-1 out → h2/M-2 in", "CAL-1 zeta out → h2/M-10 shared in"]);
    fireEvent.click(within(hall).getByRole("button", { name: /Select receiver h2\/M-10/ }));
    expect(onSelect).toHaveBeenCalledWith("h2/M-10");
  });

  it("names no hall for an instance outside every hall, and says None when nothing is linked", () => {
    const onSelect = vi.fn();
    const { rerender } = render(
      <InspectorLinks
        links={[{ from: "CAL-1", to: "CMP-2", type: "line", fromPort: "out", toPort: "in" }]}
        side="out"
        onSelect={onSelect}
      />,
    );
    expect(screen.queryByText("Site")).toBeNull();
    expect(screen.queryByRole("region")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: /Select receiver CMP-2/ }));
    expect(onSelect).toHaveBeenCalledWith("CMP-2");
    rerender(<InspectorLinks links={[]} side="out" onSelect={onSelect} />);
    expect(screen.getByText("None")).toBeInTheDocument();
  });
});
