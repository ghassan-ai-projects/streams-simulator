import { render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { EditorStore } from "../../../app/editor";
import { success } from "../../../test/outcome";
import { doc, editing, Harness, LabBridge, path } from "../../../test/editorHarness";

async function opened(plant: unknown) {
  const fake = new LabBridge();
  fake.files.set(path, success({ path, content: JSON.stringify(plant), etag: "e0" }));
  fake.actions.set("component.list", success({ components: [] }));
  const editor = new EditorStore(fake.editor);
  await editor.open(path);
  return { view: render(<Harness editor={editor} />) };
}

describe("an empty plant in the editor", () => {
  it("shows the first step next to the add form", async () => {
    const { view } = await opened({ ...doc, equipment: [], connections: [] });
    const build = screen.getByRole("complementary", { name: "Build" });
    expect(within(build).getByText(/^Add the first piece of equipment/)).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("drops the first step once the plant has equipment", async () => {
    await editing();
    expect(screen.queryByText(/^Add the first piece of equipment/)).not.toBeInTheDocument();
  });
});
