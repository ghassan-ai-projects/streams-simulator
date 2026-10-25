import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { assumptionsWire, diffWire, lineageWire } from "../../../test/provenanceFixture";
import { bridgeAdapter } from "..";

function provenance(handler: Parameters<typeof bridgeTransport>[0]) {
  const transport = bridgeTransport(handler);
  return { transport, gateway: bridgeAdapter(transport).provenance };
}

function body(transport: ReturnType<typeof bridgeTransport>, call = 0): unknown {
  const sent = transport.fetch.mock.calls[call]?.[1]?.body;
  return JSON.parse(typeof sent === "string" ? sent : "null");
}

describe("the provenance gateway", () => {
  it("asks the bridge for the lineage of the whole workspace", async () => {
    const { gateway, transport } = provenance(() => respond(200, { result: lineageWire }));
    const read = await gateway.lineage();
    expect(read.ok && read.value.plants).toHaveLength(5);
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/actions/domain.lineage");
    expect(body(transport)).toEqual({ args: {} });
  });

  it("asks for the assumptions of a domain, with the twin choices when it has them", async () => {
    const { gateway, transport } = provenance(() => respond(200, { result: assumptionsWire }));
    await gateway.assumptions({ path: "domains/a.domain.json", twin: null });
    await gateway.assumptions({ path: "domains/a.domain.json", twin: "twins/a.twin.json" });
    expect(body(transport, 0)).toEqual({ args: { path: "domains/a.domain.json" } });
    expect(body(transport, 1)).toEqual({ args: { path: "domains/a.domain.json", twin: "twins/a.twin.json" } });
  });

  it("asks for the diff of two models and says when the answer is unreadable or refused", async () => {
    const { gateway, transport } = provenance(() => respond(200, { result: diffWire }));
    expect((await gateway.diff("w-1", "domains/a.domain.json")).ok).toBe(true);
    expect(body(transport)).toEqual({ args: { a: "w-1", b: "domains/a.domain.json" } });
    const odd = provenance(() => respond(200, { result: { schema_version: "0.1" } }));
    expect(await odd.gateway.diff("a", "b")).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
    const refused = provenance(() => respond(422, { error: { code: "run_not_found", message: "no run w-9" } }));
    expect(await refused.gateway.diff("w-9", "b")).toMatchObject({
      ok: false,
      failure: { kind: "refused", diagnostic: { message: "no run w-9" } },
    });
  });

  it("composes again against the version of the plant file that is on disk", async () => {
    const summary = { channels: 3, faults: 2, effectors: 1, digest: `sha256:${"a".repeat(64)}` };
    const { gateway, transport } = provenance((url) =>
      url.startsWith("/api/file")
        ? respond(200, { path: "domains/a.domain.json", content: "{}", etag: "e7" })
        : respond(200, { result: summary }),
    );
    const composed = await gateway.compose({
      composition: "compositions/a.composition.json",
      output: "domains/a.domain.json",
    });
    expect(composed).toEqual({ ok: true, value: summary });
    expect(body(transport, 1)).toEqual({
      args: { composition: "compositions/a.composition.json", out: "domains/a.domain.json", etag: "e7" },
    });
  });

  it("composes against no version when the plant file is not there", async () => {
    const { gateway, transport } = provenance((url) =>
      url.startsWith("/api/file")
        ? respond(404, { error: { code: "file_not_found", message: "none" } })
        : respond(200, { result: { channels: 1, faults: 0, effectors: 0 } }),
    );
    const composed = await gateway.compose({ composition: "c", output: "domains/a.domain.json" });
    expect(composed).toMatchObject({ ok: true, value: { digest: "" } });
    expect(body(transport, 1)).toMatchObject({ args: { etag: "" } });
  });
});
