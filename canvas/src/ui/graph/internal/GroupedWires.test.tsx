import { fireEvent, render, renderHook, screen } from "@testing-library/react";
import { ReactFlowProvider } from "@xyflow/react";
import { act } from "react";
import { beforeEach, describe, expect, it } from "vitest";
import { plantView } from "../../../model/graph";
import { parsePlant } from "../../../model/plant";
import { stubResizeObserver } from "../../../test/dom";
import { beltComposition, beltDomain, beltSpecs, composition, domain } from "../../../test/plantFixture";
import { flowEdges } from "./flow";
import { ZoomBar } from "./GraphControls";
import { PlantGraph } from "./PlantGraph";
import { useWires } from "./useWires";

beforeEach(() => {
  stubResizeObserver();
});

const belt = parsePlant(beltDomain, beltComposition);

function Bar({ plant }: { readonly plant: typeof belt }) {
  const { wires, control } = useWires(plant);
  return (
    <ReactFlowProvider>
      <ZoomBar wires={control} />
      <output>{wires}</output>
    </ReactFlowProvider>
  );
}

describe("the group links control", () => {
  it("starts pressed on a plant with parallel links and switches to one wire per link", () => {
    render(<Bar plant={belt} />);
    const toggle = screen.getByRole("button", { name: "Group links" });
    expect(toggle).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("status")).toHaveTextContent("grouped");
    fireEvent.click(toggle);
    expect(toggle).toHaveAttribute("aria-pressed", "false");
    expect(screen.getByRole("status")).toHaveTextContent("each");
  });

  it("is not offered where no two instances share more than one link", () => {
    render(<Bar plant={parsePlant(domain, composition)} />);
    expect(screen.queryByRole("button", { name: "Group links" })).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("each");
  });

  it("forgets the choice when another plant is shown", () => {
    const { result, rerender } = renderHook(({ plant }) => useWires(plant), { initialProps: { plant: belt } });
    act(() => {
      result.current.control?.onChange("each");
    });
    expect(result.current.wires).toBe("each");
    rerender({ plant: { ...belt, id: "other" } });
    expect(result.current.wires).toBe("grouped");
  });

  it("keeps the opening choice while an edited plant gains or loses links", () => {
    const sparse: typeof belt = { ...belt, links: belt.links.filter((link) => link.fromPort !== "enabled_out") };
    const kept = renderHook(({ plant }) => useWires(plant, "keep"), { initialProps: { plant: sparse } });
    expect(kept.result.current.wires).toBe("each");
    kept.rerender({ plant: belt });
    expect(kept.result.current.wires).toBe("each");
    const followed = renderHook(({ plant }) => useWires(plant), { initialProps: { plant: sparse } });
    followed.rerender({ plant: belt });
    expect(followed.result.current.wires).toBe("grouped");
  });

  it("says in an editor that pins come back when grouping is off", () => {
    const control = { value: "grouped" as const, onChange: () => undefined };
    render(
      <ReactFlowProvider>
        <ZoomBar wires={control} linking />
      </ReactFlowProvider>,
    );
    expect(screen.getByRole("button", { name: "Group links" }).title).toMatch(/draw a link between two of them/);
  });

  it("explains the count on a grouped wire in the legend", () => {
    render(<Bar plant={belt} />);
    fireEvent.click(screen.getByRole("button", { name: "Legend" }));
    expect(screen.getByText(/drawn as one wire/)).toBeInTheDocument();
    expect(screen.queryByText("Hollow: not connected")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Group links" }));
    expect(screen.queryByText(/drawn as one wire/)).not.toBeInTheDocument();
    expect(screen.getByText("Hollow: not connected")).toBeInTheDocument();
  });
});

describe("a grouped plant graph", () => {
  const view = plantView(belt, beltSpecs, { open: null, wires: "grouped" });

  it("counts both ways on one wire, with an arrow at each end", () => {
    const [feed, drive] = flowEdges(view, new Map());
    expect(feed).toMatchObject({ label: "×2", ariaLabel: "F-1 and B-1, 2 links: 1 to B-1, 1 to F-1" });
    expect(feed?.markerStart).toMatchObject({ orient: "auto-start-reverse" });
    expect(drive).toMatchObject({ label: "×5", className: "wire family-power wire-grouped" });
  });

  it("names a one-way grouped wire by its direction and count", () => {
    const power = { ...belt, links: belt.links.filter((link) => /drag|run_cap/.test(link.fromPort ?? "")) };
    const [wire] = flowEdges(plantView(power, beltSpecs, { open: null, wires: "grouped" }), new Map());
    expect(wire).toMatchObject({ ariaLabel: "D-1 to B-1, 2 links", className: "wire family-power wire-grouped" });
    expect(wire?.markerStart).toBeUndefined();
  });

  it("lists each instance's measured channels with units instead of its ports", () => {
    render(
      <ReactFlowProvider>
        <PlantGraph view={view} viewKey="v" selected={null} onSelect={() => undefined} />
      </ReactFlowProvider>,
    );
    expect(screen.getByText("tail speed")).toBeInTheDocument();
    expect(screen.getByText("rpm")).toBeInTheDocument();
    expect(screen.getByText("winding")).toBeInTheDocument();
    expect(screen.getByText("°C")).toBeInTheDocument();
    expect(screen.getByText("+2 more")).toBeInTheDocument();
    expect(screen.queryByText("drag out")).not.toBeInTheDocument();
  });

  it("keys only the port families the plant uses in the legend", () => {
    render(
      <ReactFlowProvider>
        <ZoomBar />
        <PlantGraph view={view} viewKey="v" selected={null} onSelect={() => undefined} />
      </ReactFlowProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Legend" }));
    expect(screen.getByText("Bulk material")).toBeInTheDocument();
    expect(screen.getByText("Power")).toBeInTheDocument();
    expect(screen.getByText("Signal")).toBeInTheDocument();
    expect(screen.queryByText("Compressed air")).not.toBeInTheDocument();
  });

  it("counts the ports of an instance that has no measured channel to list, as in the editor", () => {
    const bare = { ...belt, kinds: new Map() };
    render(
      <ReactFlowProvider>
        <PlantGraph
          view={plantView(bare, beltSpecs, { open: null, wires: "grouped" })}
          viewKey="v"
          selected={null}
          onSelect={() => undefined}
        />
      </ReactFlowProvider>,
    );
    expect(screen.getByText("4 inputs · 3 outputs")).toBeInTheDocument();
    expect(screen.getByText("1 input · 1 output")).toBeInTheDocument();
  });

  it("marks a type hidden in the legend and counts it on the legend button, for this graph only", () => {
    render(
      <>
        <ReactFlowProvider>
          <ZoomBar />
        </ReactFlowProvider>
        <ReactFlowProvider>
          <ZoomBar />
        </ReactFlowProvider>
      </>,
    );
    const [first, second] = screen.getAllByRole("button", { name: "Legend" });
    if (first === undefined || second === undefined) throw new Error("two legends expected");
    fireEvent.click(first);
    const power = screen.getByRole("button", { name: "Power" });
    expect(power).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(power);
    expect(screen.getByRole("button", { name: /^Power\s*hidden$/ })).toHaveAttribute("aria-pressed", "false");
    expect(first).toHaveTextContent("Legend· 1 hidden");
    expect(second).toHaveTextContent(/^Legend$/);
    fireEvent.click(screen.getByRole("button", { name: /^Power\s*hidden$/ }));
    expect(first).toHaveTextContent(/^Legend$/);
  });
});
