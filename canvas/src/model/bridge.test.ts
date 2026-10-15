import { describe, expect, it } from "vitest";
import {
  normalisePairingCode,
  pairingBlocked,
  pairingCodeComplete,
  pairingMessage,
  parseFailure,
  parseFile,
  parsePaired,
  parseWorkspace,
} from "./bridge";

describe("bridge payloads", () => {
  it("reads a workspace listing and tolerates missing fields", () => {
    const listing = parseWorkspace({
      name: "lab",
      domains: [{ path: "domains/a.domain.json", id: "a", title: "A", instances: 4 }, null],
      compositions: "not a list",
      runs: [
        {
          world_id: "w-1",
          ended: true,
          complete: true,
          incomplete_reason: "",
          explore: false,
          seed: 7,
          plant: "a",
          duration_ns: 1.8e12,
          recorded_at: "2026-10-14T10:00:00Z",
          faults: 2,
        },
        { world_id: "w-2", ended: true, complete: false, incomplete_reason: "stopped early", seed: null },
        { world_id: "w-3" },
      ],
    });
    expect(listing.name).toBe("lab");
    expect(listing.domains).toEqual([
      { path: "domains/a.domain.json", id: "a", title: "A", instances: 4 },
      { path: "", id: "", title: "", instances: 0 },
    ]);
    expect(listing.compositions).toEqual([]);
    expect(listing.runs).toEqual([
      {
        worldId: "w-1",
        ended: true,
        complete: true,
        incompleteReason: "",
        explore: false,
        seed: 7,
        plant: "a",
        durationS: 1800,
        recordedAt: "2026-10-14T10:00:00Z",
        faults: 2,
      },
      {
        worldId: "w-2",
        ended: true,
        complete: false,
        incompleteReason: "stopped early",
        explore: false,
        seed: null,
        plant: "",
        durationS: 0,
        recordedAt: "",
        faults: 0,
      },
      {
        worldId: "w-3",
        ended: false,
        complete: false,
        incompleteReason: "",
        explore: false,
        seed: null,
        plant: "",
        durationS: 0,
        recordedAt: "",
        faults: 0,
      },
    ]);
    expect(parseWorkspace(null).runs).toEqual([]);
  });

  it("reads the bridge error envelope, or names the status when there is none", () => {
    const failure = parseFailure(422, {
      error: {
        code: "composition_invalid",
        message: "bad",
        issues: [{ pointer: "/connections/29", message: "types differ" }],
      },
    });
    expect(failure).toEqual({
      status: 422,
      code: "composition_invalid",
      message: "bad",
      issues: [{ pointer: "/connections/29", message: "types differ" }],
      result: null,
    });
    expect(
      parseFailure(422, { error: { code: "replay_mismatch", message: "no", result: { matches: false } } }),
    ).toMatchObject({ code: "replay_mismatch", result: { matches: false } });
    expect(parseFailure(502, "oops")).toEqual({
      status: 502,
      code: "bridge_failed",
      message: "The bridge answered 502.",
      issues: [],
      result: null,
    });
  });

  it("reads files and the paired flag", () => {
    expect(parseFile({ path: "p", content: "{}", etag: "e" })).toEqual({ path: "p", content: "{}", etag: "e" });
    expect(parseFile(undefined)).toEqual({ path: "", content: "", etag: "" });
    expect(parsePaired({ paired: true })).toBe(true);
    expect(parsePaired({ paired: "yes" })).toBe(false);
  });

  it("keeps only base32 characters of a typed or pasted code", () => {
    expect(normalisePairingCode("3s2q-k7mp lxq4 99")).toBe("3S2QK7MPLXQ4");
    expect(normalisePairingCode("ab1c")).toBe("ABC");
    expect(pairingCodeComplete("3S2QK7MPLXQ4")).toBe(true);
    expect(pairingCodeComplete("3S2QK7")).toBe(false);
  });

  it("explains each pairing refusal in words and blocks only the final ones", () => {
    const of = (code: string) => ({ status: 401, code, message: "raw", issues: [] });
    expect(pairingMessage(of("wrong_code"))).toMatch(/did not match/);
    expect(pairingMessage(of("pairing_locked"))).toMatch(/locked/);
    expect(pairingMessage(of("pairing_used"))).toMatch(/already used/);
    expect(pairingMessage(of("other"))).toBe("raw");
    expect(pairingBlocked(of("pairing_locked"))).toBe(true);
    expect(pairingBlocked(of("pairing_used"))).toBe(true);
    expect(pairingBlocked(of("wrong_code"))).toBe(false);
    expect(pairingBlocked(null)).toBe(false);
  });
});
