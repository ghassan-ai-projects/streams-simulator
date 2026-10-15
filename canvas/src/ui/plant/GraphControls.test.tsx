import { fireEvent, render, screen } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { describe, expect, it } from "vitest";
import { ZoomBar } from "./GraphControls";

function legend() {
  return screen.queryByRole("group", { name: "Wire colour and pin shape by port type" });
}

describe("graph controls legend", () => {
  it("gives the three icon-only zoom controls a tooltip with the name they carry", () => {
    render(
      <ReactFlowProvider>
        <ZoomBar />
      </ReactFlowProvider>,
    );
    for (const name of ["Zoom out", "Zoom in", "Fit the plant in view"]) {
      const button = screen.getByRole("button", { name });
      expect(button).toHaveAttribute("title", name);
    }
  });

  it("explains the reading marks only in a review", () => {
    const { unmount } = render(
      <ReactFlowProvider>
        <ZoomBar />
      </ReactFlowProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Legend" }));
    expect(screen.queryByText("Not reported by this run")).not.toBeInTheDocument();
    unmount();
    render(
      <ReactFlowProvider>
        <ZoomBar review />
      </ReactFlowProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Legend" }));
    expect(screen.getByText("Not reported by this run")).toBeInTheDocument();
    expect(screen.getByText("No reading yet at this time")).toBeInTheDocument();
  });

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
