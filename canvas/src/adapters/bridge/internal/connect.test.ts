import { describe, expect, it, vi } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";

function answering(status: number, body: unknown) {
  return bridgeTransport(() => respond(status, body));
}

function sentTo(t: ReturnType<typeof bridgeTransport>): { url: unknown; body: unknown } {
  const [url, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return { url, body: typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined };
}

const check = {
  domain: "arduino-thermal",
  environment: "mixed",
  gateway: true,
  instances: [
    {
      instance: "zone-01",
      origin: "real",
      kind: "thermal_sensor",
      kind_digest: "sha256:a",
      board: "uno-07",
      device: "zone-01",
    },
    { instance: "fan-01", origin: "simulated", kind: "fan" },
  ],
  implements_checked: [],
  policies: [`sha256:${"b".repeat(64)}`, `sha256:${"a".repeat(64)}`],
};

describe("connect gateway", () => {
  it("checks a plan through session.check and reads each instance's role", async () => {
    const t = answering(200, { result: check });
    const outcome = await bridgeAdapter(t).launch.checkSession("{plan}");
    expect(sentTo(t)).toEqual({ url: "/api/actions/session.check", body: { args: { content: "{plan}" } } });
    expect(outcome).toEqual({
      ok: true,
      value: {
        domain: "arduino-thermal",
        mixed: true,
        policies: [`sha256:${"b".repeat(64)}`, `sha256:${"a".repeat(64)}`],
        instances: [
          { instance: "zone-01", role: "real", kind: "thermal_sensor", board: "uno-07", device: "zone-01" },
          { instance: "fan-01", role: "simulated", kind: "fan", board: "", device: "" },
        ],
      },
    });
  });

  it("takes whether a gateway is attached from the binary's check, not from the environment word", async () => {
    const hardware = await bridgeAdapter(
      answering(200, { result: { ...check, environment: "hardware" } }),
    ).launch.checkSession("{plan}");
    expect(hardware.ok && hardware.value.mixed).toBe(true);
    const without = await bridgeAdapter(answering(200, { result: { ...check, gateway: false } })).launch.checkSession(
      "{plan}",
    );
    expect(without.ok && without.value.mixed).toBe(false);
  });

  it.each([
    [null],
    [{ instances: "all" }],
    [{ ...check, instances: [{ instance: "zone-01", origin: "hardware" }] }],
    [{ ...check, instances: [{ instance: "zone-01" }] }],
    [{ ...check, policies: undefined }],
    [{ ...check, policies: "sha256:" }],
    [{ ...check, policies: [3] }],
  ])("refuses a check shaped %j, never reading an unknown origin as simulated", async (bad) => {
    const outcome = await bridgeAdapter(answering(200, { result: bad })).launch.checkSession("{}");
    expect(outcome).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });

  it("returns a refused plan with the binary's issues, a running session as busy, and a dead bridge as unavailable", async () => {
    const issue = { pointer: "/real/0/instance", message: "no such instance", keyword: "unknown_instance" };
    const refused = answering(422, {
      error: { code: "plantwire_failed", message: "the plan does not hold", issues: [issue] },
    });
    expect(await bridgeAdapter(refused).launch.checkSession("{}")).toEqual({
      ok: false,
      failure: {
        kind: "refused",
        diagnostic: { message: "the plan does not hold", displayCode: "plantwire_failed" },
        issues: [issue],
      },
    });
    const busy = answering(422, { error: { code: "session_running", message: "still running" } });
    expect(await bridgeAdapter(busy).launch.startSession("{}")).toMatchObject({ ok: false, failure: { kind: "busy" } });
    const broken = answering(500, { error: { code: "bridge_failed", message: "boom" } });
    expect(await bridgeAdapter(broken).launch.sessionStatus()).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    const gone = bridgeTransport(() => Promise.reject(new TypeError("offline")));
    expect(await bridgeAdapter(gone).launch.stopSession()).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
  });

  it("starts, reads and stops the session by name and decodes every state with its domain and plan", async () => {
    const running = answering(200, {
      result: { run: "c-1", state: "running", domain: "arduino-thermal", plan: check },
    });
    expect(await bridgeAdapter(running).launch.startSession("{plan}")).toMatchObject({
      ok: true,
      value: { state: "running", run: "c-1", domain: "arduino-thermal", check: { mixed: true } },
    });
    expect(sentTo(running)).toEqual({ url: "/api/actions/session.start", body: { args: { content: "{plan}" } } });
    const controlled = answering(200, { result: { run: "c-1", state: "running", domain: "d", controls: true } });
    expect(await bridgeAdapter(controlled).launch.sessionStatus()).toMatchObject({
      ok: true,
      value: { controls: true },
    });
    const ended = answering(200, { result: { run: "c-1", state: "ended", result: { out: "runs/c-1" } } });
    expect(await bridgeAdapter(ended).launch.stopSession()).toEqual({
      ok: true,
      value: { state: "ended", run: "c-1", domain: "", check: null, controls: false },
    });
    for (const state of ["starting", "stopping"]) {
      const lasting = answering(200, { result: { run: "c-1", state, domain: "d" } });
      expect(await bridgeAdapter(lasting).launch.sessionStatus()).toMatchObject({ ok: true, value: { state } });
    }
    expect(sentTo(ended)).toEqual({ url: "/api/actions/session.stop", body: { args: {} } });
    const none = answering(200, { result: { state: "none" } });
    expect(await bridgeAdapter(none).launch.sessionStatus()).toEqual({ ok: true, value: { state: "none" } });
    expect(sentTo(none)).toEqual({ url: "/api/actions/session.status", body: { args: {} } });
  });

  it("reads a failed session with the command's error, a default when it said nothing, and refuses an unknown state", async () => {
    const failure = { code: "plantwire_failed", message: "host lost", issues: [{ pointer: "", message: "x" }] };
    const failed = answering(200, { result: { run: "c-2", state: "failed", failure } });
    expect(await bridgeAdapter(failed).launch.sessionStatus()).toEqual({
      ok: true,
      value: {
        state: "failed",
        run: "c-2",
        domain: "",
        check: null,
        controls: false,
        diagnostic: { message: "host lost", displayCode: "plantwire_failed" },
        issues: [{ pointer: "", message: "x" }],
      },
    });
    const silent = answering(200, { result: { run: "c-3", state: "failed" } });
    expect(await bridgeAdapter(silent).launch.sessionStatus()).toMatchObject({
      value: { diagnostic: { displayCode: "session_failed" }, issues: [] },
    });
    for (const bad of [{ state: "paused" }, "running"]) {
      expect(await bridgeAdapter(answering(200, { result: bad })).launch.sessionStatus()).toMatchObject({ ok: false });
    }
  });

  it("follows the session's run journal like a review does", () => {
    const t = bridgeTransport(() => respond(200, {}));
    const follower = { entry: vi.fn(), closed: vi.fn(), failed: vi.fn() };
    const stop = bridgeAdapter(t).launch.followSession("c-1", follower);
    const [source] = t.sources;
    source?.record({ stream: "observer", seq: 0, type: "binding", at_ns: 0, outcome: "matched", issues: [] });
    expect(t.events).toHaveBeenCalledWith("/api/runs/c-1/events");
    expect(follower.entry).toHaveBeenCalledWith({
      kind: "binding",
      atNs: 0,
      binding: { matched: true, refusal: "", reason: "", issues: [] },
    });
    stop();
    expect(source?.readyState).toBe(2);
  });
});
