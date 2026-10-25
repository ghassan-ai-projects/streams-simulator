import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it } from "vitest";
import { stubFlowDom } from "../../../test/dom";
import { mountLivePlant } from "../../../test/livePlantHarness";

beforeEach(() => {
  stubFlowDom();
});

function pin(entity: string, channel: string) {
  const outline = screen.getByRole("navigation", { name: "Outline" });
  fireEvent.click(within(outline).getByRole("button", { name: new RegExp(`^${entity}`) }));
  const inspector = screen.getByRole("complementary", { name: "Inspector" });
  fireEvent.click(within(inspector).getByRole("button", { name: `Watch ${channel}` }));
}

async function trends(pins: readonly (readonly [string, string])[]) {
  const live = await mountLivePlant((review, worldId) => {
    review.watch(worldId);
  });
  live.feed(0, 600);
  pins.forEach(([entity, channel]) => {
    pin(entity, channel);
  });
  fireEvent.click(screen.getByRole("button", { name: /^Trends/ }));
  return screen.getByRole("region", { name: "Trends" });
}

describe("the height of the trends panel", () => {
  it("is normal with two channels pinned and taller by default from three", async () => {
    const panel = await trends([
      ["CT-1", "basin temp"],
      ["CH-1", "power"],
    ]);
    expect(panel).toHaveAttribute("data-size", "normal");
    pin("P-1", "flow");
    expect(panel).toHaveAttribute("data-size", "tall");
  });

  it("steps taller and shorter from its own buttons, remembers the choice, and stops at the ends", async () => {
    const panel = await trends([["CT-1", "basin temp"]]);
    const taller = within(panel).getByRole("button", { name: "Make the trends panel taller" });
    const shorter = within(panel).getByRole("button", { name: "Make the trends panel shorter" });
    fireEvent.click(taller);
    expect(panel).toHaveAttribute("data-size", "tall");
    fireEvent.click(taller);
    expect(panel).toHaveAttribute("data-size", "full");
    fireEvent.click(taller);
    expect(panel).toHaveAttribute("data-size", "full");
    expect(taller).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(shorter);
    fireEvent.click(shorter);
    fireEvent.click(shorter);
    fireEvent.click(shorter);
    expect(panel).toHaveAttribute("data-size", "compact");
    expect(shorter).toHaveAttribute("aria-disabled", "true");
    pin("CH-1", "power");
    pin("P-1", "flow");
    expect(panel).toHaveAttribute("data-size", "compact");
  });

  it("has no accessibility violation with its size buttons", async () => {
    const panel = await trends([["CT-1", "basin temp"]]);
    expect((await axe.run(panel)).violations).toEqual([]);
  });
});
