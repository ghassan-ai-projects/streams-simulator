import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import {
  announced,
  captureResult,
  differs,
  pumpDraft,
  same,
  savedPump,
  savedPumpPath,
} from "../../../test/captureFixture";
import { bridgeWithSavedPump, cardOf, reviewed } from "../../../test/captureHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure as failed, success } from "../../../test/outcome";

describe("saving what a board announced", () => {
  it("writes the draft to the path the binary named, says where, and moves the focus to opening it", async () => {
    const target = { status: "new", target: "kinds/pump_one.kind.json" };
    const { bridge, nav, view } = await reviewed(captureResult([announced(pumpDraft, { comparison: target })]));
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    const open = await screen.findByRole("button", { name: "Open pump in the device editor" });
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/pump_one.kind.json", etag: "" });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toEqual(pumpDraft);
    expect(cardOf("pump").getByText(/Saved as kinds\/pump_one\.kind\.json/)).toBeInTheDocument();
    await waitFor(() => {
      expect(open).toHaveFocus();
    });
    expect(screen.getByRole("status")).toHaveTextContent("All decided: 1 saved, 0 skipped.");
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(open);
    expect(nav.location.hash).toBe("#/device?device=kinds%2Fpump_one.kind.json");
  });

  it("shows the chip as Saved once the device is saved, and no longer says it differs", async () => {
    await reviewed(captureResult([announced()]));
    expect(cardOf("pump").getByText("New device")).toBeInTheDocument();
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    expect(await cardOf("pump").findByText("Saved")).toBeInTheDocument();
    expect(cardOf("pump").queryByText("New device")).toBeNull();
  });

  it("saves the devices of one announcement one by one, each on its own", async () => {
    const fan = { kind: "fan", metrics: [], commands: [] };
    const { bridge } = await reviewed(captureResult([announced(), announced(fan)]));
    fireEvent.click(cardOf("fan").getByRole("button", { name: "Save as device: fan" }));
    expect(await cardOf("fan").findByText(/Saved as kinds\/fan\.kind\.json/)).toBeInTheDocument();
    expect(cardOf("pump").getByRole("button", { name: "Save as device: pump" })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("The board announced 2 devices. 1 saved, 0 skipped.");
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    expect(await cardOf("pump").findByText(/Saved as kinds\/pump\.kind\.json/)).toBeInTheDocument();
    expect(bridge.writes.map((write) => write.path)).toEqual(["kinds/fan.kind.json", "kinds/pump.kind.json"]);
  });

  it("says the device is being saved while the file is written, and cannot be skipped meanwhile", async () => {
    const { bridge } = await reviewed();
    const gates = bridge.holdWrite();
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    expect(screen.getByRole("status")).toHaveTextContent("Saving the device…");
    expect(cardOf("pump").getByText("Saving…")).toBeInTheDocument();
    const skip = cardOf("pump").getByRole("button", { name: "Skip pump" });
    expect(skip).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(skip);
    await waitFor(() => {
      expect(gates).toHaveLength(1);
    });
    gates[0]?.resolve(success({ path: "kinds/pump.kind.json", etag: "e1" }));
    expect(await screen.findByText(/Saved as kinds\/pump\.kind\.json/)).toBeInTheDocument();
  });

  it("says why a device could not be saved, and lets the author try again", async () => {
    const { bridge } = await reviewed();
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("conflict", 409, "exists")));
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("A device named pump already exists.");
    bridge.writeOutcomes.clear();
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    expect(await screen.findByText(/Saved as kinds\/pump\.kind\.json/)).toBeInTheDocument();
  });

  it("opens the saved device without writing when the board announced what is saved", async () => {
    const { bridge, nav } = await reviewed(captureResult([same()]));
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Open the saved device: pump" }));
    expect(nav.location.hash).toBe("#/device?device=kinds%2Fpump.kind.json");
    expect(bridge.writes).toEqual([]);
  });
});

describe("replacing a saved device", () => {
  const replace = () => cardOf("pump").getByRole("button", { name: "Replace the saved device: pump" });

  it("asks first, with the safe answer focused, and overwrites only when told to", async () => {
    const { bridge, view } = await reviewed(captureResult([differs()]), bridgeWithSavedPump());
    fireEvent.click(replace());
    const question = screen.getByRole("group", { name: "Replace the saved pump?" });
    expect(within(question).getByRole("button", { name: "Keep the saved device" })).toHaveFocus();
    expect(question).toHaveTextContent("kinds/pump.kind.json is overwritten");
    expect(bridge.writes).toEqual([]);
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(within(question).getByRole("button", { name: "Yes, replace it" }));
    expect(await screen.findByText(/Saved as kinds\/pump\.kind\.json/)).toBeInTheDocument();
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/pump.kind.json", etag: "k3" });
  });

  it("keeps the saved device on Keep, and on Escape, and returns the focus to the button", async () => {
    const { bridge, desk } = await reviewed(captureResult([differs()]), bridgeWithSavedPump());
    fireEvent.click(replace());
    fireEvent.click(screen.getByRole("button", { name: "Keep the saved device" }));
    expect(screen.queryByRole("group", { name: "Replace the saved pump?" })).toBeNull();
    expect(replace()).toHaveFocus();
    fireEvent.click(replace());
    fireEvent.keyDown(screen.getByRole("group", { name: "Replace the saved pump?" }), { key: "Escape" });
    expect(screen.queryByRole("group", { name: "Replace the saved pump?" })).toBeNull();
    expect(desk.capture().comingBack()).toBe(false);
    expect(replace()).toHaveFocus();
    expect(bridge.writes).toEqual([]);
  });

  it("fails for real, and says so, when the saved file changed after the comparison", async () => {
    const bridge = bridgeWithSavedPump();
    bridge.persistWrites();
    await reviewed(captureResult([differs()]), bridge);
    bridge.addFile(savedPumpPath, { ...savedPump, changed: true }, "k4");
    fireEvent.click(replace());
    fireEvent.click(screen.getByRole("button", { name: "Yes, replace it" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("changed since it was compared");
    expect(bridge.files.get(savedPumpPath)).toMatchObject({ ok: true, value: { etag: "k4" } });
    expect(bridge.writes.map((write) => write.etag)).toEqual(["k3"]);
    expect(replace()).toHaveAttribute("aria-disabled", "false");
  });
});
