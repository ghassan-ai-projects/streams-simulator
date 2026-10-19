import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { ReviewSession } from "../../../app/review";
import { createReviewSession } from "../../../app/review";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { controlReplyWire, interruptsWire, liveWire } from "../../../test/liveFixture";
import { success } from "../../../test/outcome";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { useStore } from "../../primitives";
import { InterruptLine } from "./InterruptLine";

function Line({ review }: { readonly review: ReviewSession }) {
  const state = useStore(review);
  return <InterruptLine review={review} state={state} entity="zone-01" />;
}

describe("the interrupt line of an instance", () => {
  it("follows the state the node answered to a control, as the Interrupts bar does, before a snapshot repeats it", async () => {
    const bridge = new FakeBridge();
    const timers = fakeClock();
    const review = createReviewSession(bridge.review, timers, timers);
    const known = sampleWorkspace.runs.find((run) => run.worldId === "w-2");
    review.openLive("w-2", known);
    bridge.stream(
      "w-2",
      [{ ...liveWire(10e9), interrupts: interruptsWire({ holds: [], approve: [], devices: [] }) }],
      "caught_up",
    );
    timers.advance(250);
    const view = render(<Line review={review} />);
    expect(view.container).toBeEmptyDOMElement();
    const held = [{ device: "sensor", board: "uno-07", mode: "hold", latched: false }];
    bridge.actions.set("session.control", success(controlReplyWire({ state: interruptsWire({ devices: held }) })));
    await review.controls.hold({ kind: "board", id: "uno-07" });
    await waitFor(() => {
      expect(screen.getByText("Held")).toBeInTheDocument();
    });
  });
});
