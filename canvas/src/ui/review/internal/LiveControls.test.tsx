import { fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import type { ReviewSession } from "../../../app/review";
import { reviewState } from "../../../test/reviewFixture";
import { LiveControls } from "./LiveControls";

function review(backToLive = vi.fn()): ReviewSession {
  return { backToLive } as unknown as ReviewSession;
}

const history = { entries: [], startNs: 0, endNs: 20e9, trimmedBeforeNs: null };

describe("live controls", () => {
  it("says it is catching up before it claims to follow the newest record", () => {
    const state = reviewState({ connection: "running", caughtUp: false });
    render(<LiveControls review={review()} state={state} startNs={0} now={() => 0} />);
    expect(screen.getByRole("status")).toHaveTextContent("Catching up to live…");
    expect(screen.queryByRole("button", { name: "Back to live" })).toBeNull();
  });

  it("returns to live, keeps focus on the live status and shows the age of the last update", async () => {
    const backToLive = vi.fn();
    const state = reviewState({ connection: "running", following: false, lastLiveAtMs: 1_000 });
    const view = render(<LiveControls review={review(backToLive)} state={state} startNs={0} now={() => 13_400} />);
    expect(screen.getByText("Last update 12 s ago")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Back to live" }));
    expect(backToLive).toHaveBeenCalledOnce();
    expect(document.activeElement).toBe(screen.getByRole("status"));
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says when no update came for a while, when a snapshot was unreadable and where history was trimmed", () => {
    const state = reviewState({
      connection: "running",
      following: true,
      lastLiveAtMs: 0,
      liveUnreadableNs: 75e9,
      liveHistory: { ...history, trimmedBeforeNs: 60e9 },
    });
    render(<LiveControls review={review()} state={state} startNs={0} now={() => 45_000} />);
    expect(screen.getByRole("status")).toHaveTextContent("Live · following newest record");
    expect(screen.getByText("No update for 45 s")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent(
      "A live snapshot could not be read at 00:01:15; the view is not current.",
    );
    expect(screen.getByText("History trimmed before 00:01:00")).toBeInTheDocument();
  });
});
