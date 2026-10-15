import { fireEvent, render, screen } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { ZoomBar } from "./GraphControls";

function legend() {
  return screen.queryByRole("group", { name: "Wire colour and pin shape by port type" });
}

describe("graph controls legend", () => {
  it("closes on Escape wherever focus is, and on a press outside that still reaches its target", () => {
    render(
      <ReactFlowProvider>
        <ZoomBar />
        <button type="button">Elsewhere</button>
      </ReactFlowProvider>,
    );
    const button = screen.getByRole("button", { name: "Legend" });
    fireEvent.click(button);
    expect(legend()).toBeInTheDocument();
    fireEvent.keyDown(document.body, { key: "Escape" });
    expect(legend()).not.toBeInTheDocument();
    expect(button).toHaveFocus();
    fireEvent.click(button);
    fireEvent.pointerDown(screen.getByRole("button", { name: "Elsewhere" }));
    expect(legend()).not.toBeInTheDocument();
    expect(button).toHaveAttribute("aria-expanded", "false");
  });
});
