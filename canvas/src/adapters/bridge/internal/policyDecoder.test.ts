import { describe, expect, it } from "vitest";
import { liveWire } from "../../../test/liveFixture";
import { journalEntry } from "./journalDecoder";

const digest = `sha256:${"b".repeat(64)}`;
const other = `sha256:${"c".repeat(64)}`;
const unreadable = { kind: "live-unreadable", atNs: 10e9 };

const sessionWire = (plan: Record<string, unknown>) =>
  JSON.stringify({ type: "session", stream: "observer", seq: 2, at_ns: 0, mode: "connected", plan });

const refusal = {
  receipt: "refused",
  reason: "policy_not_allowed",
  report: "unknown",
  policy_digest: digest,
};

function liveWith(change: Record<string, unknown>): string {
  const wire = liveWire();
  return JSON.stringify({ ...wire, commands: [{ ...wire.commands[0], ...change }] });
}

describe("the policy allow-list in a recorded session", () => {
  const plan = { domain: "d", instances: [{ instance: "fan-01", origin: "simulated" }] };

  it("reads the recorded list in the binary's order, an empty one, and says nothing was recorded when it is absent", () => {
    expect(journalEntry(sessionWire({ ...plan, policies: [digest, other] }))).toMatchObject({
      policies: [digest, other],
    });
    expect(journalEntry(sessionWire({ ...plan, policies: [] }))).toMatchObject({ policies: [] });
    expect(journalEntry(sessionWire(plan))).toMatchObject({ policies: null });
  });

  it.each([
    ["text", digest],
    ["an object", {}],
    ["a number in the list", [digest, 3]],
    ["null", null],
  ])("drops a session whose list is %s", (_, policies) => {
    expect(journalEntry(sessionWire({ ...plan, policies }))).toBeNull();
  });
});

describe("the policy digest of a refused command", () => {
  it("is carried on a command refused for its policy", () => {
    expect(journalEntry(liveWith(refusal))).toMatchObject({
      live: { commands: [{ receipt: "refused", reason: "policy_not_allowed", policyDigest: digest }] },
    });
  });

  it("is empty on every other command", () => {
    expect(journalEntry(liveWith({}))).toMatchObject({ live: { commands: [{ policyDigest: "" }] } });
    expect(journalEntry(liveWith({ receipt: "refused", reason: "local_control", report: "unknown" }))).toMatchObject({
      live: { commands: [{ policyDigest: "" }] },
    });
  });

  it.each([
    ["a malformed digest", { ...refusal, policy_digest: "sha256:abc" }],
    ["an upper-case digest", { ...refusal, policy_digest: `sha256:${"B".repeat(64)}` }],
    ["a digest that is not text", { ...refusal, policy_digest: 7 }],
    ["a digest on an accepted command", { policy_digest: digest }],
    ["a digest on another refusal", { ...refusal, reason: "local_control" }],
  ])("marks a command with %s as unreadable instead of showing it", (_, change) => {
    expect(journalEntry(liveWith(change))).toEqual(unreadable);
  });

  it("changes the key of the live state, so a digest that appears is a new state", () => {
    const withDigest = journalEntry(liveWith(refusal));
    const without = journalEntry(liveWith({ ...refusal, policy_digest: undefined }));
    const key = (entry: typeof withDigest) => (entry?.kind === "live" ? entry.live.key : "");
    expect(key(withDigest)).not.toBe(key(without));
  });
});
