import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../ports";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { pairingBlocked, SessionStore } from "./sessionStore";

async function refused(code: string) {
  const bridge = new FakeBridge();
  bridge.pairOutcomes.push(failed(cannedFailure(code)));
  const outcome = await bridge.session.pair("AAAABBBBCCCC");
  if (outcome.ok) throw new Error("pairing was accepted");
  return outcome.failure;
}

async function unpaired() {
  const bridge = new FakeBridge();
  const store = new SessionStore(bridge.session);
  await store.check();
  return { bridge, store };
}

describe("session", () => {
  it("starts paired when the bridge already knows this browser", async () => {
    const bridge = new FakeBridge();
    bridge.paired = true;
    const store = new SessionStore(bridge.session);
    expect(store.snapshot().phase).toBe("checking");
    await store.check();
    expect(store.snapshot()).toEqual({ phase: "paired" });
  });

  it("reports an unreachable bridge with its failure", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = failed(cannedFailure("unreachable", 0, "gone"));
    const store = new SessionStore(bridge.session);
    await store.check();
    expect(store.snapshot()).toMatchObject({ phase: "unreachable", failure: { kind: "unreachable" } });
  });

  it("pairs only with a complete code and sends the normalised code", async () => {
    const { bridge, store } = await unpaired();
    store.type("3s2q-k7mp");
    expect(store.canPair()).toBe(false);
    await store.pair();
    expect(bridge.pairedCodes).toEqual([]);
    store.type("3s2q-k7mp-lxq4");
    expect(store.canPair()).toBe(true);
    await store.pair();
    expect(bridge.pairedCodes).toEqual(["3S2QK7MPLXQ4"]);
    expect(store.snapshot()).toEqual({ phase: "paired" });
  });

  it("keeps the code after a wrong code and clears the error on the next keystroke", async () => {
    const { bridge, store } = await unpaired();
    bridge.pairOutcomes.push(failed(cannedFailure("wrong_code")));
    store.type("AAAABBBBCCCC");
    await store.pair();
    expect(store.snapshot()).toMatchObject({
      phase: "unpaired",
      code: "AAAABBBBCCCC",
      failure: { kind: "code-rejected", diagnostic: { displayCode: "wrong_code" } },
    });
    store.type("AAAABBBBCCCD");
    expect(store.snapshot()).toMatchObject({ failure: null, code: "AAAABBBBCCCD" });
  });

  it("keeps a code typed while a second status check is still answering", async () => {
    const { store } = await unpaired();
    store.type("3s2q-k7mp-lxq4");
    await store.check();
    expect(store.snapshot()).toMatchObject({ phase: "unpaired", code: "3S2QK7MPLXQ4" });
    expect(store.canPair()).toBe(true);
  });

  it("blocks typing and pairing once the bridge locks or the code was used", async () => {
    for (const code of ["pairing_locked", "pairing_used"]) {
      const { bridge, store } = await unpaired();
      bridge.pairOutcomes.push(failed(cannedFailure(code)));
      store.type("AAAABBBBCCCC");
      await store.pair();
      const expected = code === "pairing_locked" ? "locked" : "already-used";
      expect(store.snapshot()).toMatchObject({
        phase: "unpaired",
        code: "",
        failure: { kind: "pairing-closed", reason: expected, diagnostic: { displayCode: code } },
      });
      store.type("AAAABBBBCCCC");
      expect(store.snapshot()).toMatchObject({ code: "" });
      expect(store.canPair()).toBe(false);
    }
  });

  it("blocks pairing only for the refusals that close it", async () => {
    expect(pairingBlocked(await refused("pairing_locked"))).toBe(true);
    expect(pairingBlocked(await refused("pairing_used"))).toBe(true);
    expect(pairingBlocked(await refused("wrong_code"))).toBe(false);
    expect(pairingBlocked(null)).toBe(false);
  });

  it("ignores typing and pairing outside the unpaired phase", async () => {
    const bridge = new FakeBridge();
    bridge.paired = true;
    bridge.pairOutcomes.push(success(true));
    const store = new SessionStore(bridge.session);
    await store.check();
    store.type("AAAA");
    await store.pair();
    expect(store.snapshot()).toEqual({ phase: "paired" });
    expect(bridge.pairedCodes).toEqual([]);
  });
});
