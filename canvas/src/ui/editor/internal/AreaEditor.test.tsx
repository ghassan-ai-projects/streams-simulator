import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { equipmentOf } from "../../../model/editor";
import { current, doc, editing } from "../../../test/editorHarness";

describe("plant editor areas", () => {
  it("renames an area everywhere it is used, and deletes only an area with no instance and no area inside", async () => {
    const { editor } = await editing();
    const areas = screen.getByRole("region", { name: "Areas" });
    fireEvent.click(within(areas).getByRole("button", { name: "Delete Cooling" }));
    expect(current(editor).areas).toEqual(doc.areas);
    expect(within(areas).getByText("Move the 2 instances out of it first.")).toBeInTheDocument();
    fireEvent.click(within(areas).getByRole("button", { name: "Delete Yard" }));
    expect(current(editor).areas).toEqual([{ name: "Cooling" }]);
    fireEvent.change(within(areas).getByLabelText("Name of Cooling"), { target: { value: "Chilled" } });
    fireEvent.click(within(areas).getByRole("button", { name: "Rename" }));
    expect(current(editor).areas).toEqual([{ name: "Chilled" }]);
    expect(equipmentOf(current(editor)).map((entry) => entry.area)).toEqual(["Chilled", "Chilled"]);
    expect(screen.getAllByRole("option", { name: "Chilled" })).toHaveLength(2);
  });
});
