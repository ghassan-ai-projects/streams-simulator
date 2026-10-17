import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { EventMark } from "../../../model/review";
import { AlarmJump } from "./AlarmJump";

describe("alarm jump", () => {
  it("jumps to the next reported alarm and wraps at the end, selecting its source", () => {
    const marks: EventMark[] = [100, 200].map((atNs) => ({
      atNs,
      entity: "CH-1",
      kind: "alarm",
      tone: "alarm",
      label: "CH-1 hot",
      raw: "hot",
      untilNs: null,
    }));
    const onEvent = vi.fn();
    const { rerender } = render(<AlarmJump marks={marks} positionNs={150} onEvent={onEvent} />);
    fireEvent.click(screen.getByRole("button", { name: "Jump to alarm: CH-1 hot" }));
    expect(onEvent).toHaveBeenLastCalledWith(marks[1]);
    rerender(<AlarmJump marks={marks} positionNs={300} onEvent={onEvent} />);
    fireEvent.click(screen.getByRole("button", { name: "Jump to alarm: CH-1 hot" }));
    expect(onEvent).toHaveBeenLastCalledWith(marks[0]);
    rerender(<AlarmJump marks={[]} positionNs={0} onEvent={onEvent} />);
    const empty = screen.getByRole("button", { name: "Jump to alarm: none" });
    expect(empty).toHaveAttribute("aria-disabled", "true");
    fireEvent.click(empty);
    expect(onEvent).toHaveBeenCalledTimes(2);
  });
});
