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

function linkRow(): string | null | undefined {
  return screen
    .getByRole("button", { name: /^Remove link:/ })
    .closest("li")
    ?.querySelector("span")?.textContent;
}

describe("an instance's links list", () => {
  it("says in words that there are no links yet, and names the form that adds one", async () => {
    await selected({ ...doc, connections: [] }, "CH-1");
    expect(screen.getByText("No links yet. Use “Link an output” above.")).toBeInTheDocument();
  });

  it("reads one link source-output to receiver-input, whichever end the panel shows", async () => {
    await selected(doc, "CH-1");
    expect(screen.getByRole("button", { name: "Remove link: heat out to CT-1 heat in" })).toBeInTheDocument();
    expect(linkRow()).toBe("heat out→CT-1heat in");
    act(() => {
      screen.getByRole("button", { name: "CT-1" }).click();
    });
    expect(screen.getByRole("button", { name: "Remove link: CH-1 heat out to heat in" })).toBeInTheDocument();
    expect(linkRow()).toBe("CH-1heat out→heat in");
  });
});
