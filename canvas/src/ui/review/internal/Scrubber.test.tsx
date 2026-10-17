import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { report as runReport } from "../../../test/runFixture";
import { Scrubber } from "./Scrubber";

const common = {
  startNs: 0,
  endNs: 3_600_000_000_000,
  positionNs: 0,
  playing: false,
  speed: 1,
  marks: [],
  onPlay: vi.fn(),
  onPause: vi.fn(),
  onSpeed: vi.fn(),
  onSeek: vi.fn(),
  onEvent: vi.fn(),
};

describe("scrubber clock", () => {
  it("shows a placeholder for the time and total until the run is read", () => {
    render(<Scrubber {...common} report={null} />);
    expect(screen.getByText("of --:--:--")).toBeInTheDocument();
  });

  it("shows the total once the run is read", () => {
    render(<Scrubber {...common} report={runReport} />);
    expect(screen.getByText("of 01:00:00")).toBeInTheDocument();
  });
});

const idle = { playing: false, speed: 60, onPlay: vi.fn(), onPause: vi.fn(), onSpeed: vi.fn() };

describe("scrubber controls and events", () => {
  it("seeks with the slider and jumps to a disturbance", () => {
    const seek = vi.fn();
    const onEvent = vi.fn();
    render(
      <Scrubber
        {...idle}
        startNs={0}
        endNs={3_600e9}
        positionNs={60e9}
        report={runReport}
        onSeek={seek}
        onEvent={onEvent}
      />,
    );
    const slider = screen.getByRole("slider", { name: "Playback time" });
    expect(slider).toHaveAttribute("aria-valuetext", "00:01:00");
    fireEvent.change(slider, { target: { value: "120" } });
    expect(seek).toHaveBeenLastCalledWith(120e9);
    fireEvent.click(screen.getByRole("button", { name: "00:10:00 CT-1 fan failure" }));
    expect(onEvent).toHaveBeenLastCalledWith(expect.objectContaining({ atNs: 600e9 }));
    fireEvent.click(screen.getByRole("button", { name: "Next event" }));
    expect(seek).toHaveBeenLastCalledWith(300e9);
    fireEvent.click(screen.getByRole("button", { name: "Forward 10 minutes" }));
    expect(seek).toHaveBeenLastCalledWith(660e9);
    expect(screen.getByText("Moved to 00:11:00")).toBeInTheDocument();
    fireEvent.keyDown(document.body, { key: "." });
    expect(seek).toHaveBeenLastCalledWith(300e9);
    fireEvent.keyDown(slider, { key: "ArrowRight", shiftKey: true });
    expect(seek).toHaveBeenLastCalledWith(120e9);
  });

  it("selects a fault's event at its own time from its mark on the track, as its label does", () => {
    const seek = vi.fn();
    const onEvent = vi.fn();
    const { container } = render(
      <Scrubber
        {...idle}
        startNs={0}
        endNs={3_600e9}
        positionNs={0}
        report={runReport}
        onSeek={seek}
        onEvent={onEvent}
      />,
    );
    const marks = [...container.querySelectorAll<HTMLButtonElement>("button[tabindex='-1']")];
    expect(marks.length).toBeGreaterThan(0);
    marks.forEach((mark) => {
      fireEvent.click(mark);
    });
    expect(onEvent).toHaveBeenCalledWith(expect.objectContaining({ atNs: 600e9 }));
    expect(seek).not.toHaveBeenCalled();
  });

  it("plays, pauses and picks a speed", () => {
    const controls = { ...idle, onPlay: vi.fn(), onPause: vi.fn(), onSpeed: vi.fn() };
    const view = render(
      <Scrubber
        {...controls}
        startNs={0}
        endNs={60e9}
        positionNs={0}
        report={null}
        onSeek={vi.fn()}
        onEvent={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Play" }));
    expect(controls.onPlay).toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "60×" })).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(screen.getByRole("button", { name: "600×" }));
    expect(controls.onSpeed).toHaveBeenCalledWith(600);
    view.rerender(
      <Scrubber
        {...controls}
        playing
        startNs={0}
        endNs={60e9}
        positionNs={0}
        report={null}
        onSeek={vi.fn()}
        onEvent={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Pause" }));
    expect(controls.onPause).toHaveBeenCalled();
  });

  it("lists no events and disables Next event before the report arrives", () => {
    render(
      <Scrubber {...idle} startNs={0} endNs={0} positionNs={0} report={null} onSeek={vi.fn()} onEvent={vi.fn()} />,
    );
    expect(screen.getByRole("list", { name: "Events" })).toBeEmptyDOMElement();
    expect(screen.getByRole("button", { name: "Next event: none" })).toHaveAttribute("aria-disabled", "true");
  });
});
