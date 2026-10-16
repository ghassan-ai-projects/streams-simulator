import { describe, expect, it } from "vitest";
import { createOperationIdentity } from "./operationIdentity";

describe("operation identity", () => {
  it("invalidates earlier operations when a new operation begins", () => {
    const identity = createOperationIdentity();
    const first = identity.begin();
    const second = identity.begin();
    expect(identity.isCurrent(first)).toBe(false);
    expect(identity.isCurrent(second)).toBe(true);
  });

  it("invalidates the active operation on close and supports reopening", () => {
    const identity = createOperationIdentity();
    const firstVisit = identity.begin();
    identity.invalidate();
    const reopened = identity.begin();
    expect(identity.isCurrent(firstVisit)).toBe(false);
    expect(identity.isCurrent(reopened)).toBe(true);
  });
});
