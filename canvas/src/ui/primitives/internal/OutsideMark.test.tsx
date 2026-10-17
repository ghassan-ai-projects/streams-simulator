import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { OutsideMark } from "./OutsideMark";

describe("outside mark", () => {
  it("names an out-of-range value and stays absent otherwise", () => {
    const view = render(<OutsideMark text="Outside the instrument's range of 0.00 to 600 m³/h" />);
    expect(screen.getByRole("img", { name: /^Outside the instrument's range/ })).toBeInTheDocument();
    view.rerender(<OutsideMark text="" />);
    expect(screen.queryByRole("img")).toBeNull();
  });
});
