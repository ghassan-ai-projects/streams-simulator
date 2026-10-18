import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { announced, captureResult, differs, pumpDraft, same, unreadable } from "../../../test/captureFixture";
import { bridgeWithSavedPump, cardOf, reviewed } from "../../../test/captureHarness";

describe("a device the board announced beside the saved ones", () => {
  it("offers the saved device when the board announced what is already saved, and nothing to skip", async () => {
    await reviewed(captureResult([same()]));
    const card = cardOf("pump");
    expect(card.getByText("Same as pump.kind.json")).toBeInTheDocument();
    expect(card.getByText("Saved file")).toBeInTheDocument();
    expect(card.getByRole("button", { name: "Open the saved device: pump" })).toBeInTheDocument();
    expect(card.queryByRole("button", { name: /^Save as device/ })).toBeNull();
    expect(card.queryByRole("button", { name: /^Replace the saved device/ })).toBeNull();
    expect(card.queryByRole("button", { name: /^Skip/ })).toBeNull();
    expect(screen.getByRole("status")).toHaveTextContent("The board announced 1 device.");
  });

  it("says where a device differs from the saved one in words, with both values, and keeps the pointer in the fingerprint", async () => {
    await reviewed(captureResult([differs()]), bridgeWithSavedPump());
    const card = cardOf("pump");
    expect(card.getByText("Differs from pump.kind.json")).toBeInTheDocument();
    expect(card.getByText(/The first difference is in Reading flow · Unit\./)).toBeInTheDocument();
    expect(card.getByText(/Saved: "m3\/s"\. The board says: "m3\/h"\./)).toBeInTheDocument();
    expect(card.getByText("First difference at")).toBeInTheDocument();
    expect(card.getByText("/metrics/0/unit")).toBeInTheDocument();
    expect(card.queryByRole("button", { name: /^Save as device/ })).toBeNull();
  });

  it("says the whole file differs when the binary names no place", async () => {
    await reviewed(captureResult([differs("")]), bridgeWithSavedPump());
    expect(cardOf("pump").getByText("The whole file")).toBeInTheDocument();
    expect(cardOf("pump").getByText(/somewhere in the file/)).toBeInTheDocument();
  });

  it("says the saved device looks like the one announced when only the capital letters differ", async () => {
    const colliding = differs("/kind", {
      target: "kinds/Pump.kind.json",
      file: "kinds/Pump.kind.json",
      collides_with: "kinds/Pump.kind.json",
    });
    const bridge = bridgeWithSavedPump();
    bridge.addFile("kinds/Pump.kind.json", { kind: "Pump" }, "k5");
    await reviewed(captureResult([colliding]), bridge);
    const card = cardOf("pump");
    expect(card.getByText("Looks like Pump.kind.json, which differs only by capital letters.")).toBeInTheDocument();
    expect(card.getByText(/Saved: "Pump"\. The board says: "pump"\./)).toBeInTheDocument();
  });

  it("says the saved file cannot be read when the comparison could not be made, and offers only to skip", async () => {
    await reviewed(captureResult([unreadable()]));
    const card = cardOf("pump");
    expect(card.getByText("Cannot compare")).toBeInTheDocument();
    expect(card.getByText(/The saved file pump\.kind\.json cannot be read/)).toBeInTheDocument();
    expect(card.queryByRole("button", { name: /^Save as device|^Replace|^Open the saved/ })).toBeNull();
    expect(card.getByRole("button", { name: "Skip pump" })).toBeInTheDocument();
  });

  it("does not offer to replace a file it could not read, and says why", async () => {
    await reviewed(captureResult([differs()]));
    const card = cardOf("pump");
    expect(card.getByRole("alert")).toHaveTextContent("The saved device could not be read");
    expect(card.getByRole("button", { name: "Replace the saved device: pump" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
  });

  it("keeps a draft whose saved file name is not the device's, and shows the file it would replace", async () => {
    const other = {
      status: "differs",
      target: "kinds/old.kind.json",
      file: "kinds/old.kind.json",
      file_digest: "sha256:x",
      differs_at: "/kind",
    };
    const bridge = bridgeWithSavedPump();
    bridge.addFile("kinds/old.kind.json", { ...pumpDraft, kind: "old" }, "k8");
    await reviewed(captureResult([announced(pumpDraft, { comparison: other })]), bridge);
    expect(cardOf("pump").getByText("Differs from old.kind.json")).toBeInTheDocument();
  });
});
