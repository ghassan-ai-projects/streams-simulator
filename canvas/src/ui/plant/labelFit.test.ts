import { renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useLabelFit } from "./labelFit";

function measured(
  scrollWidth: number,
  clientWidth: number,
): { span: HTMLSpanElement; resize: (width: number) => void } {
  const span = document.createElement("span");
  let width = scrollWidth;
  Object.defineProperties(span, {
    scrollWidth: { get: () => width },
    clientWidth: { value: clientWidth },
  });
  return {
    span,
    resize: (next: number) => {
      width = next;
    },
  };
}

describe("label fit", () => {
  it("hides a label that is wider than its box and keeps one that fits", () => {
    const cut = measured(120, 80).span;
    const fits = measured(70, 80).span;
    renderHook(() => {
      useLabelFit({ current: cut }, 0.62);
      useLabelFit({ current: fits }, 0.62);
    });
    expect(cut.style.visibility).toBe("hidden");
    expect(fits.style.visibility).toBe("");
  });

  it("shows a label again when the zoom brings it back to fit", () => {
    const { span: label, resize } = measured(120, 80);
    const { rerender } = renderHook(
      ({ zoom }) => {
        useLabelFit({ current: label }, zoom);
      },
      {
        initialProps: { zoom: 0.5 },
      },
    );
    expect(label.style.visibility).toBe("hidden");
    resize(60);
    rerender({ zoom: 1 });
    expect(label.style.visibility).toBe("");
  });
});
