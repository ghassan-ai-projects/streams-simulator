import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { deferred } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { launchBridge, launchHarness } from "../../../test/launchFixture";
import { failure, success } from "../../../test/outcome";
import { domain } from "../../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

const check = {
  domain: "mini",
  environment: "mixed",
  instances: [
    { instance: "CH-1", origin: "real", kind: "chiller", board: "uno-07", device: "ch-1" },
    { instance: "CT-1", origin: "simulated", kind: "tower" },
  ],
};

function sessionBridge(): FakeBridge {
  const fake = launchBridge();
  fake.actions.set("session.status", success({ state: "none" }));
  fake.actions.set("session.check", success(check));
  fake.actions.set("session.start", success({ run: "c-1", state: "running" }));
  fake.actions.set("session.stop", success({ run: "c-1", state: "ended" }));
  return fake;
}

function renderSession(fake: FakeBridge, onFinished = vi.fn<(run: string) => void>()) {
  const { launch } = launchHarness(fake);
  const view = render(
    <LaunchPanel
      launch={launch}
      plant={parsePlant(domain)}
      path="domains/mini.domain.json"
      onClose={vi.fn()}
      onFinished={onFinished}
    />,
  );
  fireEvent.click(screen.getByRole("radio", { name: /^Connected session/ }));
  return { launch, view, onFinished };
}

function fill(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

function planWithARealBoard() {
  fill("Host socket", "/tmp/h.sock");
  fill("Gateway socket", "/tmp/g.sock");
  fireEvent.click(screen.getByRole("button", { name: "Add a real board" }));
  fill("Real board 1 board", "uno-07");
  fill("Real board 1 device", "ch-1");
}

describe("a connected session in the run panel", () => {
  it("says what each way of running does and starts on the seed", async () => {
    const { view } = renderSession(sessionBridge());
    expect(screen.getByRole("radio", { name: /^To completion/ })).not.toBeChecked();
    expect(screen.getByLabelText("Seed")).toHaveFocus();
    expect(screen.getByText(/runs until you stop it/)).toBeInTheDocument();
    expect(screen.getByText(/Every instance is simulated/)).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(screen.getByRole("radio", { name: /^To completion/ }));
    expect(screen.getByRole("button", { name: "Start the run" })).toBeInTheDocument();
  });

  it("checks the plan through the binary and lists each instance as a real board or simulated", async () => {
    const fake = sessionBridge();
    renderSession(fake);
    planWithARealBoard();
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    expect(screen.getByText("Checking the plan with streamsim…")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Check the plan" }));
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    const instances = await screen.findByRole("region", { name: "Instances in this session" });
    expect(
      fake.calls.filter((call) => call.name.startsWith("session.") && call.name !== "session.status"),
    ).toHaveLength(1);
    expect(within(instances).getByRole("status")).toHaveTextContent(
      "The plan holds: 1 real board and 1 simulated instance.",
    );
    expect(within(instances).getByText("CH-1 · Real board · device ch-1 on board uno-07")).toBeInTheDocument();
    const plan = JSON.parse(String(fake.calls.find((call) => call.name === "session.check")?.args.content)) as unknown;
    expect(plan).toMatchObject({
      host_socket: "/tmp/h.sock",
      gateway_socket: "/tmp/g.sock",
      real: [{ instance: "CH-1", board: "uno-07", device: "ch-1" }],
    });
  });

  it("puts each of the binary's issues on the field it names", async () => {
    const fake = sessionBridge();
    const issues = [
      {
        pointer: "/real/0/device",
        message: "board uno-07 holds no device ch-1",
        keyword: "device_missing",
        member: "device",
      },
      { pointer: "/host_socket", message: "want an absolute path", keyword: "pattern" },
      { pointer: "", message: "the domain changed on disk" },
    ];
    fake.actions.set(
      "session.check",
      failure({ ...cannedFailure("plantwire_failed", 422, "the plan does not hold"), issues }),
    );
    renderSession(fake);
    planWithARealBoard();
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("the plan does not hold (plantwire_failed)");
    expect(screen.getByLabelText("Real board 1 device")).toHaveAccessibleDescription(
      "board uno-07 holds no device ch-1",
    );
    expect(screen.getByLabelText("Host socket")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByRole("list", { name: "Other problems in the plan" })).toHaveTextContent(
      "the domain changed on disk",
    );
    expect(fake.calls.some((call) => call.name === "session.start")).toBe(false);
  });

  it("runs the session until it is stopped and opens the saved run", async () => {
    const fake = sessionBridge();
    const started = deferred<Canned<unknown>>();
    fake.handlers.set("session.start", () => started.promise);
    const { onFinished } = renderSession(fake);
    planWithARealBoard();
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    expect(await screen.findByText("Starting the session…")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    started.resolve(success({ run: "c-1", state: "running" }));
    expect(await screen.findByText("Connected session · run c-1")).toHaveFocus();
    expect(screen.getByText("Waiting for the gateway to announce its boards…")).toBeInTheDocument();
    act(() => {
      fake.stream(
        "c-1",
        [{ stream: "observer", seq: 0, type: "binding", at_ns: 0, outcome: "matched", issues: [] }],
        null,
      );
    });
    expect(screen.getByText(/announced every real board/)).toBeInTheDocument();
    expect(screen.queryByRole("radio", { name: /^To completion/ })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    expect(await screen.findByText("The session was stopped and saved as run c-1.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Open the review" }));
    expect(onFinished).toHaveBeenCalledWith("c-1");
  });

  it("shows a gateway the binary refused with the rule it broke", async () => {
    const fake = sessionBridge();
    renderSession(fake);
    planWithARealBoard();
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    await screen.findByText("Connected session · run c-1");
    const issue = {
      pointer: "/real/0/device",
      message: "board uno-07 holds no device ch-1",
      keyword: "device_missing",
    };
    const refused = {
      stream: "observer",
      seq: 0,
      type: "binding",
      at_ns: 0,
      outcome: "refused",
      refusal: "inventory_mismatch",
      reason: issue.message,
      issues: [issue],
    };
    act(() => {
      fake.stream("c-1", [refused], null);
    });
    expect(screen.getByRole("alert")).toHaveTextContent(
      "The gateway was refused (inventory_mismatch): board uno-07 holds no device ch-1.",
    );
  });

  it("says when the session failed, keeps the run's name, and goes back to the plan", async () => {
    const fake = sessionBridge();
    fake.actions.set("session.check", success({ ...check, environment: "simulated" }));
    fake.actions.set(
      "session.stop",
      success({ run: "c-1", state: "failed", failure: { code: "plantwire_failed", message: "host lost" } }),
    );
    renderSession(fake);
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    await screen.findByText("No gateway: every instance is simulated.");
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("host lost (plantwire_failed)");
    expect(screen.getByText("What ran is kept in run c-1.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Back to the session plan" }));
    expect(screen.getByLabelText("Host socket")).toBeInTheDocument();
  });

  it("adds a real board for an instance not yet bound and removes it again", () => {
    renderSession(sessionBridge());
    fireEvent.click(screen.getByRole("button", { name: "Add a real board" }));
    fireEvent.click(screen.getByRole("button", { name: "Add a real board" }));
    expect(screen.getByLabelText("Real board 1 instance")).toHaveValue("CH-1");
    expect(screen.getByLabelText("Real board 2 instance")).toHaveValue("CT-1");
    fireEvent.change(screen.getByLabelText("Real board 1 instance"), { target: { value: "MB-1" } });
    expect(screen.getByLabelText("Real board 1 instance")).toHaveValue("MB-1");
    fireEvent.click(screen.getByRole("button", { name: "Remove real board 1" }));
    expect(screen.getByLabelText("Real board 1 instance")).toHaveValue("CT-1");
    expect(screen.queryByLabelText("Real board 2 instance")).not.toBeInTheDocument();
  });

  it("picks up a session that is already running when the panel opens after a reload", async () => {
    const running = sessionBridge();
    running.actions.set("session.status", success({ run: "c-9", state: "running" }));
    renderSession(running);
    await waitFor(() => {
      expect(screen.getByText("Connected session · run c-9")).toBeInTheDocument();
    });
    expect(screen.queryByText(/Waiting for the gateway/)).not.toBeInTheDocument();
  });

  it("says a session already runs when the bridge refuses a second one", async () => {
    const fake = sessionBridge();
    fake.actions.set(
      "session.start",
      failure(cannedFailure("session_running", 422, "the connected session c-1 is still running")),
    );
    renderSession(fake);
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "the connected session c-1 is still running (session_running)",
    );
    expect(screen.queryByText(/What ran is kept/)).not.toBeInTheDocument();
  });

  it("tells another plant that a session is running elsewhere", async () => {
    const fake = sessionBridge();
    const { launch } = launchHarness(fake);
    await launch.connect.start("other", { seed: "1", hostSocket: "/tmp/h.sock", gatewaySocket: "", real: [] });
    render(<LaunchPanel launch={launch} plant={parsePlant(domain)} path="p" onClose={vi.fn()} onFinished={vi.fn()} />);
    fireEvent.click(screen.getByRole("radio", { name: /^Connected session/ }));
    expect(screen.getByText("A connected session of other is running.")).toBeInTheDocument();
  });
});
