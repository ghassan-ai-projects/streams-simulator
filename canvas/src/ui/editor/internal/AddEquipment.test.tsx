import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { success } from "../../../test/outcome";
import { doc, Harness, LabBridge, path } from "../../../test/editorHarness";
import { editorStore } from "../../../test/stores";

const components = Array.from({ length: 9 }, (_, index) => ({
  id: `lib/unit-${String(index)}`,
  kind: "equipment",
  title: `Unit ${String(index)}`,
  version: 1,
}));

async function build() {
  const fake = new LabBridge();
  fake.addFile(path, doc, "e0");
  fake.actions.set("component.list", success({ components }));
  const editor = editorStore(fake.editor);
  await editor.open(path);
  render(<Harness editor={editor} />);
  return screen.getByRole("form", { name: "Add equipment" });
}

describe("the equipment library list", () => {
  it("says how many components it holds and that the rest take a scroll", async () => {
    const form = await build();
    const list = within(form).getByLabelText("Equipment");
    expect(list).toHaveAttribute("size", "6");
    expect(within(form).getByText("9 components · scroll the list for the rest")).toBeInTheDocument();
    expect(list).toHaveAccessibleDescription("9 components · scroll the list for the rest");
  });

  it("counts what a search leaves, with no scroll cue when they all fit", async () => {
    const form = await build();
    fireEvent.change(within(form).getByLabelText("Find equipment"), { target: { value: "unit 3" } });
    expect(within(form).getByText("1 component")).toBeInTheDocument();
    expect(within(form).queryByText(/scroll the list/)).not.toBeInTheDocument();
  });
});
