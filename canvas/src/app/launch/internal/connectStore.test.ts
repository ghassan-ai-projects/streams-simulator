import { describe, expect, it, vi } from "vitest";
import { failure, success } from "../../ports";
import { firstSessionDraft } from "../../../model/launch";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { ConnectStore } from "./connectStore";

const draft = { ...firstSessionDraft, hostSocket: "/tmp/h.sock", gatewaySocket: "/tmp/g.sock" };
const check = {
  domain: "arduino-thermal",
  environment: "mixed",
  instances: [{ instance: "zone-01", origin: "real", kind: "thermal_sensor", board: "uno-07", device: "zone-01" }],
};

function rig() {
  const bridge = new FakeBridge();
  const timers = fakeClock();
  bridge.actions.set("session.check", success(check));
  bridge.actions.set(
    "session.start",
    success({ run: "c-1", state: "running", domain: "arduino-thermal", plan: check }),
  );
  bridge.actions.set(
    "session.status",
    success({ run: "c-1", state: "running", domain: "arduino-thermal", plan: check }),
  );
  bridge.actions.set("session.stop", success({ run: "c-1", state: "ended", domain: "arduino-thermal" }));
  return { bridge, timers, connect: new ConnectStore(bridge.launch, timers) };
}

describe("a connected session", () => {
  it("checks the plan through the binary and shows each instance's role", async () => {
    const { bridge, connect } = rig();
    const checked = await connect.check("arduino-thermal", draft);
    expect(checked?.instances[0]).toMatchObject({ instance: "zone-01", role: "real", board: "uno-07" });
    expect(connect.snapshot()).toMatchObject({ phase: "checked", domain: "arduino-thermal" });
    expect(JSON.parse(String(bridge.calls[0]?.args.content))).toMatchObject({
      domain: "arduino-thermal",
      gateway_socket: "/tmp/g.sock",
    });
  });

  it("keeps the binary's issues when the plan does not hold and starts nothing", async () => {
    const { bridge, connect } = rig();
    const issue = { pointer: "/real/0/device", message: "no device", keyword: "device_missing" };
    bridge.actions.set(
      "session.check",
      failure({ ...cannedFailure("plantwire_failed", 422, "the plan does not hold"), issues: [issue] }),
    );
    await connect.start("arduino-thermal", draft);
    expect(connect.snapshot()).toMatchObject({ phase: "refused", issues: [issue] });
    expect(bridge.calls.map((call) => call.name)).toEqual(["session.check"]);
  });

  it("starts after the check, follows the binding the binary journals, and stops into a saved run", async () => {
    const { bridge, connect } = rig();
    await connect.start("arduino-thermal", draft);
    expect(connect.snapshot()).toMatchObject({ phase: "live", run: "c-1", binding: null, stopping: false, note: null });
    bridge.stream(
      "c-1",
      [{ stream: "observer", seq: 1, type: "binding", at_ns: 0, outcome: "matched", issues: [] }],
      null,
    );
    expect(connect.snapshot()).toMatchObject({ binding: { matched: true } });
    const stopping = connect.stop();
    expect(connect.snapshot()).toMatchObject({ stopping: true });
    await stopping;
    expect(connect.snapshot()).toEqual({ phase: "ended", domain: "arduino-thermal", run: "c-1" });
    expect(bridge.followers.has("c-1")).toBe(false);
  });

  it("stops following the run once its binding is known, so the review's stream is the only one left", async () => {
    const { bridge, connect, timers } = rig();
    await connect.start("arduino-thermal", draft);
    expect(bridge.openFollows("c-1")).toBe(1);
    bridge.stream(
      "c-1",
      [{ stream: "observer", seq: 1, type: "binding", at_ns: 0, outcome: "matched", issues: [] }],
      null,
    );
    expect(connect.snapshot()).toMatchObject({ binding: { matched: true } });
    expect(bridge.openFollows("c-1")).toBe(0);
    timers.advance(2000);
    await Promise.resolve();
    await Promise.resolve();
    expect(bridge.openFollows("c-1")).toBe(0);
  });

  it("notices a session that died while it was watched and says why", async () => {
    const { bridge, connect, timers } = rig();
    await connect.start("arduino-thermal", draft);
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(timers.pending()).toBe(1);
    });
    expect(connect.snapshot()).toMatchObject({ phase: "live" });
    const died = { run: "c-1", state: "failed", failure: { code: "plantwire_failed", message: "host lost" } };
    bridge.actions.set("session.status", success(died));
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(connect.snapshot()).toMatchObject({ phase: "failed", run: "c-1", diagnostic: { message: "host lost" } });
    });
    expect(timers.pending()).toBe(0);
  });

  it("refuses a second session while one runs, and picks up a running one after a reload with its own domain and plan", async () => {
    const { bridge, connect } = rig();
    bridge.actions.set("session.start", failure(cannedFailure("session_running", 422, "still running")));
    await connect.start("arduino-thermal", draft);
    expect(connect.snapshot()).toMatchObject({ phase: "failed", run: null, diagnostic: { message: "still running" } });
    connect.reset();
    expect(connect.snapshot()).toEqual({ phase: "idle" });
    await connect.resume();
    expect(connect.snapshot()).toMatchObject({
      phase: "live",
      run: "c-1",
      domain: "arduino-thermal",
      check: { mixed: true, instances: [{ instance: "zone-01", role: "real" }] },
    });
    await connect.resume();
    expect(await connect.check("arduino-thermal", draft)).toBeNull();
    connect.reset();
    expect(connect.snapshot()).toMatchObject({ phase: "live" });
  });

  it("picks up a session another plant started and one that is still starting", async () => {
    const { bridge, connect, timers } = rig();
    bridge.actions.set("session.status", success({ run: "c-4", state: "starting", domain: "other" }));
    await connect.resume();
    expect(connect.snapshot()).toEqual({ phase: "starting", domain: "other", check: null });
    bridge.actions.set("session.status", success({ run: "c-4", state: "running", domain: "other" }));
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(connect.snapshot()).toMatchObject({ phase: "live", domain: "other", run: "c-4" });
    });
    expect(bridge.followers.has("c-4")).toBe(true);
  });

  it("does nothing on stop without a live session, and resumes nothing when none runs", async () => {
    const { bridge, connect } = rig();
    await connect.stop();
    bridge.actions.set("session.status", success({ state: "none" }));
    await connect.resume();
    expect(connect.snapshot()).toEqual({ phase: "idle" });
    expect(bridge.calls.map((call) => call.name)).toEqual(["session.status"]);
  });

  it("keeps a session it could not stop live with a note, still watched, and stops it on a retry", async () => {
    const { bridge, connect, timers } = rig();
    await connect.start("arduino-thermal", draft);
    bridge.actions.set("session.stop", failure(cannedFailure("bridge_failed", 500, "boom")));
    await connect.stop();
    expect(connect.snapshot()).toMatchObject({
      phase: "live",
      run: "c-1",
      stopping: false,
      note: "The session could not be stopped: boom. It is still running; try Stop again.",
    });
    expect(bridge.followers.has("c-1")).toBe(true);
    expect(timers.pending()).toBe(1);
    bridge.actions.set("session.stop", success({ run: "c-1", state: "ended", domain: "arduino-thermal" }));
    await connect.stop();
    expect(connect.snapshot()).toEqual({ phase: "ended", domain: "arduino-thermal", run: "c-1" });
    expect(timers.pending()).toBe(0);
  });

  it("notes a journal stream that closed or failed and follows it again on the next poll", async () => {
    const { bridge, connect, timers } = rig();
    await connect.start("arduino-thermal", draft);
    bridge.stream("c-1", [], "failed");
    expect(connect.snapshot()).toMatchObject({ note: "The run journal could not be followed; trying again." });
    expect(bridge.followers.has("c-1")).toBe(false);
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(bridge.followers.has("c-1")).toBe(true);
    });
    bridge.stream("c-1", [], "closed");
    expect(connect.snapshot()).toMatchObject({ note: "The run journal stopped streaming; following it again." });
  });

  it("shows a session another tab is stopping as stopping and ends it when the bridge says so", async () => {
    const { bridge, connect, timers } = rig();
    await connect.start("arduino-thermal", draft);
    bridge.actions.set("session.status", success({ run: "c-1", state: "stopping", domain: "arduino-thermal" }));
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(connect.snapshot()).toMatchObject({ phase: "live", stopping: true });
    });
    bridge.actions.set("session.status", success({ run: "c-1", state: "ended", domain: "arduino-thermal" }));
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(connect.snapshot()).toMatchObject({ phase: "ended" });
    });
  });

  it("learns from the poll that the node opened its control socket, and when it is gone again", async () => {
    const { bridge, connect, timers } = rig();
    await connect.start("arduino-thermal", draft);
    expect(connect.snapshot()).toMatchObject({ phase: "live", controls: false });
    const status = (controls: boolean) =>
      success({ run: "c-1", state: "running", domain: "arduino-thermal", controls });
    bridge.actions.set("session.status", status(true));
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(connect.snapshot()).toMatchObject({ phase: "live", controls: true, stopping: false });
    });
    bridge.actions.set("session.status", status(false));
    timers.advance(2000);
    await vi.waitFor(() => {
      expect(connect.snapshot()).toMatchObject({ phase: "live", controls: false });
    });
  });

  it("holds each plant's draft and marks its check out of date when the draft changes", async () => {
    const { connect } = rig();
    expect(connect.draftFor("arduino-thermal")).toEqual(firstSessionDraft);
    await connect.check("arduino-thermal", draft);
    connect.keepDraft("arduino-thermal", { ...draft, seed: "2" });
    expect(connect.draftFor("arduino-thermal")).toMatchObject({ seed: "2" });
    expect(connect.draftFor("other")).toEqual(firstSessionDraft);
    expect(connect.snapshot()).toMatchObject({ phase: "checked", stale: true });
  });
});
