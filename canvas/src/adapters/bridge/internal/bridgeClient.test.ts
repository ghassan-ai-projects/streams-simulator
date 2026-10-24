import { describe, expect, it, vi } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";

describe("bridge adapter wiring", () => {
  it("pairs with a JSON body and same-origin credentials", async () => {
    const t = bridgeTransport(() => respond(200, { paired: true }));
    expect(await bridgeAdapter(t).session.pair("AAAABBBBCCCC")).toEqual({ ok: true, value: true });
    const [url, init] = t.fetch.mock.calls[0] ?? [];
    expect(url).toBe("/api/session");
    expect(init).toMatchObject({ method: "POST", credentials: "same-origin", body: '{"code":"AAAABBBBCCCC"}' });
  });

  it("maps an error envelope to a failure and a network error to unreachable", async () => {
    const refused = await bridgeAdapter(
      bridgeTransport(() => respond(401, { error: { code: "wrong_code", message: "wrong" } })),
    ).session.pair("x");
    expect(refused).toMatchObject({
      ok: false,
      failure: { kind: "code-rejected", diagnostic: { displayCode: "wrong_code", message: "wrong" } },
    });
    const down = await bridgeAdapter(bridgeTransport(() => Promise.reject(new Error("down")))).session.check();
    expect(down).toMatchObject({ ok: false, failure: { kind: "unreachable" } });
    const garbage = await bridgeAdapter(bridgeTransport(() => respond(500, "not json"))).workspace.load();
    expect(garbage).toMatchObject({
      ok: false,
      failure: { kind: "unavailable", diagnostic: { displayCode: "bridge_failed" } },
    });
  });

  it("reads files by encoded path, unwraps action results and decodes the workspace", async () => {
    const t = bridgeTransport((url) =>
      url.startsWith("/api/file")
        ? respond(200, { path: "runs/w-1/run.json", content: "{}", etag: "e1" })
        : respond(200, { result: { matches: true, detail: "same" } }),
    );
    const adapters = bridgeAdapter(t);
    expect(await adapters.plant.readFile("runs/w-1/run.json")).toEqual({
      ok: true,
      value: { path: "runs/w-1/run.json", content: "{}", etag: "e1" },
    });
    expect(t.fetch.mock.calls[0]?.[0]).toBe("/api/file?path=runs%2Fw-1%2Frun.json");
    expect(await adapters.review.verify("w-1")).toEqual({
      ok: true,
      value: { matches: true, domainChanged: false, detail: "same", reason: "", file: "" },
    });
    expect(t.fetch.mock.calls[1]?.[1]).toMatchObject({ body: '{"args":{"run":"w-1"}}' });
    expect(await adapters.review.report("w-1")).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
    const empty = { name: "lab", domains: [], compositions: [], scenarios: [], kinds: [], components: [], runs: [] };
    expect(await bridgeAdapter(bridgeTransport(() => respond(200, empty))).workspace.load()).toEqual({
      ok: true,
      value: empty,
    });
  });

  it("narrows one field at a time, and refuses a response it cannot decode", async () => {
    const corrupt = await bridgeAdapter(bridgeTransport(() => respond(200, { path: "x" }))).editor.readFile("x");
    expect(corrupt).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });

  it("recovers the verdict an unsuccessful verify carries, and reads the run's domain", async () => {
    const t = bridgeTransport((url) =>
      url.includes("run.verify")
        ? respond(422, { error: { code: "replay_mismatch", message: "no", result: { matches: false, detail: "d" } } })
        : respond(200, {
            path: "runs/w-1/run.json",
            content: JSON.stringify({ domain_spec: { id: "mini" } }),
            etag: "e",
          }),
    );
    const adapters = bridgeAdapter(t);
    expect(await adapters.review.verify("w-1")).toEqual({
      ok: true,
      value: { matches: false, domainChanged: false, detail: "d", reason: "", file: "" },
    });
    expect(await adapters.review.runDomain("w-1")).toEqual({ ok: true, value: { id: "mini" } });
    expect(t.fetch.mock.calls[1]?.[0]).toBe("/api/file?path=runs%2Fw-1%2Frun.json");
  });

  it("follows a run's journal until it closes", () => {
    const t = bridgeTransport(() => respond(200, {}));
    const follower = { entry: vi.fn(), caughtUp: vi.fn(), closed: vi.fn(), failed: vi.fn() };
    const stop = bridgeAdapter(t).review.follow("w 1", follower);
    expect(t.events).toHaveBeenCalledWith("/api/runs/w%201/events");
    const source = t.sources[0];
    const record = { stream: "observer", type: "clock.advanced", seq: 1, at_ns: 30e9 };
    source?.record(record);
    expect(follower.entry.mock.calls).toEqual([[{ kind: "other", atNs: 30e9 }]]);
    source?.end("caught_up");
    expect(follower.caughtUp).toHaveBeenCalledOnce();
    source?.dispatchEvent(new MessageEvent("record", { data: "not json" }));
    expect(follower.failed).toHaveBeenCalledOnce();
    expect(source?.readyState).toBe(2);
    source?.end("closed");
    expect(follower.closed).toHaveBeenCalledOnce();
    stop();
  });

  it("fails a follower only when the stream died, not while it reconnects", () => {
    const t = bridgeTransport(() => respond(200, {}));
    const follower = { entry: vi.fn(), closed: vi.fn(), failed: vi.fn() };
    bridgeAdapter(t).review.follow("w-1", follower);
    t.sources[0]?.reconnect();
    expect(follower.failed).not.toHaveBeenCalled();
    t.sources[0]?.end("failed");
    expect(follower.failed).toHaveBeenCalledOnce();
  });

  it("listens for workspace changes and closes the stream", () => {
    const t = bridgeTransport(() => respond(200, {}));
    const changed = vi.fn();
    const stop = bridgeAdapter(t).workspace.watch(changed);
    t.sources[0]?.dispatchEvent(new Event("workspace"));
    expect(changed).toHaveBeenCalledOnce();
    stop();
    expect(t.sources[0]?.readyState).toBe(2);
  });
});
