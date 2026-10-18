import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { boardCapture } from "./captureDecoder";

const digest = "sha256:abc";
const draft = { kind: "pump", metrics: [{ name: "flow" }], commands: [] };
const fresh = { status: "new", target: "kinds/pump.kind.json" };
const announced = { kind: "pump", kind_digest: digest, draft, claims: [], comparison: fresh };

function wire(kinds: readonly object[] = [announced], over: object = {}) {
  return { environment: "hardware", bd_seq: 3, births_seen: 1, taken: { line: 1, bd_seq: 3 }, kinds, ...over };
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

  it("reads where the birth stood, what followed, each kind with its claims, issues and comparison", async () => {
    const issues = [{ pointer: "/metrics/0/unit", message: "needs a unit", keyword: "required", member: "unit" }];
    const comparison = {
      status: "differs",
      target: "kinds/pump.kind.json",
      file: "kinds/pump.kind.json",
      file_digest: "sha256:old",
      differs_at: "/kind",
      collides_with: "kinds/Pump.kind.json",
    };
    const later = [{ line: 4, bd_seq: 4, same_content: false }];
    const body = wire([{ ...announced, claims: ["/commands/0/risk_class"], issues, comparison }], {
      births_seen: 2,
      later_births: later,
      issues: [{ message: "differs", keyword: "rebirth_differs" }],
    });
    const { gateway } = capture(() => respond(200, { result: body }));
    expect(await gateway.captureKinds({ content: "x" })).toEqual({
      ok: true,
      value: {
        environment: "hardware",
        bdSeq: 3,
        birthsSeen: 2,
        taken: { line: 1, bdSeq: 3 },
        laterBirths: [{ line: 4, bdSeq: 4, sameContent: false }],
        issues: [{ pointer: "", message: "differs", keyword: "rebirth_differs" }],
        kinds: [
          {
            kind: "pump",
            digest,
            draft,
            claims: ["/commands/0/risk_class"],
            issues,
            comparison: {
              status: "differs",
              target: "kinds/pump.kind.json",
              file: "kinds/pump.kind.json",
              fileDigest: "sha256:old",
              differsAt: "/kind",
              collidesWith: "kinds/Pump.kind.json",
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
    const odd = capture(() => respond(200, { result: wire([], { environment: "elsewhere" }) }));
    expect(await odd.gateway.captureKinds({ run: "w-1" })).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
  });
});

describe("the capture document", () => {
  it("reads an announcement with no later birth and no issue, and one without kinds", () => {
    expect(boardCapture(wire())).toMatchObject({ birthsSeen: 1, laterBirths: [], issues: [], kinds: [{ claims: [] }] });
    expect(boardCapture(wire([], { environment: "simulated", bd_seq: 0 }))).toMatchObject({
      environment: "simulated",
      bdSeq: 0,
      kinds: [],
    });
  });

  it("reads a kind the same as its file, one that cannot be read, and one without a stated file digest", () => {
    const same = { status: "same", target: "kinds/pump.kind.json", file: "kinds/pump.kind.json", file_digest: digest };
    const unreadable = { status: "unreadable", target: "kinds/pump.kind.json", file: "kinds/pump.kind.json" };
    expect(boardCapture(wire([{ ...announced, comparison: same }]))).toMatchObject({
      kinds: [{ comparison: { status: "same", file: "kinds/pump.kind.json", fileDigest: digest } }],
    });
    expect(boardCapture(wire([{ ...announced, comparison: unreadable }]))).toMatchObject({
      kinds: [{ comparison: { status: "unreadable", target: "kinds/pump.kind.json" } }],
    });
    const bare = { status: "differs", target: "kinds/pump.kind.json", file: "kinds/pump.kind.json" };
    expect(boardCapture(wire([{ ...announced, comparison: bare }]))).toMatchObject({
      kinds: [{ comparison: { fileDigest: "", differsAt: "", collidesWith: "" } }],
    });
  });

  it.each([
    ["not a document", null],
    ["an environment nobody announced", wire([announced], { environment: "cloud" })],
    ["a missing environment", { bd_seq: 0, births_seen: 1, taken: { line: 1, bd_seq: 0 }, kinds: [] }],
    ["a negative sequence", wire([announced], { bd_seq: -1 })],
    ["a fractional sequence", wire([announced], { bd_seq: 1.5 })],
    ["a sequence that is text", wire([announced], { bd_seq: "0" })],
    ["no count of births", wire([announced], { births_seen: 0 })],
    ["no place of the taken birth", wire([announced], { taken: null })],
    ["a taken birth on line 0", wire([announced], { taken: { line: 0, bd_seq: 0 } })],
    ["a taken birth without a sequence", wire([announced], { taken: { line: 1 } })],
    ["later births that are not a list", wire([announced], { later_births: {} })],
    ["a later birth without its line", wire([announced], { later_births: [{ bd_seq: 1, same_content: true }] })],
    ["a later birth without its sequence", wire([announced], { later_births: [{ line: 2, same_content: true }] })],
    [
      "a later birth that does not say whether it is the same",
      wire([announced], { later_births: [{ line: 2, bd_seq: 1 }] }),
    ],
    ["kinds that are not a list", wire([announced], { kinds: {} })],
    ["a kind that is not a record", wire([[]])],
    ["a kind without a name", wire([{ ...announced, kind: "" }])],
    ["a kind without a digest", wire([{ ...announced, kind_digest: "" }])],
    ["a kind without a draft", wire([{ ...announced, draft: undefined }])],
    ["a draft that is a list", wire([{ ...announced, draft: [] }])],
    ["a kind without its claims", wire([{ ...announced, claims: undefined }])],
    ["claims that are not text", wire([{ ...announced, claims: [1] }])],
    ["a kind without a comparison", wire([{ ...announced, comparison: undefined }])],
    ["a comparison that is not a record", wire([{ ...announced, comparison: "same" }])],
    ["a comparison that is null", wire([{ ...announced, comparison: null }])],
    ["a comparison without a target", wire([{ ...announced, comparison: { status: "new" } }])],
    ["a comparison nobody defined", wire([{ ...announced, comparison: { ...fresh, status: "equal", file: "f" } }])],
    ["a same comparison without its file", wire([{ ...announced, comparison: { status: "same", target: "t" } }])],
    ["a differs comparison without its file", wire([{ ...announced, comparison: { status: "differs", target: "t" } }])],
    [
      "an unreadable comparison without its file",
      wire([{ ...announced, comparison: { status: "unreadable", target: "t" } }]),
    ],
  ])("refuses %s", (_name, value) => {
    expect(boardCapture(value)).toBeNull();
  });
});
