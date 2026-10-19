import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { cannotJudgeWire, twinReportWire } from "../../../test/twinReportFixture";
import { bridgeAdapter } from "..";

function answering(status: number, body: unknown) {
  return bridgeTransport(() => respond(status, body));
}

function sent(t: ReturnType<typeof bridgeTransport>): { url: unknown; body: unknown } {
  const [url, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return { url, body: typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined };
}

describe("the twin report asked of the bridge", () => {
  it("sends the run, the domain and the tolerances exactly as they were typed", async () => {
    const t = answering(200, { result: twinReportWire() });
    const tolerance = " sensor.temperature=2 , thermal_sensor.humidity=5 ";
    const outcome = await bridgeAdapter(t).review.twinReport({ run: "c-1", domain: "mini", tolerance });
    expect(outcome).toMatchObject({ ok: true, value: { verdict: "diverges" } });
    expect(sent(t)).toEqual({
      url: "/api/actions/run.twin_report",
      body: { args: { run: "c-1", domain: "mini", tolerance } },
    });
  });

  it("leaves the tolerance out when nothing was typed", async () => {
    for (const tolerance of ["", "   "]) {
      const t = answering(200, { result: cannotJudgeWire() });
      await bridgeAdapter(t).review.twinReport({ run: "c-1", domain: "mini", tolerance });
      expect(sent(t).body).toEqual({ args: { run: "c-1", domain: "mini" } });
    }
  });

  it("returns a report that does not agree as a result, not as a failure", async () => {
    const t = answering(200, { result: cannotJudgeWire() });
    expect(await bridgeAdapter(t).review.twinReport({ run: "c-1", domain: "mini", tolerance: "" })).toMatchObject({
      ok: true,
      value: { verdict: "cannot_judge", reason: "no_commands_applied" },
    });
  });

  it.each<[number, string, string]>([
    [422, "plantwire_failed", "refused"],
    [400, "bad_arguments", "refused"],
    [500, "bridge_failed", "unavailable"],
  ])("shows the binary's refusal (%s %s) as the %s failure with its words", async (status, code, kind) => {
    const t = answering(status, { error: { code, message: "invalid tolerance: no such metric" } });
    expect(await bridgeAdapter(t).review.twinReport({ run: "c-1", domain: "mini", tolerance: "x=1" })).toMatchObject({
      ok: false,
      failure: { kind, diagnostic: { displayCode: code, message: "invalid tolerance: no such metric" } },
    });
  });

  it("tells a bridge that did not answer from an answer it cannot read", async () => {
    const gone = bridgeTransport(() => Promise.reject(new TypeError("offline")));
    expect(await bridgeAdapter(gone).review.twinReport({ run: "c-1", domain: "mini", tolerance: "" })).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    const garbled = answering(200, { result: { schema_version: "0.1", verdict: "perhaps" } });
    expect(await bridgeAdapter(garbled).review.twinReport({ run: "c-1", domain: "mini", tolerance: "" })).toMatchObject(
      {
        ok: false,
        failure: { kind: "unreadable" },
      },
    );
  });
});
