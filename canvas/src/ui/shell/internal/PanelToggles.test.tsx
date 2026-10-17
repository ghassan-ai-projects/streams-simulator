import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { stubCheckVisibility } from "../../../test/dom";
import { createPanelStore } from "../../../app/shell";
import { PanelToggles } from "./PanelToggles";

function preferences() {
  const values = new Map<string, string>();
  return {
    read: (key: string) => values.get(key) ?? null,
    write: (key: string, value: string) => values.set(key, value),
  };
}

describe("panel toggles", () => {
  it("hides a panel with its key and moves focus off a control the panel hid", async () => {
    const panels = createPanelStore(preferences());
    render(
      <>
        <h1 tabIndex={-1}>Plant</h1>
        <PanelToggles panels={panels} right="Details" />
        <button type="button">Inside the outline</button>
      </>,
    );
    const inside = screen.getByRole("button", { name: "Inside the outline" });
    stubCheckVisibility((node) => node !== inside);
    inside.focus();
    fireEvent.keyDown(inside, { key: "[" });
    expect(panels.snapshot().left).toBe(false);
    expect(screen.getByRole("button", { name: "Outline panel" })).toHaveAttribute("aria-pressed", "false");
    await act(() => new Promise((resolve) => requestAnimationFrame(resolve)));
    expect(screen.getByRole("heading", { name: "Plant" })).toHaveFocus();
  });

  it("hands the right toggle to its own press when one is given, and leaves the panel to it", () => {
    const panels = createPanelStore(preferences());
    const press = vi.fn();
    render(<PanelToggles panels={panels} right="Build" onRightPress={press} />);
    fireEvent.click(screen.getByRole("button", { name: "Build panel" }));
    expect(press).toHaveBeenCalledTimes(1);
    expect(panels.snapshot().right).toBe(true);
  });

  it("names the right panel after what it shows", () => {
    render(<PanelToggles panels={createPanelStore(preferences())} right="Build" />);
    expect(screen.getByRole("button", { name: "Build panel" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByRole("button", { name: "Details panel" })).not.toBeInTheDocument();
  });
});
