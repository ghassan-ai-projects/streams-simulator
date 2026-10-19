import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import type { ConnectSession, ConnectState } from "../../../app/launch";
import { createReviewSession } from "../../../app/review";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { controlReplyWire, interruptsWire, liveWire } from "../../../test/liveFixture";
import { failure, success } from "../../../test/outcome";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { Interrupts } from "./Interrupts";

function knownRun() {
  const run = sampleWorkspace.runs.find((entry) => entry.worldId === "w-2");
  if (run === undefined) throw new Error("missing fixture");
  return run;
}

const offered: ConnectState = {
  phase: "live",
  domain: "mini",
  run: "c-1",
  check: null,
  binding: null,
  stopping: false,
  controls: true,
  note: null,
};

const fanWire = {
  device: "fan",
  board: "mega-01",
  instance: "fan-01",
  environment: "simulated",
  boot_id: "1",
  health: "simulated",
  reboots: 0,
  readings: [],
};

function snapshot(interrupts?: object) {
  const wire = liveWire(10e9);
  return { ...wire, devices: [...wire.devices, fanWire], ...(interrupts === undefined ? {} : { interrupts }) };
}

const [waitingFan] = interruptsWire().pending;
const busy = interruptsWire({
  holds: [
    { kind: "board", id: "uno-07", latched: false },
    { kind: "device", id: "fan", latched: true },
  ],
  pending: [
    { ...waitingFan, deadline_ns: 130e9 },
    {
      ...waitingFan,
      command_id: "cmd-8",
      device: "sensor",
      command: "calibrate",
      deadline_ns: 1_700_000_045e9,
      clock: "wall",
    },
  ],
});

function mount(wire: object | null = snapshot(busy), connection: ConnectState = offered) {
  const bridge = new FakeBridge();
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  review.openLive("w-2", knownRun());
  if (wire !== null) {
    bridge.stream("w-2", [wire], "caught_up");
    timers.advance(250);
  }
  const connect = { snapshot: () => connection, subscribe: () => () => undefined } as unknown as ConnectSession;
  const view = render(
    <Interrupts review={review} state={review.snapshot()} connect={connect} now={() => 1_700_000_000_000} />,
  );
  return { bridge, view };
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
    expect(within(waiting).getByText("cmd-7 · expires in 2 min (run time)")).toBeInTheDocument();
    expect(within(waiting).getByText("cmd-8 · expires in 45 s")).toBeInTheDocument();
    expect(screen.getByText(/latched, until you release it/)).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says nothing is set, and that nothing waits, when the node holds nothing", () => {
    mount(snapshot(interruptsWire({ holds: [], approve: [], pending: [], devices: [] })));
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
    expect(controlCallsOf(bridge)).toEqual([{ name: "session.control", args: { op: "approve", command_id: "cmd-7" } }]);
  });

  it("denies a command by its id", async () => {
    const { bridge } = mount();
    bridge.actions.set("session.control", success(controlReplyWire({ op: "deny" })));
    fireEvent.click(screen.getByRole("button", { name: "Deny cmd-7" }));
    await waitFor(() => {
      expect(controlCallsOf(bridge)).toEqual([{ name: "session.control", args: { op: "deny", command_id: "cmd-7" } }]);
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

  it("waits for the node's first state before it offers anything, and says why when it cannot", () => {
    mount(snapshot());
    expect(screen.getByText("Waiting for the node to publish its interrupt state…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Emergency stop" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Hold" })).toBeDisabled();
  });

  it("says the controls are not available when the session has no control socket", () => {
    mount(null, { ...offered, controls: false });
    expect(screen.getByText(/Interrupt controls are not available: the node has not opened/)).toBeInTheDocument();
    expect(within(region()).getAllByRole("option")).toHaveLength(1);
  });
});
