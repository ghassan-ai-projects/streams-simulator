import { describe, expect, it } from "vitest";
import { changeDrive, chooseInstrument, chooseResponse, renameTwin, setArgument } from "../../../model/devices";
import { deferred, flushPromises } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { choicesPath, coolerPath, domainPath, refusal, twinBridge } from "../../../test/twinFixture";
import type { TwinDraft } from "./twinDraft";
import { builtOf, isStale, problemsOf } from "./twinSelectors";
import { createTwinSession } from "./twinStore";

const drive = { kind: "cooler", metric: "temperature", gain: -0.5, timeConstantS: 120 };
const at = { kind: "cooler", command: "set_duty", index: 0 };
const made = { id: "cooler-twin", digest: "d", channels: 1, faults: 1, effectors: 1, implements: [] };

function ready(twin: ReturnType<typeof createTwinSession>): TwinDraft {
  const state = twin.snapshot();
  if (state.phase !== "ready") throw new Error(state.phase);
  return state;
}

async function opened(fake: FakeBridge = twinBridge()) {
  const twin = createTwinSession(fake.devices);
  await twin.open(coolerPath);
  return { twin, fake };
}

function driven(twin: ReturnType<typeof createTwinSession>, timeConstantS = 120) {
  twin.edit((choices) => chooseResponse(choices, "cooler", "set_duty", "first_order"));
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
    expect(built?.made).toMatchObject({ id: "cooler-twin", implements: ["cooler"] });
    expect(ready(twin).issues).toEqual([]);
  });

  it("updates the plant it made over the version it made, and notices when the choices change after", async () => {
    const { twin, fake } = await opened();
    await twin.create();
    twin.edit((choices) => renameTwin(choices, "cooler-twin"));
    expect(isStale(ready(twin))).toBe(false);
    twin.edit((choices) => chooseInstrument(choices, "cooler", "temperature", "none"));
    expect(isStale(ready(twin))).toBe(true);
    expect(builtOf(ready(twin))).toBeNull();
    await twin.create();
    expect(twinEtags(fake)).toEqual(["", "d1"]);
    expect(fake.writes.map((write) => write.etag)).toEqual(["", "e1"]);
    expect(isStale(ready(twin))).toBe(false);
  });

  it("does not ask for a twin the form still has to finish", async () => {
    const { twin, fake } = await opened();
    twin.edit((choices) => setArgument(choices, "cooler", "set_duty", "duty", ""));
    expect(problemsOf(ready(twin))).toHaveLength(1);
    await twin.create();
    expect(fake.writes).toHaveLength(0);
    expect(twinEtags(fake)).toEqual([]);
  });

  it("puts each issue of the binary on the choice it names", async () => {
    const { twin, fake } = await opened();
    const pointer = "/kinds/0/commands/set_duty/drives/0/time_constant_s";
    fake.handlers.set("kind.twin", () => refusal(pointer, "must be above zero"));
    driven(twin, -5);
    await twin.create();
    const draft = ready(twin);
    expect(draft.build).toMatchObject({ kind: "refused" });
    expect(draft.issues).toEqual([
      {
        scope: "command",
        kind: "cooler",
        key: "set_duty",
        field: "time_constant_s",
        spot: "0",
        message: "must be above zero",
      },
    ]);
    expect(fake.calls.some((call) => call.name === "domain.validate")).toBe(false);
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
    await twin.replace();
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
    fake.actions.set("domain.validate", refusal("/kinds/0/implements", "differs from the kind file"));
    await twin.create();
    expect(ready(twin).build).toMatchObject({
      kind: "built",
      implemented: { kind: "failed", failure: { kind: "refused" } },
      lint: { kind: "unavailable" },
    });
    expect(fake.calls.some((call) => call.name === "domain.lint")).toBe(false);
  });

  it("keeps the findings of a plausibility check that found errors, and goes without one it cannot run", async () => {
    const { twin, fake } = await opened();
    const report = { errors: 1, warnings: 0, findings: [{ severity: "error", entity: "e1", detail: "pinned" }] };
    fake.actions.set("domain.lint", failure({ ...cannedFailure("domain_lint", 422, "errors"), result: report }));
    await twin.create();
    expect(ready(twin).build).toMatchObject({ lint: { kind: "done", report: { errors: 1 } } });
    fake.actions.set("domain.lint", failure(cannedFailure("bridge_failed", 500, "broken")));
    await twin.create();
    expect(ready(twin).build).toMatchObject({ lint: { kind: "unavailable" } });
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
