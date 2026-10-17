import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { kindChecked } from "./kindIssues";

const path = "kinds/fan.kind.json";
const summary = { path, kind: "fan", kind_digest: "sha256:abc", metrics: 2, commands: 1 };

function refusal(message: string, issues: readonly { pointer: string; message: string }[] = []) {
  return respond(422, { error: { code: "kind_failed", message, issues } });
}

function devices(handler: Parameters<typeof bridgeTransport>[0]) {
  const transport = bridgeTransport(handler);
  return { transport, gateway: bridgeAdapter(transport).devices };
}

describe("devices adapter", () => {
  it("reads and writes a kind file by its workspace path", async () => {
    const { transport, gateway } = devices((url) =>
      url.startsWith("/api/file?")
        ? respond(200, { path, content: "{}", etag: "e1" })
        : respond(200, { path, etag: "e2" }),
    );
    expect(await gateway.readFile(path)).toEqual({ ok: true, value: { path, content: "{}", etag: "e1" } });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/file?path=kinds%2Ffan.kind.json");
    expect(await gateway.writeFile(path, "{}", "e1")).toEqual({ ok: true, value: { path, etag: "e2" } });
    expect(transport.fetch.mock.calls[1]?.[1]).toMatchObject({
      method: "PUT",
      body: JSON.stringify({ path, content: "{}", etag: "e1" }),
    });
  });

  it("asks the binary to validate the saved file and reads its summary", async () => {
    const { transport, gateway } = devices(() => respond(200, { result: { valid: true, kinds: [summary] } }));
    expect(await gateway.check(path)).toEqual({
      ok: true,
      value: { kind: "fan", digest: "sha256:abc", metrics: 2, commands: 1 },
    });
    expect(transport.fetch.mock.calls[0]?.[0]).toBe("/api/actions/kind.validate");
    expect(transport.fetch.mock.calls[0]?.[1]).toMatchObject({ body: JSON.stringify({ args: { path } }) });
  });

  it("keeps the issues the binary names with their pointers", async () => {
    const issues = [{ pointer: "/metrics/0/name", message: "string does not match pattern" }];
    const { gateway } = devices(() => refusal("kind file fails its schema", issues));
    expect(await gateway.check(path)).toMatchObject({
      ok: false,
      failure: { kind: "refused", diagnostic: { displayCode: "kind_failed" }, issues },
    });
  });

  it("refuses with its message alone when the binary names no located issue", async () => {
    const { gateway } = devices(() =>
      respond(422, {
        error: { code: "kind_failed", message: "kind file is not JSON", issues: [{ message: "not JSON" }] },
      }),
    );
    expect(await gateway.check(path)).toMatchObject({
      ok: false,
      failure: { kind: "refused", diagnostic: { message: "kind file is not JSON" }, issues: [] },
    });
  });

  it.each([
    ["a missing file", respond(404, { error: { code: "file_not_found", message: "no file" } }), "missing"],
    ["a changed file", respond(409, { error: { code: "conflict", message: "changed" } }), "conflict"],
    ["a bridge fault", respond(500, "broken"), "unavailable"],
    [
      "a refusal that is not a validation result",
      respond(400, { error: { code: "bad_path", message: "no" } }),
      "unreadable",
    ],
  ])("maps %s on read and write", async (_name, response, kind) => {
    const { gateway } = devices(() => response.clone());
    expect(await gateway.readFile(path)).toMatchObject({ ok: false, failure: { kind } });
    expect(await gateway.writeFile(path, "{}", "")).toMatchObject({ ok: false, failure: { kind } });
  });

  it("does not take a 422 on a read as a validation result", async () => {
    const { gateway } = devices(() => refusal("nope"));
    expect(await gateway.readFile(path)).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });

  it("reports an unreachable bridge and an answer it cannot decode", async () => {
    const down = devices(() => Promise.reject(new Error("down")));
    expect(await down.gateway.check(path)).toMatchObject({ ok: false, failure: { kind: "unavailable" } });
    const odd = devices(() => respond(200, { result: { valid: false } }));
    expect(await odd.gateway.check(path)).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });
});

describe("kind check payload", () => {
  it("defaults what it is not told, and refuses anything that is not a valid verdict", () => {
    expect(kindChecked({ valid: true, kinds: [{ kind: "fan" }] })).toEqual({
      kind: "fan",
      digest: "",
      metrics: 0,
      commands: 0,
    });
    expect(kindChecked({ valid: true })).toEqual({ kind: "", digest: "", metrics: 0, commands: 0 });
    expect(kindChecked({ valid: false })).toBeNull();
    expect(kindChecked(null)).toBeNull();
  });
});
