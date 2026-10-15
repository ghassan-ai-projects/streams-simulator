import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { KeyHelp } from "./KeyHelp";

describe("key help", () => {
  it("opens from a visible Keyboard button, grouped by where each key works, and returns focus on Escape", () => {
    render(<KeyHelp />);
    const button = screen.getByRole("button", { name: "Keyboard" });
    button.focus();
    fireEvent.click(button);
    const dialog = screen.getByRole("dialog", { name: "Keyboard" });
    expect(within(dialog).getByRole("heading", { name: "Keyboard" })).toHaveFocus();
    for (const scope of ["Everywhere", "Plant graph", "Run review", "Editor"]) {
      expect(within(dialog).getByRole("region", { name: scope })).toBeInTheDocument();
    }
    expect(within(dialog).queryByText(/\(graph\)|Previous and next event/)).toBeNull();
    fireEvent.keyDown(window, { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(button).toHaveFocus();
  });

  it("lists only the keys that work on the screen it is opened from", () => {
    render(<KeyHelp screen="plant" />);
    fireEvent.click(screen.getByRole("button", { name: "Keyboard" }));
    const dialog = screen.getByRole("dialog", { name: "Keyboard" });
    expect(within(dialog).getByRole("region", { name: "Plant graph" })).toBeInTheDocument();
    expect(within(dialog).queryByRole("region", { name: "Run review" })).toBeNull();
    expect(within(dialog).queryByRole("region", { name: "Editor" })).toBeNull();
  });

  it("dims the page and closes when the dimmed page is clicked", () => {
    render(<KeyHelp />);
    fireEvent.keyDown(window, { key: "?" });
    const scrim = document.querySelector("[aria-hidden='true'][class*='scrim']");
    expect(scrim).not.toBeNull();
    if (scrim !== null) fireEvent.click(scrim);
    expect(screen.queryByRole("dialog")).toBeNull();
  });
});
