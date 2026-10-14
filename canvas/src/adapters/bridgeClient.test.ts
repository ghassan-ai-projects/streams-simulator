import { describe, expect, it, vi } from "vitest";
import type { Transport } from "./bridgeClient";
import { bridgeClient } from "./bridgeClient";

function respond(status: number, body: unknown): Response {
  return new Response(typeof body === "string" ? body : JSON.stringify(body), { status });
}

function transport(handler: (url: string, init?: RequestInit) => Response | Promise<Response>) {
  const listeners = new Map<string, (event?: Event) => void>();
  const close = vi.fn();
  const events = vi.fn(
    () =>
      ({
        addEventListener: (name: string, fn: (event?: Event) => void) => listeners.set(name, fn),
        close,
        readyState: 2,
      }) as unknown as EventSource,
  );
  const fetch = vi.fn((url: string, init?: RequestInit) => Promise.resolve(handler(url, init)));
  return { fetch, events, listeners, close } satisfies Transport & Record<string, unknown>;
}

describe("bridge client", () => {
  it("pairs with a JSON body and same-origin credentials", async () => {
    const t = transport(() => respond(200, { paired: true }));
    const outcome = await bridgeClient(t).pair("AAAABBBBCCCC");
    expect(outcome).toEqual({ ok: true, value: true });
    const [url, init] = t.fetch.mock.calls[0] ?? [];
    expect(url).toBe("/api/session");
    expect(init).toMatchObject({ method: "POST", credentials: "same-origin", body: '{"code":"AAAABBBBCCCC"}' });
  });

  it("maps an error envelope to a failure and a network error to unreachable", async () => {
    const refused = await bridgeClient(
      transport(() => respond(401, { error: { code: "wrong_code", message: "wrong" } })),
    ).pair("x");
    expect(refused).toMatchObject({ ok: false, failure: { status: 401, code: "wrong_code" } });
    const down = await bridgeClient(transport(() => Promise.reject(new Error("down")))).session();
    expect(down).toMatchObject({ ok: false, failure: { code: "unreachable" } });
    const garbage = await bridgeClient(transport(() => respond(500, "not json"))).workspace();
    expect(garbage).toMatchObject({ ok: false, failure: { code: "bridge_failed", status: 500 } });
  });

  it("reads files by encoded path and unwraps action results", async () => {
    const t = transport((url) =>
      url.startsWith("/api/file")
        ? respond(200, { path: "runs/w-1/run.json", content: "{}", etag: "e1" })
        : respond(200, { result: { ok: 1 } }),
    );
    const client = bridgeClient(t);
    expect(await client.readFile("runs/w-1/run.json")).toEqual({
      ok: true,
      value: { path: "runs/w-1/run.json", content: "{}", etag: "e1" },
    });
    expect(t.fetch.mock.calls[0]?.[0]).toBe("/api/file?path=runs%2Fw-1%2Frun.json");
    expect(await client.action("run.report", { run: "w-1" })).toEqual({ ok: true, value: { ok: 1 } });
    expect(t.fetch.mock.calls[1]?.[1]).toMatchObject({ body: '{"args":{"run":"w-1"}}' });
    expect(await client.workspace()).toMatchObject({ ok: true, value: { runs: [] } });
  });

  it("follows a run's journal until it closes, and reports a dead stream", () => {
    const t = transport(() => respond(200, {}));
    const follower = { record: vi.fn(), closed: vi.fn(), failed: vi.fn() };
    const stop = bridgeClient(t).followRun("w 1", follower);
    expect(t.events).toHaveBeenCalledWith("/api/runs/w%201/events");
    t.listeners.get("record")?.(new MessageEvent("record", { data: '{"seq":1}' }));
    t.listeners.get("record")?.(new Event("record"));
    expect(follower.record.mock.calls).toEqual([['{"seq":1}'], [""]]);
    t.listeners.get("closed")?.();
    expect(follower.closed).toHaveBeenCalledOnce();
    expect(t.close).toHaveBeenCalled();
    t.listeners.get("error")?.();
    expect(follower.failed).toHaveBeenCalledOnce();
    stop();
  });

  it("listens for workspace changes and closes the stream", () => {
    const t = transport(() => respond(200, {}));
    const changed = vi.fn();
    const stop = bridgeClient(t).watchWorkspace(changed);
    t.listeners.get("workspace")?.();
    expect(changed).toHaveBeenCalledOnce();
    stop();
    expect(t.close).toHaveBeenCalledOnce();
  });
});
