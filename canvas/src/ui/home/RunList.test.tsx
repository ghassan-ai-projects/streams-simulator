import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { success } from "../../model/bridge";
import { FakeBridge, sampleWorkspace } from "../../test/fakeBridge";
import { show } from "../../test/homeHarness";

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
});
