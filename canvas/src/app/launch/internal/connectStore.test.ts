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
  bridge.actions.set("session.start", success({ run: "c-1", state: "running" }));
  bridge.actions.set("session.status", success({ run: "c-1", state: "running" }));
  bridge.actions.set("session.stop", success({ run: "c-1", state: "ended", result: { out: "runs/c-1" } }));
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
    expect(connect.snapshot()).toMatchObject({ phase: "live", run: "c-1", binding: null, stopping: false });
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

  it("refuses a second session while one runs, and picks up a running one after a reload", async () => {
    const { bridge, connect } = rig();
    bridge.actions.set("session.start", failure(cannedFailure("session_running", 422, "still running")));
    await connect.start("arduino-thermal", draft);
    expect(connect.snapshot()).toMatchObject({ phase: "failed", run: null, diagnostic: { message: "still running" } });
    connect.reset();
    expect(connect.snapshot()).toEqual({ phase: "idle" });
    await connect.resume("arduino-thermal");
    expect(connect.snapshot()).toMatchObject({ phase: "live", run: "c-1", check: null });
    await connect.resume("arduino-thermal");
    expect(await connect.check("arduino-thermal", draft)).toBeNull();
    connect.reset();
    expect(connect.snapshot()).toMatchObject({ phase: "live" });
  });

  it("does nothing on stop without a live session, and resumes nothing when none runs", async () => {
    const { bridge, connect } = rig();
    await connect.stop();
    bridge.actions.set("session.status", success({ state: "none" }));
    await connect.resume("arduino-thermal");
    expect(connect.snapshot()).toEqual({ phase: "idle" });
    expect(bridge.calls.map((call) => call.name)).toEqual(["session.status"]);
  });

  it("reports a stop the bridge could not carry out", async () => {
    const { bridge, connect } = rig();
    await connect.start("arduino-thermal", draft);
    bridge.actions.set("session.stop", failure(cannedFailure("bridge_failed", 500, "boom")));
    await connect.stop();
    expect(connect.snapshot()).toMatchObject({ phase: "failed", diagnostic: { message: "boom" } });
  });
});
