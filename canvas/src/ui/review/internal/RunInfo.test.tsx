import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { RunInfo } from "./RunInfo";

describe("run info", () => {
  it("opens on click, closes on Escape with focus back on its button, and closes when focus leaves", () => {
    render(
      <>
        <RunInfo lines={["Explore: the truth is visible.", "factory-utilities · seed 7"]} />
        <button type="button">Elsewhere</button>
      </>,
    );
    const button = screen.getByRole("button", { name: "About this run" });
    expect(screen.queryByText("factory-utilities · seed 7")).not.toBeVisible();
    fireEvent.click(button);
    expect(button).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByText("factory-utilities · seed 7")).toBeVisible();
    fireEvent.keyDown(button, { key: "Escape" });
    expect(button).toHaveAttribute("aria-expanded", "false");
    expect(button).toHaveFocus();
    fireEvent.click(button);
    fireEvent.blur(button, { relatedTarget: screen.getByRole("button", { name: "Elsewhere" }) });
    expect(button).toHaveAttribute("aria-expanded", "false");
  });
});
