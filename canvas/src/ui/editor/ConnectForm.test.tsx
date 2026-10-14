import { act, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { EditorStore } from "../../app/editorStore";
import { success } from "../../model/bridge";
import { choose, doc, Harness, LabBridge, path } from "../../test/editorHarness";

describe("link form", () => {
  it("names the missing target when nothing takes what an output gives", async () => {
    const fake = new LabBridge();
    const alone = { ...doc, equipment: [doc.equipment[0]], connections: [] };
    fake.files.set(path, success({ path, content: JSON.stringify(alone), etag: "d1" }));
    fake.actions.set("component.list", success({ components: [] }));
    const editor = new EditorStore(fake);
    await editor.open(path);
    render(<Harness editor={editor} />);
    act(() => {
      editor.select("CH-1");
    });
    expect(within(screen.getByLabelText("To")).getByRole("option")).toHaveTextContent("Choose an output first");
    choose("Output", "heat_out");
    expect(within(screen.getByLabelText("To")).getByRole("option")).toHaveTextContent("No instance takes it");
    expect(
      screen.getByText(
        "No instance takes heat rejection: CH-1's heat out is heat rejection. The other instances here have no inputs.",
      ),
    ).toBeInTheDocument();
    expect(within(screen.getByLabelText("Input")).getByRole("option")).toHaveTextContent("Choose an instance first");
  });
});
