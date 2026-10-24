import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
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
