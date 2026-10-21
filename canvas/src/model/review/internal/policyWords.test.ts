import { describe, expect, it } from "vitest";
import { acceptanceDetail, copiedWords, copyFailedWords, digestListedWords, refusalWords } from "./sessionWords";
import type { SessionCommand } from "./sessionLive";

const refused: SessionCommand = {
  sentAtNs: 1,
  updatedAtNs: 2,
  commandId: "cmd-2",
  device: "fan",
  environment: "simulated",
  command: "set_cooling",
  bootId: "none",
  receipt: "refused",
  reason: "policy_not_allowed",
  report: "unknown",
  effect: "unknown",
  effectEvidence: "",
  policyDigest: "",
};

describe("a refusal for the policy", () => {
  it("says which side refused it: this session for a simulated device, the gateway for a real board", () => {
    expect(acceptanceDetail(refused)).toBe("Policy not allowed by this session");
    expect(acceptanceDetail({ ...refused, environment: "hardware" })).toBe("Policy not allowed by the gateway");
    expect(acceptanceDetail({ ...refused, environment: "unknown" })).toBe("Policy not allowed");
    expect(refusalWords("policy_not_allowed")).toBe("Policy not allowed");
  });

  it("leaves every other refusal and an accepted command as before", () => {
    expect(acceptanceDetail({ ...refused, reason: "stale_epoch" })).toBe("Stale authority epoch");
    expect(acceptanceDetail({ ...refused, receipt: "accepted", reason: "" })).toBe("");
    expect(acceptanceDetail({ ...refused, reason: "local_control" })).toBe("");
  });

  it("words the copy and the draft steps", () => {
    expect(copiedWords("cmd-2")).toBe("Copied the policy digest of cmd-2.");
    expect(copyFailedWords).toBe("Could not copy; select the digest instead.");
    expect(digestListedWords("mini")).toContain("draft session plan of mini");
    expect(digestListedWords("mini")).toContain("a running session keeps its list");
  });
});
