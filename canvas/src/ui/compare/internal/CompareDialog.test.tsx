import { act, fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it } from "vitest";
import { stubFlowDom } from "../../../test/dom";
import { mountApp } from "../../../test/appHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { reportWire } from "../../../test/runFixture";
import { miniWorkspace, reviewHash } from "../../../test/screenFixture";
import { bridgeWith, reviewing, run } from "../../../test/reviewActionsHarness";

beforeEach(() => {
  stubFlowDom();
});

describe("Compare with…", () => {
  async function opened() {
    const app = await reviewing();
    const opener = screen.getByRole("button", { name: "Compare with…" });
    opener.focus();
    fireEvent.click(opener);
    const dialog = await screen.findByRole("dialog", { name: "Compare with another run" });
    return { ...app, dialog, opener };
  }

  it("lists the other ended runs of the same plant and names the two roles", async () => {
    const { dialog } = await opened();
    const choices = within(dialog).getByRole("list", { name: "Runs to use as the reference" });
    expect(
      within(choices)
        .getAllByRole("button")
        .map((button) => button.textContent),
    ).toEqual([expect.stringContaining("Baseline")]);
    expect(choices).toHaveTextContent("w-3 · Explore · seed 11 · 30 min · complete");
    expect(dialog).toHaveTextContent("Reference (A)");
    expect(dialog).toHaveTextContent("This run (B)");
    expect(dialog).not.toHaveTextContent("w-8");
  });

  it("compares with a is the reference and b is this run, then shows what differs, the limits and the numbers", async () => {
    const { dialog, bridge } = await opened();
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    await within(dialog).findByRole("region", { name: "What differs" });
    expect(within(dialog).queryByRole("list", { name: "Runs to use as the reference" })).toBeNull();
    expect(bridge.calls.at(-1)).toEqual({ name: "run.compare", args: { a: "w-3", b: "w-1" } });
    expect(within(dialog).getByRole("region", { name: "Effector calls" })).toHaveTextContent(
      "2 effector calls are the same in both runs",
    );
    const limits = within(dialog).getByRole("region", { name: "Comparability" });
    expect(limits.querySelector("[data-code='single_run_per_side']")).toHaveTextContent("seed noise");
    const sections = within(dialog)
      .getAllByRole("region")
      .map((section) => section.getAttribute("aria-label"));
    expect(sections.indexOf("Comparability")).toBeLessThan(sections.indexOf("Whole-run results"));
    expect(sections.indexOf("What differs")).toBeLessThan(sections.indexOf("Comparability"));
    const kpis = within(dialog).getByRole("region", { name: "KPIs" });
    expect(kpis).toHaveTextContent("Whole run: reference 30:00, this run 28:00 simulated");
    expect(kpis).toHaveTextContent("−0.69");
    expect(kpis).toHaveTextContent("−13.6 %");
    expect(dialog).toHaveTextContent("Event-window comparisons are not available");
  });

  it("goes back to the list of references on request", async () => {
    const { dialog, stores } = await opened();
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    await within(dialog).findByRole("region", { name: "What differs" });
    fireEvent.click(within(dialog).getByRole("button", { name: "Choose another reference" }));
    expect(stores.review.compare.snapshot()).toEqual({ phase: "idle" });
    expect(within(dialog).getByRole("list", { name: "Runs to use as the reference" })).toBeInTheDocument();
  });

  it("never ranks the runs or colours a difference as good or bad", async () => {
    const { dialog } = await opened();
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    await within(dialog).findByRole("region", { name: "Whole-run results" });
    expect(dialog.textContent).not.toMatch(/better|worse|winner|best|improv|degrad|score/i);
  });

  it("shows the binary's refusal word for word", async () => {
    const { dialog, bridge } = await opened();
    const message = "run B has not ended: its truth is still sealed";
    bridge.handlers.set("run.compare", () => failure(cannedFailure("run_not_ended", 422, message)));
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    expect(await within(dialog).findByRole("alert")).toHaveTextContent(message);
  });

  it("closes on Escape, clears the comparison and returns focus to the button that opened it", async () => {
    const { dialog, opener, stores } = await opened();
    fireEvent.keyDown(dialog, { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(document.activeElement).toBe(opener);
    expect(stores.review.compare.snapshot()).toEqual({ phase: "idle" });
  });

  it("still closes on Escape after choosing a reference replaced the button that had focus", async () => {
    const { dialog, opener } = await opened();
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    await within(dialog).findByRole("region", { name: "What differs" });
    expect(document.activeElement).toBe(dialog);
    fireEvent.keyDown(document.body, { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(document.activeElement).toBe(opener);
  });

  it("keeps focus inside while open, wrapping Tab and Shift+Tab", async () => {
    const { dialog } = await opened();
    expect(document.activeElement).toBe(dialog);
    const buttons = within(dialog).getAllByRole("button");
    const last = buttons.at(-1);
    last?.focus();
    fireEvent.keyDown(last ?? dialog, { key: "Tab" });
    expect(document.activeElement).toBe(buttons[0]);
    fireEvent.keyDown(buttons[0] ?? dialog, { key: "Tab", shiftKey: true });
    expect(document.activeElement).toBe(last);
  });

  it("offers a way out when there is nothing to compare with", async () => {
    const { bridge } = bridgeWith();
    bridge.workspaceOutcome = success({ ...miniWorkspace, runs: [run("w-1")] });
    bridge.handlers.set("run.report", () => success(reportWire));
    mountApp(bridge, reviewHash());
    await screen.findByRole("region", { name: "This run" });
    fireEvent.click(screen.getByRole("button", { name: "Compare with…" }));
    expect(await screen.findByText("No other ended run of this plant yet.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Close comparison" }));
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("has no accessibility violation with a comparison on screen", async () => {
    const { dialog } = await opened();
    await act(async () => {
      fireEvent.click(within(dialog).getByRole("button", { name: /Baseline/ }));
      await Promise.resolve();
    });
    await within(dialog).findByRole("region", { name: "Whole-run results" });
    expect((await axe.run(dialog)).violations).toEqual([]);
  });
});
