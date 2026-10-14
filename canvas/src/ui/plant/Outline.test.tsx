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
});
