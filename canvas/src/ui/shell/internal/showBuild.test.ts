import { act } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PanelStore } from "../../../app/shell";
import { editing } from "../../../test/editorHarness";
import { showBuild } from "./showBuild";

describe("the Build button", () => {
  it("shows the Build panel in place of a selected instance, then toggles it when nothing is selected", async () => {
    const { editor } = await editing();
    const panels = new PanelStore({ read: () => null, write: () => undefined });
    panels.toggle("right");
    expect(panels.snapshot().right).toBe(false);
    act(() => {
      editor.select("CH-1");
    });
    showBuild(editor, panels);
    expect(editor.snapshot()).toMatchObject({ selected: null });
    expect(panels.snapshot().right).toBe(true);
    showBuild(editor, panels);
    expect(panels.snapshot().right).toBe(false);
  });
});
