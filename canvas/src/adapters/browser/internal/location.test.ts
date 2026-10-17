import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { hashLocation } from "..";

describe("hash location", () => {
  beforeEach(() => {
    window.location.hash = "";
  });

  afterEach(() => {
    window.location.hash = "";
  });

  it("reads the current hash", () => {
    window.location.hash = "#/runs";
    expect(hashLocation(window).read()).toBe("#/runs");
  });

  it("writes a new hash into history and replaces it when asked", () => {
    const location = hashLocation(window);
    const length = window.history.length;
    location.write("#/one");
    expect(window.location.hash).toBe("#/one");
    expect(window.history.length).toBe(length + 1);
    location.write("#/two", true);
    expect(window.location.hash).toBe("#/two");
    expect(window.history.length).toBe(length + 1);
  });

  it("does nothing when the hash is already current", () => {
    const location = hashLocation(window);
    location.write("#/same");
    const replace = vi.spyOn(window.history, "replaceState");
    const length = window.history.length;
    location.write("#/same", true);
    location.write("#/same");
    expect(replace).not.toHaveBeenCalled();
    expect(window.history.length).toBe(length);
  });

  it("notifies on hash changes until unsubscribed", () => {
    const listener = vi.fn();
    const unsubscribe = hashLocation(window).subscribe(listener);
    window.dispatchEvent(new HashChangeEvent("hashchange"));
    expect(listener).toHaveBeenCalledOnce();
    unsubscribe();
    window.dispatchEvent(new HashChangeEvent("hashchange"));
    expect(listener).toHaveBeenCalledOnce();
  });
});
