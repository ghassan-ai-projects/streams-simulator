import { render } from "@testing-library/react";
import { useRef } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ResizeWatcher } from "../../../test/dom";
import { stubResizeObserver } from "../../../test/dom";
import { useResizeObserver } from "./useResizeObserver";

function Probe({ callback }: { readonly callback: () => void }) {
  const ref = useRef<HTMLDivElement>(null);
  useResizeObserver(ref, callback);
  return <div ref={ref} data-testid="probe" />;
}

describe("useResizeObserver", () => {
  let observers: ResizeWatcher[] = [];

  beforeEach(() => {
    observers = stubResizeObserver();
  });

  it("observes the element, calls the latest callback and disconnects on unmount", () => {
    const first = vi.fn();
    const second = vi.fn();
    const view = render(<Probe callback={first} />);
    expect(observers).toHaveLength(1);
    expect(observers[0]?.observed).toEqual([view.getByTestId("probe")]);
    view.rerender(<Probe callback={second} />);
    expect(observers).toHaveLength(1);
    observers[0]?.trigger();
    expect(first).not.toHaveBeenCalled();
    expect(second).toHaveBeenCalledTimes(1);
    view.unmount();
    expect(observers[0]?.disconnected).toBe(true);
  });

  it("does nothing where ResizeObserver does not exist", () => {
    vi.stubGlobal("ResizeObserver", undefined);
    expect(() => render(<Probe callback={vi.fn()} />)).not.toThrow();
  });
});
