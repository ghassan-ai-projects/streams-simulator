import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { announced, captureResult, differs, pumpDraft, same } from "../../../test/captureFixture";
import { cardOf, reviewed } from "../../../test/captureHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure as failed, success } from "../../../test/outcome";

describe("saving what a board announced", () => {
  it("writes the draft as a new device file, says where, and moves the focus to opening it", async () => {
    const { bridge, nav, view } = await reviewed();
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Save as device: pump" }));
    const open = await screen.findByRole("button", { name: "Open pump in the device editor" });
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/pump.kind.json", etag: "" });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toEqual(pumpDraft);
    expect(cardOf("pump").getByText(/Saved as kinds\/pump\.kind\.json/)).toBeInTheDocument();
    await waitFor(() => {
      expect(open).toHaveFocus();
    });
    expect(screen.getByRole("status")).toHaveTextContent("All decided: 1 saved, 0 skipped.");
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(open);
    expect(nav.location.hash).toBe("#/device?device=kinds%2Fpump.kind.json");
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

  it("does not write a device whose name cannot be a file name", async () => {
    const { bridge } = await reviewed(captureResult([announced({ kind: "a/b" })]));
    const save = cardOf("a/b").getByRole("button", { name: "Save as device: a/b" });
    expect(save).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(save);
    expect(bridge.writes).toEqual([]);
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

  it("skips a device without writing anything", async () => {
    const { bridge } = await reviewed();
    fireEvent.click(cardOf("pump").getByRole("button", { name: "Skip pump" }));
    expect(screen.getByText("Skipped. Nothing was saved for pump.")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("All decided: 0 saved, 1 skipped.");
    expect(bridge.writes).toEqual([]);
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
    const { bridge, view } = await reviewed(captureResult([differs()]));
    bridge.addFile("kinds/pump.kind.json", { kind: "pump" }, "k3");
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
    const { bridge } = await reviewed(captureResult([differs()]));
    fireEvent.click(replace());
    fireEvent.click(screen.getByRole("button", { name: "Keep the saved device" }));
    expect(screen.queryByRole("group", { name: "Replace the saved pump?" })).toBeNull();
    expect(replace()).toHaveFocus();
    fireEvent.click(replace());
    fireEvent.keyDown(screen.getByRole("group", { name: "Replace the saved pump?" }), { key: "Escape" });
    expect(screen.queryByRole("group", { name: "Replace the saved pump?" })).toBeNull();
    expect(screen.getByRole("heading", { level: 1, name: "Capture from a board" })).toBeInTheDocument();
    expect(replace()).toHaveFocus();
    expect(bridge.writes).toEqual([]);
  });

  it("says the saved device changed when the file moved on, and writes nothing over it", async () => {
    const { bridge } = await reviewed(captureResult([differs()]));
    bridge.addFile("kinds/pump.kind.json", { kind: "pump" }, "k3");
    bridge.writeOutcomes.set("kinds/pump.kind.json", failed(cannedFailure("conflict", 409, "changed")));
    fireEvent.click(replace());
    fireEvent.click(screen.getByRole("button", { name: "Yes, replace it" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("changed since it was compared");
    expect(replace()).toHaveAttribute("aria-disabled", "false");
  });
});
