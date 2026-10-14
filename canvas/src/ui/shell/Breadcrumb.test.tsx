import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Breadcrumb, CrumbSlot } from "./Breadcrumb";

describe("breadcrumb", () => {
  it("renders in place when there is no top bar slot, with the current page as the heading", () => {
    const go = vi.fn();
    render(<Breadcrumb trail={[{ label: "Workspace", onGo: go }]} here="Factory" heading={null} tag="Editing" />);
    const nav = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(within(nav).getByRole("heading", { level: 1, name: "Factory" })).toHaveAttribute("aria-current", "location");
    expect(within(nav).getByText("Editing")).toBeInTheDocument();
    fireEvent.click(within(nav).getByRole("button", { name: "Workspace" }));
    expect(go).toHaveBeenCalledOnce();
  });

  it("renders into the top bar slot when one is provided", () => {
    const slot = document.createElement("div");
    document.body.append(slot);
    render(
      <CrumbSlot value={slot}>
        <section aria-label="Screen">
          <Breadcrumb trail={[]} here="Workspace" heading={null} />
        </section>
      </CrumbSlot>,
    );
    expect(within(slot).getByRole("heading", { level: 1, name: "Workspace" })).toBeInTheDocument();
    expect(within(screen.getByRole("region", { name: "Screen" })).queryByRole("heading")).toBeNull();
    slot.remove();
  });
});
