import { act, fireEvent, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { serveCapture } from "../../../test/captureFixture";
import { FakeBridge } from "../../../test/fakeBridge";
import { success } from "../../../test/outcome";

describe("the capture route", () => {
  it("shows the capture screen at its own address, and keeps the review when the author comes back from the editor", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    bridge.persistWrites();
    serveCapture(bridge);
    const { location } = mountApp(bridge, "#/capture");
    expect(await screen.findByRole("heading", { level: 1, name: "Capture from a board" })).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Announcement text"), { target: { value: "{}" } });
    fireEvent.click(screen.getByRole("button", { name: "Read the announcement" }));
    fireEvent.click(await screen.findByRole("button", { name: "Save as device: pump" }));
    fireEvent.click(await screen.findByRole("button", { name: "Open pump in the device editor" }));
    expect(location.hash).toBe("#/device?device=kinds%2Fpump.kind.json");
    expect(screen.queryByRole("heading", { level: 2, name: "Review what the board announced" })).toBeNull();
    act(() => {
      location.hash = "#/capture";
      location.fire();
    });
    expect(
      await screen.findByRole("heading", { level: 2, name: "Review what the board announced" }),
    ).toBeInTheDocument();
    expect(screen.getByText(/Saved as kinds\/pump\.kind\.json/)).toBeInTheDocument();
  });

  it("goes back to the workspace from the top bar", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    const { location } = mountApp(bridge, "#/capture");
    await screen.findByRole("heading", { level: 1, name: "Capture from a board" });
    fireEvent.click(screen.getByRole("button", { name: "Streams Simulator, workspace home" }));
    expect(location.hash).toBe("#/");
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
  });
});
