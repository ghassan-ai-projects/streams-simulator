import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TopBar } from "./TopBar";

describe("top bar", () => {
  it("offers a skip past itself to the screen", () => {
    render(
      <>
        <TopBar crumbs={vi.fn()} help={null} preference="system" onCycleTheme={vi.fn()} onWorkspace={vi.fn()} />
        <main id="main" tabIndex={-1} aria-label="Workspace" />
      </>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Skip to the screen" }));
    expect(screen.getByRole("main", { name: "Workspace" })).toHaveFocus();
  });

  it("says on screen that the button changes the theme and shows its current value", () => {
    render(<TopBar crumbs={vi.fn()} help={null} preference="dark" onCycleTheme={vi.fn()} onWorkspace={vi.fn()} />);
    const button = screen.getByRole("button", { name: "Theme: dark" });
    expect(button).toHaveTextContent("Theme: Dark");
  });
});
