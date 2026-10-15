import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeAll, describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../model/plant";
import { domain } from "../../test/plantFixture";
import { Outline } from "./Outline";

describe("outline", () => {
  beforeAll(() => {
    Element.prototype.scrollIntoView = vi.fn();
  });

  it("marks the selection as current and says how many instances a filter keeps", () => {
    const onSelect = vi.fn();
    render(<Outline plant={parsePlant(domain)} hall={null} selected="CH-1" onSelect={onSelect} />);
    const outline = screen.getByRole("navigation", { name: "Outline" });
    expect(within(outline).getByRole("button", { name: /^CH-1/ })).toHaveAttribute("aria-current", "true");
    expect(within(outline).getByRole("button", { name: /^CT-1/ })).not.toHaveAttribute("aria-current");
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    fireEvent.change(screen.getByLabelText("Filter instances"), { target: { value: "CT" } });
    expect(screen.getByRole("status")).toHaveTextContent("1 instance matches.");
    fireEvent.change(screen.getByLabelText("Filter instances"), { target: { value: "zzz" } });
    expect(screen.getByRole("status")).toHaveTextContent('No instance matches "zzz".');
    fireEvent.change(screen.getByLabelText("Filter instances"), { target: { value: "" } });
    fireEvent.click(within(outline).getByRole("button", { name: /^CT-1/ }));
    expect(onSelect).toHaveBeenCalledWith("CT-1");
  });

  it("says a plant without instances has none yet", () => {
    render(
      <Outline plant={parsePlant({ id: "empty", instances: [] })} hall={null} selected={null} onSelect={vi.fn()} />,
    );
    expect(screen.getByRole("status")).toHaveTextContent("No instances yet.");
  });

  it("marks instances the binary flagged", () => {
    render(
      <Outline plant={parsePlant(domain)} hall={null} selected={null} onSelect={vi.fn()} flagged={new Set(["CT-1"])} />,
    );
    expect(screen.getByRole("button", { name: /^CT-1.*has issues$/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^CH-1/ })).not.toHaveTextContent("has issues");
  });

  it("indents areas under their parent and shows an empty area only when asked", () => {
    const plant = parsePlant({
      id: "tree",
      site: {
        name: "Site",
        areas: [{ name: "Site" }, { name: "Plant room", parent: "Site" }, { name: "Roof", parent: "Site" }],
      },
      instances: [{ kind: "pump", names: ["P-1"], area: "Plant room" }],
    });
    const { rerender } = render(<Outline plant={plant} hall={null} selected={null} onSelect={vi.fn()} />);
    expect(screen.getByText("Plant room")).toBeInTheDocument();
    expect(screen.queryByText("Roof")).not.toBeInTheDocument();
    rerender(<Outline plant={plant} hall={null} selected={null} onSelect={vi.fn()} showEmptyAreas />);
    const titles = screen.getAllByRole("group").map((group) => group.querySelector("summary")?.textContent);
    expect(titles).toEqual(["Site0 · empty", "Plant room1", "Roof0 · empty"]);
    expect(screen.getAllByRole("group").map((group) => group.getAttribute("style"))).toEqual([
      "margin-left: calc(var(--space-3) * 0);",
      "margin-left: calc(var(--space-3) * 1);",
      "margin-left: calc(var(--space-3) * 1);",
    ]);
    expect(screen.getAllByText("0 · empty")).toHaveLength(2);
  });

  it("lists the populated child areas of a site before its empty ones, each empty area on one row", () => {
    const plant = parsePlant({
      id: "tree",
      site: {
        name: "Site",
        areas: [
          { name: "Site" },
          { name: "Hall", parent: "Site" },
          { name: "Substation", parent: "Site" },
          { name: "House", parent: "Site" },
          { name: "Cabinet", parent: "House" },
          { name: "Yard", parent: "Site" },
        ],
      },
      instances: [
        { kind: "compressor", names: ["C-1", "C-2", "C-3"], area: "House" },
        { kind: "meter", names: ["M-1"], area: "Cabinet" },
        { kind: "pump", names: ["P-1"], area: "Yard" },
      ],
    });
    render(<Outline plant={plant} hall={null} selected={null} onSelect={vi.fn()} showEmptyAreas />);
    const titles = screen.getAllByRole("group").map((group) => group.querySelector("summary")?.textContent);
    expect(titles).toEqual(["Site0 · empty", "House3", "Cabinet1", "Yard1", "Hall0 · empty", "Substation0 · empty"]);
    const empty = screen.getAllByText("0 · empty");
    expect(empty).toHaveLength(3);
    expect(empty.every((note) => note.closest("summary")?.querySelectorAll("span").length === 2)).toBe(true);
    expect(screen.getByText("Substation")).toHaveAttribute("title", "Substation");
  });
});
