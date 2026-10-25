import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { comparison, comparisonWire } from "../../../test/compareFixture";
import { reportWire } from "../../../test/runFixture";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";

function sent(transport: ReturnType<typeof bridgeTransport>, index = 0): unknown {
  const init = transport.fetch.mock.calls[index]?.[1];
  const body = init?.body;
  return JSON.parse(typeof body === "string" ? body : "null");
}

describe("the report of a run with its replay checked", () => {
  it("asks for the replay only when the caller wants it", async () => {
    const transport = bridgeTransport(() => respond(200, { result: reportWire }));
    const review = bridgeAdapter(transport).review;
    await review.report("w-1");
    await review.report("w-1", { replay: true });
    expect(sent(transport, 0)).toEqual({ args: { run: "w-1" } });
    expect(sent(transport, 1)).toEqual({ args: { run: "w-1", replay: true } });
  });
});

describe("the handover of a run", () => {
  it("asks the binary for the markdown report and returns its text untouched", async () => {
    const transport = bridgeTransport(() =>
      respond(200, { result: { format: "markdown", markdown: "# Run w-1\n\n## Open questions\n" } }),
    );
    const handover = await bridgeAdapter(transport).review.handover("w-1", { replay: true });
    expect(handover).toEqual({ ok: true, value: "# Run w-1\n\n## Open questions\n" });
    expect(sent(transport)).toEqual({ args: { run: "w-1", format: "markdown", replay: true } });
  });

  it("does not ask for a replay the caller did not want", async () => {
    const transport = bridgeTransport(() => respond(200, { result: { format: "markdown", markdown: "x" } }));
    await bridgeAdapter(transport).review.handover("w-1", { replay: false });
    expect(sent(transport)).toEqual({ args: { run: "w-1", format: "markdown" } });
  });

  it("keeps the binary's refusal message", async () => {
    const refusal = { error: { code: "run_not_ended", message: "run w-1 has not ended" } };
    const handover = await bridgeAdapter(bridgeTransport(() => respond(422, refusal))).review.handover("w-1", {
      replay: false,
    });
    expect(handover).toMatchObject({
      ok: false,
      failure: { kind: "refused", diagnostic: { displayCode: "run_not_ended", message: "run w-1 has not ended" } },
    });
  });

  it("refuses an answer that is not a markdown document", async () => {
    const transport = bridgeTransport(() => respond(200, { result: { format: "json", markdown: "x" } }));
    expect(await bridgeAdapter(transport).review.handover("w-1", { replay: false })).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
  });
});

describe("the comparison of two runs", () => {
  it("sends the reference as a and this run as b and decodes the document", async () => {
    const transport = bridgeTransport(() => respond(200, { result: comparisonWire }));
    const compared = await bridgeAdapter(transport).review.compare("w-1", "w-2");
    expect(compared).toEqual({ ok: true, value: comparison });
    expect(sent(transport)).toEqual({ args: { a: "w-1", b: "w-2" } });
  });

  it("shows the binary's refusal of a run that has not ended, word for word", async () => {
    const message = "run B has not ended: its truth is still sealed";
    const refusal = { error: { code: "run_not_ended", message } };
    const compared = await bridgeAdapter(bridgeTransport(() => respond(422, refusal))).review.compare("w-1", "w-2");
    expect(compared).toMatchObject({ ok: false, failure: { kind: "refused", diagnostic: { message } } });
  });

  it("tells a server fault from a refusal and an unreadable document from both", async () => {
    const fault = bridgeAdapter(bridgeTransport(() => respond(500, { error: { code: "x", message: "boom" } })));
    expect(await fault.review.compare("a", "b")).toMatchObject({ ok: false, failure: { kind: "unavailable" } });
    const garbled = bridgeAdapter(bridgeTransport(() => respond(200, { result: { plan: 1 } })));
    expect(await garbled.review.compare("a", "b")).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
    const down = bridgeAdapter(bridgeTransport(() => Promise.reject(new Error("down"))));
    expect(await down.review.compare("a", "b")).toMatchObject({ ok: false, failure: { kind: "unavailable" } });
  });
});
