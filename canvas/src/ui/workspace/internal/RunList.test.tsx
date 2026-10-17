import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { success } from "../../../test/outcome";
import { FakeBridge, sampleWorkspace } from "../../../test/fakeBridge";
import { show, navigation } from "../../../test/homeHarness";

describe("runs on home", () => {
  it("reads a run that did not finish as Did not finish, with its reason, and never as checked", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({
      ...sampleWorkspace,
      runs: [
        {
          worldId: "w-6",
          ended: true,
          complete: false,
          incompleteReason: "the world was destroyed before its run ended",
          explore: false,
          seed: 3,
          plant: "factory-utilities",
          durationS: 600,
          recordedAt: "2026-10-14T10:00:00Z",
        },
      ],
    });
    await show(bridge);
    const row = within(screen.getByRole("region", { name: "Runs" })).getByRole("listitem");
    expect(within(row).getByText("Did not finish")).toHaveAttribute(
      "title",
      "the world was destroyed before its run ended",
    );
    expect(row).toHaveTextContent(
      /w-6 · 14 Oct 2026, \d\d:00 · 10 min · seed 3 · Did not finish: the world was destroyed before its run ended/,
    );
    expect(row).not.toHaveTextContent(/Verified|Not checked/);
  });

  it("gives a run with no recorded end its own chip, not the success tokens Verified uses", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({
      ...sampleWorkspace,
      runs: [
        {
          worldId: "w-open",
          ended: false,
          complete: false,
          incompleteReason: "",
          explore: false,
          seed: 7,
          plant: "factory-utilities",
          durationS: 1800,
          recordedAt: "2026-10-15T10:00:00Z",
        },
        {
          worldId: "w-good",
          ended: true,
          complete: true,
          incompleteReason: "",
          explore: false,
          seed: 7,
          plant: "factory-utilities",
          durationS: 1800,
          recordedAt: "2026-10-15T09:00:00Z",
        },
      ],
    });
    await show(bridge, navigation(), (worldId) => (worldId === "w-good" ? "verified" : null));
    const unfinished = screen.getByText("Did not finish");
    expect(unfinished).toHaveClass(/unfinished/);
    expect(unfinished).not.toHaveClass(/live|verified/);
    const verified = screen.getByText("Verified");
    expect(verified).toHaveClass(/verified/);
    expect(verified).not.toHaveClass(/unfinished/);
  });

  it("names the one run left to show in the singular", async () => {
    const bridge = new FakeBridge();
    const runs = Array.from({ length: 21 }, (_, i) => ({
      worldId: `w-${String(i)}`,
      ended: true,
      complete: true,
      incompleteReason: "",
      explore: false,
      seed: null,
      plant: "factory-utilities",
      durationS: 60,
      recordedAt: `2026-10-14T10:${String(i).padStart(2, "0")}:00Z`,
    }));
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs });
    await show(bridge);
    expect(screen.getByRole("button", { name: "Show 1 more run" })).toBeInTheDocument();
  });
});
