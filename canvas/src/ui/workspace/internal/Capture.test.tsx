import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { announced, captureResult, differs, pumpDraft, refuseCapture, same } from "../../../test/captureFixture";
import { captureCalls, cardOf, heading, openCapture, paste, reviewed } from "../../../test/captureHarness";
import { FakeBridge } from "../../../test/fakeBridge";

const read = () => screen.findByRole("heading", { level: 2, name: heading });

describe("capturing from a board", () => {
  it("opens from the Devices list on a screen of its own, with the heading focused", async () => {
    const { view } = await openCapture();
    const title = screen.getByRole("heading", { level: 1, name: "Capture from a board" });
    expect(title).toHaveFocus();
    expect(screen.queryByRole("region", { name: "Devices" })).toBeNull();
    expect(screen.getByLabelText("Announcement text")).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("keeps Read off until there is text to read, and never sends an empty text", async () => {
    const { bridge } = await openCapture();
    const button = screen.getByRole("button", { name: "Read the announcement" });
    expect(button).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(button);
    expect(captureCalls(bridge)).toEqual([]);
  });

  it("reads the pasted text and asks to review what the board announced, with the claim said plainly", async () => {
    const { bridge, view } = await reviewed();
    expect(captureCalls(bridge)).toEqual([{ name: "kind.capture", args: { content: '{"type":"NBIRTH"}' } }]);
    expect(screen.getByRole("heading", { level: 2, name: heading })).toHaveFocus();
    expect(screen.getByText("Real board")).toBeInTheDocument();
    expect(screen.getByText("Announcement number 2")).toBeInTheDocument();
    expect(
      screen.getByText("A board's announcement is a claim. Nothing is saved until you review and accept it."),
    ).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("The board announced 1 device.");
    expect(bridge.writes).toEqual([]);
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("shows a simulated announcement as Simulated, not as a real board", async () => {
    await reviewed(captureResult([announced()], { environment: "simulated", bd_seq: 0 }));
    expect(screen.getByText("Simulated")).toBeInTheDocument();
    expect(screen.queryByText("Real board")).toBeNull();
  });

  it("says each announced device, what it reads and what it accepts, with its digest in the details", async () => {
    await reviewed();
    const card = cardOf("pump");
    expect(card.getByText("New device")).toBeInTheDocument();
    expect(card.getByText("2 readings · 2 commands")).toBeInTheDocument();
    expect(card.getByText("Number · m3/h · 0 to 120")).toBeInTheDocument();
    expect(card.getByText("Number · % · 0 to 100 · setpoint")).toBeInTheDocument();
    expect(card.getByText("takes pct · Medium risk")).toBeInTheDocument();
    expect(card.getByText("no arguments · Low risk · safe stop")).toBeInTheDocument();
    expect(card.getByText("sha256:aaa")).toBeInTheDocument();
    expect(card.getByText("Digest")).toBeInTheDocument();
    expect(card.queryByText("Saved file")).toBeNull();
  });

  it("lists every device of a board that announced several", async () => {
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

  it("marks a draft that still breaks a rule as not yet valid, places the issue in words, and still offers to save it", async () => {
    const issues = [{ pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" }];
    await reviewed(captureResult([announced(pumpDraft, { issues })]));
    const card = cardOf("pump");
    expect(card.getByText("Not yet valid")).toBeInTheDocument();
    expect(card.getByText("Reading flow · Unit")).toBeInTheDocument();
    expect(card.getByText(/Unit is required/)).toBeInTheDocument();
    expect(card.getByText("needs a unit")).toBeInTheDocument();
    expect(card.getByRole("button", { name: "Save as device: pump" })).toHaveAttribute("aria-disabled", "false");
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

describe("a refused announcement", () => {
  async function refused(keyword = "kind_digest_mismatch") {
    const bridge = new FakeBridge();
    const opened = await openCapture(undefined, bridge);
    refuseCapture(bridge, keyword, "plantwire: the birth is refused");
    paste();
    await screen.findByText("The board's announcement was not accepted");
    return opened;
  }

  it("shows the binary's sentence and the plain words, writes nothing, and puts the focus on the first issue", async () => {
    const { bridge, view } = await refused();
    expect(screen.getByRole("alert")).toHaveTextContent("plantwire: the birth is refused");
    const issue = screen.getByText(/does not match its own fingerprint, so it cannot be trusted/).closest("li");
    expect(issue).toHaveFocus();
    expect(within(issue as HTMLElement).getByText("does not hold")).toBeInTheDocument();
    expect(screen.getByText("Nothing was saved.")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: heading })).toBeNull();
    expect(bridge.writes).toEqual([]);
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("can be tried again, and reviews the announcement once it holds", async () => {
    const { bridge } = await refused("no_birth");
    expect(screen.getByText("No announcement from a board was found")).toBeInTheDocument();
    bridge.actions.set("kind.capture", { ok: true, value: captureResult() });
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(await read()).toHaveFocus();
    expect(captureCalls(bridge)).toHaveLength(2);
  });

  it("puts the focus on the sentence when the binary names no issue", async () => {
    const bridge = new FakeBridge();
    await openCapture(undefined, bridge);
    bridge.actions.delete("kind.capture");
    paste();
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveFocus();
    expect(alert).toHaveTextContent(/unknown_action/);
    expect(screen.queryByRole("list", { name: "What was wrong" })).toBeNull();
  });
});

describe("the other cards of a board", () => {
  it("offers the saved device when the board announced what is already saved", async () => {
    await reviewed(captureResult([same()]));
    const card = cardOf("pump");
    expect(card.getByText("Same as pump.kind.json")).toBeInTheDocument();
    expect(card.getByText("Saved file")).toBeInTheDocument();
    expect(card.getByRole("button", { name: "Open the saved device: pump" })).toBeInTheDocument();
    expect(card.queryByRole("button", { name: /^Save as device/ })).toBeNull();
    expect(card.queryByRole("button", { name: /^Replace the saved device/ })).toBeNull();
  });

  it("says where a device differs from the saved one in words, and keeps the pointer in the details", async () => {
    await reviewed(captureResult([differs()]));
    const card = cardOf("pump");
    expect(card.getByText("Differs from pump.kind.json")).toBeInTheDocument();
    expect(card.getByText(/The first difference is in Reading flow · Unit\./)).toBeInTheDocument();
    expect(card.getByText("First difference at")).toBeInTheDocument();
    expect(card.getByText("/metrics/0/unit")).toBeInTheDocument();
    expect(card.queryByRole("button", { name: /^Save as device/ })).toBeNull();
  });

  it("says the whole file differs when the binary names no place", async () => {
    await reviewed(captureResult([differs("")]));
    expect(cardOf("pump").getByText("The whole file")).toBeInTheDocument();
    expect(cardOf("pump").getByText(/somewhere in the file/)).toBeInTheDocument();
  });

  it("says a saved digest is not reported when the binary gives none", async () => {
    const comparison = { status: "same", file: "kinds/pump.kind.json" };
    await reviewed(captureResult([announced(pumpDraft, { comparison })]));
    expect(cardOf("pump").getByText("Not reported")).toBeInTheDocument();
  });
});
