import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { isDirty } from "../../../app/editor";
import { createNavigationStore } from "../../../app/shell";
import { editing, path, current } from "../../../test/editorHarness";
import { success } from "../../../test/outcome";
import { memoryLocation } from "../../../test/location";
import { PendingLeave } from "./LeaveDialog";

async function dirtyEditor() {
  const { editor, fake } = await editing();
  act(() => {
    editor.edit((doc) => ({ ...doc, title: "Lab two" }));
  });
  const location = memoryLocation(`#/edit?composition=${encodeURIComponent(path)}`);
  const navigation = createNavigationStore(location);
  navigation.guardUnsaved(() => isDirty(editor.snapshot()));
  render(<PendingLeave title="Lab two" editor={editor} navigation={navigation} />);
  act(() => {
    navigation.go({ screen: "workspace" });
  });
  return { editor, navigation, fake };
}

describe("leaving the editor with unsaved changes", () => {
  it("asks first and Stay keeps the editor open", async () => {
    const { navigation } = await dirtyEditor();
    expect(screen.getByRole("heading", { name: "Leave with unsaved changes?" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Stay" })).toHaveFocus();
    fireEvent.click(screen.getByRole("button", { name: "Stay" }));
    expect(screen.queryByRole("heading", { name: "Leave with unsaved changes?" })).toBeNull();
    expect(navigation.snapshot()).toEqual({ screen: "edit", composition: path });
  });

  it("Escape stays too", async () => {
    await dirtyEditor();
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("Discard and leave drops the unsaved change and goes", async () => {
    const { editor, navigation } = await dirtyEditor();
    fireEvent.click(screen.getByRole("button", { name: "Discard and leave" }));
    expect(navigation.snapshot()).toEqual({ screen: "workspace" });
    expect(current(editor).title).toBe("Lab");
  });

  it("Save and leave saves first and goes once the composition is saved", async () => {
    const { editor, navigation } = await dirtyEditor();
    fireEvent.click(screen.getByRole("button", { name: "Save and leave" }));
    await vi.waitFor(() => {
      expect(navigation.snapshot()).toEqual({ screen: "workspace" });
    });
    expect(isDirty(editor.snapshot())).toBe(false);
  });

  it("cannot discard while the save is still being written", async () => {
    const { fake } = await dirtyEditor();
    const write = fake.holdWrite();
    fireEvent.click(screen.getByRole("button", { name: "Save and leave" }));
    await vi.waitFor(() => {
      expect(screen.getByRole("button", { name: "Discard and leave" })).toBeDisabled();
    });
    await vi.waitFor(() => {
      expect(write).toHaveLength(1);
    });
    write[0]?.resolve(success({ path, etag: "c2" }));
    await vi.waitFor(() => {
      expect(screen.queryByRole("dialog")).toBeNull();
    });
  });
});
