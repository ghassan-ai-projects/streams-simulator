import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { deferred } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { launchHarness } from "../../../test/launchFixture";
import { failure, success } from "../../../test/outcome";
import { domain } from "../../../test/plantFixture";
import { check, fill, planWithARealBoard, renderSession, sessionBridge } from "../../../test/sessionHarness";
import { LaunchPanel } from "./LaunchPanel";

describe("a running connected session in the run panel", () => {
  it("runs the session until it is stopped and opens the saved run", async () => {
    const fake = sessionBridge();
    const started = deferred<Canned<unknown>>();
    fake.handlers.set("session.start", () => started.promise);
    const { onFinished } = renderSession(fake);
    planWithARealBoard();
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    expect(await screen.findByText("Starting the session…")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    started.resolve(success({ run: "c-1", state: "running", domain: "mini", plan: check }));
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
    const panel = screen.getByRole("complementary");
    expect(within(panel).getByText("Connected session · run c-1", { exact: true })).toBeInTheDocument();
    expect(within(panel).getByRole("region", { name: "Instances in this session" })).toBeInTheDocument();
    expect(within(panel).getByRole("button", { name: "Stop and save" })).toBeInTheDocument();
    expect(within(panel).queryByRole("group", { name: "How this run goes" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    expect(await screen.findByText("The session was stopped and saved as run c-1.")).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Open the review" }));
    expect(onFinished).toHaveBeenCalledWith("c-1");
  });

  it("shows a gateway the binary refused, marks its real boards not bound, and says the next step", async () => {
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
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("The gateway was refused (inventory_mismatch), so its real boards were not bound.");
    expect(alert).toHaveTextContent("Stop the session, fix the board or device, then start it again.");
    expect(within(alert).getByRole("list", { name: "Why the gateway was refused" })).toHaveTextContent(
      "The board holds no such device.",
    );
    const instances = screen.getByRole("region", { name: "Instances in this session" });
    expect(within(instances).getByText("CH-1 · Not bound: refused · not announced to the host")).toBeInTheDocument();
    expect(within(instances).queryByText(/Real · device/)).not.toBeInTheDocument();
  });

  it("keeps a session it could not stop running, says so, and stops it on a retry", async () => {
    const fake = sessionBridge();
    renderSession(fake);
    fill("Host socket", "/tmp/h.sock");
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    await screen.findByText("Connected session · run c-1");
    fake.actions.set("session.stop", failure(cannedFailure("bridge_failed", 500, "boom")));
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("The session could not be stopped: boom.");
    expect(screen.getByText("Connected session · run c-1")).toBeInTheDocument();
    fake.actions.set("session.stop", success({ run: "c-1", state: "ended", domain: "mini" }));
    fireEvent.click(screen.getByRole("button", { name: "Stop and save" }));
    expect(await screen.findByText("The session was stopped and saved as run c-1.")).toBeInTheDocument();
  });

  it("says when the session failed, keeps the run's name, and goes back to the plan", async () => {
    const fake = sessionBridge();
    const simulated = { ...check, environment: "simulated" };
    fake.actions.set("session.check", success(simulated));
    fake.actions.set("session.start", success({ run: "c-1", state: "running", domain: "mini", plan: simulated }));
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
    expect(screen.getByRole("alert")).toHaveFocus();
    expect(screen.getByText("What ran is kept in run c-1.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Back to the session plan" }));
    expect(screen.getByLabelText("Host socket")).toBeInTheDocument();
  });

  it("picks up a session that is already running when the panel opens after a reload, with its plan", async () => {
    const running = sessionBridge();
    running.actions.set("session.status", success({ run: "c-9", state: "running", domain: "mini", plan: check }));
    renderSession(running);
    await waitFor(() => {
      expect(screen.getByText("Connected session · run c-9")).toBeInTheDocument();
    });
    expect(screen.getByText("CH-1 · Real · device ch-1 on board uno-07")).toBeInTheDocument();
    expect(screen.getByText("Waiting for the gateway to announce its boards…")).toBeInTheDocument();
  });

  it("tells this plant a session of another plant runs when the panel opens after a reload", async () => {
    const running = sessionBridge();
    running.actions.set("session.status", success({ run: "c-9", state: "running", domain: "other" }));
    renderSession(running);
    expect(await screen.findByText("A connected session of other is running.")).toBeInTheDocument();
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
    fake.actions.set("session.start", success({ run: "c-1", state: "running", domain: "other" }));
    await launch.connect.start("other", {
      seed: "1",
      hostSocket: "/tmp/h.sock",
      gatewaySocket: "",
      real: [],
      nextKey: 0,
    });
    render(<LaunchPanel launch={launch} plant={parsePlant(domain)} path="p" onClose={vi.fn()} onFinished={vi.fn()} />);
    fireEvent.click(screen.getByRole("radio", { name: /^Connected session/ }));
    expect(screen.getByText("A connected session of other is running.")).toBeInTheDocument();
  });
});
