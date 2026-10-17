import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { Route } from "../../../model/shell";
import { PanelStore } from "../../../app/shell";
import { KeyHelp } from "./KeyHelp";
import { PanelToggles } from "./PanelToggles";
import { rightPanelName } from "./panelName";

describe("key help", () => {
  it("opens from a visible Keyboard button, grouped by where each key works, and returns focus on Escape", () => {
    render(<KeyHelp screen="run" />);
    const button = screen.getByRole("button", { name: "Keyboard" });
    button.focus();
    fireEvent.click(button);
    const dialog = screen.getByRole("dialog", { name: "Keyboard" });
    expect(within(dialog).getByRole("heading", { name: "Keyboard" })).toHaveFocus();
    for (const scope of ["Everywhere", "Run review"]) {
      expect(within(dialog).getByRole("region", { name: scope })).toBeInTheDocument();
    }
    const review = within(dialog).getByRole("region", { name: "Run review" });
    expect(within(review).getByText("go to the previous event, the next event")).toBeInTheDocument();
    expect(within(review).getByText("play or pause")).toBeInTheDocument();
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

  it("gives the right panel one name per screen, in the help and on the toggle", () => {
    for (const [where, name] of Object.entries(rightPanelName) as readonly (readonly [Route["screen"], string])[]) {
      const toggle = render(
        <PanelToggles panels={new PanelStore({ read: () => null, write: () => undefined })} right={name} />,
      );
      const button = screen.getByRole("button", { name: `${name} panel` });
      expect(button).toHaveAttribute("title", `Hide the ${name.toLowerCase()} panel (])`);
      toggle.unmount();

      const view = render(<KeyHelp screen={where} />);
      fireEvent.click(screen.getByRole("button", { name: "Keyboard" }));
      expect(screen.getByText(`hide or show the outline panel, the ${name.toLowerCase()} panel`)).toBeInTheDocument();
      expect(screen.queryByText(/the details panel/)).toBeNull();
      view.unmount();
    }
  });

  it("dims the page and closes when the dimmed page is clicked", () => {
    render(<KeyHelp screen="plant" />);
    fireEvent.keyDown(window, { key: "?" });
    const scrim = document.querySelector("[aria-hidden='true'][class*='scrim']");
    expect(scrim).not.toBeNull();
    if (scrim !== null) fireEvent.click(scrim);
    expect(screen.queryByRole("dialog")).toBeNull();
  });
});
