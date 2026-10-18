import { render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { InstanceOrigin } from "../../../model/graph";
import { parsePlant } from "../../../model/plant";
import { stubScrollIntoView } from "../../../test/dom";
import { domain } from "../../../test/plantFixture";
import { OriginMark } from "./OriginMark";
import { Outline } from "./Outline";

const real: InstanceOrigin = { instance: "CH-1", kind: "real", board: "uno-07", device: "ch-1", bootId: "61" };
const simulated: InstanceOrigin = {
  instance: "CT-1",
  kind: "simulated",
  board: "mega-01",
  device: "ct-1",
  bootId: "b-1",
};

describe("origin marks", () => {
  beforeEach(() => {
    stubScrollIntoView();
  });

  it("tells a real board from a simulated instance by shape, colour and words", async () => {
    const { container } = render(
      <p>
        <OriginMark origin={real} />
        <OriginMark origin={simulated} variant="glyph" />
        <OriginMark origin={null} />
        <OriginMark origin={{ ...real, kind: "unknown" }} voice="silent" />
      </p>,
    );
    const [tag, glyph, unknown] = container.querySelectorAll("[data-origin]");
    expect(tag).toHaveAttribute("data-origin", "real");
    expect(tag).toHaveTextContent("Real");
    expect(tag).toHaveAttribute("title", "Real · board uno-07 · boot 61");
    expect(glyph).toHaveAttribute("data-origin", "simulated");
    expect(glyph).toHaveTextContent("Simulated · board mega-01");
    expect(tag?.querySelector("svg")?.outerHTML).not.toBe(glyph?.querySelector("svg")?.outerHTML);
    expect(unknown).toHaveAttribute("title", "Origin unknown: the run recorded no origin for it");
    expect(unknown?.querySelector(".sr-only")).toBeNull();
    expect(container.querySelectorAll("[data-origin]")).toHaveLength(3);
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it("marks each outline row with the origin of its instance", () => {
    const origins = new Map<string, InstanceOrigin>([
      [real.instance, real],
      [simulated.instance, simulated],
    ]);
    render(<Outline plant={parsePlant(domain)} hall={null} selected={null} onSelect={vi.fn()} origins={origins} />);
    const outline = screen.getByRole("navigation", { name: "Outline" });
    expect(within(outline).getByRole("button", { name: /^CH-1.*Real · board uno-07 · boot 61$/ })).toBeInTheDocument();
    expect(within(outline).getByRole("button", { name: /^CT-1.*Simulated · board mega-01$/ })).toBeInTheDocument();
    expect(within(outline).getByRole("button", { name: /^MB-1/ }).querySelector("[data-origin]")).toBeNull();
  });
});
