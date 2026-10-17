import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { campusReviewBridge, reviewBridge, reviewHash, streamJournal } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

async function reviewing(bridge = reviewBridge(), node = /^CH-1, chiller/) {
  const app = mountApp(bridge, reviewHash());
  await screen.findByRole("region", { name: "Run w-1" });
  act(() => {
    streamJournal(bridge);
    app.timers.advance(250);
  });
  await screen.findByRole("group", { name: node });
  return { ...app, bridge };
}

describe("the review timeline", () => {
  it("plays and pauses from the button and says which it is doing", async () => {
    const app = await reviewing();
    fireEvent.click(screen.getByRole("button", { name: "Play" }));
    expect(app.stores.review.snapshot().playing).toBe(true);
    expect(screen.getByRole("button", { name: "Pause" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Pause" }));
    expect(app.stores.review.snapshot().playing).toBe(false);
  });

  it("changes the playback speed", async () => {
    const app = await reviewing();
    fireEvent.click(screen.getByRole("button", { name: "600×" }));
    expect(app.stores.review.snapshot().speed).toBe(600);
    expect(screen.getByRole("button", { name: "600×" })).toHaveAttribute("aria-pressed", "true");
  });

  it("seeks with the slider", async () => {
    const app = await reviewing();
    fireEvent.change(screen.getByRole("slider", { name: "Playback time" }), { target: { value: "120" } });
    expect(app.stores.review.snapshot().positionNs).toBe(120e9);
  });

  it("jumps to a disturbance and selects the instance it concerns", async () => {
    const app = await reviewing();
    fireEvent.click(screen.getByRole("button", { name: "00:10:00 CT-1 fan failure" }));
    expect(app.stores.review.snapshot()).toMatchObject({ positionNs: 600e9, selected: "CT-1" });
    expect(screen.getByRole("button", { name: "Inspect · CT-1" })).toBeInTheDocument();
  });
});

describe("a campus run", () => {
  const hallNode = /^h01, /;

  it("opens a hall on the graph and returns to the whole run from the crumb", async () => {
    const app = await reviewing(campusReviewBridge(), hallNode);
    fireEvent.doubleClick(screen.getByRole("group", { name: hallNode }));
    const trail = screen.getByRole("navigation", { name: "Breadcrumb" });
    expect(screen.getByRole("heading", { name: "h01" })).toBeInTheDocument();
    fireEvent.click(within(trail).getByRole("button", { name: "Run w-1" }));
    expect(screen.getByRole("heading", { name: "Run w-1" })).toBeInTheDocument();
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "review" });
  });

  it("explains a hall picked on the graph instead of showing readings", async () => {
    await reviewing(campusReviewBridge(), hallNode);
    fireEvent.click(screen.getByRole("group", { name: hallNode }));
    expect(
      await screen.findByText("Hall h01 holds 2 instances. Open it to see each one's readings."),
    ).toBeInTheDocument();
  });

  it("returns from a hall with Escape on the graph", async () => {
    await reviewing(campusReviewBridge(), hallNode);
    fireEvent.doubleClick(screen.getByRole("group", { name: hallNode }));
    const flow = screen.getByRole("application", { name: "Plant graph" });
    fireEvent.keyDown(flow, { key: "Escape" });
    expect(screen.getByRole("heading", { name: "Run w-1" })).toBeInTheDocument();
  });
});
