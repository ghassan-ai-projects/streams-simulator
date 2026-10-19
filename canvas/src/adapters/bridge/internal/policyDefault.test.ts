import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";

const a = `sha256:${"a".repeat(64)}`;
const b = `sha256:${"b".repeat(64)}`;
const path = "settings/default.policies.json";

function holding(content: string, etag = "e1") {
  return bridgeTransport(() => respond(200, { path, content, etag }));
}

function sentBody(init: RequestInit | undefined): string {
  const body = init?.body;
  if (typeof body !== "string") throw new Error("the save did not send a text body");
  return body;
}

describe("the workspace policy default", () => {
  it("reads the saved list in the order it was saved, with the etag to save from", async () => {
    const t = holding(JSON.stringify({ policies: [b, a] }));
    expect(await bridgeAdapter(t).launch.readPolicyDefault()).toEqual({
      ok: true,
      value: { state: "present", policies: [b, a], etag: "e1" },
    });
    expect(t.fetch.mock.calls[0]?.[0]).toBe(`/api/file?path=${encodeURIComponent(path)}`);
  });

  it("reads a default with no file as absent", async () => {
    const t = bridgeTransport(() => respond(404, { error: { code: "not_found", message: "no file" } }));
    expect(await bridgeAdapter(t).launch.readPolicyDefault()).toEqual({ ok: true, value: { state: "absent" } });
  });

  it.each([
    ["is not JSON", "{"],
    ["is not an object", "[]"],
    ["has no list", "{}"],
    ["has a list that holds a number", '{"policies":[1]}'],
    ["has a list that is text", '{"policies":"x"}'],
    ["holds more than the list", `{"policies":[],"note":"x"}`],
  ])("reads a default that %s as unreadable, keeping the etag", async (_, content) => {
    expect(await bridgeAdapter(holding(content)).launch.readPolicyDefault()).toEqual({
      ok: true,
      value: { state: "unreadable", etag: "e1" },
    });
  });

  it("fails when the file answer is not a file, or the bridge does not answer or fails", async () => {
    const odd = bridgeTransport(() => respond(200, { nothing: true }));
    expect(await bridgeAdapter(odd).launch.readPolicyDefault()).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
    const broken = bridgeTransport(() => respond(500, { error: { code: "bridge_failed", message: "boom" } }));
    expect(await bridgeAdapter(broken).launch.readPolicyDefault()).toMatchObject({
      ok: false,
      failure: { kind: "unavailable", diagnostic: { message: "boom" } },
    });
    const refused = bridgeTransport(() => respond(400, { error: { code: "bad_path", message: "no" } }));
    expect(await bridgeAdapter(refused).launch.readPolicyDefault()).toMatchObject({ failure: { kind: "unreadable" } });
  });

  it("writes the list as the file the bridge reads back, from the etag it holds", async () => {
    const t = bridgeTransport(() => respond(200, { path, etag: "e2" }));
    expect(await bridgeAdapter(t).launch.savePolicyDefault([a, b], "e1")).toEqual({ ok: true, value: { etag: "e2" } });
    const [url, init] = t.fetch.mock.calls[0] ?? [];
    expect(url).toBe("/api/file");
    expect(init?.method).toBe("PUT");
    const sent = JSON.parse(sentBody(init)) as { path: string; content: string; etag: string };
    expect(sent).toMatchObject({ path, etag: "e1" });
    expect(JSON.parse(sent.content)).toEqual({ policies: [a, b] });
    expect(sent.content.endsWith("\n")).toBe(true);
  });

  it("writes an empty list to clear the default, and a first write from the empty etag", async () => {
    const t = bridgeTransport(() => respond(200, { path, etag: "e1" }));
    await bridgeAdapter(t).launch.savePolicyDefault([], "");
    const sent = JSON.parse(sentBody(t.fetch.mock.calls[0]?.[1])) as { content: string; etag: string };
    expect(JSON.parse(sent.content)).toEqual({ policies: [] });
    expect(sent.etag).toBe("");
  });

  it("reports a default that changed elsewhere as a conflict, and a dead bridge as unavailable", async () => {
    const stale = bridgeTransport(() => respond(409, { error: { code: "conflict", message: "changed on disk" } }));
    expect(await bridgeAdapter(stale).launch.savePolicyDefault([a], "old")).toMatchObject({
      ok: false,
      failure: { kind: "conflict", diagnostic: { displayCode: "conflict" } },
    });
    const down = bridgeTransport(() => {
      throw new TypeError("offline");
    });
    expect(await bridgeAdapter(down).launch.savePolicyDefault([a], "")).toMatchObject({
      failure: { kind: "unavailable" },
    });
    const odd = bridgeTransport(() => respond(200, { path }));
    expect(await bridgeAdapter(odd).launch.savePolicyDefault([a], "")).toMatchObject({
      failure: { kind: "unreadable" },
    });
  });
});
