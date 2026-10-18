import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { boardCapture } from "./captureDecoder";

const digest = "sha256:abc";
const draft = { kind: "pump", metrics: [{ name: "flow" }], commands: [] };
const announced = { kind: "pump", kind_digest: digest, draft };

function wire(kinds: readonly object[] = [announced], over: object = {}) {
  return { environment: "hardware", bd_seq: 3, kinds, ...over };
}

function capture(handler: Parameters<typeof bridgeTransport>[0]) {
  const transport = bridgeTransport(handler);
  return { transport, gateway: bridgeAdapter(transport).devices };
}

describe("capturing the kinds a board announced", () => {
  it("sends the pasted text, or the recorded run, and nothing else", async () => {
    const { transport, gateway } = capture(() => respond(200, { result: wire() }));
    await gateway.captureKinds({ content: '{"type":"NBIRTH"}' });
    await gateway.captureKinds({ run: "w-1" });
    const [first, second] = transport.fetch.mock.calls;
    expect(first?.[0]).toBe("/api/actions/kind.capture");
    expect(first?.[1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ args: { content: '{"type":"NBIRTH"}' } }),
    });
    expect(second?.[1]).toMatchObject({ body: JSON.stringify({ args: { run: "w-1" } }) });
  });

  it("reads the environment, the sequence and each kind with its draft, issues and comparison", async () => {
    const issues = [{ pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" }];
    const comparison = {
      status: "differs",
      file: "kinds/pump.kind.json",
      file_digest: "sha256:old",
      differs_at: "/kind",
    };
    const body = wire([{ ...announced, issues, comparison }]);
    const { gateway } = capture(() => respond(200, { result: body }));
    expect(await gateway.captureKinds({ content: "x" })).toEqual({
      ok: true,
      value: {
        environment: "hardware",
        bdSeq: 3,
        kinds: [
          {
            kind: "pump",
            digest,
            draft,
            issues,
            comparison: {
              status: "differs",
              file: "kinds/pump.kind.json",
              fileDigest: "sha256:old",
              differsAt: "/kind",
            },
          },
        ],
      },
    });
  });

  it("keeps the binary's issues and its sentence when the birth does not hold", async () => {
    const issues = [{ pointer: "/kinds/0/kind_digest", message: "does not hold", keyword: "kind_digest_mismatch" }];
    const refused = respond(422, { error: { code: "kind_failed", message: "the birth is refused", issues } });
    const { gateway } = capture(() => refused.clone());
    expect(await gateway.captureKinds({ content: "x" })).toMatchObject({
      ok: false,
      failure: { kind: "refused", diagnostic: { message: "the birth is refused", displayCode: "kind_failed" }, issues },
    });
  });

  it.each([
    ["a run without a journal", respond(404, { error: { code: "file_not_found", message: "no journal" } }), "missing"],
    ["text over the cap", respond(413, { error: { code: "too_large", message: "too large" } }), "unreadable"],
    ["a bridge fault", respond(500, "broken"), "unavailable"],
  ])("maps %s", async (_name, response, kind) => {
    const { gateway } = capture(() => response.clone());
    expect(await gateway.captureKinds({ run: "w-1" })).toMatchObject({ ok: false, failure: { kind } });
  });

  it("does not take an unreachable bridge or an unreadable document for a capture", async () => {
    const down = capture(() => Promise.reject(new Error("down")));
    expect(await down.gateway.captureKinds({ run: "w-1" })).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    const odd = capture(() => respond(200, { result: { environment: "elsewhere", bd_seq: 0, kinds: [] } }));
    expect(await odd.gateway.captureKinds({ run: "w-1" })).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
  });
});

describe("the capture document", () => {
  it("takes a kind without a comparison as a new device, and an empty announcement as no kinds", () => {
    expect(boardCapture(wire())).toMatchObject({ kinds: [{ comparison: { status: "new" }, issues: [] }] });
    expect(boardCapture(wire([], { environment: "simulated", bd_seq: 0 }))).toEqual({
      environment: "simulated",
      bdSeq: 0,
      kinds: [],
    });
  });

  it("reads a kind the same as its file and one that is new beside a kinds folder", () => {
    const same = { status: "same", file: "kinds/pump.kind.json", file_digest: digest };
    expect(boardCapture(wire([{ ...announced, comparison: same }]))).toMatchObject({
      kinds: [{ comparison: { status: "same", file: "kinds/pump.kind.json", fileDigest: digest } }],
    });
    expect(boardCapture(wire([{ ...announced, comparison: { status: "new" } }]))).toMatchObject({
      kinds: [{ comparison: { status: "new" } }],
    });
  });

  it.each([
    ["not a document", null],
    ["an environment nobody announced", wire([announced], { environment: "cloud" })],
    ["a missing environment", { bd_seq: 0, kinds: [] }],
    ["a negative sequence", wire([announced], { bd_seq: -1 })],
    ["a fractional sequence", wire([announced], { bd_seq: 1.5 })],
    ["a sequence that is text", wire([announced], { bd_seq: "0" })],
    ["kinds that are not a list", wire([announced], { kinds: {} })],
    ["a kind that is not a record", wire([[]])],
    ["a kind without a name", wire([{ ...announced, kind: "" }])],
    ["a kind without a digest", wire([{ ...announced, kind_digest: "" }])],
    ["a kind without a draft", wire([{ kind: "pump", kind_digest: digest }])],
    ["a draft that is a list", wire([{ ...announced, draft: [] }])],
    ["a comparison that is not a record", wire([{ ...announced, comparison: "same" }])],
    ["a comparison that is null", wire([{ ...announced, comparison: null }])],
    ["a comparison nobody defined", wire([{ ...announced, comparison: { status: "equal" } }])],
    ["a same comparison without its file", wire([{ ...announced, comparison: { status: "same" } }])],
    ["a differs comparison without its file", wire([{ ...announced, comparison: { status: "differs" } }])],
  ])("refuses %s", (_name, value) => {
    expect(boardCapture(value)).toBeNull();
  });
});
