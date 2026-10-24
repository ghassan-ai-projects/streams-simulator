import { act, fireEvent, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { stubFlowDom } from "../../../test/dom";
import { mountApp } from "../../../test/appHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { reportWire } from "../../../test/runFixture";
import { reviewBridge, reviewHash } from "../../../test/screenFixture";
import { reviewing } from "../../../test/reviewActionsHarness";

beforeEach(() => {
  stubFlowDom();
});

describe("Check replay now", () => {
  it("asks the binary to replay, then shows the new replay status and drops the question", async () => {
    const { asked } = await reviewing();
    expect(screen.getByText("Replay in this report: Not checked in this report")).toBeInTheDocument();
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Check replay now" }));
      await Promise.resolve();
    });
    expect(await screen.findByText("Replay in this report: Reproduces")).toBeInTheDocument();
    expect(asked.at(-1)).toEqual({ run: "w-1", replay: true });
    expect(document.querySelector("[data-code='replay_not_checked']")).toBeNull();
    expect(document.querySelector("[data-code='stopped_early']")).not.toBeNull();
  });

  it("shows the binary's failure word for word and keeps the question", async () => {
    const { bridge } = await reviewing();
    bridge.handlers.set("run.report", () =>
      failure(cannedFailure("replay_failed", 500, "the recording is unreadable")),
    );
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Check replay now" }));
      await Promise.resolve();
    });
    expect(await screen.findByRole("alert")).toHaveTextContent("the recording is unreadable");
    expect(screen.getByRole("button", { name: "Check replay now" })).toBeEnabled();
  });
});

describe("Export handover", () => {
  it("downloads the binary's markdown as <run id>-handover.md and says so", async () => {
    const { asked, saved } = await reviewing();
    fireEvent.click(screen.getByRole("button", { name: "Export handover" }));
    expect(await screen.findByText("Handover saved as r-1-handover.md.")).toBeInTheDocument();
    expect(asked.at(-1)).toEqual({ run: "w-1", format: "markdown" });
    expect(saved).toEqual([{ name: "r-1-handover.md", text: "# Run w-1\n\n## Open questions\n" }]);
  });

  it("says it next to the button, in a status region, and takes the note away after a few seconds", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      await reviewing();
      fireEvent.click(screen.getByRole("button", { name: "Export handover" }));
      const note = await screen.findByText("Saved r-1-handover.md");
      expect(note.closest("[role='status']")).not.toBeNull();
      expect(note.closest("[role='status']")?.parentElement).toContainElement(
        screen.getByRole("button", { name: "Export handover" }),
      );
      await act(async () => {
        await vi.advanceTimersByTimeAsync(8100);
      });
      expect(screen.queryByText("Saved r-1-handover.md")).toBeNull();
      expect(screen.getByText("Handover saved as r-1-handover.md.")).toBeInTheDocument();
    } finally {
      vi.useRealTimers();
    }
  });

  it("names the file after the world when the report carries no run id", async () => {
    const { saved } = await reviewing(undefined, "");
    fireEvent.click(screen.getByRole("button", { name: "Export handover" }));
    expect(await screen.findByText("Handover saved as w-1-handover.md.")).toBeInTheDocument();
    expect(saved.at(-1)?.name).toBe("w-1-handover.md");
  });

  it("asks for the replay too when it was already checked in this view", async () => {
    const { asked } = await reviewing();
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Check replay now" }));
      await Promise.resolve();
    });
    await screen.findByText("Replay in this report: Reproduces");
    fireEvent.click(screen.getByRole("button", { name: "Export handover" }));
    await screen.findByText("Handover saved as r-1-handover.md.");
    expect(asked.at(-1)).toEqual({ run: "w-1", format: "markdown", replay: true });
  });

  it("shows the binary's refusal word for word", async () => {
    const { bridge } = await reviewing();
    bridge.handlers.set("run.report", () => failure(cannedFailure("run_not_ended", 422, "run w-1 has not ended")));
    fireEvent.click(screen.getByRole("button", { name: "Export handover" }));
    const refusal = await screen.findByRole("alert");
    expect(refusal).toHaveTextContent("run w-1 has not ended");
    expect(refusal.parentElement).toContainElement(screen.getByRole("button", { name: "Export handover" }));
  });
});

describe("a run that has not ended", () => {
  const message = "The run has not ended, so its truth is still sealed: this report withholds the faults.";

  async function unsealed() {
    const bridge = reviewBridge();
    bridge.handlers.set("run.report", () =>
      success({
        run: { ...reportWire.run, ended: false, complete: false, mode: "evaluate" },
        verification: { status: "not_checked" },
        counts: {},
        replay: { status: "not_checked" },
        open_questions: [{ code: "evaluate_unsealed", message }],
      }),
    );
    const saver = { save: vi.fn(() => true) };
    mountApp(bridge, reviewHash(), undefined, saver);
    await screen.findByRole("region", { name: "This run" });
    return { bridge, saver };
  }

  it("cannot be exported or compared, and says why in the binary's words", async () => {
    const { bridge, saver } = await unsealed();
    const exporting = screen.getByRole("button", { name: "Export handover" });
    const comparing = screen.getByRole("button", { name: "Compare with…" });
    for (const button of [exporting, comparing]) {
      expect(button).toHaveAttribute("aria-disabled", "true");
      expect(button).toHaveAttribute("title", message);
      expect(button).toHaveAccessibleDescription(message);
    }
    fireEvent.click(exporting);
    fireEvent.click(comparing);
    expect(saver.save).not.toHaveBeenCalled();
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(bridge.calls.filter((call) => call.name === "run.compare")).toEqual([]);
    expect(bridge.calls.filter((call) => call.args.format === "markdown")).toEqual([]);
  });

  it("shows nothing sealed anywhere on the page", async () => {
    await unsealed();
    expect(document.body.textContent).not.toMatch(/fan failure|heat wave|interlock|tower__fan/);
  });
});
