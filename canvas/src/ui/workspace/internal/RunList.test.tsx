import { screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { success } from "../../../test/outcome";
import { FakeBridge } from "../../../test/fakeBridge";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { show, navigation } from "../../../test/workspaceHarness";

function rowOf(worldId: string): HTMLElement {
  const row = screen.getByRole("button", { name: `Review ${worldId}` }).closest("li");
  if (row === null) throw new Error(`no row for ${worldId}`);
  return row;
}

describe("runs in the workspace", () => {
  it("labels a connected session as one, running while it runs, and never as Evaluate", async () => {
    const bridge = new FakeBridge();
    const session = {
      worldId: "c-1",
      ended: false,
      complete: false,
      running: true,
      incompleteReason: "",
      explore: false,
      seed: 1,
      plant: "factory-utilities",
      durationS: 0,
      recordedAt: "2026-10-14T10:00:00Z",
      faults: 0,
    };
    bridge.workspaceOutcome = success({
      ...sampleWorkspace,
      runs: [
        { ...session, connected: "running" as const },
        { ...session, worldId: "c-2", ended: true, complete: true, running: false, connected: "saved" as const },
      ],
    });
    await show(bridge);
    expect(within(rowOf("c-1")).getByText("Connected session · running")).toBeInTheDocument();
    expect(within(rowOf("c-1")).getByText("Running")).toBeInTheDocument();
    expect(within(rowOf("c-2")).getByText("Connected session")).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Runs" })).not.toHaveTextContent(/Evaluate|Did not finish/);
  });

  it("reads a run that did not finish as Did not finish, with its reason, and never as checked", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({
      ...sampleWorkspace,
      runs: [
        {
          worldId: "w-6",
          ended: true,
          complete: false,
          running: false,
          incompleteReason: "the world was destroyed before its run ended",
          explore: false,
          seed: 3,
          plant: "factory-utilities",
          durationS: 600,
          recordedAt: "2026-10-14T10:00:00Z",
          faults: 0,
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
          running: false,
          incompleteReason: "",
          explore: false,
          seed: 7,
          plant: "factory-utilities",
          durationS: 1800,
          recordedAt: "2026-10-15T10:00:00Z",
          faults: 0,
        },
        {
          worldId: "w-good",
          ended: true,
          complete: true,
          running: false,
          incompleteReason: "",
          explore: false,
          seed: 7,
          plant: "factory-utilities",
          durationS: 1800,
          recordedAt: "2026-10-15T09:00:00Z",
          faults: 0,
        },
      ],
    });
    await show(bridge, navigation(), (worldId) => (worldId === "w-good" ? "verified" : null));
    const open = rowOf("w-open");
    const good = rowOf("w-good");
    expect(within(open).getByText("Did not finish")).toBeInTheDocument();
    expect(within(open).queryByText("Verified")).not.toBeInTheDocument();
    expect(within(good).getByText("Verified")).toBeInTheDocument();
    expect(within(good).queryByText("Did not finish")).not.toBeInTheDocument();
  });

  it("reads an unfinished run that is running as Running and never as Did not finish", async () => {
    const bridge = new FakeBridge();
    const open = {
      worldId: "w-live",
      ended: false,
      complete: false,
      running: true,
      incompleteReason: "",
      explore: false,
      seed: 7,
      plant: "factory-utilities",
      durationS: 1800,
      recordedAt: "2026-10-15T10:00:00Z",
      faults: 0,
    };
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs: [open] });
    await show(bridge);
    const live = rowOf("w-live");
    expect(within(live).getByText("Running")).toBeInTheDocument();
    expect(live).not.toHaveTextContent(/Did not finish/);
    expect(live).not.toHaveTextContent(/Connected session/);
  });

  it("says a stopped run was stopped, with the length it reached out of the length it was asked for", async () => {
    const bridge = new FakeBridge();
    const stopped = {
      worldId: "w-cut",
      ended: true,
      complete: true,
      running: false,
      incompleteReason: "",
      explore: false,
      seed: 7,
      plant: "factory-utilities",
      durationS: 540,
      requestedS: 1800,
      terminationReason: "stopped_by_operator",
      recordedAt: "2026-10-15T10:00:00Z",
      faults: 1,
    };
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs: [stopped] });
    await show(bridge);
    expect(rowOf("w-cut")).toHaveTextContent("9 of 30 min · stopped · seed 7 · 1 fault");
    expect(rowOf("w-cut")).not.toHaveTextContent(/9 min/);
  });

  it("names the one run left to show in the singular", async () => {
    const bridge = new FakeBridge();
    const runs = Array.from({ length: 21 }, (_, i) => ({
      worldId: `w-${String(i)}`,
      ended: true,
      complete: true,
      running: false,
      incompleteReason: "",
      explore: false,
      seed: null,
      plant: "factory-utilities",
      durationS: 60,
      recordedAt: `2026-10-14T10:${String(i).padStart(2, "0")}:00Z`,
      faults: 0,
    }));
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs });
    await show(bridge);
    expect(screen.getByRole("button", { name: "Show 1 more run" })).toBeInTheDocument();
  });
});
