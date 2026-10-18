import { act, fireEvent, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { announced, captureResult, pumpDraft } from "../../../test/captureFixture";
import { captureCalls, cardOf, heading, openCapture, paste, reviewed } from "../../../test/captureHarness";
import { flushPromises } from "../../../test/deferred";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { failure as failed } from "../../../test/outcome";

describe("capturing from a board", () => {
  it("opens on a screen of its own, with the heading focused", async () => {
    const { view } = await openCapture();
    expect(screen.getByRole("heading", { level: 1, name: "Capture from a board" })).toHaveFocus();
    expect(screen.getByLabelText("Announcement text")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("keeps Read off until there is text, says why, and never sends an empty text", async () => {
    const { bridge } = await openCapture();
    const button = screen.getByRole("button", { name: "Read the announcement" });
    expect(button).toHaveAttribute("aria-disabled", "true");
    expect(button).toHaveAccessibleDescription("Paste an announcement first.");
    fireEvent.click(button);
    expect(captureCalls(bridge)).toEqual([]);
    fireEvent.change(screen.getByLabelText("Announcement text"), { target: { value: "{}" } });
    expect(button).not.toHaveAttribute("aria-describedby");
  });

  it("reads the pasted text and asks to review what the board announced, with the claim said plainly", async () => {
    const { bridge, view } = await reviewed();
    expect(captureCalls(bridge)).toEqual([{ name: "kind.capture", args: { content: '{"type":"NBIRTH"}' } }]);
    expect(screen.getByRole("heading", { level: 2, name: heading })).toHaveFocus();
    expect(screen.getByText("Real board")).toBeInTheDocument();
    expect(screen.getByText("Board restart count: 2")).toBeInTheDocument();
    expect(
      screen.getByText("A board's announcement is a claim. Nothing is saved until you review and accept it."),
    ).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("The board announced 1 device.");
    expect(bridge.writes).toEqual([]);
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("folds the form away once there is a review, so the cards come first, and opens it again on request", async () => {
    await reviewed();
    expect(screen.queryByLabelText("Announcement text")).toBeNull();
    expect(screen.getByText("Read from the pasted text.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Read another announcement" }));
    expect(screen.getByLabelText("Announcement text")).toHaveFocus();
    expect(screen.getByRole("article", { name: "pump" })).toBeInTheDocument();
  });

  it("drops the old review when the text is cleared", async () => {
    await reviewed();
    fireEvent.click(screen.getByRole("button", { name: "Read another announcement" }));
    const text = screen.getByLabelText("Announcement text");
    fireEvent.change(text, { target: { value: "{}" } });
    expect(screen.getByRole("article", { name: "pump" })).toBeInTheDocument();
    fireEvent.change(text, { target: { value: "" } });
    expect(screen.queryByRole("article", { name: "pump" })).toBeNull();
    expect(screen.queryByRole("heading", { level: 2, name: heading })).toBeNull();
  });

  it("shows a simulated announcement as Simulated, and says nothing of restarts when there were none", async () => {
    await reviewed(captureResult([announced()], { environment: "simulated", bd_seq: 0 }));
    expect(screen.getByText("Simulated")).toBeInTheDocument();
    expect(screen.queryByText("Real board")).toBeNull();
    expect(screen.queryByText(/restart count/)).toBeNull();
  });

  it("says each announced device, what it reads and accepts, and what the board only claimed", async () => {
    const claims = [
      "/commands/0/risk_class",
      "/commands/0/safe_stop",
      "/commands/0/args_schema/properties/pct/minimum",
    ];
    await reviewed(captureResult([announced(pumpDraft, { claims })]));
    const card = cardOf("pump");
    expect(card.getByText("New device")).toBeInTheDocument();
    expect(card.getByText("2 readings · 2 commands")).toBeInTheDocument();
    expect(card.getByText("Number · m3/h · 0 to 120")).toBeInTheDocument();
    expect(card.getByText("Number · % · 0 to 100 · setpoint")).toBeInTheDocument();
    expect(card.getByText("takes pct · Medium risk")).toBeInTheDocument();
    expect(card.getByText("no arguments · Low risk · safe stop")).toBeInTheDocument();
    expect(
      card.getByText("Taken from the board's announcement, not verified: risk class, safe stop, argument limits."),
    ).toBeInTheDocument();
  });

  it("keeps the fingerprint and the raw draft out of the way, and shows the draft exactly as it would be saved", async () => {
    await reviewed();
    const card = cardOf("pump");
    expect(card.getByText("Fingerprint")).toBeInTheDocument();
    expect(card.getByText("sha256:aaa")).toBeInTheDocument();
    expect(card.getByText("Raw draft")).toBeInTheDocument();
    const raw = card.getByText(/"kind": "pump"/);
    expect(raw.tagName).toBe("PRE");
    expect(JSON.parse(raw.textContent)).toEqual(pumpDraft);
    expect(card.queryByText("Saved file")).toBeNull();
  });

  it("lists every device of a board that announced several, and says so when it announced nothing", async () => {
    const fan = { kind: "fan", metrics: [], commands: [] };
    await reviewed(captureResult([announced(), announced(fan)]));
    expect(screen.getAllByRole("article")).toHaveLength(2);
    expect(cardOf("fan").getByText("This device reports nothing.")).toBeInTheDocument();
    expect(cardOf("fan").getByText("This device accepts no commands.")).toBeInTheDocument();
  });

  it("says so when the board announced nothing", async () => {
    await reviewed(captureResult([]));
    expect(screen.getByText("The board announced no devices.")).toBeInTheDocument();
  });

  it("marks a draft that breaks a rule as not yet valid, places the issue in words, and still offers to save it", async () => {
    const issues = [{ pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" }];
    await reviewed(captureResult([announced(pumpDraft, { issues })]));
    const card = cardOf("pump");
    expect(card.getByText("Not yet valid")).toBeInTheDocument();
    expect(card.getByText("Reading flow · Unit")).toBeInTheDocument();
    expect(card.getByText(/Unit is required/)).toBeInTheDocument();
    expect(card.getByText("needs a unit")).toBeInTheDocument();
    expect(card.getByRole("button", { name: "Save as device: pump" })).toHaveAttribute("aria-disabled", "false");
  });

  it("does not call a draft invalid for what the comparison with the saved file found", async () => {
    const issues = [{ pointer: "/kind", message: "collides", keyword: "kind_name_collision", member: "kind" }];
    await reviewed(captureResult([announced(pumpDraft, { issues })]));
    expect(cardOf("pump").queryByText("Not yet valid")).toBeNull();
    expect(cardOf("pump").queryByText("collides")).toBeNull();
  });

  it("words a draft the board broke beyond reading", async () => {
    const issues = [{ pointer: "", message: "broken", keyword: "kind_draft" }];
    await reviewed(captureResult([announced({ kind: "odd", metrics: "none" }, { issues })]));
    const card = cardOf("odd");
    expect(card.getByText("The announced device could not be read.")).toBeInTheDocument();
    expect(card.getByText("Could not be read")).toBeInTheDocument();
    expect(card.getByText(/does not meet the rules for a device file/)).toBeInTheDocument();
  });
});

describe("what a log holds beside the announcement taken", () => {
  it("says which announcement was taken, that later ones differ, and what else the scan found", async () => {
    const issues = [
      { message: "line 4 differs", keyword: "rebirth_differs" },
      { message: "line 5 unreadable", keyword: "birth_line_unreadable" },
    ];
    const later = [
      { line: 4, bd_seq: 3, same_content: false },
      { line: 6, bd_seq: 4, same_content: true },
    ];
    const { view } = await reviewed(captureResult([announced()], { births_seen: 3, later_births: later, issues }));
    const notice = screen.getByRole("region", { name: "About this announcement" });
    expect(notice).toHaveTextContent("This log holds 3 announcements; showing the first (line 1). Later ones differ.");
    expect(notice).toHaveTextContent("A line that names an announcement could not be read.");
    expect(notice).toHaveTextContent("line 4 differs");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says nothing when the text held one clean announcement", async () => {
    await reviewed();
    expect(screen.queryByRole("region", { name: "About this announcement" })).toBeNull();
  });
});

describe("while the workspace changes underneath", () => {
  it("keeps the review when listing the workspace fails meanwhile", async () => {
    const { bridge } = await reviewed();
    bridge.workspaceOutcome = failed(cannedFailure("bridge_failed", 500, "down"));
    await act(async () => {
      bridge.changeOnDisk();
      await flushPromises();
    });
    expect(screen.getByRole("article", { name: "pump" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 2, name: heading })).toBeInTheDocument();
  });

  it("waits for a workspace that has not been listed yet, with nothing to read from", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = failed(cannedFailure("bridge_failed", 500, "down"));
    await openCapture(undefined, bridge);
    expect(screen.getByText("Loading the workspace…")).toBeInTheDocument();
    expect(screen.queryByLabelText("Announcement text")).toBeNull();
  });
});

describe("reading again", () => {
  it("shows the new review, with the new heading focused", async () => {
    const { bridge } = await reviewed();
    fireEvent.click(screen.getByRole("button", { name: "Read another announcement" }));
    paste('{"type":"NBIRTH","again":true}');
    expect(await screen.findByRole("heading", { level: 2, name: heading })).toHaveFocus();
    expect(captureCalls(bridge)).toHaveLength(2);
  });
});
