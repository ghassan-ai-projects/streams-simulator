import { describe, expect, it } from "vitest";
import { inTextField, ownsSpace, plainKey } from "./keys";

function element(html: string): HTMLElement {
  document.body.innerHTML = html;
  const found = document.body.querySelector("[data-pick]");
  if (!(found instanceof HTMLElement)) throw new Error("no element");
  return found;
}

describe("keyboard targets", () => {
  it("knows text fields, where typing must never trigger a shortcut", () => {
    expect(inTextField(element(`<input data-pick>`))).toBe(true);
    expect(inTextField(element(`<input type="search" data-pick>`))).toBe(true);
    expect(inTextField(element(`<textarea data-pick></textarea>`))).toBe(true);
    expect(inTextField(element(`<select data-pick></select>`))).toBe(true);
    expect(inTextField(element(`<input type="range" data-pick>`))).toBe(false);
    expect(inTextField(element(`<button data-pick></button>`))).toBe(false);
    expect(inTextField(null)).toBe(false);
  });

  it("leaves Space to controls that use it, but not to the playback slider", () => {
    expect(ownsSpace(element(`<input type="range" data-pick>`))).toBe(false);
    expect(ownsSpace(element(`<h1 tabindex="-1" data-pick>x</h1>`))).toBe(false);
    expect(ownsSpace(element(`<button data-pick></button>`))).toBe(true);
    expect(ownsSpace(element(`<details><summary data-pick>x</summary></details>`))).toBe(true);
    expect(ownsSpace(element(`<div role="application"><span data-pick>x</span></div>`))).toBe(true);
    expect(ownsSpace(element(`<input type="checkbox" data-pick>`))).toBe(true);
    expect(ownsSpace(document)).toBe(false);
  });

  it("tells a plain key from a chord", () => {
    expect(plainKey({ ctrlKey: false, metaKey: false, altKey: false })).toBe(true);
    expect(plainKey({ ctrlKey: true, metaKey: false, altKey: false })).toBe(false);
    expect(plainKey({ ctrlKey: false, metaKey: false, altKey: true })).toBe(false);
  });
});
