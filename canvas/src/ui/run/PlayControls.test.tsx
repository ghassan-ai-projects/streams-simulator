import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PlayControls } from "./PlayControls";

describe("play controls", () => {
  it("announces play and pause only when they change, wherever focus is", () => {
    const handlers = { onPlay: vi.fn(), onPause: vi.fn(), onSpeed: vi.fn() };
    const view = render(<PlayControls playing={false} speed={60} {...handlers} />);
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    view.rerender(<PlayControls playing speed={60} {...handlers} />);
    expect(screen.getByRole("status")).toHaveTextContent("Playing at 60×");
    view.rerender(<PlayControls playing={false} speed={60} {...handlers} />);
    expect(screen.getByRole("status")).toHaveTextContent("Paused");
  });
});
