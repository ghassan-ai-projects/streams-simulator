import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MetaParts } from "./MetaParts";

describe("metadata parts", () => {
  it("glues each separator to the part before it, so a wrap never strands a lone dot", () => {
    const { container } = render(<MetaParts parts={["w-13", "Evaluate", "30 min", "seed 1", "15 Oct 2026, 02:35"]} />);
    expect(container.textContent).toBe("w-13 · Evaluate · 30 min · seed 1 · 15 Oct 2026, 02:35");
    const bound = [...container.querySelectorAll("span")];
    expect(bound).toHaveLength(5);
    expect(bound[0]?.textContent).toBe("w-13 ·");
    expect(bound[1]?.textContent).toBe("Evaluate ·");
    expect(bound[3]?.textContent).toBe("seed 1 ·");
    expect(bound[4]?.textContent).toBe("15 Oct 2026, 02:35");
  });

  it("keeps every part on the same no-wrap span class and drops empty parts", () => {
    const { container } = render(<MetaParts parts={["w-6", "", "Did not finish: the run did not close."]} />);
    const bound = [...container.querySelectorAll("span")];
    expect(bound).toHaveLength(2);
    expect(bound[0]?.className).not.toBe("");
    expect(bound[0]?.className).toBe(bound[1]?.className);
    expect(container.textContent).toBe("w-6 · Did not finish: the run did not close.");
  });
});
