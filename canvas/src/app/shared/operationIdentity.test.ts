import { describe, expect, it } from "vitest";
import { createOperationIdentity } from "./operationIdentity";

describe("operation identity", () => {
  it("invalidates earlier operations when a new operation begins", () => {
    const identity = createOperationIdentity();
    const first = identity.begin();
    const second = identity.begin();
    expect(first.live()).toBe(false);
    expect(second.live()).toBe(true);
  });

  it("invalidates the active operation and supports beginning again", () => {
    const identity = createOperationIdentity();
    const firstVisit = identity.begin();
    identity.invalidate();
    expect(firstVisit.live()).toBe(false);
    const reopened = identity.begin();
    expect(firstVisit.live()).toBe(false);
    expect(reopened.live()).toBe(true);
  });

  it("never resurrects an earlier operation when a later one begins and ends", () => {
    const identity = createOperationIdentity();
    const a = identity.begin();
    const b = identity.begin();
    identity.invalidate();
    const again = identity.begin();
    expect(a.live()).toBe(false);
    expect(b.live()).toBe(false);
    expect(again.live()).toBe(true);
  });

  it("reports the active operation, or none once invalidated", () => {
    const identity = createOperationIdentity();
    expect(identity.current()).toBeNull();
    const operation = identity.begin();
    expect(identity.current()).toBe(operation);
    identity.invalidate();
    expect(identity.current()).toBeNull();
  });
});
