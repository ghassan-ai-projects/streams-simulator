import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import type { ReviewSession } from "../../../app/review";
import { createReviewSession } from "../../../app/review";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { busyInterrupts, offeredSession, sessionSnapshot } from "../../../test/interruptFixture";
import { controlReplyWire, interruptsWire } from "../../../test/liveFixture";
import { failure, success } from "../../../test/outcome";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { deferred } from "../../../test/deferred";
import type { Canned } from "../../../test/fakeFailures";
import { useStore } from "../../primitives";
import { Interrupts } from "./Interrupts";

function knownRun() {
  const run = sampleWorkspace.runs.find((entry) => entry.worldId === "w-2");
  if (run === undefined) throw new Error("missing fixture");
  return run;
}

const digest = `sha256:${"a".repeat(64)}`;

function Harness({ review, connect }: { readonly review: ReviewSession; readonly connect: ConnectSession }) {
  const state = useStore(review);
  return <Interrupts review={review} state={state} connect={connect} now={() => 1_700_000_000_000} />;
}

function mount(wire: object | null = sessionSnapshot(busyInterrupts), connection: ConnectState = offeredSession) {
  const bridge = new FakeBridge();
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  review.openLive("w-2", knownRun());
  const stream = (next: object) => {
    act(() => {
      bridge.stream("w-2", [next], null);
      timers.advance(250);
    });
  };
  if (wire !== null) {
    bridge.stream("w-2", [wire], "caught_up");
    timers.advance(250);
  }
  const connect = { snapshot: () => connection, subscribe: () => () => undefined } as unknown as ConnectSession;
  const view = render(<Harness review={review} connect={connect} />);
  return { bridge, view, review, stream, timers };
}

const region = () => screen.getByRole("region", { name: "Interrupts" });

function controlCallsOf(bridge: FakeBridge) {
  return bridge.calls.filter((call) => call.name === "session.control");
}

describe("the interrupt controls on the live screen", () => {
  it("shows what is held, latched and approving in words, and the commands that wait with their deadline", async () => {
    const { view } = mount();
    const active = within(region()).getByRole("list", { name: "Active interrupts" });
    expect(within(active).getByText("Board uno-07: Held")).toBeInTheDocument();
    expect(within(active).getByText("Device fan: Held · latched by emergency stop")).toBeInTheDocument();
    expect(within(active).getByText("All devices: Approving each command")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "2 commands waiting for approval" })).toBeInTheDocument();
    const waiting = screen.getByRole("list", { name: "Waiting commands" });
    expect(within(waiting).getByText("set_fan on fan")).toBeInTheDocument();
    expect(within(waiting).getByText("cmd-7 · expires in about 2 min (run time)")).toBeInTheDocument();
    expect(within(waiting).getByText("cmd-8 · expires in about 45 s")).toBeInTheDocument();
    expect(within(waiting).getAllByText("Arguments: duty_permille: 300, lease_ms: 5000")).toHaveLength(2);
    expect(screen.getByText(/latched, until you release it/)).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says nothing is set, and that nothing waits, when the node holds nothing", () => {
    mount(sessionSnapshot(interruptsWire({ holds: [], approve: [], pending: [], devices: [] })));
    expect(screen.getByText("No holds or approval modes are set.")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "No commands wait for approval." })).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Waiting commands" })).toBeNull();
  });

  it("offers all devices, each board and each device, and acts on the chosen scope only", async () => {
    const { bridge } = mount();
    bridge.actions.set("session.control", success(controlReplyWire()));
    const scope = screen.getByLabelText("Scope");
    expect(
      within(scope)
        .getAllByRole("option")
        .map((option) => option.textContent),
    ).toEqual(["All devices", "Board mega-01", "Board uno-07", "Device fan", "Device sensor"]);
    fireEvent.change(scope, { target: { value: "board:mega-01" } });
    fireEvent.click(screen.getByRole("button", { name: "Hold" }));
    expect(await screen.findByText("Hold on board mega-01 was applied.")).toBeInTheDocument();
    expect(controlCallsOf(bridge)).toEqual([
      { name: "session.control", args: { op: "hold", scope: { kind: "board", id: "mega-01" } } },
    ]);
  });

  it("shows a hold and an approval mode as pressed and turns the mode off with the same button", async () => {
    const { bridge } = mount();
    bridge.actions.set("session.control", success(controlReplyWire({ op: "approve_mode" })));
    expect(screen.getByRole("button", { name: "Hold" })).toHaveAttribute("aria-pressed", "false");
    const approve = screen.getByRole("button", { name: "Approve each command" });
    expect(approve).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(approve);
    await waitFor(() => {
      expect(bridge.calls.at(-1)?.args).toEqual({ op: "approve_mode", scope: { kind: "all" }, enabled: false });
    });
    fireEvent.change(screen.getByLabelText("Scope"), { target: { value: "board:uno-07" } });
    expect(screen.getByRole("button", { name: "Hold" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Approve each command" })).toHaveAttribute("aria-pressed", "false");
    fireEvent.click(screen.getByRole("button", { name: "Release" }));
    await waitFor(() => {
      expect(bridge.calls.at(-1)?.args).toEqual({ op: "release", scope: { kind: "board", id: "uno-07" } });
    });
  });

  it("sends the emergency stop in one press and says what it stopped and that the scope stays held", async () => {
    const { bridge } = mount();
    const reply = controlReplyWire({ op: "estop", stopped: ["fan"], skipped: ["sensor"] });
    bridge.actions.set("session.control", success(reply));
    const stop = screen.getByRole("button", { name: "Emergency stop" });
    expect(stop).toHaveAccessibleDescription(/latched, until you release it/);
    fireEvent.click(stop);
    expect(await screen.findByText(/Safe stop sent to fan\. sensor declares no safe stop\./)).toBeInTheDocument();
    expect(controlCallsOf(bridge)).toEqual([
      { name: "session.control", args: { op: "estop", scope: { kind: "all" } } },
    ]);
  });

  it("approves or denies a waiting command and puts focus on the waiting heading", async () => {
    const { bridge } = mount();
    bridge.actions.set(
      "session.control",
      success(controlReplyWire({ op: "approve", state: interruptsWire({ pending: [] }) })),
    );
    fireEvent.click(screen.getByRole("button", { name: "Approve cmd-7" }));
    expect(await screen.findByRole("heading", { name: "No commands wait for approval." })).toHaveFocus();
    expect(controlCallsOf(bridge)).toEqual([
      { name: "session.control", args: { op: "approve", command_id: "cmd-7", content_digest: digest } },
    ]);
  });

  it("denies a command by its id", async () => {
    const { bridge } = mount();
    bridge.actions.set("session.control", success(controlReplyWire({ op: "deny" })));
    fireEvent.click(screen.getByRole("button", { name: "Deny cmd-7" }));
    await waitFor(() => {
      expect(controlCallsOf(bridge)).toEqual([
        { name: "session.control", args: { op: "deny", command_id: "cmd-7", content_digest: digest } },
      ]);
    });
  });

  it("states a refusal and a failure plainly", async () => {
    const { bridge } = mount();
    bridge.actions.set("session.control", success(controlReplyWire({ result: "refused", reason: "unknown_command" })));
    fireEvent.click(screen.getByRole("button", { name: "Deny cmd-7" }));
    expect(await screen.findByText(/Deny cmd-7 was refused: That command no longer waits/)).toHaveAttribute(
      "data-tone",
      "refused",
    );
    bridge.actions.set("session.control", failure(cannedFailure("control_gone", 422, "words")));
    fireEvent.click(screen.getByRole("button", { name: "Release" }));
    const failed = await screen.findByText(/Release of all devices: The node's control socket is not there/);
    expect(failed).toHaveAttribute("data-tone", "failed");
  });

  it("waits for the node's first state before it offers hold or approval, but never holds back the emergency stop", () => {
    mount(sessionSnapshot());
    expect(screen.getByText("Waiting for the node to publish its interrupt state…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Emergency stop" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "Hold" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Release" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Approve each command" })).toBeDisabled();
  });

  it("says the controls are not available when the session has no control socket, and disables every one", () => {
    mount(null, { ...offeredSession, controls: false });
    expect(screen.getByText(/Interrupt controls are not available: the node has not opened/)).toBeInTheDocument();
    expect(within(region()).getAllByRole("option")).toHaveLength(1);
    for (const name of ["Emergency stop", "Hold", "Release", "Approve each command"]) {
      expect(screen.getByRole("button", { name })).toBeDisabled();
    }
  });

  it("disables every control, and says why, while the session is stopping", () => {
    mount(sessionSnapshot(busyInterrupts), { ...offeredSession, stopping: true });
    expect(screen.getByText("The session is stopping, so the node takes no more controls.")).toBeInTheDocument();
    for (const name of ["Emergency stop", "Hold", "Release", "Approve each command", "Approve cmd-7", "Deny cmd-7"]) {
      expect(screen.getByRole("button", { name })).toBeDisabled();
    }
  });

  it("keeps a picked scope that left the session, labelled, and offers no hold, release or approval for it", () => {
    const { review } = mount();
    act(() => {
      review.controls.pick({ kind: "board", id: "gone" });
    });
    const scope = screen.getByLabelText("Scope");
    expect(scope).toHaveValue("board:gone");
    expect(within(scope).getByRole("option", { name: "Board gone (not in this session)" })).toBeInTheDocument();
    for (const name of ["Hold", "Release", "Approve each command"]) {
      expect(screen.getByRole("button", { name })).toBeDisabled();
    }
    expect(screen.getByRole("button", { name: "Emergency stop" })).toBeEnabled();
    expect(screen.getByText(/This scope is not in the session now/)).toBeInTheDocument();
    fireEvent.change(scope, { target: { value: "device:fan" } });
    expect(review.controls.snapshot().picked).toEqual({ kind: "device", id: "fan" });
    expect(screen.getByRole("button", { name: "Hold" })).toBeEnabled();
  });

  it("sends the emergency stop even while another control waits for the node, and refuses the others meanwhile", async () => {
    const { bridge } = mount();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("session.control", (args) => {
      if (args.op === "estop") return success(controlReplyWire({ op: "estop", stopped: ["fan"] }));
      return gate.promise;
    });
    fireEvent.click(screen.getByRole("button", { name: "Approve each command" }));
    expect(screen.getByRole("button", { name: "Approve cmd-7" })).toBeDisabled();
    expect(region().querySelector("[role='status'][aria-busy='true']")).not.toBeNull();
    expect(screen.getByRole("button", { name: "Emergency stop" })).toBeEnabled();
    fireEvent.click(screen.getByRole("button", { name: "Hold" }));
    expect(region().querySelector("[role='status'][aria-busy='true']")).toHaveTextContent(
      "Hold on all devices: not sent: another control is still waiting.",
    );
    fireEvent.click(screen.getByRole("button", { name: "Emergency stop" }));
    await waitFor(() => {
      expect(controlCallsOf(bridge).map((call) => call.args.op)).toEqual(["approve_mode", "estop"]);
    });
    gate.resolve(success(controlReplyWire({ op: "approve_mode" })));
    expect(await screen.findByText(/Safe stop sent to fan\./)).toBeInTheDocument();
  });

  it("keeps the note that a control was not sent visible after the control in flight has answered", async () => {
    const { bridge, review } = mount();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("session.control", () => gate.promise);
    fireEvent.click(screen.getByRole("button", { name: "Hold" }));
    await act(async () => {
      await review.controls.release({ kind: "all", id: "" });
    });
    expect(
      screen.getByText(/Release of all devices: not sent: another control is still waiting\./),
    ).toBeInTheDocument();
    gate.resolve(success(controlReplyWire()));
    expect(
      await screen.findByText(
        /Hold on all devices was applied\..*Release of all devices: not sent: another control is still waiting\./,
      ),
    ).toBeInTheDocument();
  });

  it("says which devices an emergency stop could not reach", async () => {
    const { bridge } = mount();
    bridge.actions.set("session.control", success(controlReplyWire({ op: "estop", unreachable: ["pump"] })));
    fireEvent.click(screen.getByRole("button", { name: "Emergency stop" }));
    const line = await screen.findByText(/The stop could NOT be sent to pump: they may still be energised\./);
    expect(line).toHaveAttribute("data-tone", "failed");
  });

  it("shows a pressed toggle by a visible word as well as by colour", () => {
    mount();
    const approve = screen.getByRole("button", { name: "Approve each command" });
    expect(within(approve).getByText("On")).toBeInTheDocument();
    expect(within(screen.getByRole("button", { name: "Hold" })).getByText("On")).not.toBeVisible();
    fireEvent.change(screen.getByLabelText("Scope"), { target: { value: "board:uno-07" } });
    expect(within(screen.getByRole("button", { name: "Hold" })).getByText("On")).toBeInTheDocument();
  });

  it("names the wider scope that already covers the chosen one", () => {
    mount();
    fireEvent.change(screen.getByLabelText("Scope"), { target: { value: "device:sensor" } });
    expect(screen.getByText("Device sensor is also covered by Board uno-07: Held.")).toBeInTheDocument();
  });

  it("announces a command that arrives, politely, and not the ones already waiting", () => {
    const { stream } = mount(sessionSnapshot(interruptsWire({ pending: [] })));
    expect(screen.queryByText(/A command is waiting for approval/)).toBeNull();
    stream({ ...sessionSnapshot(busyInterrupts), at_ns: 11e9, seq: 12 });
    const announcement = screen.getByText("A command is waiting for approval: calibrate on sensor (cmd-8).");
    expect(announcement).toHaveAttribute("role", "status");
    expect(announcement).toHaveClass("sr-only");
  });
});
