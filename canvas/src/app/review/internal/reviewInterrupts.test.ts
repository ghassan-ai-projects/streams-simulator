import { describe, expect, it } from "vitest";
import { success } from "../../ports";
import { flushPromises } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { controlReplyWire, interruptWire, interruptsWire, liveWire } from "../../../test/liveFixture";
import { reportWire } from "../../../test/runFixture";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { createReviewSession } from "./reviewSession";

function setup() {
  const bridge = new FakeBridge();
  bridge.actions.set("run.report", success(reportWire));
  const timers = fakeClock();
  const review = createReviewSession(bridge.review, timers, timers);
  const known = sampleWorkspace.runs.find((run) => run.worldId === "w-2");
  if (known === undefined) throw new Error("missing fixture");
  review.openLive("w-2", known);
  const stream = (...entries: object[]) => {
    bridge.stream("w-2", entries, "caught_up");
    timers.advance(250);
  };
  return { bridge, review, stream };
}

describe("interrupts in a live review", () => {
  it("marks the operator's controls and the node's decisions on the timeline, once each", () => {
    const { review, stream } = setup();
    stream(
      liveWire(10e9),
      interruptWire({ event: "control", op: "hold", scope: { kind: "board", id: "uno-07" } }, 11e9),
      interruptWire({ event: "verdict", command_id: "cmd-7", verdict: "queue" }, 12e9),
      interruptWire({ event: "resolve", command_id: "cmd-7", resolution: "forward", waiting: [] }, 13e9),
    );
    const marks = review.snapshot().liveMarks.filter((mark) => mark.kind === "interrupt");
    expect(marks.map((mark) => [mark.atNs, mark.label])).toEqual([
      [11e9, "Operator: Hold on board uno-07"],
      [13e9, "cmd-7 approved and handed on"],
    ]);
    expect(review.snapshot().liveHistory?.entries).toHaveLength(1);
  });

  it("does not clear the note about an unreadable snapshot, and notes a record it cannot read", () => {
    const { review, stream } = setup();
    stream({ ...liveWire(10e9), host: "lost" });
    expect(review.snapshot().liveUnreadableNs).toBe(10e9);
    stream(interruptWire({ event: "control", op: "hold", scope: { kind: "all", id: "" } }, 11e9));
    expect(review.snapshot().liveUnreadableNs).toBe(10e9);
    stream(liveWire(12e9));
    expect(review.snapshot().liveUnreadableNs).toBeNull();
    stream(interruptWire({ event: "teleport" }, 13e9));
    expect(review.snapshot().liveUnreadableNs).toBeNull();
    expect(review.snapshot().interruptUnreadableNs).toBe(13e9);
  });

  it("keeps a snapshot with an interrupt state as a state of its own at the cursor", () => {
    const { review, stream } = setup();
    stream({ ...liveWire(10e9), interrupts: interruptsWire() });
    expect(review.snapshot().live?.interrupts).toMatchObject({ holds: [{ latched: false }] });
    stream({ ...liveWire(11e9), interrupts: interruptsWire({ holds: [] }) });
    expect(review.snapshot().liveHistory?.entries).toHaveLength(2);
  });

  it("shows nothing of a finished session's controls to the next session", async () => {
    const { bridge, review, stream } = setup();
    bridge.actions.set("session.control", success(controlReplyWire()));
    stream({ ...liveWire(10e9), interrupts: interruptsWire({ holds: [] }) });
    review.controls.pick({ kind: "board", id: "uno-07" });
    await review.controls.hold({ kind: "board", id: "uno-07" });
    expect(review.controls.snapshot()).toMatchObject({ run: "w-2", seen: { run: "w-2" } });
    review.close();
    review.openLive("w-3");
    expect(review.controls.snapshot()).toMatchObject({
      run: "w-3",
      seen: null,
      last: null,
      sending: [],
      picked: { kind: "all", id: "" },
    });
  });

  it("keeps a session's controls while its screens come and go", async () => {
    const { bridge, review } = setup();
    bridge.actions.set("session.control", success(controlReplyWire()));
    review.controls.pick({ kind: "board", id: "uno-07" });
    await review.controls.hold({ kind: "board", id: "uno-07" });
    review.openLive("w-2");
    review.close();
    expect(review.controls.snapshot()).toMatchObject({ run: "w-2", picked: { kind: "board", id: "uno-07" } });
  });

  it("sends a control and remembers the newest snapshot the answer was read against", async () => {
    const { bridge, review, stream } = setup();
    bridge.actions.set("session.control", success(controlReplyWire()));
    stream({ ...liveWire(10e9), interrupts: interruptsWire({ holds: [] }) });
    await review.controls.hold({ kind: "board", id: "uno-07" });
    await flushPromises();
    const newest = review.snapshot().liveHistory?.entries.at(-1);
    expect(newest?.key).not.toBe("");
    expect(review.controls.snapshot().seen?.seenKey).toBe(newest?.key);
  });
});
