import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure } from "../../../test/outcome";
import { fill, planWithARealBoard, renderSession, sessionBridge } from "../../../test/sessionHarness";

describe("the session plan in the run panel", () => {
  it("says what each way of running does, leaves focus on the choice, and keeps the draft across modes", async () => {
    const { view } = renderSession(sessionBridge());
    expect(screen.getByRole("radio", { name: /^To completion/ })).not.toBeChecked();
    expect(screen.getByLabelText("Seed")).not.toHaveFocus();
    fill("Host socket", "/tmp/kept.sock");
    expect(screen.getByText(/runs until you stop it/)).toBeInTheDocument();
    expect(screen.getByText(/Every instance is simulated/)).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
    fireEvent.click(screen.getByRole("radio", { name: /^To completion/ }));
    expect(screen.getByRole("button", { name: "Start the run" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Start the session" })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("radio", { name: /^Connected session/ }));
    expect(screen.getByLabelText("Host socket")).toHaveValue("/tmp/kept.sock");
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
    expect(within(instances).getByText("CH-1 · Real · device ch-1 on board uno-07")).toBeInTheDocument();
    const plan = JSON.parse(String(fake.calls.find((call) => call.name === "session.check")?.args.content)) as unknown;
    expect(plan).toMatchObject({
      host_socket: "/tmp/h.sock",
      gateway_socket: "/tmp/g.sock",
      real: [{ instance: "CH-1", board: "uno-07", device: "ch-1" }],
    });
  });

  it("puts each of the binary's issues on the field it names in plain words, focuses the first, and lists the rest", async () => {
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
      { pointer: "/kinds", message: "a kind file is unreadable", keyword: "kind_unreadable" },
    ];
    fake.actions.set(
      "session.check",
      failure({ ...cannedFailure("plantwire_failed", 422, "the plan does not hold"), issues }),
    );
    renderSession(fake);
    planWithARealBoard();
    fireEvent.click(screen.getByRole("button", { name: "Start the session" }));
    const banner = await screen.findByRole("alert");
    expect(banner).toHaveTextContent("streamsim refused the plan: fix what is marked below and check it again.");
    expect(within(banner).getByText("the plan does not hold (plantwire_failed)")).not.toBeVisible();
    expect(screen.getByLabelText("Real board 1 device")).toHaveAccessibleDescription("The board holds no such device.");
    expect(screen.getByLabelText("Host socket")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getByLabelText("Host socket")).toHaveFocus();
    const all = screen.getByRole("list", { name: "Problems in the plan" });
    expect(within(all).getAllByRole("listitem")).toHaveLength(issues.length);
    expect(all).toHaveTextContent("The board holds no such device.");
    expect(all).toHaveTextContent("the domain changed on disk");
    expect(all).toHaveTextContent("a kind file is unreadable");
    expect(screen.getByText("For developers")).toBeInTheDocument();
    expect(fake.calls.some((call) => call.name === "session.start")).toBe(false);
    fill("Host socket", "/tmp/other.sock");
    expect(screen.getByLabelText("Host socket")).not.toHaveAttribute("aria-invalid");
    expect(screen.getByText(/The plan changed since this check/)).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.queryByText(/streamsim refused the plan/)).not.toBeInTheDocument();
  });

  it("adds a real board for an instance not yet bound, focuses it, and removes it again", () => {
    renderSession(sessionBridge());
    fireEvent.click(screen.getByRole("button", { name: "Add a real board" }));
    expect(screen.getByLabelText("Real board 1 instance")).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Add a real board" }));
    expect(screen.getByLabelText("Real board 1 instance")).toHaveValue("CH-1");
    expect(screen.getByLabelText("Real board 2 instance")).toHaveValue("CT-1");
    expect(screen.getByLabelText("Real board 2 instance")).toHaveFocus();
    fireEvent.change(screen.getByLabelText("Real board 1 instance"), { target: { value: "MB-1" } });
    expect(screen.getByLabelText("Real board 1 instance")).toHaveValue("MB-1");
    fireEvent.click(screen.getByRole("button", { name: "Remove real board 1" }));
    expect(screen.getByLabelText("Real board 1 instance")).toHaveValue("CT-1");
    expect(screen.queryByLabelText("Real board 2 instance")).not.toBeInTheDocument();
  });

  it("adds no real board once every instance has one, and says why", () => {
    renderSession(sessionBridge());
    const add = screen.getByRole("button", { name: "Add a real board" });
    fireEvent.click(add);
    fireEvent.click(add);
    fireEvent.click(add);
    fireEvent.click(add);
    expect(screen.queryByLabelText("Real board 4 instance")).not.toBeInTheDocument();
    expect(add).toHaveAttribute("aria-disabled", "true");
    expect(add).toHaveAttribute("title", "Every instance of the plant already has a real board.");
  });
});
