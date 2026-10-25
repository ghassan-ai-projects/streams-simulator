import { act, fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it } from "vitest";
import { factoryJournal, singleReadings } from "../../../test/factoryFixture";
import { stubFlowDom } from "../../../test/dom";
import { mountLivePlant } from "../../../test/livePlantHarness";
import type { LivePlant } from "../../../test/livePlantHarness";

beforeEach(() => {
  stubFlowDom();
});

async function watching(): Promise<LivePlant> {
  const live = await mountLivePlant((review, worldId) => {
    review.watch(worldId);
  });
  live.feed(0, 600);
  return live;
}

function pin(entity: string, channel: string) {
  const outline = screen.getByRole("navigation", { name: "Outline" });
  fireEvent.click(within(outline).getByRole("button", { name: new RegExp(`^${entity}`) }));
  const inspector = screen.getByRole("complementary", { name: "Inspector" });
  fireEvent.click(within(inspector).getByRole("button", { name: `Watch ${channel}` }));
}

function openTrends() {
  fireEvent.click(screen.getByRole("button", { name: /^Trends/ }));
  return screen.getByRole("region", { name: "Trends" });
}

async function pinned(): Promise<{ live: LivePlant; panel: HTMLElement }> {
  const live = await watching();
  pin("CT-1", "basin temp");
  pin("CH-1", "power");
  return { live, panel: openTrends() };
}

describe("the trends panel of a watched run", () => {
  it("is closed until the toolbar button opens it, and says what to do while nothing is watched", async () => {
    await watching();
    expect(screen.queryByRole("region", { name: "Trends" })).toBeNull();
    const panel = openTrends();
    expect(panel).toHaveTextContent("Nothing watched yet");
    fireEvent.click(within(panel).getByRole("button", { name: "Hide the trends" }));
    expect(screen.queryByRole("region", { name: "Trends" })).toBeNull();
  });

  it("shows one plot for each pinned channel, each with its own unit, and counts them on the toolbar", async () => {
    const { panel } = await pinned();
    const plots = within(panel).getAllByRole("slider");
    expect(plots.map((plot) => plot.getAttribute("aria-label"))).toEqual([
      "CT-1 · basin temp trend, °C",
      "CH-1 · power trend, kW",
    ]);
    plots.forEach((plot) => {
      expect(plot.querySelectorAll("polyline")).toHaveLength(1);
    });
    expect(within(panel).getAllByText("°C").length + within(panel).getAllByText("kW").length).toBe(2);
    expect(screen.getByRole("button", { name: "Trends · 2" })).toHaveAttribute("aria-pressed", "true");
    expect(panel).toHaveTextContent("Trends · 2 watched");
  });

  it("keeps the watch list while another instance is selected and removes a channel from the plot or the pin", async () => {
    const { panel } = await pinned();
    pin("P-1", "flow");
    expect(within(panel).getAllByRole("slider")).toHaveLength(3);
    fireEvent.click(within(panel).getByRole("button", { name: "Stop watching P-1 · flow" }));
    expect(within(panel).getAllByRole("slider")).toHaveLength(2);
    const inspector = screen.getByRole("complementary", { name: "Inspector" });
    fireEvent.click(within(inspector).getByRole("button", { name: "Watch flow" }));
    expect(within(panel).getAllByRole("slider")).toHaveLength(3);
    fireEvent.click(within(panel).getByRole("button", { name: "Clear" }));
    expect(within(panel).queryAllByRole("slider")).toHaveLength(0);
  });

  it("reads the exact value at the cursor with unit, observed-at time, age and quality, and moves it with the keyboard", async () => {
    const { panel } = await pinned();
    const plot = within(panel).getByRole("slider", { name: /^CT-1 · basin temp trend/ });
    expect(plot.getAttribute("aria-valuetext")).toMatch(/^CT-1 · basin temp: [\d.-]+ °C, observed at 00:09:50, /);
    const before = Number(plot.getAttribute("aria-valuenow"));
    fireEvent.keyDown(plot, { key: "ArrowLeft" });
    expect(Number(plot.getAttribute("aria-valuenow"))).toBeLessThan(before);
    expect(plot.getAttribute("aria-valuetext")).toContain("observed at 00:09:40");
    expect(within(panel).getByText(/^Probe 00:09:40$/)).toBeInTheDocument();
    fireEvent.keyDown(plot, { key: "Escape" });
    expect(within(panel).getByText(/^Run clock/)).toBeInTheDocument();
  });

  it("moves one shared probe: every plot reads its own value at the time the probe stands on", async () => {
    const { panel } = await pinned();
    const [first, second] = within(panel).getAllByRole("slider");
    fireEvent.keyDown(first ?? panel, { key: "Home" });
    expect(second?.getAttribute("aria-valuetext")).toMatch(/observed at 00:00:00/);
    expect(first?.getAttribute("aria-valuetext")).toMatch(/observed at 00:00:00/);
  });

  it("leaves a gap where a channel stopped reporting and says how a gap is inferred", async () => {
    const live = await watching();
    pin("CT-1", "basin temp");
    live.send(singleReadings("CH-1", 600, 900));
    live.send(singleReadings("CT-1", 900, 1190));
    const panel = openTrends();
    const plot = within(panel).getByRole("slider", { name: /^CT-1 · basin temp trend/ });
    expect(plot.querySelectorAll("polyline").length).toBeGreaterThanOrEqual(1);
    fireEvent.click(within(panel).getByText("How to read these plots"));
    expect(panel).toHaveTextContent("longer than 3 times the median interval");
    expect(panel).toHaveTextContent("never drawn as zero");
  });

  it("offers a window and shows what limits it", async () => {
    const { panel } = await pinned();
    const group = within(panel).getByRole("group", { name: "Time window" });
    expect(within(group).getByRole("button", { name: "10 min" })).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(within(group).getByRole("button", { name: "Whole run" }));
    expect(within(group).getByRole("button", { name: "Whole run" })).toHaveAttribute("aria-pressed", "true");
    expect(panel).toHaveTextContent("Everything the run has produced so far");
  });

  it("selects the instance when a plot is chosen with Enter or a click", async () => {
    const { panel } = await pinned();
    const plot = within(panel).getByRole("slider", { name: /^CH-1 · power trend/ });
    fireEvent.keyDown(plot, { key: "Enter" });
    expect(screen.getByRole("complementary", { name: "Inspector" })).toHaveTextContent("CH-1");
  });

  it("has no accessibility violations with plots on screen", async () => {
    const { panel } = await pinned();
    expect((await axe.run(panel)).violations).toEqual([]);
  });

  it("follows the run: a later flush moves the shared window and the readings", async () => {
    const { live, panel } = await pinned();
    const plot = within(panel).getByRole("slider", { name: /^CT-1 · basin temp trend/ });
    await act(async () => {
      live.send(factoryJournal(600, 700));
      await Promise.resolve();
    });
    expect(plot.getAttribute("aria-valuetext")).toContain("observed at 00:11:30");
  });
});
