import { act, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { EditorStore } from "../../app/editorStore";
import { success } from "../../model/bridge";
import { doc, Harness, LabBridge, path } from "../../test/editorHarness";

async function selected(plant: unknown, id: string) {
  const fake = new LabBridge();
  fake.files.set(path, success({ path, content: JSON.stringify(plant), etag: "e0" }));
  fake.actions.set("component.list", success({ components: [] }));
  const editor = new EditorStore(fake);
  await editor.open(path);
  render(<Harness editor={editor} />);
  act(() => {
    editor.select(id);
  });
}

describe("an instance's links list", () => {
  it("says in words that there are no links yet, and names the form that adds one", async () => {
    await selected({ ...doc, connections: [] }, "CH-1");
    expect(screen.getByText("No links yet. Use “Link an output” above.")).toBeInTheDocument();
  });
});
