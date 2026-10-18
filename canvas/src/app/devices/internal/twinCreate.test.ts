import { describe, expect, it } from "vitest";
import { changeDrive, chooseInstrument, chooseResponse, renameTwin, setArgument } from "../../../model/devices";
import { deferred, flushPromises } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { choicesPath, cooler, coolerPath, domainPath, refusal, twinBridge, withPlant } from "../../../test/twinFixture";
import type { TwinDraft } from "./twinDraft";
import { builtOf, isStale, problemsOf, verdictOf } from "./twinSelectors";
import { createTwinSession } from "./twinStore";

const drive = { kind: coolerPath, metric: "m1", gain: -0.5, timeConstantS: 120 };
const at = { kind: coolerPath, command: "c1", index: 0 };
const made = { id: "cooler-sensor-twin", digest: "d", channels: 1, faults: 1, effectors: 1, implements: [] };

function ready(twin: ReturnType<typeof createTwinSession>): TwinDraft {
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function opened(fake: FakeBridge = twinBridge()) {
  const twin = createTwinSession(fake.devices, () => cooler.kind);
  await twin.open(coolerPath);
  return { twin, fake };
}

function driven(twin: ReturnType<typeof createTwinSession>, timeConstantS = 120) {
  twin.edit((choices) => chooseResponse(choices, coolerPath, "c1", "first_order"));
  twin.edit((choices) => changeDrive(choices, at, { ...drive, timeConstantS }));
}

function twinEtags(fake: FakeBridge) {
  return fake.calls.filter((call) => call.name === "kind.twin").map((call) => call.args.etag);
}

describe("creating the twin", () => {
  it("saves the choices, asks the binary, then checks the plant it made", async () => {
    const { twin, fake } = await opened();
    await twin.create();
    expect(fake.writes[0]).toMatchObject({ path: choicesPath, etag: "" });
    const asked = fake.calls.filter((call) => call.name === "kind.twin");
    expect(asked[0]?.args).toEqual({ choices: choicesPath, out: domainPath, etag: "" });
    expect(fake.calls.map((call) => call.name).slice(-3)).toEqual(["kind.twin", "domain.validate", "domain.lint"]);
    const built = builtOf(ready(twin));
    expect(built).toMatchObject({ path: domainPath, implemented: { kind: "passed" }, lint: { kind: "done" } });
    expect(built?.made).toMatchObject({ id: "cooler-sensor-twin", implements: ["cooler"] });
    expect(built && verdictOf(built)).toBe("ready");
    expect(ready(twin).issues).toEqual([]);
    expect(ready(twin).onDisk).toBe(true);
  });

  it("updates the plant it made over the version it made, and notices when the choices change after", async () => {
    const { twin, fake } = await opened();
    await twin.create();
    twin.edit((choices) => renameTwin(choices, "cooler-sensor-twin"));
    expect(isStale(ready(twin))).toBe(false);
    twin.edit((choices) => chooseInstrument(choices, coolerPath, "m1", "none"));
    twin.edit((choices) => chooseInstrument(choices, coolerPath, "m1", "streamsim/thermistor@1"));
    expect(isStale(ready(twin))).toBe(true);
    expect(builtOf(ready(twin))).toBeNull();
    await twin.create();
    expect(twinEtags(fake)).toEqual(["", "d1"]);
    expect(fake.writes.map((write) => write.etag)).toEqual(["", "e1"]);
    expect(isStale(ready(twin))).toBe(false);
  });

  it("does not ask for a twin the form still has to finish", async () => {
    const { twin, fake } = await opened();
    twin.edit((choices) => setArgument(choices, coolerPath, "c1", "duty", ""));
    expect(problemsOf(ready(twin))).toHaveLength(1);
    await twin.create();
    expect(fake.writes).toHaveLength(0);
    expect(twinEtags(fake)).toEqual([]);
  });

  it("does not start a second creation while one is under way", async () => {
    const { twin, fake } = await opened();
    await Promise.all([twin.create(), twin.create()]);
    expect(twinEtags(fake)).toEqual([""]);
  });

  it("puts each issue of the binary on the choice it names, in plain words, and drops it once the value changes", async () => {
    const { twin, fake } = await opened();
    const pointer = "/kinds/0/commands/set_duty/drives/0/time_constant_s";
    fake.handlers.set("kind.twin", () => refusal({ pointer, message: "must be > 0", keyword: "exclusiveMinimum" }));
    driven(twin, 5);
    await twin.create();
    const draft = ready(twin);
    expect(draft.build).toMatchObject({ kind: "refused" });
    expect(draft.issues).toEqual([
      {
        scope: "command",
        kind: coolerPath,
        key: "c1",
        field: "time_constant_s",
        spot: "0",
        pointer,
        keyword: "exclusiveMinimum",
        message: "must be > 0",
        text: "Must be above zero",
      },
    ]);
    expect(fake.calls.some((call) => call.name === "domain.validate")).toBe(false);
    twin.edit((choices) => renameTwin(choices, "other-name"));
    expect(ready(twin).issues).toHaveLength(1);
    twin.edit((choices) => changeDrive(choices, at, { timeConstantS: 30 }));
    expect(ready(twin).issues).toEqual([]);
    await twin.create();
    expect(ready(twin).issues).toHaveLength(1);
  });

  it("asks before replacing a plant it did not make, then replaces it over its version", async () => {
    const { twin, fake } = await opened();
    fake.handlers.set("kind.twin", () => failure(cannedFailure("conflict", 409, "changed")));
    await twin.create();
    expect(ready(twin).build).toEqual({ kind: "exists" });
    fake.handlers.delete("kind.twin");
    fake.actions.set("kind.twin", success(made));
    fake.addFile(domainPath, { id: "mine" }, "d7");
    await Promise.all([twin.replace(), twin.replace()]);
    expect(twinEtags(fake)).toEqual(["", "d7"]);
    expect(builtOf(ready(twin))).not.toBeNull();
    await twin.replace();
    expect(twinEtags(fake)).toHaveLength(2);
  });

  it("replaces without a version when the plant has gone", async () => {
    const { twin, fake } = await opened();
    fake.handlers.set("kind.twin", () => failure(cannedFailure("conflict", 409, "changed")));
    await twin.create();
    fake.handlers.delete("kind.twin");
    fake.actions.set("kind.twin", success(made));
    await twin.replace();
    expect(twinEtags(fake)).toEqual(["", ""]);
  });

  it("says when the binary could not be asked", async () => {
    const { twin, fake } = await opened();
    fake.handlers.set("kind.twin", () => failure(cannedFailure("bridge_failed", 500, "broken")));
    await twin.create();
    expect(ready(twin).build).toMatchObject({ kind: "failed", failure: { kind: "unavailable" } });
  });

  it("says when the plant does not match its device, and then does not check it further", async () => {
    const { twin, fake } = await opened();
    fake.actions.set(
      "domain.validate",
      refusal({ pointer: "/kinds/0/implements", message: "differs", keyword: "implements_drift" }),
    );
    await twin.create();
    expect(ready(twin).build).toMatchObject({
      kind: "built",
      implemented: { kind: "failed", failure: { kind: "refused" } },
      lint: { kind: "unavailable" },
    });
    expect(fake.calls.some((call) => call.name === "domain.lint")).toBe(false);
    const built = builtOf(ready(twin));
    expect(built && verdictOf(built)).toBe("failed");
  });

  it("keeps the findings of a plausibility check that found errors, and goes without one it cannot run", async () => {
    const { twin, fake } = await opened();
    const report = { errors: 1, warnings: 0, findings: [{ severity: "error", entity: "e1", detail: "pinned" }] };
    fake.actions.set("domain.lint", failure({ ...cannedFailure("domain_lint", 422, "errors"), result: report }));
    await twin.create();
    expect(ready(twin).build).toMatchObject({ lint: { kind: "done", report: { errors: 1 } } });
    const built = builtOf(ready(twin));
    expect(built && verdictOf(built)).toBe("failed");
    fake.actions.set("domain.lint", failure(cannedFailure("bridge_failed", 500, "broken")));
    await twin.create();
    expect(ready(twin).build).toMatchObject({ lint: { kind: "unavailable" } });
    const again = builtOf(ready(twin));
    expect(again && verdictOf(again)).toBe("ready");
  });

  it("stops following a creation once the screen was left", async () => {
    const { twin, fake } = await opened();
    const gate = deferred<Canned<unknown>>();
    fake.handlers.set("kind.twin", () => gate.promise);
    const creating = twin.create();
    await flushPromises();
    twin.close();
    gate.resolve(success(made));
    await creating;
    expect(twin.snapshot()).toEqual({ phase: "idle" });
  });
});

describe("a twin that was made before", () => {
  it("is found on disk when the behaviour opens, and checked against its device", async () => {
    const fake = withPlant(twinBridge());
    const { twin } = await opened(fake);
    const draft = ready(twin);
    expect(draft.onDisk).toBe(true);
    const built = builtOf(draft);
    expect(built).toMatchObject({
      path: domainPath,
      made: null,
      implemented: { kind: "passed" },
      lint: { kind: "done" },
    });
    expect(built && verdictOf(built)).toBe("ready");
    await twin.create();
    expect(twinEtags(fake)).toEqual(["d5"]);
  });

  it("is reported as failing when the plant on disk no longer matches its device", async () => {
    const fake = withPlant(twinBridge());
    fake.actions.set(
      "domain.validate",
      refusal({ pointer: "/kinds/0/implements", message: "x", keyword: "implements_drift" }),
    );
    const { twin } = await opened(fake);
    const built = builtOf(ready(twin));
    expect(built && verdictOf(built)).toBe("failed");
    expect(built?.lint).toEqual({ kind: "unavailable" });
  });

  it("is not offered when the plant that carries its name is not on disk", async () => {
    const { twin } = await opened();
    expect(ready(twin).onDisk).toBe(false);
    expect(builtOf(ready(twin))).toBeNull();
  });
});
