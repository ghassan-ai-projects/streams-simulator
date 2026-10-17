import { describe, expect, it } from "vitest";
import { bridgeTransport, respond } from "../../../test/bridgeTransport";
import { bridgeAdapter } from "..";

function answering(result: unknown) {
  return bridgeTransport(() => respond(200, { result }));
}

function lastCall(t: ReturnType<typeof bridgeTransport>): { url: unknown; body: unknown } {
  const [url, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return { url, body: typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined };
}

const compose = { composition: "plants/p.composition.json", output: "plants/p.json", etag: "e1" };

describe("launch gateway", () => {
  it("creates a world and reads its id", async () => {
    const t = answering({ world_id: "w-9" });
    const outcome = await bridgeAdapter(t).launch.createWorld({ domain: "plants/p.json", seed: 7, explore: true });
    expect(outcome).toEqual({ ok: true, value: { worldId: "w-9" } });
    expect(lastCall(t)).toEqual({
      url: "/api/actions/world.create",
      body: { args: { domain: "plants/p.json", seed: 7, explore: true } },
    });
  });

  it.each([[{}], [{ world_id: "" }], [{ world_id: 4 }], [null]])("refuses a created world shaped %j", async (bad) => {
    const outcome = await bridgeAdapter(answering(bad)).launch.createWorld({ domain: "d", seed: 1, explore: false });
    expect(outcome).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });

  it("injects a fault, ends a run and destroys a world with the wire names", async () => {
    const injected = answering({ fault_id: "f1", simulated: true });
    expect(await bridgeAdapter(injected).launch.inject("w-1", "CH-1", "stuck")).toEqual({
      ok: true,
      value: undefined,
    });
    expect(lastCall(injected)).toEqual({
      url: "/api/actions/fault.inject",
      body: { args: { world_id: "w-1", entity_id: "CH-1", fault: "stuck" } },
    });
    const ended = answering({ run_artifact_path: "runs/w-1/run.json" });
    expect(await bridgeAdapter(ended).launch.endRun("w-1")).toEqual({ ok: true, value: undefined });
    expect(lastCall(ended)).toEqual({ url: "/api/actions/run.end", body: { args: { world_id: "w-1" } } });
    const destroyed = answering({ destroyed: true });
    await bridgeAdapter(destroyed).launch.destroyWorld("w-1");
    expect(lastCall(destroyed)).toEqual({ url: "/api/actions/world.destroy", body: { args: { world_id: "w-1" } } });
  });

  it("reports an unsimulated fault and a failed end as failures", async () => {
    const notInjected = answering({ fault_id: "f1", simulated: false });
    expect(await bridgeAdapter(notInjected).launch.inject("w", "e", "stuck")).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
    const unended = bridgeTransport(() => respond(404, { error: { code: "unknown_world", message: "gone" } }));
    expect(await bridgeAdapter(unended).launch.endRun("w")).toMatchObject({ ok: false });
  });

  it("invokes an effector and reads the acknowledgement", async () => {
    const t = answering({ accepted: true, reason: "", effect_eta_ns: 2e9 });
    const outcome = await bridgeAdapter(t).launch.invoke("w-1", "CH-1", "setpoint", { value: 6 });
    expect(outcome).toEqual({ ok: true, value: { accepted: true, reason: "", effectEtaNs: 2e9 } });
    expect(lastCall(t).body).toEqual({
      args: { world_id: "w-1", entity_id: "CH-1", effector: "setpoint", args: { value: 6 } },
    });
  });

  it("swallows a failed destroy rather than throwing", async () => {
    const t = bridgeTransport(() => Promise.reject(new Error("down")));
    await expect(bridgeAdapter(t).launch.destroyWorld("w-1")).resolves.toBeUndefined();
  });
});

describe("plant and editor gateways", () => {
  it.each([
    ["plant", (a: ReturnType<typeof bridgeAdapter>) => a.plant.compose(compose)],
    ["editor", (a: ReturnType<typeof bridgeAdapter>) => a.editor.compose(compose)],
  ])("composes through the %s gateway", async (_name, run) => {
    const t = answering({ channels: 4, faults: 2, effectors: 1 });
    expect(await run(bridgeAdapter(t))).toEqual({ ok: true, value: { channels: 4, faults: 2, effectors: 1 } });
    expect(lastCall(t)).toEqual({
      url: "/api/actions/domain.compose",
      body: { args: { composition: compose.composition, out: compose.output, etag: "e1" } },
    });
  });

  it.each([
    ["an example", { kind: "example", id: "mini" }, "copy", { example: "mini" }],
    ["a file", { kind: "file", path: "plants/a.json" }, "copy", { from: "plants/a.json" }],
    ["an empty start", { kind: "example", id: "mini" }, "empty", { example: "mini", empty: true }],
  ] as const)("generates a composition from %s", async (_name, source, mode, expected) => {
    const document = {
      schema_version: "0.1",
      id: "n",
      version: "1",
      title: "N",
      stresses: "load",
      axes: {},
      profiles: [],
      ground_truth: {},
    };
    const t = answering({ composition: document });
    const outcome = await bridgeAdapter(t).plant.generateComposition({ source, id: "n", title: "N", mode });
    expect(outcome).toEqual({ ok: true, value: document });
    expect(lastCall(t)).toEqual({
      url: "/api/actions/component.new",
      body: { args: { id: "n", title: "N", ...expected } },
    });
  });

  it("refuses a generated composition that is not a valid document", async () => {
    const outcome = await bridgeAdapter(answering({ composition: { id: "n" } })).plant.generateComposition({
      source: { kind: "example", id: "mini" },
      id: "n",
      title: "N",
      mode: "copy",
    });
    expect(outcome).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });

  it.each(["plant", "editor"] as const)("writes a file through the %s gateway with its etag", async (gateway) => {
    const t = bridgeTransport(() => respond(200, { path: "plants/p.json", etag: "e2" }));
    const outcome = await bridgeAdapter(t)[gateway].writeFile("plants/p.json", "{}", "e1");
    expect(outcome).toEqual({ ok: true, value: { path: "plants/p.json", etag: "e2" } });
    expect(t.fetch.mock.calls[0]?.[1]).toMatchObject({
      method: "PUT",
      body: '{"path":"plants/p.json","content":"{}","etag":"e1"}',
    });
  });

  it("reads the library, a component, its presets and the examples through their actions", async () => {
    const shown = {
      entry: { class: "plant/chiller", tags: ["hvac"], title: "Chiller" },
      document: { presets: { a: {} }, ports: [] },
    };
    const t = bridgeTransport((url) =>
      respond(
        200,
        url.endsWith("component.list")
          ? { result: { components: [{ kind: "equipment", id: "chiller" }] } }
          : url.endsWith("component.examples")
            ? { result: { examples: [{ id: "mini" }] } }
            : { result: shown },
      ),
    );
    const adapters = bridgeAdapter(t);
    expect(await adapters.editor.library()).toMatchObject({ ok: true, value: [{ ref: "chiller@1" }] });
    expect(await adapters.editor.component("chiller@1")).toMatchObject({ ok: true, value: { category: "cool" } });
    expect(await adapters.editor.presets("chiller@1")).toMatchObject({ ok: true, value: { options: [{ name: "a" }] } });
    expect(await adapters.plant.component("chiller@1")).toMatchObject({ ok: true, value: { title: "Chiller" } });
    expect(await adapters.plant.examples()).toMatchObject({ ok: true, value: [{ id: "mini", title: "mini" }] });
    expect(lastCall(t).body).toEqual({ args: {} });
  });

  it("refuses a library response that lacks components", async () => {
    expect(await bridgeAdapter(answering({ components: "all" })).editor.library()).toMatchObject({
      ok: false,
      failure: { kind: "unreadable" },
    });
  });
});

describe("review gateway", () => {
  it("fails the follower when the stream closes with an error", () => {
    const t = bridgeTransport(() => respond(200, {}));
    const failed = { entry: () => undefined, closed: () => undefined, failed: () => undefined };
    let count = 0;
    bridgeAdapter(t).review.follow("w-1", { ...failed, failed: () => (count += 1) });
    const source = t.events.mock.results[0]?.value as { readyState: number };
    t.listeners.get("error")?.();
    expect(count).toBe(1);
    source.readyState = 0;
    t.listeners.get("error")?.();
    expect(count).toBe(1);
  });
});
