import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { journalWire } from "../../../test/runFixture";
import { reviewBridge, reviewHash } from "../../../test/screenFixture";

beforeEach(() => {
  stubFlowDom();
});

function origin(instance: string, environment: string, board: string, bootId: string) {
  return {
    stream: "observer",
    type: "origin",
    seq: 1,
    at_ns: 0,
    instance,
    device: instance,
    board,
    environment,
    boot_id: bootId,
  };
}

async function sessionReview() {
  const bridge = reviewBridge();
  const app = mountApp(bridge, reviewHash());
  await screen.findByRole("region", { name: "Run w-1" });
  const [opened, ...rest] = journalWire;
  act(() => {
    bridge.stream("w-1", [
      opened ?? {},
      origin("CH-1", "hardware", "uno-07", "61"),
      origin("CT-1", "simulated", "mega-01", "mega-01-boot-1"),
      ...rest,
    ]);
    app.timers.advance(250);
  });
  return app;
}

describe("a reviewed connected session", () => {
  it("marks the real board and the simulated instance in the outline and on the timeline", async () => {
    await sessionReview();
    const outline = screen.getByRole("navigation", { name: "Outline" });
    expect(within(outline).getByRole("button", { name: /^CH-1.*Real · board uno-07 · boot 61$/ })).toBeInTheDocument();
    expect(within(outline).getByRole("button", { name: /^CT-1.*Simulated · board mega-01$/ })).toBeInTheDocument();
    const events = await screen.findByRole("list", { name: "Events" });
    expect(events.querySelector("[data-origin='real']")).not.toBeNull();
    expect(events.querySelector("[data-origin='simulated']")).not.toBeNull();
  });

  it("heads the inspector with the board and boot and never shows the twin's values as the board's", async () => {
    await sessionReview();
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /^CH-1/ }));
    const observed = screen.getByRole("region", { name: "Observed · CH-1" });
    expect(within(observed).getByText("Real · board uno-07 · boot 61", { selector: "p" })).toBeInTheDocument();
    expect(within(observed).getByText(/This instance is a real board/)).toBeInTheDocument();
    expect(within(observed).queryByText("95.1")).not.toBeInTheDocument();
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /^CT-1/ }));
    expect(
      within(screen.getByRole("region", { name: "Observed · CT-1" })).getByText("Simulated · board mega-01", {
        selector: "p",
      }),
    ).toBeInTheDocument();
  });
});
