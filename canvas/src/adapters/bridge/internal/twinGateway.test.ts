import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { domainChecked, instrumentRefs, instrumentSpec, lintReport, twinMade } from "./twinDecoders";

function devices(handler: Parameters<typeof bridgeTransport>[0]) {
  const transport = bridgeTransport(handler);
  return { transport, gateway: bridgeAdapter(transport).devices };
}

const issues = [{ pointer: "/kinds/0/commands/set/drives/0/gain", message: "a drive with gain 0 moves nothing" }];
const refusal = respond(422, { error: { code: "kind_failed", message: "refused", issues } });

describe("the device list", () => {
  it("reads the kind files of the workspace", async () => {
    const wire = {
      name: "lab",
      domains: [],
      compositions: [],
      scenarios: [],
      kinds: [{ path: "kinds/a.kind.json", kind: "a", title: "A" }, { path: "kinds/b.kind.json" }],
      components: [],
      runs: [],
    };
    const { gateway, transport } = devices(() => respond(200, wire));
    expect(await gateway.kinds()).toEqual({
      ok: true,
      value: [
        { path: "kinds/a.kind.json", kind: "a", title: "A" },
        { path: "kinds/b.kind.json", kind: "", title: "" },
      ],
    });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/workspace");
  });

  it("says when the workspace listing cannot be read", async () => {
    const { gateway } = devices(() => respond(200, { name: 3 }));
    expect(await gateway.kinds()).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });
});

describe("the instrument library", () => {
  it("lists the instruments of the component library with a reference to each", async () => {
    const components = [
      { id: "streamsim/rtd", version: 2, kind: "instrument", title: "RTD" },
      { id: "streamsim/bare", kind: "instrument" },
      { id: "", kind: "instrument" },
      { id: "streamsim/pump", version: 1, kind: "equipment", title: "Pump" },
    ];
    const { gateway, transport } = devices(() => respond(200, { result: { components } }));
    expect(await gateway.instruments()).toEqual({
      ok: true,
      value: [
        { ref: "streamsim/rtd@2", title: "RTD" },
        { ref: "streamsim/bare@1", title: "streamsim/bare" },
      ],
    });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/actions/component.list");
    expect(instrumentRefs({ components: "no" })).toBeNull();
    expect(instrumentRefs(null)).toBeNull();
  });

  it("reads the unit and the range of one instrument", async () => {
    const document = { unit_family: "temperature", unit: "degC", range: { min: -50, max: 200 } };
    const { gateway, transport } = devices(() => respond(200, { result: { entry: {}, document } }));
    expect(await gateway.instrument("streamsim/rtd@1")).toEqual({
      ok: true,
      value: { family: "temperature", unit: "degC", min: -50, max: 200 },
    });
    expect(transport.fetch.mock.calls[0]?.[1]).toMatchObject({
      body: JSON.stringify({ args: { ref: "streamsim/rtd@1" } }),
    });
    expect(instrumentSpec({ document: { unit: "A", range: { min: "x" } } })).toEqual({
      family: "",
      unit: "A",
      min: null,
      max: null,
    });
    expect(instrumentSpec({ document: 1 })).toBeNull();
    expect(instrumentSpec(null)).toBeNull();
  });
});

describe("making a twin", () => {
  it("asks the binary to scaffold the twin from the choices file, over the version of the plant it holds", async () => {
    const result = {
      id: "cooler-twin",
      digest: "sha256:t",
      channels: 3,
      faults: 2,
      effectors: 1,
      implements: ["cooler", ""],
    };
    const { gateway, transport } = devices(() => respond(200, { result }));
    const out = await gateway.twin({ choices: "twins/c.twin.json", out: "domains/c.domain.json", etag: "d1" });
    expect(out).toEqual({
      ok: true,
      value: { id: "cooler-twin", digest: "sha256:t", channels: 3, faults: 2, effectors: 1, implements: ["cooler"] },
    });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/actions/kind.twin");
    expect(transport.fetch.mock.calls[0]?.[1]).toMatchObject({
      body: JSON.stringify({ args: { choices: "twins/c.twin.json", out: "domains/c.domain.json", etag: "d1" } }),
    });
    expect(twinMade({ id: "", digest: "d" })).toBeNull();
    expect(twinMade({ id: "x" })).toEqual({
      id: "x",
      digest: "",
      channels: 0,
      faults: 0,
      effectors: 0,
      implements: [],
    });
    expect(twinMade(null)).toBeNull();
  });

  it("keeps the issues the binary places in the choices file", async () => {
    const { gateway } = devices(() => refusal.clone());
    const input = { choices: "twins/c.twin.json", out: "domains/c.domain.json", etag: "" };
    expect(await gateway.twin(input)).toMatchObject({ ok: false, failure: { kind: "refused", issues } });
  });

  it.each([
    ["a plant changed on disk", respond(409, { error: { code: "conflict", message: "changed" } }), "conflict"],
    ["a missing choices file", respond(404, { error: { code: "file_not_found", message: "none" } }), "missing"],
    ["a bridge fault", respond(500, "broken"), "unavailable"],
  ])("maps %s", async (_name, response, kind) => {
    const { gateway } = devices(() => response.clone());
    const input = { choices: "twins/c.twin.json", out: "domains/c.domain.json", etag: "" };
    expect(await gateway.twin(input)).toMatchObject({ ok: false, failure: { kind } });
  });
});

describe("checking the plant that was made", () => {
  it("validates it against the kind files of the workspace", async () => {
    const result = { id: "cooler-twin", digest: "sha256:t", valid: true };
    const { gateway, transport } = devices(() => respond(200, { result }));
    expect(await gateway.validateDomain("domains/c.domain.json")).toEqual({
      ok: true,
      value: { id: "cooler-twin", digest: "sha256:t" },
    });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/actions/domain.validate");
    expect(domainChecked({ valid: false })).toBeNull();
    expect(domainChecked({ valid: true })).toEqual({ id: "", digest: "" });
    expect(domainChecked(null)).toBeNull();
  });

  it("keeps the issues of a refused plant", async () => {
    const { gateway } = devices(() => refusal.clone());
    expect(await gateway.validateDomain("domains/c.domain.json")).toMatchObject({
      ok: false,
      failure: { kind: "refused", issues },
    });
  });

  it("reads the plausibility report, also when it found errors", async () => {
    const report = { errors: 0, warnings: 1, findings: [{ severity: "warning", entity: "a", detail: "slow" }, {}] };
    const ok = devices(() => respond(200, { result: report }));
    expect(await ok.gateway.lint("domains/c.domain.json")).toEqual({
      ok: true,
      value: {
        errors: 0,
        warnings: 1,
        findings: [
          { severity: "warning", entity: "a", detail: "slow" },
          { severity: "", entity: "", detail: "" },
        ],
      },
    });
    expect(ok.transport.fetch.mock.calls[0]?.[0]).toBe("/api/actions/domain.lint");
    const failing = { ...report, errors: 2 };
    const failed = devices(() => respond(422, { error: { code: "domain_lint", message: "errors", result: failing } }));
    expect(await failed.gateway.lint("domains/c.domain.json")).toMatchObject({ ok: true, value: { errors: 2 } });
  });

  it("says when the plausibility check cannot be read or did not run", async () => {
    const odd = devices(() => respond(200, { result: { errors: 0 } }));
    expect(await odd.gateway.lint("domains/c.domain.json")).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
    const bare = devices(() => respond(422, { error: { code: "domain_lint", message: "errors" } }));
    expect(await bare.gateway.lint("domains/c.domain.json")).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
    const down = devices(() => respond(500, "broken"));
    expect(await down.gateway.lint("domains/c.domain.json")).toMatchObject({
      ok: false,
      failure: { kind: "unavailable" },
    });
    expect(lintReport(null)).toBeNull();
  });
});
