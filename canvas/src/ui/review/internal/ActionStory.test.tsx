import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { stubScrollIntoView } from "../../../test/dom";
import { report } from "../../../test/runFixture";
import { Events } from "./ReportEvents";

function actions() {
  return { pick: vi.fn(), seek: vi.fn() };
}

function mount(picked: { atNs: number; entity: string } | null = null, calls = actions()) {
  stubScrollIntoView();
  const view = render(<Events report={report} picked={picked} actions={calls} />);
  return { ...view, calls };
}

const startCall = /^Open 00:20:00 CH-1 start/;

describe("the action story of an effector call", () => {
  it("opens nothing until a call is selected", () => {
    mount();
    expect(screen.queryByRole("region", { name: "Action story" })).toBeNull();
  });

  it("selects the call, moves the cursor to it and shows the exact request and the reported outcome", () => {
    const { calls } = mount();
    fireEvent.click(screen.getByRole("button", { name: startCall }));
    expect(calls.pick).toHaveBeenCalledWith(1_200e9, "CH-1");
    const story = screen.getByRole("region", { name: "Action story" });
    expect(story).toHaveTextContent("start on CH-1");
    expect(story).toHaveTextContent("chiller_id = CH-1");
    expect(story).toHaveTextContent("level = 0.5");
    expect(story).toHaveTextContent("00:20:00");
    expect(story).toHaveTextContent("applied · 300 ms");
    expect(document.querySelector("tr[aria-current='true']")).toHaveTextContent("00:20:00");
  });

  it("marks the selected row and shows a refusal with its reason", () => {
    mount();
    fireEvent.click(screen.getByRole("button", { name: /^Open 00:21:40 CH-1 stop/ }));
    expect(screen.getByRole("region", { name: "Action story" })).toHaveTextContent("refused: interlock");
    const marked = document.querySelectorAll("tr[aria-current='true']");
    expect(marked).toHaveLength(1);
    expect(marked[0]).toHaveTextContent("00:21:40");
  });

  it("says a call the operator gave during the run was applied by the operator, or requested if refused", () => {
    stubScrollIntoView();
    const calls = report.effectorCalls.slice(0, 2).map((call) => ({ ...call, immediate: true }));
    render(<Events report={{ ...report, effectorCalls: calls }} picked={null} actions={actions()} />);
    fireEvent.click(screen.getByRole("button", { name: startCall }));
    expect(screen.getByRole("region", { name: "Action story" })).toHaveTextContent(
      "Applied by the operator during the run",
    );
    fireEvent.click(screen.getByRole("button", { name: /^Open 00:21:40 CH-1 stop/ }));
    expect(screen.getByRole("region", { name: "Action story" })).toHaveTextContent(
      "Requested by the operator during the run",
    );
  });

  it("moves the cursor a minute before or after the call, and keeps the story open", () => {
    const { calls } = mount();
    fireEvent.click(screen.getByRole("button", { name: startCall }));
    fireEvent.click(screen.getByRole("button", { name: "Show 1 min before" }));
    expect(calls.seek).toHaveBeenLastCalledWith(1_140e9);
    fireEvent.click(screen.getByRole("button", { name: "Show 1 min after" }));
    expect(calls.seek).toHaveBeenLastCalledWith(1_260e9);
    expect(screen.getByRole("region", { name: "Action story" })).toBeInTheDocument();
  });

  it("keeps the story when the cursor moves away from the call, and hides it on request", () => {
    const calls = actions();
    const { rerender } = mount({ atNs: 1_200e9, entity: "CH-1" }, calls);
    expect(screen.getByRole("region", { name: "Action story" })).toBeInTheDocument();
    rerender(<Events report={report} picked={null} actions={calls} />);
    expect(screen.getByRole("region", { name: "Action story" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Hide" }));
    expect(screen.queryByRole("region", { name: "Action story" })).toBeNull();
  });

  it("selects the call from anywhere in its row, not only from the time", () => {
    const { calls } = mount();
    const row = screen.getByRole("button", { name: startCall }).closest("tr");
    expect(row).not.toBeNull();
    fireEvent.click(within(row as HTMLElement).getByText("start"));
    expect(calls.pick).toHaveBeenCalledTimes(1);
    expect(calls.pick).toHaveBeenCalledWith(1_200e9, "CH-1");
    expect(screen.getByRole("region", { name: "Action story" })).toBeInTheDocument();
  });

  it("moves the cursor to a disturbance from anywhere in its row and shows no call story", () => {
    const { calls } = mount();
    const row = screen.getByRole("button", { name: /^Open 00:10:00 CT-1/ }).closest("tr");
    fireEvent.click(within(row as HTMLElement).getByText("CT-1"));
    expect(calls.pick).toHaveBeenCalledWith(600e9, "CT-1");
    expect(screen.queryByRole("region", { name: "Action story" })).toBeNull();
  });

  it("says when each disturbance ends, or that it did not end before the run did", () => {
    const ending = report.disturbances.map((d) =>
      d.name === "tower__fan_failure"
        ? { ...d, clearedNs: 840e9, clearedBy: "fault.clear" }
        : d.name === "drop"
          ? { ...d, untilNs: 480e9 }
          : d,
    );
    render(<Events report={{ ...report, disturbances: ending }} picked={null} actions={actions()} />);
    const table = screen.getByRole("region", { name: "Disturbances" });
    expect(within(table).getByRole("columnheader", { name: "Ends" })).toBeInTheDocument();
    const cells = (id: RegExp) => within(within(table).getByRole("button", { name: id }).closest("tr") as HTMLElement);
    expect(cells(/^Open 00:10:00 CT-1/).getByText("00:14:00")).toHaveAttribute(
      "title",
      "Ends at 00:14:00 by fault.clear",
    );
    expect(cells(/^Open 00:05:00 delivery/).getByText("00:08:00")).toBeInTheDocument();
    expect(cells(/^Open 00:15:00 weather/).getByText("–")).toHaveAttribute("title", "Did not end before the run ended");
  });

  it("opens the story of a call picked on the timeline, and of no disturbance", () => {
    const { rerender } = mount({ atNs: 600e9, entity: "CT-1" });
    expect(screen.queryByRole("region", { name: "Action story" })).toBeNull();
    rerender(<Events report={report} picked={{ atNs: 1_200e9, entity: "CH-1" }} actions={actions()} />);
    expect(screen.getByRole("region", { name: "Action story" })).toBeInTheDocument();
  });

  it("makes no claim about what a call changed", () => {
    mount();
    fireEvent.click(screen.getByRole("button", { name: startCall }));
    const text = screen.getByRole("region", { name: "Action story" }).textContent;
    expect(text).not.toMatch(/fixed|caused|improved|resolved|because/i);
    expect(text).toContain("does not say what the call changed");
  });

  it("is operable by keyboard and has no accessibility violation", async () => {
    const { container } = mount();
    const button = screen.getByRole("button", { name: startCall });
    button.focus();
    expect(document.activeElement).toBe(button);
    fireEvent.click(button);
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it("offers no selection where no actions are given", () => {
    render(<Events report={report} picked={null} />);
    expect(screen.queryByRole("button", { name: startCall })).toBeNull();
  });
});
