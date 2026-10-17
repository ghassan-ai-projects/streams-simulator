import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useCutBottom } from "./useCutBottom";

function Probe() {
  const { ref, cut } = useCutBottom<HTMLDivElement>();
  return (
    <div ref={ref} data-testid="probe" data-cut={cut ? "true" : undefined}>
      <details>
        <summary>Faults</summary>
        <p>Available fault types</p>
      </details>
    </div>
  );
}

function metrics(node: HTMLElement, client: number, total: number, top: number) {
  Object.defineProperty(node, "clientHeight", { value: client, configurable: true });
  Object.defineProperty(node, "scrollHeight", { value: total, configurable: true });
  Object.defineProperty(node, "scrollTop", { value: top, configurable: true });
}

describe("the cut cue", () => {
  it("cues while content continues below, and clears at the bottom or when it all fits", () => {
    render(<Probe />);
    const probe = screen.getByTestId("probe");
    expect(probe).not.toHaveAttribute("data-cut");
    metrics(probe, 200, 500, 0);
    fireEvent.scroll(probe);
    expect(probe).toHaveAttribute("data-cut", "true");
    metrics(probe, 200, 500, 300);
    fireEvent.scroll(probe);
    expect(probe).not.toHaveAttribute("data-cut");
    metrics(probe, 500, 500, 0);
    fireEvent.scroll(probe);
    expect(probe).not.toHaveAttribute("data-cut");
  });

  it("updates when a native disclosure changes the content height without a scroll or resize", async () => {
    render(<Probe />);
    const probe = screen.getByTestId("probe");
    metrics(probe, 200, 200, 0);
    expect(probe).not.toHaveAttribute("data-cut");
    metrics(probe, 200, 500, 0);
    fireEvent.click(screen.getByText("Faults"));
    await waitFor(() => {
      expect(probe).toHaveAttribute("data-cut", "true");
    });
    metrics(probe, 200, 200, 0);
    fireEvent.click(screen.getByText("Faults"));
    await waitFor(() => {
      expect(probe).not.toHaveAttribute("data-cut");
    });
  });

  it("updates when a child changes its text without a parent render", async () => {
    render(<Probe />);
    const probe = screen.getByTestId("probe");
    metrics(probe, 200, 500, 0);
    screen.getByText("Available fault types").textContent = "An expanded fault description";
    await waitFor(() => {
      expect(probe).toHaveAttribute("data-cut", "true");
    });
  });
});
