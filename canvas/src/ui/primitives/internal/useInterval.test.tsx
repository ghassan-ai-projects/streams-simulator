import { act, render } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useInterval } from "./useInterval";

function Probe({ callback, active }: { readonly callback: () => void; readonly active: boolean }) {
  useInterval(callback, active ? 1000 : null);
  return null;
}

describe("useInterval", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("calls the latest callback every period while active", () => {
    const first = vi.fn();
    const second = vi.fn();
    const view = render(<Probe callback={first} active />);
    act(() => {
      vi.advanceTimersByTime(2000);
    });
    expect(first).toHaveBeenCalledTimes(2);
    view.rerender(<Probe callback={second} active />);
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(first).toHaveBeenCalledTimes(2);
    expect(second).toHaveBeenCalledTimes(1);
  });

  it("stops when inactive or unmounted", () => {
    const callback = vi.fn();
    const view = render(<Probe callback={callback} active={false} />);
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(callback).not.toHaveBeenCalled();
    view.rerender(<Probe callback={callback} active />);
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(callback).toHaveBeenCalledTimes(1);
    view.unmount();
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(callback).toHaveBeenCalledTimes(1);
  });
});
