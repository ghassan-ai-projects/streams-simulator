import { fireEvent, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { announced, captureResult } from "../../../test/captureFixture";
import { cardOf, heading, reviewed } from "../../../test/captureHarness";

const fan = { kind: "fan", metrics: [], commands: [] };
const valve = { kind: "valve", metrics: [], commands: [] };

async function three() {
  return reviewed(captureResult([announced(), announced(fan), announced(valve)]));
}

describe("skipping a device", () => {
  it("writes nothing, says so, and moves the focus to the next device's first action", async () => {
    const { bridge } = await three();
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Skip pump" }));
    expect(screen.getByText(/Skipped\. Nothing will be saved for pump\./)).toBeInTheDocument();
    expect(cardOf("fan").getByRole("button", { name: "Save as device: fan" })).toHaveFocus();
    fireEvent.click(cardOf("fan").getByRole("button", { name: "Skip fan" }));
    expect(cardOf("valve").getByRole("button", { name: "Save as device: valve" })).toHaveFocus();
    expect(screen.getByRole("status")).toHaveTextContent("The board announced 3 devices. 0 saved, 2 skipped.");
    expect(bridge.writes).toEqual([]);
  });

  it("moves the focus to the heading of the review when no device after it is left", async () => {
    await three();
    fireEvent.click(cardOf("valve").getByRole("button", { name: "Skip valve" }));
    expect(screen.getByRole("heading", { level: 2, name: heading })).toHaveFocus();
  });

  it("takes a skip back with Undo, and puts the focus on that device's first action", async () => {
    await three();
    fireEvent.click(cardOf("fan").getByRole("button", { name: "Skip fan" }));
    fireEvent.click(cardOf("fan").getByRole("button", { name: "Undo skipping fan" }));
    expect(screen.queryByText(/Skipped\./)).toBeNull();
    expect(cardOf("fan").getByRole("button", { name: "Save as device: fan" })).toHaveFocus();
  });

  it("is done when every device is saved or skipped", async () => {
    await reviewed(captureResult([announced()]));
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Skip pump" }));
    expect(screen.getByRole("status")).toHaveTextContent("All decided: 0 saved, 1 skipped.");
  });
});
