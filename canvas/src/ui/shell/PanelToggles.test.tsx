import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { PanelStore } from "../../app/panelStore";
import { PanelToggles } from "./PanelToggles";

function preferences() {
  const values = new Map<string, string>();
  return {
    read: (key: string) => values.get(key) ?? null,
    write: (key: string, value: string) => values.set(key, value),
  };
}

describe("panel toggles", () => {
  afterEach(() => {
    Reflect.deleteProperty(HTMLElement.prototype, "checkVisibility");
  });

  it("hides a panel with its key and moves focus off a control the panel hid", async () => {
    const panels = new PanelStore(preferences());
    render(
      <>
        <h1 tabIndex={-1}>Plant</h1>
        <PanelToggles panels={panels} />
        <button type="button">Inside the outline</button>
      </>,
    );
    const inside = screen.getByRole("button", { name: "Inside the outline" });
    Object.defineProperty(HTMLElement.prototype, "checkVisibility", {
      configurable: true,
      value(this: HTMLElement) {
        return this !== inside;
      },
    });
    inside.focus();
    fireEvent.keyDown(inside, { key: "[" });
    expect(panels.snapshot().left).toBe(false);
    expect(screen.getByRole("button", { name: "Outline panel" })).toHaveAttribute("aria-pressed", "false");
    await act(() => new Promise((resolve) => requestAnimationFrame(resolve)));
    expect(screen.getByRole("heading", { name: "Plant" })).toHaveFocus();
  });

  it("names the right panel after what it shows", () => {
    render(<PanelToggles panels={new PanelStore(preferences())} right="Build" />);
    expect(screen.getByRole("button", { name: "Build panel" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByRole("button", { name: "Details panel" })).not.toBeInTheDocument();
  });
});
