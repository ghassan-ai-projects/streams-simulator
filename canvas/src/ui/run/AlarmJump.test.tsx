import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type { EventMark } from "../../model/timeline";
import { AlarmJump } from "./AlarmJump";

it("jumps to the next reported alarm and wraps at the end, selecting the source when available", () => {
  const marks: EventMark[] = [100, 200].map((atNs) => ({
    atNs,
    entity: "CH-1",
    tone: "alarm",
    label: "CH-1 hot",
    raw: "hot",
    untilNs: null,
  }));
  const onSeek = vi.fn();
  const onEvent = vi.fn();
  const { rerender } = render(<AlarmJump marks={marks} positionNs={150} onSeek={onSeek} onEvent={onEvent} />);
  fireEvent.click(screen.getByRole("button", { name: "Jump to alarm: CH-1 hot" }));
  expect(onEvent).toHaveBeenLastCalledWith(marks[1]);
  expect(onSeek).not.toHaveBeenCalled();
  rerender(<AlarmJump marks={marks} positionNs={300} onSeek={onSeek} onEvent={undefined} />);
  fireEvent.click(screen.getByRole("button", { name: "Jump to alarm: CH-1 hot" }));
  expect(onSeek).toHaveBeenLastCalledWith(100);
  rerender(<AlarmJump marks={[]} positionNs={0} onSeek={onSeek} onEvent={undefined} />);
  const empty = screen.getByRole("button", { name: "Jump to alarm: none" });
  expect(empty).toHaveAttribute("aria-disabled", "true");
  fireEvent.click(empty);
  expect(onSeek).toHaveBeenCalledTimes(1);
});
