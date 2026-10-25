import { screen } from "@testing-library/react";
import { mountApp } from "./appHarness";
import { comparisonWire } from "./compareFixture";
import type { RunEntry } from "../model/workspace";
import { success } from "./outcome";
import { reportWire } from "./runFixture";
import { miniWorkspace, reviewBridge, reviewHash } from "./screenFixture";

const replayQuestion = { code: "replay_not_checked", message: "The recording was not replayed." };
const stopped = {
  code: "stopped_early",
  message: "The operator stopped the run at 00:09:00 of the requested 00:30:00.",
};

export function run(worldId: string, extra: Partial<RunEntry> = {}): RunEntry {
  return {
    worldId,
    ended: true,
    complete: true,
    running: false,
    incompleteReason: "",
    explore: false,
    seed: 7,
    plant: "mini",
    durationS: 1800,
    recordedAt: "2026-10-14T10:00:00Z",
    faults: 0,
    ...extra,
  };
}

export function bridgeWith(openQuestions: unknown[] = [replayQuestion, stopped], runId = "r-1") {
  const bridge = reviewBridge();
  bridge.workspaceOutcome = success({
    ...miniWorkspace,
    runs: [
      run("w-1"),
      run("w-3", { name: "Baseline", explore: true, seed: 11 }),
      run("w-2", { ended: false, complete: false }),
      run("w-8", { plant: "other" }),
    ],
  });
  const asked: Record<string, unknown>[] = [];
  bridge.handlers.set("run.report", (args) => {
    asked.push({ ...args });
    if (args.format === "markdown")
      return success({ format: "markdown", markdown: "# Run w-1\n\n## Open questions\n" });
    const checked = args.replay === true;
    return success({
      ...reportWire,
      run: { ...reportWire.run, run_id: runId },
      replay: { status: checked ? "reproduces" : "not_checked" },
      open_questions: checked ? openQuestions.filter((each) => each !== replayQuestion) : openQuestions,
    });
  });
  bridge.handlers.set("run.compare", () => success(comparisonWire));
  return { bridge, asked };
}

export async function reviewing(openQuestions?: unknown[], runId?: string) {
  const { bridge, asked } = bridgeWith(openQuestions, runId);
  const saved: { name: string; text: string }[] = [];
  const saver = {
    save: (name: string, text: string) => {
      saved.push({ name, text });
      return true;
    },
  };
  const app = mountApp(bridge, reviewHash(), undefined, saver);
  await screen.findByRole("region", { name: "This run" });
  return { ...app, bridge, asked, saved };
}
