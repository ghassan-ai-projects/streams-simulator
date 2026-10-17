import { render, screen } from "@testing-library/react";
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
