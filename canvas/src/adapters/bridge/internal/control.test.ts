import { describe, expect, it } from "vitest";
import type { ControlRequest } from "../../../model/review";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { controlReplyWire } from "../../../test/liveFixture";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";

function answering(status: number, body: unknown) {
  return bridgeTransport(() => respond(status, body));
}

function sent(t: ReturnType<typeof bridgeTransport>): { url: unknown; body: unknown } {
  const [url, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return { url, body: typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined };
}

describe("a control sent through the bridge", () => {
  it.each<[ControlRequest, Record<string, unknown>]>([
    [
      { op: "hold", scope: { kind: "board", id: "mega-01" } },
      { op: "hold", scope: { kind: "board", id: "mega-01" } },
    ],
    [
      { op: "release", scope: { kind: "all", id: "" } },
      { op: "release", scope: { kind: "all" } },
    ],
    [
      { op: "estop", scope: { kind: "device", id: "fan" } },
      { op: "estop", scope: { kind: "device", id: "fan" } },
    ],
    [
      { op: "approve_mode", scope: { kind: "all", id: "" }, enabled: false },
      { op: "approve_mode", scope: { kind: "all" }, enabled: false },
    ],
    [
      { op: "approve", commandId: "cmd-7" },
      { op: "approve", command_id: "cmd-7" },
    ],
    [
      { op: "deny", commandId: "cmd-7" },
      { op: "deny", command_id: "cmd-7" },
    ],
    [{ op: "state" }, { op: "state" }],
  ])("sends %j as the node's own request document and nothing else", async (request, args) => {
    const t = answering(200, { result: controlReplyWire() });
    expect(await bridgeAdapter(t).review.control(request)).toMatchObject({ ok: true, value: { result: "applied" } });
    expect(sent(t)).toEqual({ url: "/api/actions/session.control", body: { args } });
  });

  it("returns the node's refusal as a reply, not as a failure", async () => {
    const refused = controlReplyWire({ result: "refused", reason: "unknown_device" });
    const t = answering(200, { result: refused });
    expect(await bridgeAdapter(t).review.control({ op: "hold", scope: { kind: "device", id: "x" } })).toMatchObject({
      ok: true,
      value: { result: "refused", reason: "unknown_device" },
    });
  });

  it.each([
    ["no_session", 422, "no-session"],
    ["control_unavailable", 422, "no-controls"],
    ["control_gone", 422, "gone"],
    ["control_timeout", 422, "timeout"],
    ["control_unreadable", 422, "unreadable"],
    ["bridge_failed", 500, "unavailable"],
    ["bad_arguments", 400, "unreadable"],
  ])("maps the bridge's %s to the failure %s", async (code, status, kind) => {
    const t = answering(status, { error: { code, message: "words" } });
    expect(await bridgeAdapter(t).review.control({ op: "state" })).toMatchObject({
      ok: false,
      failure: { kind, diagnostic: { displayCode: code, message: "words" } },
    });
  });

  it("tells a bridge that did not answer from an answer it cannot read", async () => {
    const gone = bridgeTransport(() => Promise.reject(new TypeError("offline")));
    expect(await bridgeAdapter(gone).review.control({ op: "state" })).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    for (const result of [{ result: "applied" }, { result: "later", state: {} }]) {
      const t = answering(200, { result });
      expect(await bridgeAdapter(t).review.control({ op: "state" })).toMatchObject({
        ok: false,
        failure: { kind: "unreadable" },
      });
    }
  });
});
